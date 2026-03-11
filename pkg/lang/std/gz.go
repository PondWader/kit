package std

import (
	"compress/gzip"
	"io"

	"github.com/PondWader/kit/pkg/lang/values"
)

var Gz = values.Of(gzDecode)

func gzDecode(src values.Value) (values.Value, error) {
	srcObj, ok := src.ToObject()
	if !ok {
		return values.Nil, values.NewError("expected readable i/o object as argument to gz")
	}

	r, ok := srcObj.Binding.(io.Reader)
	if !ok {
		return values.Nil, values.NewError("expected readable i/o object as argument to gz")
	}

	gzReader, err := gzip.NewReader(r)
	if err != nil {
		return values.Nil, err
	}

	obj := values.ObjectFromStruct(PendingGz{r: gzReader})
	return values.Of(obj), nil
}

type PendingGz struct {
	r *gzip.Reader
}

func (g PendingGz) Text() (values.Value, error) {
	defer g.r.Close()

	body, err := io.ReadAll(g.r)
	if err != nil {
		return values.Nil, err
	}
	return values.Of(string(body)), nil
}

func (g PendingGz) Read(p []byte) (n int, err error) {
	return g.r.Read(p)
}
