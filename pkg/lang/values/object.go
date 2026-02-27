package values

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"unicode"
)

var ErrKeyNotFound = errors.New("key does not exist")

type Object struct {
	Binding any
	m       map[string]Value
	i       []*Interface
}

func NewObject() *Object {
	return &Object{nil, make(map[string]Value), nil}
}

func ObjectFromMap(m map[string]Value) *Object {
	return &Object{nil, m, nil}
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

func (o *Object) TagInterface(iface *Interface) {
	if iface == nil {
		return
	}
	if slices.Contains(o.i, iface) {
		return
	}
	o.i = append(o.i, iface)
}

func (o *Object) Implements(iface *Interface) bool {
	if iface == nil {
		return false
	}
	return slices.Contains(o.i, iface)
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
		panic("expected struct in ObjectFromStruct but got " + val.Kind().String())
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

		fieldName := pascalToSnakeCase(field.Name)

		if fieldValue.Kind() == reflect.Struct {
			// Convert the struct to an object
			obj.Put(fieldName, ObjectFromStruct(fieldValue.Interface()).Val())
		} else {
			// Convert field value to Value and add to object
			obj.Put(fieldName, Of(fieldValue.Interface()))
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

		methodName := pascalToSnakeCase(method.Name)

		// Wrap method as Function and add to object
		obj.Put(methodName, Value{Function{methodValue}})
	}

	return obj
}

func pascalToSnakeCase(s string) string {
	runes := []rune(s)
	var result []rune
	for i, r := range runes {
		if unicode.IsUpper(r) {
			if i > 0 {
				if unicode.IsLower(runes[i-1]) {
					result = append(result, '_')
				} else if unicode.IsUpper(runes[i-1]) && i+1 < len(runes) && unicode.IsLower(runes[i+1]) {
					result = append(result, '_')
				}
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}
