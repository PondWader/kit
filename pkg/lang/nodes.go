package lang

import (
	"fmt"
	"strings"

	"github.com/PondWader/kit/pkg/lang/values"
)

type Node interface {
	Eval(*Environment) (values.Value, *values.Error)
	String() string
}

type NodeExport struct {
	Decl NodeDeclaration
}

func (n NodeExport) Eval(e *Environment) (values.Value, *values.Error) {
	v, err := n.Decl.Eval(e)
	if err != nil {
		return values.Nil, err
	}
	e.Export(n.Decl.Name, v)
	return v, nil
}

func (n NodeExport) String() string {
	return fmt.Sprintf("export %s", n.Decl.String())
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

func (n NodeDeclaration) String() string {
	return fmt.Sprintf("%s = %s", n.Name, n.Value.String())
}

type NodeList struct {
	Elements []Node
}

func (n NodeList) Eval(e *Environment) (values.Value, *values.Error) {
	list := values.NewList(len(n.Elements))
	for i, el := range n.Elements {
		el, err := el.Eval(e)
		if err != nil {
			return values.Nil, err
		}
		list.Set(i, el)
	}
	return list.Val(), nil
}

func (n NodeList) String() string {
	var parts []string
	for _, el := range n.Elements {
		parts = append(parts, el.String())
	}
	return fmt.Sprintf("[%s]", strings.Join(parts, ", "))
}

type NodeObject struct {
	Body []Node
}

func (n NodeObject) Eval(e *Environment) (values.Value, *values.Error) {
	child := e.NewChild()
	if err := child.Execute(n.Body); err != nil {
		return values.Nil, err
	}
	obj := values.ObjectFromMap(child.Vars)
	return obj.Val(), nil
}

func (n NodeObject) String() string {
	var parts []string
	for _, node := range n.Body {
		parts = append(parts, node.String())
	}
	if len(parts) > 0 {
		return fmt.Sprintf("{ %s }", strings.Join(parts, "; "))
	}
	return "{}"
}

type NodeBlock struct {
	Body           []Node
	IsFunctionBody bool
}

func (n NodeBlock) Eval(e *Environment) (values.Value, *values.Error) {
	child := e.NewChild()
	if n.IsFunctionBody {
		child.control = NewExec()
		child.control.ReturnAllowed = true
	}
	if err := child.Execute(n.Body); err != nil {
		return values.Nil, err
	}
	if n.IsFunctionBody {
		return child.control.ReturnVal, nil
	}
	return values.Nil, nil
}

func (n NodeBlock) String() string {
	var parts []string
	for _, node := range n.Body {
		parts = append(parts, node.String())
	}
	if len(parts) > 0 {
		return fmt.Sprintf("{ %s }", strings.Join(parts, "; "))
	}
	return "{}"
}

type NodeLiteral struct {
	Value values.Value
}

func (n NodeLiteral) Eval(e *Environment) (values.Value, *values.Error) {
	return n.Value, nil
}

func (n NodeLiteral) String() string {
	return fmt.Sprintf("%v", n.Value)
}

type NodeIdentifier struct {
	Ident string
}

func (n NodeIdentifier) Eval(e *Environment) (values.Value, *values.Error) {
	return e.Get(n.Ident)
}

func (n NodeIdentifier) String() string {
	return n.Ident
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
		sb.WriteString(v.String())
	}
	return values.Of(sb.String()), nil
}

func (n NodeString) String() string {
	if len(n.Parts) == 1 {
		return fmt.Sprintf("\"%s\"", n.Parts[0].String())
	}
	return fmt.Sprintf("\"<string with %d parts>\"", len(n.Parts))
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

func (n NodeCall) String() string {
	if n.Arg == nil {
		return fmt.Sprintf("%s()", n.Fn.String())
	}
	return fmt.Sprintf("%s(%s)", n.Fn.String(), n.Arg.String())
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

func (n NodeKeyAccess) String() string {
	return fmt.Sprintf("%s.%s", n.Val.String(), n.Key)
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

func (n NodeFunction) String() string {
	if n.ArgName != "" {
		return fmt.Sprintf("fn(%s) %s", n.ArgName, n.Body.String())
	}
	return fmt.Sprintf("fn() %s", n.Body.String())
}

type NodeReturn struct {
	Val Node
}

func (n NodeReturn) Eval(e *Environment) (values.Value, *values.Error) {
	v, err := n.Val.Eval(e)
	if err != nil {
		return values.Nil, err
	}
	return v, e.Return(v)
}

func (n NodeReturn) String() string {
	return "return " + n.Val.String()
}
