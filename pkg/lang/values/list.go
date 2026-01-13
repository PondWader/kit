package values

type List []Value

func (l List) Val() Value {
	return Value{l}
}
