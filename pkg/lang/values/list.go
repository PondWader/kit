package values

import (
	"math"
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

func (l *List) Filter(predicate Value) (*List, *Error) {
	fn, ok := predicate.ToFunction()
	if !ok {
		return nil, FmtTypeError("filter", KindFunction)
	}
	var filtered []Value
	for _, v := range l.s {
		result, err := fn.Call(v)
		if err != nil {
			return nil, err
		}
		if b, ok := result.ToBool(); ok && b {
			filtered = append(filtered, v)
		} else if !ok {
			return nil, FmtTypeError("filter(... -> ?)", KindBool)
		}
	}
	out := ListOf(filtered)
	for i, v := range filtered {
		out.Set(i, v)
	}
	return out, nil
}

func (l *List) Length() Value {
	return Of(l.Size())
}

func (l *List) Slice(spec Value) (*List, *Error) {
	o, ok := spec.ToObject()
	if !ok {
		return nil, FmtTypeError("slice", KindObject)
	}

	startIdx, err := resolveSliceIndex(o.Get("start"), 0, l.Size(), "start")
	if err != nil {
		return nil, err
	}
	endIdx, err := resolveSliceIndex(o.Get("end"), l.Size(), l.Size(), "end")
	if err != nil {
		return nil, err
	}
	if endIdx <= startIdx {
		return NewList(0), nil
	}

	out := make([]Value, endIdx-startIdx)
	copy(out, l.s[startIdx:endIdx])
	return &List{s: out}, nil
}

func resolveSliceIndex(v Value, defaultIdx, size int, key string) (int, *Error) {
	if v.Obj == nil {
		return defaultIdx, nil
	}

	n, ok := v.ToNumber()
	if !ok {
		return 0, FmtTypeError("slice."+key, KindNumber)
	}
	if math.Trunc(n) != n {
		return 0, NewError("slice." + key + " must be an integer")
	}

	idx := int(n)
	if idx < 0 {
		idx += size
		if idx < 0 {
			return 0, nil
		}
	}
	if idx > size {
		return size, nil
	}
	return idx, nil
}

func (l *List) Get(key string) Value {
	switch key {
	case "map":
		return Of(l.Map)
	case "filter":
		return Of(l.Filter)
	case "length":
		return Of(l.Length)
	case "slice":
		return Of(l.Slice)
	default:
		return Nil
	}
}

func (l *List) AsSlice() []Value {
	return l.s
}

func ListOf(s any) *List {
	if vs, ok := s.([]Value); ok {
		return &List{vs}
	}

	v := reflect.ValueOf(s)
	l := NewList(v.Len())
	for i := range v.Len() {
		l.Set(i, Of(v.Index(i).Interface()))
	}
	return l
}
