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

	ModLoader func(modName string) (*Environment, error)

	VarBoundary bool
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
	e.SetScoped("xz", std.Xz)
	e.SetScoped("ar", std.Ar)
	e.SetScoped("Error", std.Error)
	e.SetScoped("error", std.NewError)
}

func (e *Environment) Export(name string, value values.Value) {
	e.Exports[name] = value
}

func (e *Environment) Get(name string) (values.Value, *values.Error) {
	env, v := e.getVarEnv(name, false)
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
	env, _ := e.getVarEnv(name, true)
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

func (e *Environment) ExecuteReader(r io.Reader) *values.Error {
	prog, err := Parse(r)
	if err != nil {
		return values.GoError(err)
	}
	return e.Execute(prog)
}

func (e *Environment) getVarEnv(name string, settable bool) (*Environment, values.Value) {
	v, ok := e.Vars[name]
	if ok {
		return e, v
	} else if e.parent == nil || (settable && e.VarBoundary) {
		return nil, values.Nil
	}
	return e.parent.getVarEnv(name, settable)
}

func (e *Environment) NewChild() *Environment {
	child := NewEnv()
	child.parent = e
	child.control = e.control
	return child
}

func (e *Environment) Return(v values.Value) *values.Error {
	if e.control == nil || !e.control.ReturnAllowed {
		return values.NewError("return not allowed in this context")
	}
	e.control.ReturnVal = v
	e.control.Returned = true
	return nil
}

func (e *Environment) Import(modName string) *values.Error {
	if e.ModLoader == nil {
		return values.NewError("import \"" + modName + "\" not found")
	}
	mod, err := e.ModLoader(modName)
	if err != nil {
		return values.GoError(err)
	}
	e.Set(modName, values.ObjectFromMap(mod.Exports).Val())
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

type Binding interface {
	Load(env *Environment)
}

func (e *Environment) Enable(b Binding) {
	if b != nil {
		b.Load(e)
	}
}
