package std

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/PondWader/kit/pkg/lang/values"
)

var Fetch = values.Of(fetch)

var client = &http.Client{}

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

	if res.StatusCode >= 300 {
		return values.Nil, values.NewError("received error status: " + res.Status)
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

	body, err := io.ReadAll(f.res.Body)
	if err != nil {
		return values.Nil, err
	}

	return values.Of(string(body)), nil
}

func (f PendingFetch) Json() (values.Value, error) {
	defer f.res.Body.Close()

	dec := json.NewDecoder(f.res.Body)
	var parsed any
	if err := dec.Decode(&parsed); err != nil {
		return values.Nil, err
	}

	return jsonToValue(parsed), nil
}

func jsonToValue(v any) values.Value {
	switch v := v.(type) {
	case map[string]any:
		obj := values.NewObject()
		for key, val := range v {
			obj.Put(key, jsonToValue(val))
		}
		return obj.Val()
	case []any:
		list := values.NewList(len(v))
		for i, val := range v {
			list.Set(i, jsonToValue(val))
		}
		return list.Val()
	case string:
		return values.Of(v)
	case float64:
		return values.Of(v)
	case bool:
		return values.Of(v)
	default:
		return values.Nil
	}
}

func (f PendingFetch) Read(p []byte) (n int, err error) {
	return f.res.Body.Read(p)
}
