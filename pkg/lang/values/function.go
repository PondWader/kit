package values

import (
	"reflect"
)

type Function struct {
	f reflect.Value
}

func (f Function) Call(args ...Value) (Value, *Error) {
	if len(args) != f.f.Type().NumIn() {
		return Nil, NewError("incorrect arg count")
	}

	reflectArgs := make([]reflect.Value, len(args))
	for i, arg := range args {
		reflectArgs[i] = reflect.ValueOf(arg)
	}

	results := f.f.Call(reflectArgs)

	var val Value = Nil
	var err any
	if len(results) == 0 {
		return Nil, nil
	} else if len(results) == 1 {
		unknown := results[0].Interface()
		if e, ok := unknown.(*Error); ok {
			return val, e
		} else if e, ok := unknown.(error); ok {
			return val, GoError(e)
		}
		return Of(unknown), nil
	} else if len(results) == 2 {
		err = results[1].Interface()
		val = Of(results[0].Interface())
	}

	if e, ok := err.(*Error); ok {
		return val, e
	} else if e, ok := err.(error); ok {
		return val, GoError(e)
	}

	return val, nil
}
