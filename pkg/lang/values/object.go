package values

import (
	"errors"
	"fmt"
	"reflect"
	"unicode"
)

var ErrKeyNotFound = errors.New("key does not exist")

type Object struct {
	Binding any
	m       map[string]Value
}

func NewObject() *Object {
	return &Object{nil, make(map[string]Value)}
}

func ObjectFromMap(m map[string]Value) *Object {
	return &Object{nil, m}
}

func (o *Object) Val() Value {
	return Value{o}
}

func (o *Object) Put(key string, val Value) {
	o.m[key] = val
}

func (o *Object) Get(key string) Value {
	return o.m[key]
}

func (o *Object) GetString(key string) (string, error) {
	v, ok := o.m[key]
	if !ok {
		return "", fmt.Errorf("%w: looking for string value called \"%s\"", ErrKeyNotFound, key)
	}
	str, ok := v.ToString()
	if !ok {
		return "", errors.New("expected string value called \"" + key + "\" is of type " + v.Kind().String())
	}
	return str.String(), nil
}

func ObjectFromStruct(v any) *Object {
	obj := NewObject()
	obj.Binding = v
	val := reflect.ValueOf(v)

	// Handle pointer to struct
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}

	// Ensure it's a struct
	if val.Kind() != reflect.Struct {
		return obj
	}

	typ := val.Type()

	// Add fields
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Convert field name to camelCase if not an acronym
		fieldName := field.Name
		if len(fieldName) == 1 || (unicode.IsUpper(rune(fieldName[0])) && !unicode.IsUpper(rune(fieldName[1]))) {
			fieldName = string(unicode.ToLower(rune(fieldName[0]))) + fieldName[1:]
		}

		if fieldValue.Kind() == reflect.Struct {
			// Convert the struct to an object
			obj.Put(fieldName, ObjectFromStruct(fieldValue.Interface()).Val())
		} else {
			// Convert field value to Value and add to object
			obj.Put(fieldName, Value{fieldValue.Interface()})
		}
	}

	// Add methods
	for i := 0; i < val.NumMethod(); i++ {
		method := typ.Method(i)
		methodValue := val.Method(i)

		// Skip unexported methods
		if !method.IsExported() {
			continue
		}

		// Convert method name to camelCase if not an acronym
		methodName := method.Name
		if len(methodName) == 1 || (unicode.IsUpper(rune(methodName[0])) && !unicode.IsUpper(rune(methodName[1]))) {
			methodName = string(unicode.ToLower(rune(methodName[0]))) + methodName[1:]
		}

		// Wrap method as Function and add to object
		obj.Put(methodName, Value{Function{methodValue}})
	}

	return obj
}
