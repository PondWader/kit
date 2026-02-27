package std

import "github.com/PondWader/kit/pkg/lang/values"

var Error = values.NewInterfaceWith("Error", map[string]values.Kind{
	"message": values.KindString,
}, map[string]struct{}{}).Val()

var NewError = values.Of(newError)

func newError(message values.Value) (values.Value, error) {
	msg, ok := message.ToString()
	if !ok {
		return values.Nil, values.FmtTypeError("NewError", values.KindString)
	}

	iface, ok := Error.ToInterface()
	if !ok {
		return values.Nil, values.NewError("std Error interface is invalid")
	}

	obj := values.NewObject()
	obj.Put("message", values.Of(msg))
	obj.TagInterface(iface)

	return obj.Val(), nil
}
