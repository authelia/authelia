package configuration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			[]string{"template", "template"},
			"duplicate filter named 'template'",
		},
		{
			"ShouldErrorOnDuplicateFilterNameCaps",
			[]string{"TEMPLATE", "template"},
			"duplicate filter named 'template'",
		},
		{
			"ShouldNotErrorOnValidFilters",
			[]string{"template"},
			"",
		},
		{
			"ShouldErrorOnExpandEnvFilter",
			[]string{"expand-env"},
			"invalid filter named 'expand-env'",
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
			actual, theError := NewFileFilters(tc.have, "", "")

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

func TestTemplateBytesFilter(t *testing.T) {
	t.Run("ShouldReturnFilterName", func(t *testing.T) {
		assert.Equal(t, filterTemplate, NewTemplateFileFilter("", "").Name())
	})

	t.Run("ShouldRenderWithDefaultDelimiters", func(t *testing.T) {
		filter := NewTemplateFileFilter("", "")

		out, err := filter.Filter([]byte(`hello {{ "world" }}`))

		require.NoError(t, err)
		assert.Equal(t, "hello world", string(out))
	})

	t.Run("ShouldRenderWithCustomDelimiters", func(t *testing.T) {
		filter := NewTemplateFileFilter("<%", "%>")

		out, err := filter.Filter([]byte(`hello <% "world" %> with {{ braces }} preserved`))

		require.NoError(t, err)
		assert.Equal(t, "hello world with {{ braces }} preserved", string(out))
	})

	t.Run("ShouldReturnErrorOnParseFailure", func(t *testing.T) {
		filter := NewTemplateFileFilter("", "")

		out, err := filter.Filter([]byte("{{ if }}"))

		assert.Nil(t, out)
		assert.ErrorContains(t, err, "missing value for if")
	})

	t.Run("ShouldReturnErrorOnExecuteFailure", func(t *testing.T) {
		filter := NewTemplateFileFilter("", "")

		out, err := filter.Filter([]byte(`{{ template "missing" }}`))

		assert.Nil(t, out)
		assert.ErrorContains(t, err, `template "missing"`)
	})

	t.Run("ShouldForwardCustomDelimitersThroughNewFileFilters", func(t *testing.T) {
		filters, err := NewFileFilters([]string{filterTemplate}, "<%", "%>")
		require.NoError(t, err)
		require.Len(t, filters, 1)

		out, err := filters[0].Filter([]byte(`<% "rendered" %>`))

		require.NoError(t, err)
		assert.Equal(t, "rendered", string(out))
	})
}

func TestFilteredFile_Read(t *testing.T) {
	f := FilteredFileProvider("/tmp/does-not-matter")

	m, err := f.Read()

	assert.Nil(t, m)
	assert.EqualError(t, err, "filtered file provider does not support this method")
}
