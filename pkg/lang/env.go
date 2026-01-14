package lang

import (
	"io"

	"github.com/PondWader/kit/pkg/lang/std"
	"github.com/PondWader/kit/pkg/lang/values"
)

type Environment struct {
	Exports map[string]values.Value
	Vars    map[string]values.Value

	control *ExecutionControl
	parent  *Environment
}

func NewEnv() *Environment {
	return &Environment{
		Exports: make(map[string]values.Value),
		Vars:    make(map[string]values.Value),
		control: &ExecutionControl{},
	}
}

func Execute(r io.Reader) (*Environment, error) {
	env := NewEnv()
	prog, err := Parse(r)
	if err != nil {
		return nil, err
	}
	if err := env.Execute(prog); err != nil {
		return nil, err
	}
	return env, nil
}

func (e *Environment) LoadStd() {
	e.SetScoped("fetch", std.Fetch)
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

func (e *Environment) GetExport(name string) (values.Value, error) {
	v, ok := e.Exports[name]
	if ok {
		return v, nil
	}
	if e.parent != nil {
		return e.parent.GetExport(name)
	}
	return values.Nil, values.NewError("export named " + name + " does not exist")
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
		if e.control.Returned {
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
	child.control = e.control
	return e
}

func (e *Environment) Return(v values.Value) *values.Error {
	if e.control == nil || !e.control.ReturnAllowed {
		return values.NewError("return not allowed in this context")
	}
	e.control.ReturnVal = v
	e.control.Returned = true
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
