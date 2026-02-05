package values

import (
	"strings"
)

type String string

func (s String) TrimWhitespace() Value {
	return Of(strings.TrimSpace(string(s)))
}

func (s String) CutPrefixBefore(sep Value) (Value, *Error) {
	sepStr, ok := sep.ToString()
	if !ok {
		return Nil, FmtTypeError("cut_prefix_before", KindString)
	}
	_, str, _ := strings.Cut(string(s), sepStr.String())
	return Of(str), nil
}

func (s String) CutSuffixAfter(sep Value) (Value, *Error) {
	sepStr, ok := sep.ToString()
	if !ok {
		return Nil, FmtTypeError("cut_suffix_after", KindString)
	}
	idx := strings.LastIndex(string(s), sepStr.String())
	if idx == -1 {
		return Of(s), nil
	}
	return Of(s[:idx]), nil
}

func (s String) Split(sep Value) (Value, *Error) {
	sepStr, ok := sep.ToString()
	if !ok {
		return Nil, FmtTypeError("split", KindString)
	}
	return Of(strings.Split(string(s), sepStr.String())), nil
}

func (s String) Get(key string) Value {
	switch key {
	case "trim_whitespace":
		return Of(s.TrimWhitespace)
	case "inclusive_remove_until":
		return Of(s.CutPrefixBefore)
	case "inclusive_remove_after":
		return Of(s.CutSuffixAfter)
	case "split":
		return Of(s.Split)
	default:
		return Nil
	}
}

func (s String) Val() Value {
	return Value{s}
}

func (s String) String() string {
	return string(s)
}
