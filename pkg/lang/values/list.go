package values

import (
	"reflect"
)

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

func (l *List) Size() int {
	return len(l.s)
}

func (l *List) Val() Value {
	return Value{l}
}

func (l *List) Map(mapper Value) (*List, *Error) {
	fn, ok := mapper.ToFunction()
	if !ok {
		return nil, FmtTypeError("map", KindFunction)
	}
	new := NewList(l.Size())
	for i, v := range l.s {
		nv, err := fn.Call(v)
		if err != nil {
			return nil, err
		}
		new.Set(i, nv)
	}
	return new, nil
}

func (l *List) Get(key string) Value {
	switch key {
	case "map":
		return Of(l.Map)
	default:
		return Nil
	}
}

func (l *List) AsSlice() []Value {
	return l.s
}

func ListOf(s any) *List {
	v := reflect.ValueOf(s)
	l := NewList(v.Len())
	for i := range v.Len() {
		l.Set(i, Of(v.Index(i).Interface()))
	}
	return l
}
