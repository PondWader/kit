package values

type Kind uint8

const (
	KindUnknownKind = iota
	KindNumber
	KindString
	KindBool
	KindObject
	KindList
	KindFunction
	KindNil
)

func (k Kind) String() string {
	switch k {
	case KindNumber:
		return "number"
	case KindString:
		return "string"
	case KindBool:
		return "bool"
	case KindObject:
		return "object"
	case KindList:
		return "list"
	case KindFunction:
		return "function"
	case KindNil:
		return "nil"
	default:
		return "???"
	}
}
