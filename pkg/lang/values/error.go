package values

type Error struct {
	Msg String
}

func NewError(msg string) *Error {
	return &Error{Msg: String(msg)}
}

func (s Error) Get(key string) Value {
	switch key {
	case "msg":
		return Of(s.Msg)
	default:
		return Nil
	}
}

func (e Error) Error() string {
	return string(e.Msg)
}
