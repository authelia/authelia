package configuration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFileFilters(t *testing.T) {
	testCases := []struct {
		name   string
		have   []string
		expect string
	}{
		{
			"ShouldErrorOnInvalidFilterName",
			[]string{"abc"},
			"invalid filter named 'abc'",
		},
		{
			"ShouldErrorOnInvalidFilterNameWithDuplicates",
			[]string{"abc", "abc"},
			"invalid filter named 'abc'",
		},
		{
			"ShouldErrorOnInvalidFilterNameWithDuplicatesCaps",
			[]string{"ABC", "abc"},
			"invalid filter named 'abc'",
		},
		{
			"ShouldErrorOnDuplicateFilterName",
			[]string{"expand-env", "expand-env"},
			"duplicate filter named 'expand-env'",
		},
		{
			"ShouldErrorOnDuplicateFilterNameCaps",
			[]string{"expand-ENV", "expand-env"},
			"duplicate filter named 'expand-env'",
		},
		{
			"ShouldNotErrorOnValidFilters",
			[]string{"expand-env", "template"},
			"",
		},
		{
			"ShouldNotErrorOnExpandEnvFilter",
			[]string{"expand-env"},
			"",
		},
		{
			"ShouldNotErrorOnExpandEnvFilterCaps",
			[]string{"EXPAND-env"},
			"",
		},
		{
			"ShouldNotErrorOnTemplateFilter",
			[]string{"template"},
			"",
		},
		{
			"ShouldNotErrorOnTemplateFilterCaps",
			[]string{"TEMPLATE"},
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, theError := NewFileFilters(tc.have)

			switch tc.expect {
			case "":
				assert.NoError(t, theError)
				assert.Len(t, actual, len(tc.have))
			default:
				assert.EqualError(t, theError, tc.expect)
			}
		})
	}
}
