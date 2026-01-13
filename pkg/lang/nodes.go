package lang

import (
	"github.com/PondWader/kit/pkg/lang/env"
	"github.com/PondWader/kit/pkg/lang/values"
)

type Node interface {
	Eval(*env.Environment) values.Value
}

type NodeExport struct {
	Decl NodeDeclaration
}

func (n NodeExport) Eval(e *env.Environment) {

}

type NodeDeclaration struct {
	Name  string
	Value Node
}

type NodeList struct {
	Elements []Node
}

type NodeObject struct {
	Body []Node
}

type NodeLiteral struct {
	Value any
}

type NodeString struct{}

type NodeCall struct{}

type NodeKeyAccess struct{}

type NodeFunction struct{}

type NodeAnonymousFunction struct{}
