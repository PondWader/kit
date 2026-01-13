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
	_ = req

	resp := PendingFetch{req}

	obj := values.ObjectFromStruct(resp)
	return values.Of(obj), nil
}

type PendingFetch struct {
	req *http.Request
}

func (f PendingFetch) Text() (values.Value, error) {
	resp, err := client.Do(f.req)
	if err != nil {
		return values.Nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return values.Nil, values.NewError("received error status: " + resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return values.Nil, err
	}

	return values.Of(string(body)), nil
}
