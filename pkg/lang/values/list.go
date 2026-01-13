package values

import "reflect"

type List struct {
	s []Value
}

func NewList(size int) *List {
	return &List{
		s: make([]Value, size),
	}
}

func (l *List) Set(idx int, v Value) {
	l.s[idx] = v
}

func (l *List) Val() Value {
	return Value{l}
}

func ListOf(s any) *List {
	v := reflect.ValueOf(s)
	l := NewList(v.Len())
	for i := range v.Len() {
		l.Set(i, Of(v.Index(i)))
	}
	return l
}
