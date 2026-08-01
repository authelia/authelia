package utils

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderTable(t *testing.T) {
	testCases := []struct {
		name     string
		headers  []string
		rows     [][]string
		expected string
	}{
		{
			name:    "ShouldRenderHeadersOnlyWithNoRows",
			headers: []string{"Name", "Value"},
			rows:    nil,
			expected: "" +
				"+------+-------+\n" +
				"| Name | Value |\n" +
				"+------+-------+\n" +
				"+------+-------+\n",
		},
		{
			name:    "ShouldRenderSingleColumn",
			headers: []string{"Username"},
			rows:    [][]string{{"johndoe"}, {"harry"}},
			expected: "" +
				"+----------+\n" +
				"| Username |\n" +
				"+----------+\n" +
				"| johndoe  |\n" +
				"| harry    |\n" +
				"+----------+\n",
		},
		{
			name:    "ShouldExpandColumnWidthToWidestCell",
			headers: []string{"A", "B"},
			rows:    [][]string{{"short", "x"}, {"a-much-longer-value", "y"}},
			expected: "" +
				"+---------------------+---+\n" +
				"| A                   | B |\n" +
				"+---------------------+---+\n" +
				"| short               | x |\n" +
				"| a-much-longer-value | y |\n" +
				"+---------------------+---+\n",
		},
		{
			name:    "ShouldRenderNoColumnsWithEmptyHeaders",
			headers: []string{},
			rows:    nil,
			expected: "" +
				"+\n" +
				"|\n" +
				"+\n" +
				"+\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer

			require.NoError(t, RenderTable(&buf, tc.headers, tc.rows))
			assert.Equal(t, tc.expected, buf.String())
		})
	}
}

type errWriter struct{}

func (errWriter) Write(_ []byte) (n int, err error) {
	return 0, errors.New("write error")
}

func TestRenderTableShouldReturnWriterError(t *testing.T) {
	err := RenderTable(errWriter{}, []string{"A"}, [][]string{{"1"}})

	assert.EqualError(t, err, "write error")
}
