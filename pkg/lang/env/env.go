package env

type Environment struct {
	IsTopLevel bool
	Exports    map[string]any
	Values     map[string]any
}
