package values

import (
	"reflect"
)

func Of(v any) Value {
	switch v := v.(type) {
	case Value:
		return v
	case string:
		return Value{Obj: String(v)}
	case String:
		return Value{Obj: v}
	case int:
		return Value{Obj: float64(v)}
	case float64:
		return Value{Obj: float64(v)}
	case bool, *Object, *List, nil:
		return Value{Obj: v}
	case Object, List:
		return Value{Obj: &v}
	default:
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Func {
			return Value{Obj: Function{f: rv}}
		}
		panic(reflect.TypeOf(v).String() + " is not a valid kitlang type")
	}
}

type Value struct{ Obj any }

var Nil = Value{}

type Keyable interface {
	Get(key string) Value
}

type Callable interface {
	Call(args ...Value) (Value, *Error)
}

func (v Value) Kind() Kind {
	switch v.Obj.(type) {
	case float64:
		return KindNumber
	case String:
		return KindString
	case *Object:
		return KindObject
	case *List:
		return KindList
	case Function:
		return KindFunction
	case nil:
		return KindNil
	default:
		return KindUnknownKind
	}
}

func (v Value) Get(key string) (Value, *Error) {
	if v.Obj == nil {
		return Nil, NewError("key \"" + key + "\" does not exist on value")
	}
	k, ok := v.Obj.(Keyable)
	if !ok {
		return Nil, NewError("key \"" + key + "\" does not exist on value")
	}
	return k.Get(key), nil
}

func (v Value) IsCallable() bool {
	_, ok := v.Obj.(Callable)
	return ok
}

func (v Value) Call(args ...Value) (Value, *Error) {
	c, ok := v.Obj.(Callable)
	if !ok {
		return Nil, NewError("value is not callable")
	}
	return c.Call(args...)
}

func (v Value) Stringify() String {
	str, ok := v.ToString()
	if ok {
		return str
	}
	toString, err := v.Get("to_string")
	if err == nil && toString.IsCallable() {
		str, err := toString.Call()
		if str, ok := str.ToString(); err == nil && ok {
			return str
		}
	}
	return String(v.Kind().String())
}

func (v Value) ToString() (String, bool) {
	s, ok := v.Obj.(String)
	return s, ok
}

func (v Value) ToNumber() {

}

func (v Value) ToBool() (b bool, ok bool) {
	b, ok = v.Obj.(bool)
	return
}

func (v Value) ToList() (l *List, ok bool) {
	l, ok = v.Obj.(*List)
	return
}

func (v Value) ToObject() (o *Object, ok bool) {
	o, ok = v.Obj.(*Object)
	return
}
