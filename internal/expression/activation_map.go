package expression

import "github.com/google/cel-go/interpreter"

func NewMapActivation(parent interpreter.Activation, values map[string]any) *MapActivation {
	return &MapActivation{
		parent: parent,
		values: values,
	}
}

type MapActivation struct {
	parent interpreter.Activation
	values map[string]any
}

func (a *MapActivation) ResolveName(name string) (object any, found bool) {
	if object, found = a.values[name]; found {
		return object, true
	}

	if a.parent != nil {
		return a.parent.ResolveName(name)
	}

	return nil, false
}

func (a *MapActivation) Parent() interpreter.Activation {
	return a.parent
}
