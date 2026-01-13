package lang

import (
	"github.com/PondWader/kit/pkg/lang/values"
)

type Environment struct {
	Exports map[string]values.Value
	Vars    map[string]values.Value

	Exec   *ExecutionControl
	parent *Environment
}

func NewEnv() *Environment {
	return &Environment{
		Exports: make(map[string]values.Value),
		Vars:    make(map[string]values.Value),
		Exec:    &ExecutionControl{},
	}
}

func (e *Environment) Export(name string, value values.Value) {
	e.Exports[name] = value
}

func (e *Environment) Get(name string) (values.Value, *values.Error) {
	env, v := e.getVarEnv(name)
	if env == nil {
		return v, values.NewError(name + " does not exist in scope")
	}
	return v, nil
}

func (e *Environment) Set(name string, value values.Value) {
	env, _ := e.getVarEnv(name)
	if env != nil {
		env.Vars[name] = value
	} else {
		e.Vars[name] = value
	}
}

func (e *Environment) SetScoped(name string, value values.Value) {
	e.Vars[name] = value
}

func (e *Environment) Execute(prog []Node) *values.Error {
	for _, n := range prog {
		if _, err := n.Eval(e); err != nil {
			return err
		}
		if e.Exec.Returned {
			return nil
		}
	}
	return nil
}

func (e *Environment) getVarEnv(name string) (*Environment, values.Value) {
	v, ok := e.Vars[name]
	if ok {
		return e, v
	} else if e.parent == nil {
		return nil, values.Nil
	}
	return e.parent.getVarEnv(name)
}

func (e *Environment) NewChild() *Environment {
	child := NewEnv()
	child.parent = e
	child.Exec = e.Exec
	return e
}

func (e *Environment) Return(v values.Value) *values.Error {
	if e.Exec == nil || !e.Exec.ReturnAllowed {
		return values.NewError("return not allowed in this context")
	}
	e.Exec.ReturnVal = v
	e.Exec.Returned = true
	return nil
}

type ExecutionControl struct {
	ReturnAllowed bool
	ReturnVal     values.Value
	Returned      bool
}

func NewExec() *ExecutionControl {
	return &ExecutionControl{}
}
