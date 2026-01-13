package values

type Error struct {
	Msg   String
	Cause error
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

func (e Error) Unwrap() error {
	return e.Cause
}

func GoError(err error) *Error {
	return &Error{
		Msg:   String(err.Error()),
		Cause: err,
	}
}

func FmtTypeError(fnName string, expectedType Kind) *Error {
	return NewError("expected " + expectedType.String() + " as argument to " + fnName)
}
