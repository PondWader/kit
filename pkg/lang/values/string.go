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

func (s String) StartsWith(prefix Value) (Value, *Error) {
	suffixStr, ok := prefix.ToString()
	if !ok {
		return Nil, FmtTypeError("starts_with", KindString)
	}
	return Of(strings.HasPrefix(string(s), suffixStr.String())), nil
}

func (s String) EndsWith(suffix Value) (Value, *Error) {
	suffixStr, ok := suffix.ToString()
	if !ok {
		return Nil, FmtTypeError("ends_with", KindString)
	}
	return Of(strings.HasSuffix(string(s), suffixStr.String())), nil
}

func (s String) RemovePrefix(prefix Value) (Value, *Error) {
	prefixStr, ok := prefix.ToString()
	if !ok {
		return Nil, FmtTypeError("remove_prefix", KindString)
	}
	return Of(strings.TrimPrefix(string(s), prefixStr.String())), nil
}

func (s String) RemoveSuffix(suffix Value) (Value, *Error) {
	suffixStr, ok := suffix.ToString()
	if !ok {
		return Nil, FmtTypeError("remove_suffix", KindString)
	}
	return Of(strings.TrimSuffix(string(s), suffixStr.String())), nil
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
	case "starts_with":
		return Of(s.StartsWith)
	case "ends_with":
		return Of(s.EndsWith)
	case "remove_prefix":
		return Of(s.RemovePrefix)
	case "remove_suffix":
		return Of(s.RemoveSuffix)
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
