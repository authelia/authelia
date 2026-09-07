package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertAttributeValueBoolean(t *testing.T) {
	testCases := []struct {
		name        string
		value       any
		expected    bool
		expectedErr string
	}{
		{name: "ShouldPassThroughBool", value: true, expected: true},
		{name: "ShouldParseTrueVariant", value: "TRUE", expected: true},
		{name: "ShouldParseTVariant", value: "t", expected: true},
		{name: "ShouldParseOneVariant", value: "1", expected: true},
		{name: "ShouldParseYesVariant", value: "yes", expected: true},
		{name: "ShouldParseYVariant", value: "Y", expected: true},
		{name: "ShouldParseFalseVariant", value: "FALSE", expected: false},
		{name: "ShouldParseFVariant", value: "f", expected: false},
		{name: "ShouldParseZeroVariant", value: "0", expected: false},
		{name: "ShouldParseNoVariant", value: "no", expected: false},
		{name: "ShouldParseNVariant", value: "N", expected: false},
		{name: "ShouldParseEmptyStringAsFalse", value: "", expected: false},
		{name: "ShouldErrorOnInvalidString", value: "maybe", expectedErr: "invalid boolean value: maybe"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := ConvertAttributeValue(tc.value, AttributeValueTypeBoolean, false)

			if tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestConvertAttributeValueInteger(t *testing.T) {
	testCases := []struct {
		name        string
		value       any
		expected    int64
		expectedErr string
	}{
		{name: "ShouldConvertInt", value: int(42), expected: 42},
		{name: "ShouldPassThroughInt64", value: int64(42), expected: 42},
		{name: "ShouldConvertFloat64FromJSON", value: float64(42), expected: 42},
		{name: "ShouldParseNumericString", value: "42", expected: 42},
		{name: "ShouldErrorOnInvalidString", value: "abc", expectedErr: `strconv.ParseInt: parsing "abc": invalid syntax`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := ConvertAttributeValue(tc.value, AttributeValueTypeInteger, false)

			if tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestConvertAttributeValueString(t *testing.T) {
	testCases := []struct {
		name      string
		valueType string
		value     any
		expected  string
	}{
		{name: "ShouldPassThroughStringWithExplicitType", valueType: AttributeValueTypeString, value: "hello", expected: "hello"},
		{name: "ShouldPassThroughStringWithEmptyType", valueType: "", value: "hello", expected: "hello"},
		{name: "ShouldStringifyNonStringValue", valueType: AttributeValueTypeString, value: 42, expected: "42"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := ConvertAttributeValue(tc.value, tc.valueType, false)

			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestConvertAttributeValueUnknownTypePassesThrough(t *testing.T) {
	actual, err := ConvertAttributeValue(42, "unknown-type", false)

	require.NoError(t, err)
	assert.Equal(t, 42, actual)
}

func TestConvertAttributeValueMultiValued(t *testing.T) {
	testCases := []struct {
		name     string
		value    any
		expected []interface{}
	}{
		{
			name:     "ShouldPassThroughInterfaceSlice",
			value:    []interface{}{"a", "b"},
			expected: []interface{}{"a", "b"},
		},
		{
			name:     "ShouldConvertStringSlice",
			value:    []string{"a", "b"},
			expected: []interface{}{"a", "b"},
		},
		{
			name:     "ShouldWrapScalarValue",
			value:    "solo",
			expected: []interface{}{"solo"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := ConvertAttributeValue(tc.value, AttributeValueTypeString, true)

			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
