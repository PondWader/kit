package values

type Error struct {
	Msg   String
	Cause error
}

func NewError(msg string) *Error {
	return &Error{Msg: String(msg)}
}

func (e *Error) Get(key string) Value {
	if e == nil {
		return Nil
	}

	switch key {
	case "msg":
		return Of(e.Msg)
	default:
		return Nil
	}
}

func (e *Error) Error() string {
	if e == nil {
		return "<nil error>"
	}

	return string(e.Msg)
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}

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
