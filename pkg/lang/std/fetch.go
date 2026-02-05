package std

import (
	"io"
	"net/http"
	"time"

	"github.com/PondWader/kit/pkg/lang/values"
)

var Fetch = values.Of(fetch)

var client = &http.Client{
	Timeout: time.Minute,
}

func fetch(url values.Value) (values.Value, error) {
	urlStr, ok := url.ToString()
	if !ok {
		return values.Nil, values.FmtTypeError("fetch", values.KindString)
	}

	req, err := http.NewRequest("GET", urlStr.String(), nil)
	if err != nil {
		return values.Nil, err
	}
	req.Header.Set("User-Agent", "Kit Package Manager")

	res, err := client.Do(req)
	if err != nil {
		return values.Nil, err
	}

	resp := PendingFetch{req, res}

	obj := values.ObjectFromStruct(resp)
	return values.Of(obj), nil
}

type PendingFetch struct {
	req *http.Request
	res *http.Response
}

func (f PendingFetch) Text() (values.Value, error) {
	defer f.res.Body.Close()

	if f.res.StatusCode >= 300 {
		return values.Nil, values.NewError("received error status: " + f.res.Status)
	}

	body, err := io.ReadAll(f.res.Body)
	if err != nil {
		return values.Nil, err
	}

	return values.Of(string(body)), nil
}

func (f PendingFetch) Read(p []byte) (n int, err error) {
	return f.res.Body.Read(p)
}
