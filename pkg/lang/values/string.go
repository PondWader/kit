package values

import "strings"

type String string

func (s String) Trim() Value {
	return Of(strings.TrimSpace(string(s)))
}

func (s String) CutPrefixBefore(sep string) Value {
	_, str, _ := strings.Cut(string(s), sep)
	return Of(str)
}

func (s String) CutSuffixAfter(sep string) Value {
	idx := strings.LastIndex(string(s), sep)
	if idx == -1 {
		return Of(s)
	}
	return Of(s[:idx])
}

func (s String) Get(key string) Value {
	switch key {
	case "trim":
		return Of(s.Trim)
	case "cut_prefix_before":
		return Of(s.CutPrefixBefore)
	case "cut_suffix_after":
		return Of(s.CutSuffixAfter)
	default:
		return Nil
	}
}

func (s String) Val() Value {
	return Value{s}
}
