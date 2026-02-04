package kit

import (
	"io"
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
	return values.ObjectFromStruct(tar{})
}

func (b *installBinding) Load(env *lang.Environment) {
	env.Set("sys", b.CreateSys().Val())
	env.Set("tar", b.CreateTar().Val())
}

type tar struct {
	Gz tarGz
}

type tarGz struct{}

func (tgz tarGz) Extract(src values.Value) (values.Value, *values.Error) {
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
	obj.Put("to", values.Of(func(dst values.Value) *values.Error {
		_ = r
		return nil
	}))
	return obj.Val(), nil
}
