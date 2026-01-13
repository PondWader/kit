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

	if len(results) == 0 {
		return Nil, nil
	} else if len(results) == 1 {
		return Nil, results[0].Interface().(*Error)
	} else if len(results) == 2 {
		err := results[1].Interface().(*Error)
		return Of(results[0].Interface()), err
	}
	return Nil, nil
}
