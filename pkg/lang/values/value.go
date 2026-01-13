package values

import (
	"reflect"

	"github.com/PondWader/kit/pkg/lang/env"
)

func Of(v any) Value {
	switch v := v.(type) {
	case string:
		return Value{Obj: String(v)}
	case String:
		return Value{Obj: v}
	case int:
		return Value{Obj: float64(v)}
	case float64:
		return Value{Obj: float64(v)}
	case bool, *Object, List, nil:
		return Value{Obj: v}
	case Object:
		return Value{Obj: &v}
	default:
		panic(reflect.TypeOf(v).String() + " is not a valid kitlang type")
	}
}

type Value struct{ Obj any }

var Nil = Value{}

type Keyable interface {
	Get(key string) Value
}

type Callable interface {
	Call(e *env.Environment, args ...any) Value
}

func (v Value) Get(key string) Value {
	if v.Obj == nil {
		return v
	}
	k, ok := v.Obj.(Keyable)
	if !ok {
		return Value{}
	}
	return k.Get(key)
}

func (v Value) Call(e *env.Environment, args ...any) Value {
	c, ok := v.Obj.(Callable)
	if !ok {
		panic(NewError("value is not callable"))
	}
	return c.Call(e, args)
}

func (v Value) ToString() {}

func (v Value) ToNumber() {

}

func (v Value) ToBool() (b bool, ok bool) {
	b, ok = v.Obj.(bool)
	return
}

func (v Value) ToList() {

}

func (v Value) ToObject() {

}
