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
	case int:
		return Value{Obj: float64(v)}
	case float64:
		return Value{Obj: float64(v)}
	case bool, *Object, *List, nil, Function, String:
		return Value{Obj: v}
	case Object, List:
		return Value{Obj: &v}
	default:
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Func {
			return Value{Obj: Function{f: rv}}
		} else if rv.Kind() == reflect.Slice {
			return Value{Obj: ListOf(v)}
		} else if rv.Kind() == reflect.Struct {
			return Value{Obj: ObjectFromStruct(v)}
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
	member := k.Get(key)
	if member == Nil {
		return Nil, NewError("key \"" + key + "\" does not exist on value")
	}
	return member, nil
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

func (v Value) String() string {
	return string(v.Stringify())
}

func (v Value) ToString() (String, bool) {
	s, ok := v.Obj.(String)
	return s, ok
}

func (v Value) ToNumber() (n float64, ok bool) {
	n, ok = v.Obj.(float64)
	return
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

func (v Value) ToFunction() (f Function, ok bool) {
	f, ok = v.Obj.(Function)
	return
}

func (v Value) Equals(o Value) (bool, *Error) {
	if v.Kind() != o.Kind() {
		return false, NewError("values in equals must be of the same kind: got " + v.Kind().String() + " and " + o.Kind().String())
	}

	return reflect.ValueOf(v.Obj).Equal(reflect.ValueOf(o.Obj)), nil
}
