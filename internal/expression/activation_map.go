package expression

import "cel.dev/cel-go/interpreter"

// NewMapActivation returns a *MapActivation which resolves names from the given values before the parent.
func NewMapActivation(parent interpreter.Activation, values map[string]any) *MapActivation {
	return &MapActivation{
		parent: parent,
		values: values,
	}
}

// MapActivation is an interpreter.Activation which resolves names from a map.
type MapActivation struct {
	parent interpreter.Activation
	values map[string]any
}

// ResolveName returns the value of the named attribute.
func (a *MapActivation) ResolveName(name string) (object any, found bool) {
	if object, found = a.values[name]; found {
		return object, true
	}

	if a.parent != nil {
		return a.parent.ResolveName(name)
	}

	return nil, false
}

// Parent returns the parent activation.
func (a *MapActivation) Parent() interpreter.Activation {
	return a.parent
}
