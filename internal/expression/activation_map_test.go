package expression

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapActivationResolveName(t *testing.T) {
	parent := NewMapActivation(nil, map[string]any{"parent_key": "parent_value"})

	testCases := []struct {
		name          string
		activation    *MapActivation
		resolve       string
		expectedValue any
		expectedFound bool
	}{
		{
			"ShouldResolveFromValues",
			NewMapActivation(nil, map[string]any{"key": "value"}),
			"key",
			"value",
			true,
		},
		{
			"ShouldReturnNotFoundWithoutParent",
			NewMapActivation(nil, map[string]any{"key": "value"}),
			"missing",
			nil,
			false,
		},
		{
			"ShouldFallBackToParent",
			NewMapActivation(parent, map[string]any{"key": "value"}),
			"parent_key",
			"parent_value",
			true,
		},
		{
			"ShouldPreferValuesOverParent",
			NewMapActivation(parent, map[string]any{"parent_key": "child_value"}),
			"parent_key",
			"child_value",
			true,
		},
		{
			"ShouldReturnNotFoundWithParentMiss",
			NewMapActivation(parent, map[string]any{"key": "value"}),
			"missing",
			nil,
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, found := tc.activation.ResolveName(tc.resolve)

			assert.Equal(t, tc.expectedValue, actual)
			assert.Equal(t, tc.expectedFound, found)
		})
	}
}

func TestMapActivationParent(t *testing.T) {
	testCases := []struct {
		name       string
		activation *MapActivation
		expectNil  bool
	}{
		{
			"ShouldReturnNilParent",
			NewMapActivation(nil, map[string]any{}),
			true,
		},
		{
			"ShouldReturnParent",
			NewMapActivation(NewMapActivation(nil, map[string]any{}), map[string]any{}),
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectNil {
				assert.Nil(t, tc.activation.Parent())
			} else {
				assert.NotNil(t, tc.activation.Parent())
			}
		})
	}
}
