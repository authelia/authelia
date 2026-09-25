// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package configuration

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
			actual, theError := NewFileFilters(nil, tc.have...)

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

func TestNewFileFiltersValuesFiles(t *testing.T) {
	testCases := []struct {
		name        string
		valuesFiles []string
		filters     []string
		expected    map[string]any
		err         string
	}{
		{
			"ShouldLoadValuesForTemplateFilter",
			[]string{"./test_resources/config_values.values.yml"},
			[]string{"template"},
			map[string]any{"Example": map[string]any{"Value": "light"}},
			"",
		},
		{
			"ShouldLoadValuesForTemplateFilterWithExpandEnvFilter",
			[]string{"./test_resources/config_values.values.yml"},
			[]string{"expand-env", "template"},
			map[string]any{"Example": map[string]any{"Value": "light"}},
			"",
		},
		{
			"ShouldErrorOnBadValuesFileForTemplateFilter",
			[]string{"./test_resources/missing.yml"},
			[]string{"template"},
			nil,
			"error reading values file: open ./test_resources/missing.yml: no such file or directory",
		},
		{
			"ShouldIgnoreValuesWithoutAFilterWhichUtilizesThem",
			[]string{"./test_resources/missing.yml"},
			[]string{"expand-env"},
			nil,
			"",
		},
		{
			"ShouldIgnoreValuesWithoutAnyFilters",
			[]string{"./test_resources/missing.yml"},
			nil,
			nil,
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := NewFileFilters(tc.valuesFiles, tc.filters...)

			if tc.err != "" {
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, actual)

				var errValues *FilterValuesError

				assert.True(t, errors.As(err, &errValues))
				assert.EqualError(t, errors.Unwrap(errValues), tc.err)

				return
			}

			assert.NoError(t, err)
			require.Len(t, actual, len(tc.filters))

			for _, filter := range actual {
				if filter.Name() != filterTemplate {
					continue
				}

				template, ok := filter.(*TemplateBytesFilter)

				require.True(t, ok)
				assert.Equal(t, tc.expected, template.data.Values)
			}
		})
	}
}

func TestTemplateFilterMissingKeys(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		expected string
		err      string
	}{
		{
			"ShouldRenderExistingKey",
			"value: {{ .Values.Example.Value }}",
			"value: light",
			"",
		},
		{
			"ShouldErrorOnMissingTopLevelKey",
			"value: {{ .Values.Missing }}",
			"",
			"template: config.template:1:17: executing \"config.template\" at <.Values.Missing>: map has no entry for key \"Missing\"",
		},
		{
			"ShouldErrorOnMissingNestedKey",
			"value: {{ .Values.Example.Missing }}",
			"",
			"template: config.template:1:17: executing \"config.template\" at <.Values.Example.Missing>: map has no entry for key \"Missing\"",
		},
		{
			"ShouldErrorOnMissingAutheliaKey",
			"value: {{ .Authelia.Build.Tga }}",
			"",
			"template: config.template:1:19: executing \"config.template\" at <.Authelia.Build.Tga>: map has no entry for key \"Tga\"",
		},
		{
			"ShouldAllowDefaultForMissingKeyViaIndex",
			"value: {{ index .Values \"Missing\" | default \"fallback\" }}",
			"value: fallback",
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filters, err := NewFileFilters([]string{"./test_resources/config_values.values.yml"}, "template")
			require.NoError(t, err)
			require.Len(t, filters, 1)

			out, err := filters[0].Filter([]byte(tc.have))

			if tc.err != "" {
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, out)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, string(out))
			}
		})
	}
}

func TestTemplateFilterShouldRenderLargeJSONIntegersExactly(t *testing.T) {
	filters, err := NewFileFilters([]string{"./test_resources/config_values.numbers.large.json"}, "template")
	require.NoError(t, err)
	require.Len(t, filters, 1)

	out, err := filters[0].Filter([]byte("big: {{ .Values.Big }}\nmax: {{ .Values.Max }}\n"))
	require.NoError(t, err)

	assert.Equal(t, "big: 9223372036854775808\nmax: 18446744073709551615\n", string(out))
}

func TestLoadValuesFile(t *testing.T) {
	testCases := []struct {
		name     string
		path     string
		expected map[string]any
		err      string
	}{
		{
			"ShouldReturnNilOnEmptyPath",
			"",
			nil,
			"",
		},
		{
			"ShouldLoadYAML",
			"./test_resources/config_values.values.yml",
			map[string]any{"Example": map[string]any{"Value": "light"}},
			"",
		},
		{
			"ShouldLoadYAMLLong",
			"./test_resources/config_values.values.yaml",
			map[string]any{"Example": map[string]any{"Value": "light"}},
			"",
		},
		{
			"ShouldLoadJSON",
			"./test_resources/config_values.values.json",
			map[string]any{"Example": map[string]any{"Value": "light"}},
			"",
		},
		{
			"ShouldLoadJSONOmittingTopLevelSchema",
			"./test_resources/config_values.schema.json",
			map[string]any{"Example": map[string]any{"$schema": "nested", "Value": "light"}},
			"",
		},
		{
			"ShouldLoadTOML",
			"./test_resources/config_values.values.toml",
			map[string]any{"Example": map[string]any{"Value": "light"}},
			"",
		},
		{
			"ShouldNormalizeNonStringKeys",
			"./test_resources/config_values.keys.yml",
			map[string]any{
				"Example": map[string]any{
					"1":      "one",
					"Nested": map[string]any{"2": "two", "Keep": "keep"},
				},
				"List": []any{map[string]any{"3": "three"}},
			},
			"",
		},
		{
			"ShouldLoadJSONNumbersWithoutFloatConversion",
			"./test_resources/config_values.numbers.json",
			map[string]any{
				"Integer": int64(3600000),
				"Float":   1.5,
				"Nested":  map[string]any{"List": []any{int64(10000000)}},
			},
			"",
		},
		{
			"ShouldNormalizeTopLevelKeysSameAsNestedKeys",
			"./test_resources/config_values.keys.toplevel.yml",
			map[string]any{
				"1":      "one",
				"true":   "yes",
				"null":   "nothing",
				"Nested": map[string]any{"1": "one", "true": "yes", "null": "nothing"},
			},
			"",
		},
		{
			"ShouldErrorOnTopLevelKeysCollidingAfterNormalization",
			"./test_resources/config_values.keys.toplevel.collision.yml",
			nil,
			"error parsing values file './test_resources/config_values.keys.toplevel.collision.yml': duplicate key '1' after normalizing keys to strings",
		},
		{
			"ShouldErrorOnTopLevelNonMapping",
			"./test_resources/config_values.list.yml",
			nil,
			"error parsing values file './test_resources/config_values.list.yml': the top-level value must be a mapping but it's a []interface {}",
		},
		{
			"ShouldLoadJSONIntegersAboveMaxInt64WithoutRounding",
			"./test_resources/config_values.numbers.large.json",
			map[string]any{
				"Big": uint64(9223372036854775808),
				"Max": uint64(18446744073709551615),
			},
			"",
		},
		{
			"ShouldErrorOnJSONIntegersAboveMaxUint64",
			"./test_resources/config_values.numbers.overflow.json",
			nil,
			"error parsing values file './test_resources/config_values.numbers.overflow.json': integer '18446744073709551616' is out of range",
		},
		{
			"ShouldErrorOnKeysCollidingAfterNormalization",
			"./test_resources/config_values.keys.collision.yml",
			nil,
			"error parsing values file './test_resources/config_values.keys.collision.yml': duplicate key '1' after normalizing keys to strings",
		},
		{
			"ShouldErrorOnMissingFile",
			"./test_resources/this-file-does-not-exist.yml",
			nil,
			"error reading values file: open ./test_resources/this-file-does-not-exist.yml: no such file or directory",
		},
		{
			"ShouldErrorOnUnsupportedExtension",
			"./test_resources/config_values.values.ini",
			nil,
			"error parsing values file './test_resources/config_values.values.ini': unsupported extension '.ini': must be one of '.yml', '.yaml', '.json', or '.toml'",
		},
		{
			"ShouldErrorOnUnrecognizedExtensionWithDotInName",
			"./test_resources/config_values.values",
			nil,
			"error parsing values file './test_resources/config_values.values': unsupported extension '.values': must be one of '.yml', '.yaml', '.json', or '.toml'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := loadValuesFile(tc.path)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, actual)
			} else {
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, actual)
			}
		})
	}

	t.Run("ShouldErrorOnUpperCaseExtension", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "values.YML")

		require.NoError(t, os.WriteFile(path, []byte("Example: value\n"), 0600))

		actual, err := loadValuesFile(path)

		assert.EqualError(t, err, fmt.Sprintf("error parsing values file '%s': unsupported extension '.YML': must be one of '.yml', '.yaml', '.json', or '.toml'", path))
		assert.Nil(t, actual)
	})
}

func TestLoadValuesFiles(t *testing.T) {
	testCases := []struct {
		name     string
		paths    []string
		expected map[string]any
		err      string
	}{
		{
			"ShouldReturnNilForNilSlice",
			nil,
			nil,
			"",
		},
		{
			"ShouldReturnNilForEmptySlice",
			[]string{},
			nil,
			"",
		},
		{
			"ShouldLoadSingleFile",
			[]string{"./test_resources/config_values.values.yml"},
			map[string]any{"Example": map[string]any{"Value": "light"}},
			"",
		},
		{
			"ShouldOverlayLaterFileOverEarlier",
			[]string{
				"./test_resources/config_values.values.yml",
				"./test_resources/config_values.overlay.yml",
			},
			map[string]any{
				"Example": map[string]any{"Value": "dark"},
				"Extra":   map[string]any{"Added": "yes"},
			},
			"",
		},
		{
			"ShouldOverlayEarlierFileOverLater",
			[]string{
				"./test_resources/config_values.overlay.yml",
				"./test_resources/config_values.values.yml",
			},
			map[string]any{
				"Example": map[string]any{"Value": "light"},
				"Extra":   map[string]any{"Added": "yes"},
			},
			"",
		},
		{
			"ShouldOverlayAcrossFormats",
			[]string{
				"./test_resources/config_values.values.json",
				"./test_resources/config_values.overlay.yml",
			},
			map[string]any{
				"Example": map[string]any{"Value": "dark"},
				"Extra":   map[string]any{"Added": "yes"},
			},
			"",
		},
		{
			"ShouldDeepMergeNormalizedNonStringKeys",
			[]string{
				"./test_resources/config_values.keys.yml",
				"./test_resources/config_values.keys.overlay.yml",
			},
			map[string]any{
				"Example": map[string]any{
					"1":      "uno",
					"Nested": map[string]any{"2": "two", "4": "four", "Keep": "keep"},
				},
				"List": []any{map[string]any{"3": "three"}},
			},
			"",
		},
		{
			"ShouldStopOnFirstError",
			[]string{
				"./test_resources/config_values.values.yml",
				"./test_resources/missing.yml",
			},
			nil,
			"error reading values file: open ./test_resources/missing.yml: no such file or directory",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := loadValuesFiles(tc.paths)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, actual)
			} else {
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, actual)
			}
		})
	}
}

func TestNormalizeValues(t *testing.T) {
	testCases := []struct {
		name     string
		have     any
		expected any
	}{
		{
			"ShouldNotModifyScalar",
			"abc",
			"abc",
		},
		{
			"ShouldNotModifyNil",
			nil,
			nil,
		},
		{
			"ShouldNormalizeMapWithNonStringKeys",
			map[any]any{1: "a", true: "b", "c": "d"},
			map[string]any{"1": "a", "true": "b", "c": "d"},
		},
		{
			"ShouldNormalizeNestedMapWithNonStringKeys",
			map[string]any{"a": map[any]any{1: map[any]any{2: "b"}}},
			map[string]any{"a": map[string]any{"1": map[string]any{"2": "b"}}},
		},
		{
			"ShouldNormalizeMapsWithinSlices",
			map[string]any{"a": []any{map[any]any{1: "b"}, "c"}},
			map[string]any{"a": []any{map[string]any{"1": "b"}, "c"}},
		},
		{
			"ShouldNotModifyNormalizedValues",
			map[string]any{"a": map[string]any{"b": []any{"c"}}},
			map[string]any{"a": map[string]any{"b": []any{"c"}}},
		},
		{
			"ShouldNormalizeNilKeyToNull",
			map[any]any{nil: "a"},
			map[string]any{"null": "a"},
		},
		{
			"ShouldErrorOnCollidingKeys",
			map[any]any{1: "a", "1": "b"},
			nil,
		},
		{
			"ShouldErrorOnNestedCollidingKeys",
			map[string]any{"a": []any{map[any]any{true: "b", "true": "c"}}},
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := normalizeValues(tc.have)

			if tc.expected == nil && tc.have != nil {
				assert.ErrorContains(t, err, "after normalizing keys to strings")
				assert.Nil(t, actual)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, actual)
			}
		})
	}
}

func TestMergeValues(t *testing.T) {
	testCases := []struct {
		name     string
		dst      map[string]any
		src      map[string]any
		expected map[string]any
	}{
		{
			"ShouldAddNewKey",
			map[string]any{"a": 1},
			map[string]any{"b": 2},
			map[string]any{"a": 1, "b": 2},
		},
		{
			"ShouldOverwriteScalar",
			map[string]any{"a": 1},
			map[string]any{"a": 2},
			map[string]any{"a": 2},
		},
		{
			"ShouldDeepMergeNestedMaps",
			map[string]any{"a": map[string]any{"x": 1, "y": 2}},
			map[string]any{"a": map[string]any{"y": 20, "z": 30}},
			map[string]any{"a": map[string]any{"x": 1, "y": 20, "z": 30}},
		},
		{
			"ShouldReplaceMapWithScalar",
			map[string]any{"a": map[string]any{"x": 1}},
			map[string]any{"a": "string"},
			map[string]any{"a": "string"},
		},
		{
			"ShouldReplaceScalarWithMap",
			map[string]any{"a": "string"},
			map[string]any{"a": map[string]any{"x": 1}},
			map[string]any{"a": map[string]any{"x": 1}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mergeValues(tc.dst, tc.src)

			assert.Equal(t, tc.expected, tc.dst)
		})
	}
}
