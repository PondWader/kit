package lang

import (
	"strings"

	"github.com/PondWader/kit/pkg/lang/values"
)

type Node interface {
	Eval(*Environment) (values.Value, *values.Error)
}

type NodeExport struct {
	Decl NodeDeclaration
}

func (n NodeExport) Eval(e *Environment) (values.Value, *values.Error) {
	v, err := n.Decl.Value.Eval(e)
	if err != nil {
		return values.Nil, err
	}
	e.Export(n.Decl.Name, v)
	return v, nil
}

type NodeDeclaration struct {
	Name  string
	Value Node
}

func (n NodeDeclaration) Eval(e *Environment) (values.Value, *values.Error) {
	v, err := n.Value.Eval(e)
	if err != nil {
		return values.Nil, err
	}
	e.Set(n.Name, v)
	return v, nil
}

type NodeList struct {
	Elements []Node
}

func (n NodeList) Eval(e *Environment) (values.Value, *values.Error) {
	list := make(values.List, len(n.Elements))
	for i, el := range n.Elements {
		el, err := el.Eval(e)
		if err != nil {
			return values.Nil, err
		}
		list[i] = el
	}
	return list.Val(), nil
}

type NodeObject struct {
	Body []Node
}

func (n NodeObject) Eval(e *Environment) (values.Value, *values.Error) {
	child := e.NewChild()
	if err := child.Execute(n.Body); err != nil {
		return values.Nil, err
	}
	obj := (*values.Object)(&child.Vars)
	return obj.Val(), nil
}

type NodeBlock struct {
	Body           []Node
	IsFunctionBody bool
}

func (n NodeBlock) Eval(e *Environment) (values.Value, *values.Error) {
	child := e.NewChild()
	if n.IsFunctionBody {
		child.Exec = NewExec()
	}
	if err := child.Execute(n.Body); err != nil {
		return values.Nil, err
	}
	if n.IsFunctionBody {
		return child.Exec.ReturnVal, nil
	}
	return values.Nil, nil
}

type NodeLiteral struct {
	Value values.Value
}

func (n NodeLiteral) Eval(e *Environment) (values.Value, *values.Error) {
	return n.Value, nil
}

type NodeIdentifier struct {
	Ident string
}

func (n NodeIdentifier) Eval(e *Environment) (values.Value, *values.Error) {
	return e.Get(n.Ident)
}

type NodeString struct {
	Parts []Node
}

func (n NodeString) Eval(e *Environment) (values.Value, *values.Error) {
	// Fast-path for single part
	if len(n.Parts) == 1 {
		v, err := n.Parts[0].Eval(e)
		if err != nil {
			return values.Nil, err
		}
		return v.Stringify().Val(), nil
	}

	var sb strings.Builder
	for _, part := range n.Parts {
		v, err := part.Eval(e)
		if err != nil {
			return values.Nil, err
		}
		sb.WriteString(string(v.Stringify()))
	}
	return values.Of(sb.String()), nil
}

type NodeCall struct {
	Fn  Node
	Arg Node
}

func (n NodeCall) Eval(e *Environment) (values.Value, *values.Error) {
	fn, err := n.Fn.Eval(e)
	if err != nil {
		return values.Nil, err
	} else if n.Arg == nil {
		return fn.Call()
	}

	arg, err := n.Arg.Eval(e)
	if err != nil {
		return values.Nil, err
	}
	return fn.Call(arg)
}

type NodeKeyAccess struct {
	Val Node
	Key string
}

func (n NodeKeyAccess) Eval(e *Environment) (values.Value, *values.Error) {
	val, err := n.Val.Eval(e)
	if err != nil {
		return values.Nil, err
	}
	return val.Get(n.Key)
}

type NodeFunction struct {
	ArgName string
	Body    Node
}

func (n NodeFunction) Eval(e *Environment) (values.Value, *values.Error) {
	if n.ArgName != "" {
		return values.Of(func(arg values.Value) (values.Value, *values.Error) {
			c := e.NewChild()
			c.SetScoped(n.ArgName, arg)
			return n.Body.Eval(c)
		}), nil
	}
	return values.Of(func() (values.Value, *values.Error) {
		return n.Body.Eval(e)
	}), nil
}
