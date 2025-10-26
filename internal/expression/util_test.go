package expression

import (
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNativeValues(t *testing.T) {
	testCases := []struct {
		name       string
		expression string
		have       map[string]any
		expected   any
	}{
		{
			"ShouldHandleBasicCase",
			"true",
			map[string]any{},
			true,
		},
		{
			"ShouldHandleComplexCaseSliceMapOutput",
			`groups.map(i, {"name": i, "oidcID": i})`,
			map[string]any{
				"groups": []string{"abc", "123"},
			},
			[]any{
				map[string]any{"name": "abc", "oidcID": "abc"},
				map[string]any{"name": "123", "oidcID": "123"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := getStandardCELEnvOpts()

			env, err := cel.NewEnv(opts...)
			require.NoError(t, err)

			ast, issues := env.Compile(tc.expression)
			require.NoError(t, issues.Err())

			program, err := env.Program(ast)
			require.NoError(t, err)

			result, _, err := program.Eval(tc.have)
			require.NoError(t, err)

			assert.Equal(t, tc.expected, toNativeValue(result))
		})
	}
}
