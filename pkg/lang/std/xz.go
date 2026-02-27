package std

import (
	"io"

	"github.com/PondWader/kit/pkg/lang/values"
	"github.com/ulikunitz/xz"
)

var Xz = values.Of(xzDecode)

func xzDecode(src values.Value) (values.Value, error) {
	srcObj, ok := src.ToObject()
	if !ok {
		return values.Nil, values.NewError("expected readable i/o object as argument to xz")
	}

	r, ok := srcObj.Binding.(io.Reader)
	if !ok {
		return values.Nil, values.NewError("expected readable i/o object as argument to xz")
	}

	xzReader, err := xz.NewReader(r)
	if err != nil {
		return values.Nil, err
	}

	obj := values.ObjectFromStruct(PendingXz{r: xzReader})
	return values.Of(obj), nil
}

type PendingXz struct {
	r io.Reader
}

func (x PendingXz) Text() (values.Value, error) {
	body, err := io.ReadAll(x.r)
	if err != nil {
		return values.Nil, err
	}
	return values.Of(string(body)), nil
}

func (x PendingXz) Read(p []byte) (n int, err error) {
	return x.r.Read(p)
}
