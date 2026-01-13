package values

import (
	"reflect"
	"unicode"
)

type Object struct {
	m map[string]Value
}

func NewObject() *Object {
	return &Object{make(map[string]Value)}
}

func ObjectFromMap(m map[string]Value) *Object {
	return &Object{m}
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

func ObjectFromStruct(v any) *Object {
	obj := NewObject()
	val := reflect.ValueOf(v)

	// Handle pointer to struct
	if val.Kind() == reflect.Ptr {
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

		// Convert field value to Value and add to object
		obj.Put(fieldName, Value{fieldValue.Interface()})
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
