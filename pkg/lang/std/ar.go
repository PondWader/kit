package std

import (
	"bytes"
	"errors"
	"io"

	debar "github.com/PondWader/kit/internal/ar"
	"github.com/PondWader/kit/pkg/lang/values"
)

var Ar = values.Of(arDecode)

func arDecode(src values.Value) (values.Value, error) {
	srcObj, ok := src.ToObject()
	if !ok {
		return values.Nil, values.NewError("expected readable i/o object as argument to ar")
	}

	r, ok := srcObj.Binding.(io.Reader)
	if !ok {
		return values.Nil, values.NewError("expected readable i/o object as argument to ar")
	}

	arReader, err := debar.NewArReader(r)
	if err != nil {
		return values.Nil, err
	}

	archive := arArchive{state: &arArchiveState{
		reader: arReader,
		files:  make(map[string][]byte),
	}}
	return values.Of(values.ObjectFromStruct(archive)), nil
}

type arArchive struct {
	state *arArchiveState
}

type arArchiveState struct {
	reader *debar.ArReader
	files  map[string][]byte
	done   bool
}

func (a arArchive) File(name values.Value) (values.Value, error) {
	nameStr, ok := name.ToString()
	if !ok {
		return values.Nil, values.FmtTypeError("ar(...).file", values.KindString)
	}
	if a.state == nil {
		return values.Nil, values.NewError("ar archive is invalid")
	}

	fileName := nameStr.String()

	contents, ok := a.state.files[fileName]
	if ok {
		return values.Of(values.ObjectFromStruct(arFile{r: bytes.NewReader(contents)})), nil
	}

	for !a.state.done {
		hdr, err := a.state.reader.Next()
		if errors.Is(err, io.EOF) {
			a.state.done = true
			break
		}
		if err != nil {
			return values.Nil, err
		}

		fileContents, err := io.ReadAll(a.state.reader)
		if err != nil {
			return values.Nil, err
		}

		a.state.files[hdr.Name] = fileContents
		if hdr.Name == fileName {
			contents = fileContents
			return values.Of(values.ObjectFromStruct(arFile{r: bytes.NewReader(contents)})), nil
		}
	}

	return values.Nil, values.NewError("file \"" + fileName + "\" not found in ar archive")

}

type arFile struct {
	r *bytes.Reader
}

func (f arFile) Read(p []byte) (n int, err error) {
	return f.r.Read(p)
}
