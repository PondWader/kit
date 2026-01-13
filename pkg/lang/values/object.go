package values

type Object map[string]Value

func (o *Object) Val() Value {
	return Value{o}
}
