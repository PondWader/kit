package kit

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/PondWader/kit/pkg/lang"
	"github.com/PondWader/kit/pkg/lang/values"
)

type installBinding struct {
	RootDir *os.Root
}

func (b *installBinding) CreateSys() *values.Object {
	o := values.NewObject()
	o.Put("OS", values.Of(runtime.GOOS))
	o.Put("ARCH", values.Of(runtime.GOARCH))
	return o
}

func (b *installBinding) CreateTar() *values.Object {
	return values.ObjectFromStruct(tarBinding{b: b, Gz: tarGzBinding{b: b}})
}

func (b *installBinding) Load(env *lang.Environment) {
	env.Set("sys", b.CreateSys().Val())
	env.Set("tar", b.CreateTar().Val())
	env.Set("link_bin_dir", values.Of(b.LinkBinDir))
}

func (b *installBinding) LinkBinDir(dirV values.Value) error {
	dir, ok := dirV.ToString()
	if !ok {
		return values.FmtTypeError("link_bin_dir", values.KindString)
	}
	_ = dir
	return nil
}

type tarBinding struct {
	b  *installBinding
	Gz tarGzBinding
}

type tarGzBinding struct {
	b *installBinding
}

func (tgz tarGzBinding) Extract(src values.Value) (values.Value, *values.Error) {
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
	obj.Put("from_archive_dir", values.Of(func(dst values.Value) (values.Value, error) {
		dir, ok := dst.ToString()
		if !ok {
			return values.Nil, values.FmtTypeError("tar.gz.extract(...).from_archive_dir", values.KindString)
		}
		archiveDir = string(dir)
		return obj.Val(), nil
	}))

	obj.Put("to", values.Of(func(dst values.Value) error {
		dstStr, ok := dst.ToString()
		if !ok {
			return values.FmtTypeError("tar.gz.extract(...).to", values.KindString)
		}
		gr, err := gzip.NewReader(r)
		if err != nil {
			return err
		}

		resolvedDst := filepath.Join(".", string(dstStr))
		root, err := tgz.b.RootDir.OpenRoot(resolvedDst)
		if err != nil {
			return err
		}

		return extractTar(tar.NewReader(gr), archiveDir, root)
	}))
	return obj.Val(), nil
}

func extractTar(tr *tar.Reader, archiveRoot string, dst *os.Root) error {
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target, err := filepath.Rel(filepath.Join(".", archiveRoot), hdr.Name)
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
