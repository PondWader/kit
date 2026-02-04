package kit

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/PondWader/kit/pkg/lang"
	"github.com/PondWader/kit/pkg/lang/values"
)

type installBinding struct{}

func (b *installBinding) CreateSys() *values.Object {
	o := values.NewObject()
	o.Put("OS", values.Of(runtime.GOOS))
	o.Put("ARCH", values.Of(runtime.GOARCH))
	return o
}

func (b *installBinding) CreateTar() *values.Object {
	return values.ObjectFromStruct(tarBinding{})
}

func (b *installBinding) Load(env *lang.Environment) {
	env.Set("sys", b.CreateSys().Val())
	env.Set("tar", b.CreateTar().Val())
}

type tarBinding struct {
	Gz tarGzBinding
}

type tarGzBinding struct{}

func (tgz tarGzBinding) Extract(src values.Value) (values.Value, *values.Error) {
	// TODO: should have a better interface system so doesn't need to use bindings
	srcObj, ok := src.ToObject()
	if !ok {
		return values.Nil, values.NewError("expected readable i/o object as argument to tar.gz.extract")
	}
	r, ok := srcObj.Binding.(io.Reader)
	if !ok {
		return values.Nil, values.NewError("expected readable i/o object as argument to tar.gz.extract")
	}

	obj := values.NewObject()
	obj.Put("to", values.Of(func(dst values.Value) error {
		dstStr, ok := dst.ToString()
		if !ok {
			return values.FmtTypeError("tar.gz.extract.to", values.KindString)
		}
		_ = dstStr
		_ = r
		// TODO: resolve path to install path
		// return extractTar(tar.NewReader(r), string(dstStr))
		return values.NewError("not implemented")
	}))
	return obj.Val(), nil
}

func extractTar(tr *tar.Reader, dst string) error {
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dst, hdr.Name)

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(hdr.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := extractTarFile(tr, target, os.FileMode(hdr.Mode)); err != nil {
				return err
			}
		}
	}
	return nil
}

func extractTarFile(r io.Reader, target string, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, r); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
