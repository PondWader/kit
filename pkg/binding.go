package kit

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"io"
	"strings"
	"unsafe"

	"io/fs"
	"os"
	"path/filepath"
	"runtime"

	"github.com/PondWader/kit/pkg/lang"
	"github.com/PondWader/kit/pkg/lang/values"
	"github.com/ulikunitz/xz"
)

type mountBinding struct {
	MountDir string
}

type installBinding struct {
	RootDir *os.Root
	Install *mountBinding

	mountSetup []func(m *Mount) error
}

func (b *installBinding) CreateSys() *values.Object {
	o := values.NewObject()
	o.Put("OS", values.Of(runtime.GOOS))
	o.Put("ARCH", values.Of(runtime.GOARCH))
	return o
}

func (b *installBinding) CreateTar() *values.Object {
	return values.ObjectFromStruct(tarBinding{
		b: b,
		Gz: tarLayer{b: b, newReader: func(r io.Reader) (io.Reader, error) {
			return gzip.NewReader(r)
		}},
		Xz: tarLayer{b: b, newReader: func(r io.Reader) (io.Reader, error) {
			return xz.NewReader(r)
		}},
	})
}

func (b *installBinding) CreateZip() *values.Object {
	return values.ObjectFromStruct(zipBinding{b: b})
}

func (b *installBinding) CreateFs() *values.Object {
	o := values.NewObject()
	o.Put("file", values.Of(func(path values.Value) (values.Value, *values.Error) {
		pathStr, ok := path.ToString()
		if !ok {
			return values.Nil, values.FmtTypeError("fs.file", values.KindString)
		}
		return b.createFileBuilder(string(pathStr)), nil
	}))
	o.Put("dir", values.Of(func(path values.Value) (values.Value, *values.Error) {
		pathStr, ok := path.ToString()
		if !ok {
			return values.Nil, values.FmtTypeError("fs.dir", values.KindString)
		}
		return b.createDirBuilder(string(pathStr)), nil
	}))
	return o
}

func (b *installBinding) createFileBuilder(path string) values.Value {
	obj := values.NewObject()
	resolvedPath := filepath.Join(".", path)

	obj.Put("name", values.Of(filepath.Base(path)))
	obj.Put("path", values.Of(path))

	var fh *os.File

	obj.Put("create_with_perms", values.Of(func(mode values.Value) (values.Value, error) {
		perm, ok := mode.ToNumber()
		if !ok {
			return values.Nil, values.FmtTypeError("fs.file(...).create_with_perms", values.KindNumber)
		}
		f, err := b.RootDir.OpenFile(resolvedPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fs.FileMode(perm))
		if err != nil {
			return values.Nil, err
		}
		fh = f
		return obj.Val(), nil
	}))

	obj.Put("write_and_close", values.Of(func(content values.Value) error {
		contentStr, ok := content.ToString()
		if !ok {
			return values.FmtTypeError("fs.file(...).write_and_close", values.KindString)
		}
		contentBytes := unsafe.Slice(unsafe.StringData(contentStr.String()), len(contentStr.String()))
		if _, err := fh.Write(contentBytes); err != nil {
			fh.Close()
			return err
		}
		return fh.Close()
	}))

	obj.Put("symlink", values.Of(func(target values.Value) error {
		targetStr, ok := target.ToString()
		if !ok {
			return values.FmtTypeError("fs.file(...).symlink", values.KindString)
		}
		return os.Symlink(string(targetStr), filepath.Join(b.RootDir.Name(), resolvedPath))
	}))

	return obj.Val()
}

func (b *installBinding) createDirBuilder(path string) values.Value {
	obj := values.NewObject()
	resolvedPath := filepath.Join(".", path)

	obj.Put("read_entries", values.Of(func() (*values.List, error) {
		entries, err := b.RootDir.FS().(fs.ReadDirFS).ReadDir(resolvedPath)
		if err != nil {
			return nil, err
		}
		list := values.NewList(len(entries))
		for i, entry := range entries {
			entryPath := filepath.Join(path, entry.Name())
			list.Set(i, b.createFileBuilder(entryPath))
		}
		return list, nil
	}))

	return obj.Val()
}

func (b *installBinding) Load(env *lang.Environment) {
	env.Set("sys", b.CreateSys().Val())
	env.Set("tar", b.CreateTar().Val())
	env.Set("zip", b.CreateZip().Val())
	env.Set("fs", b.CreateFs().Val())
	env.Set("link_bin_dir", values.Of(b.LinkBinDir))
	env.Set("link_bin_file", values.Of(b.LinkBinFile))
	if b.Install != nil {
		env.Set("install", values.ObjectFromStruct(b.Install).Val())
	}
	env.ModLoader = b.loadMod
}

func (b *installBinding) LinkBinDir(dirV values.Value) error {
	dir, ok := dirV.ToString()
	if !ok {
		return values.FmtTypeError("link_bin_dir", values.KindString)
	}

	dirPath := filepath.Join(".", string(dir))
	entries, err := b.RootDir.FS().(fs.ReadDirFS).ReadDir(dirPath)
	if err != nil {
		return err
	}

	b.mountSetup = append(b.mountSetup, func(m *Mount) error {
		for _, entry := range entries {
			if err := m.LinkBin(filepath.Join(dirPath, entry.Name()), entry.Name()); err != nil {
				return err
			}
		}
		return nil
	})

	return nil
}

func (b *installBinding) LinkBinFile(pathV values.Value) error {
	path, ok := pathV.ToString()
	if !ok {
		return values.FmtTypeError("link_bin_file", values.KindString)
	}

	relPath := filepath.Join(".", string(path))

	b.mountSetup = append(b.mountSetup, func(m *Mount) error {
		return m.LinkBin(relPath, filepath.Base(path.String()))
	})

	return nil
}

func (b *installBinding) SetupMount(m *Mount) error {
	for _, fn := range b.mountSetup {
		if err := fn(m); err != nil {
			return err
		}
	}
	return nil
}

type tarBinding struct {
	b  *installBinding
	Gz tarLayer
	Xz tarLayer
}

type tarLayer struct {
	b         *installBinding
	newReader func(r io.Reader) (io.Reader, error)
}

func (tl tarLayer) Extract(src values.Value) (values.Value, *values.Error) {
	srcObj, ok := src.ToObject()
	if !ok {
		return values.Nil, values.NewError("expected readable i/o object as argument to tar.gz.extract")
	}
	// TODO: should have a better interface system so doesn't need to use bindings
	r, ok := srcObj.Binding.(io.Reader)
	if !ok {
		return values.Nil, values.NewError("expected readable i/o object as argument to tar.gz.extract")
	}

	obj := values.NewObject()

	var archiveDir string
	var skipBaseDir bool
	var ignoreDirs []string
	obj.Put("from_archive_dir", values.Of(func(dst values.Value) (values.Value, error) {
		dir, ok := dst.ToString()
		if !ok {
			return values.Nil, values.FmtTypeError("tar.gz.extract(...).from_archive_dir", values.KindString)
		}
		archiveDir = string(dir)
		return obj.Val(), nil
	}))

	obj.Put("skipping_base_dir", values.Of(func() values.Value {
		skipBaseDir = true
		return obj.Val()
	}))

	obj.Put("ignoring_dir", values.Of(func(dir values.Value) (values.Value, *values.Error) {
		dirStr, ok := dir.ToString()
		if !ok {
			return values.Nil, values.FmtTypeError("tar.gz.extract(...).ignoring_dir", values.KindString)
		}
		ignoreDirs = append(ignoreDirs, string(dirStr))
		return obj.Val(), nil
	}))

	obj.Put("to", values.Of(func(dst values.Value) error {
		dstStr, ok := dst.ToString()
		if !ok {
			return values.FmtTypeError("tar.gz.extract(...).to", values.KindString)
		}
		gr, err := tl.newReader(r)
		if err != nil {
			return err
		}

		resolvedDst := filepath.Join(".", string(dstStr))
		root, err := tl.b.RootDir.OpenRoot(resolvedDst)
		if err != nil {
			return err
		}

		return extractTar(tar.NewReader(gr), archiveDir, skipBaseDir, ignoreDirs, root)
	}))
	return obj.Val(), nil
}

type zipBinding struct {
	b *installBinding
}

func (z zipBinding) Extract(src values.Value) (values.Value, *values.Error) {
	srcObj, ok := src.ToObject()
	if !ok {
		return values.Nil, values.NewError("expected readable i/o object as argument to zip.extract")
	}
	// TODO: should have a better interface system so doesn't need to use bindings
	r, ok := srcObj.Binding.(io.Reader)
	if !ok {
		return values.Nil, values.NewError("expected readable i/o object as argument to zip.extract")
	}

	obj := values.NewObject()

	var archiveDir string
	obj.Put("from_archive_dir", values.Of(func(dst values.Value) (values.Value, error) {
		dir, ok := dst.ToString()
		if !ok {
			return values.Nil, values.FmtTypeError("zip.extract(...).from_archive_dir", values.KindString)
		}
		archiveDir = string(dir)
		return obj.Val(), nil
	}))

	obj.Put("to", values.Of(func(dst values.Value) error {
		dstStr, ok := dst.ToString()
		if !ok {
			return values.FmtTypeError("zip.extract(...).to", values.KindString)
		}

		contents, err := io.ReadAll(r)
		if err != nil {
			return err
		}

		zr, err := zip.NewReader(bytes.NewReader(contents), int64(len(contents)))
		if err != nil {
			return err
		}

		resolvedDst := filepath.Join(".", string(dstStr))
		root, err := z.b.RootDir.OpenRoot(resolvedDst)
		if err != nil {
			return err
		}

		return extractZip(zr, archiveDir, root)
	}))

	return obj.Val(), nil
}

func extractTar(tr *tar.Reader, archiveRoot string, skipBaseDir bool, ignoreDirs []string, dst *os.Root) error {
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		name := hdr.Name
		if skipBaseDir {
			_, after, found := strings.Cut(name, "/")
			if !found || after == "" {
				continue
			}
			name = after
		}

		ignored := false
		for _, dir := range ignoreDirs {
			if name == dir || strings.HasPrefix(name, dir+"/") {
				ignored = true
				break
			}
		}
		if ignored {
			continue
		}

		target, err := filepath.Rel(filepath.Join(".", archiveRoot), name)
		if err != nil {
			return err
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := dst.MkdirAll(target, os.FileMode(hdr.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := extractTarFile(tr, dst, target, os.FileMode(hdr.Mode)); err != nil {
				return err
			}
		}
	}
	return nil
}

func extractTarFile(r io.Reader, dst *os.Root, name string, mode os.FileMode) error {
	if err := dst.MkdirAll(filepath.Dir(name), 0o755); err != nil {
		return err
	}
	f, err := dst.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, r); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func extractZip(zr *zip.Reader, archiveRoot string, dst *os.Root) error {
	for _, file := range zr.File {
		target, err := filepath.Rel(filepath.Join(".", archiveRoot), file.Name)
		if err != nil {
			return err
		}

		if file.FileInfo().IsDir() {
			if err := dst.MkdirAll(target, file.Mode()); err != nil {
				return err
			}
			continue
		}

		r, err := file.Open()
		if err != nil {
			return err
		}

		if err := extractTarFile(r, dst, target, file.Mode()); err != nil {
			r.Close()
			return err
		}

		if err := r.Close(); err != nil {
			return err
		}
	}

	return nil
}
