package values

import "maps"

type Interface struct {
	Name    string
	fields  map[string]Kind
	methods map[string]struct{}
}

func NewInterface(name string) *Interface {
	return NewInterfaceWith(name, nil, nil)
}

func NewInterfaceWith(name string, fields map[string]Kind, methods map[string]struct{}) *Interface {
	fieldsCopy := make(map[string]Kind, len(fields))
	maps.Copy(fieldsCopy, fields)

	methodsCopy := make(map[string]struct{}, len(methods))
	maps.Copy(methodsCopy, methods)

	return &Interface{
		Name:    name,
		fields:  fieldsCopy,
		methods: methodsCopy,
	}
}

func (i *Interface) AddField(name string, kind Kind) {
	if i == nil {
		return
	}
	i.fields[name] = kind
}

func (i *Interface) RemoveField(name string) {
	if i == nil {
		return
	}
	delete(i.fields, name)
}

func (i *Interface) FieldKind(name string) (Kind, bool) {
	if i == nil {
		return KindUnknownKind, false
	}
	kind, ok := i.fields[name]
	return kind, ok
}

func (i *Interface) Fields() map[string]Kind {
	if i == nil {
		return map[string]Kind{}
	}
	out := make(map[string]Kind, len(i.fields))
	for name, kind := range i.fields {
		out[name] = kind
	}
	return out
}

func (i *Interface) AddMethod(name string) {
	if i == nil {
		return
	}
	i.methods[name] = struct{}{}
}

func (i *Interface) RemoveMethod(name string) {
	if i == nil {
		return
	}
	delete(i.methods, name)
}

func (i *Interface) HasMethod(name string) bool {
	if i == nil {
		return false
	}
	_, ok := i.methods[name]
	return ok
}

func (i *Interface) Methods() []string {
	if i == nil {
		return nil
	}
	methods := make([]string, 0, len(i.methods))
	for name := range i.methods {
		methods = append(methods, name)
	}
	return methods
}

func (i *Interface) Val() Value {
	return Value{Obj: i}
}

func (i *Interface) ValidateObject(obj *Object) *Error {
	if i == nil {
		return NewError("interface cannot be nil")
	}
	if obj == nil {
		return NewError("expected object instance for interface " + i.Name)
	}

	for fieldName, fieldKind := range i.fields {
		v := obj.Get(fieldName)
		if v == Nil {
			return NewError("object does not satisfy interface " + i.Name + ": missing field " + fieldName)
		}
		if fieldKind != KindUnknownKind && v.Kind() != fieldKind {
			return NewError("object does not satisfy interface " + i.Name + ": field " + fieldName + " expected " + fieldKind.String() + " but got " + v.Kind().String())
		}
	}

	for methodName := range i.methods {
		v := obj.Get(methodName)
		if v == Nil {
			return NewError("object does not satisfy interface " + i.Name + ": missing method " + methodName)
		}
		if !v.IsCallable() {
			return NewError("object does not satisfy interface " + i.Name + ": method " + methodName + " is not callable")
		}
	}

	return nil
}
