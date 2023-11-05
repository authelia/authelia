package templates

import (
	"crypto/sha1" //nolint:gosec
	"crypto/sha256"
	"crypto/sha512"
	"hash"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFuncGetEnv(t *testing.T) {
	testCases := []struct {
		name     string
		have     map[string]string
		expected map[string]string
	}{
		{"ShouldGetEnv",
			map[string]string{
				"AN_ENV":      "a",
				"ANOTHER_ENV": "b",
			},
			map[string]string{
				"AN_ENV":      "a",
				"ANOTHER_ENV": "b",
			},
		},
		{"ShouldNotGetSecretEnv",
			map[string]string{
				"AUTHELIA_ENV_SECRET": "a",
				"ANOTHER_ENV":         "b",
			},
			map[string]string{
				"AUTHELIA_ENV_SECRET": "",
				"ANOTHER_ENV":         "b",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for key, value := range tc.have {
				t.Setenv(key, value)
			}

			for key, expected := range tc.expected {
				assert.Equal(t, expected, FuncGetEnv(key))
			}
		})
	}
}

func TestFuncExpandEnv(t *testing.T) {
	testCases := []struct {
		name     string
		env      map[string]string
		have     string
		expected string
	}{
		{"ShouldExpandEnv",
			map[string]string{
				"AN_ENV":      "a",
				"ANOTHER_ENV": "b",
			},
			"This is ${AN_ENV} and ${ANOTHER_ENV}",
			"This is a and b",
		},
		{"ShouldNotExpandSecretEnv",
			map[string]string{
				"AUTHELIA_ENV_SECRET":   "a",
				"X_AUTHELIA_ENV_SECRET": "a",
				"ANOTHER_ENV":           "b",
			},
			"This is ${AUTHELIA_ENV_SECRET} and ${ANOTHER_ENV} without ${X_AUTHELIA_ENV_SECRET}",
			"This is  and b without ",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for key, value := range tc.env {
				t.Setenv(key, value)
			}

			assert.Equal(t, tc.expected, FuncExpandEnv(tc.have))
		})
	}
}

func TestFuncHashSum(t *testing.T) {
	testCases := []struct {
		name     string
		new      func() hash.Hash
		have     []string
		expected []string
	}{
		{"ShouldHashSHA1", sha1.New, []string{"abc", "123", "authelia"}, []string{"616263da39a3ee5e6b4b0d3255bfef95601890afd80709", "313233da39a3ee5e6b4b0d3255bfef95601890afd80709", "61757468656c6961da39a3ee5e6b4b0d3255bfef95601890afd80709"}},
		{"ShouldHashSHA256", sha256.New, []string{"abc", "123", "authelia"}, []string{"616263e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", "313233e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", "61757468656c6961e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"}},
		{"ShouldHashSHA512", sha512.New, []string{"abc", "123", "authelia"}, []string{"616263cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e", "313233cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e", "61757468656c6961cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, len(tc.have), len(tc.expected))

			h := FuncHashSum(tc.new)

			for i := 0; i < len(tc.have); i++ {
				assert.Equal(t, tc.expected[i], h(tc.have[i]))
			}
		})
	}
}

func TestFuncStringReplace(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		old, new string
		expected string
	}{
		{"ShouldReplaceSingle", "ABC123", "123", "456", "ABC456"},
		{"ShouldReplaceMultiple", "123ABC123123", "123", "456", "456ABC456456"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, FuncStringReplace(tc.old, tc.new, tc.have))
		})
	}
}

func TestFuncStringContains(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		substr   string
		expected bool
	}{
		{"ShouldMatchNormal", "abc123", "c12", true},
		{"ShouldNotMatchWrongCase", "abc123", "C12", false},
		{"ShouldNotMatchNotContains", "abc123", "xyz", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, FuncStringContains(tc.substr, tc.have))
		})
	}
}

func TestFuncStringHasPrefix(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		substr   string
		expected bool
	}{
		{"ShouldMatchNormal", "abc123", "abc", true},
		{"ShouldNotMatchWrongCase", "abc123", "ABC", false},
		{"ShouldNotMatchNotPrefix", "abc123", "123", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, FuncStringHasPrefix(tc.substr, tc.have))
		})
	}
}

func TestFuncStringHasSuffix(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		substr   string
		expected bool
	}{
		{"ShouldMatchNormal", "abc123xyz", "xyz", true},
		{"ShouldNotMatchWrongCase", "abc123xyz", "XYZ", false},
		{"ShouldNotMatchNotSuffix", "abc123xyz", "123", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, FuncStringHasSuffix(tc.substr, tc.have))
		})
	}
}

func TestFuncStringTrimAll(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		cutset   string
		expected string
	}{
		{"ShouldTrimSuffix", "abc123xyz", "xyz", "abc123"},
		{"ShouldTrimPrefix", "xyzabc123", "xyz", "abc123"},
		{"ShouldNotTrimMiddle", "abcxyz123", "xyz", "abcxyz123"},
		{"ShouldNotTrimWrongCase", "xyzabcxyz123xyz", "XYZ", "xyzabcxyz123xyz"},
		{"ShouldNotTrimWrongChars", "abc123xyz", "456", "abc123xyz"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, FuncStringTrimAll(tc.cutset, tc.have))
		})
	}
}

func TestFuncStringTrimPrefix(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		cutset   string
		expected string
	}{
		{"ShouldNotTrimSuffix", "abc123xyz", "xyz", "abc123xyz"},
		{"ShouldTrimPrefix", "xyzabc123", "xyz", "abc123"},
		{"ShouldNotTrimMiddle", "abcxyz123", "xyz", "abcxyz123"},
		{"ShouldNotTrimWrongCase", "xyzabcxyz123xyz", "XYZ", "xyzabcxyz123xyz"},
		{"ShouldNotTrimWrongChars", "abc123xyz", "456", "abc123xyz"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, FuncStringTrimPrefix(tc.cutset, tc.have))
		})
	}
}

func TestFuncStringTrimSuffix(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		cutset   string
		expected string
	}{
		{"ShouldTrimSuffix", "abc123xyz", "xyz", "abc123"},
		{"ShouldNotTrimPrefix", "xyzabc123", "xyz", "xyzabc123"},
		{"ShouldNotTrimMiddle", "abcxyz123", "xyz", "abcxyz123"},
		{"ShouldNotTrimWrongCase", "xyzabcxyz123xyz", "XYZ", "xyzabcxyz123xyz"},
		{"ShouldNotTrimWrongChars", "abc123xyz", "456", "abc123xyz"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, FuncStringTrimSuffix(tc.cutset, tc.have))
		})
	}
}

func TestFuncElemsJoin(t *testing.T) {
	testCases := []struct {
		name     string
		have     any
		sep      string
		expected string
	}{
		{"ShouldNotJoinNonElements", "abc123xyz", "xyz", "abc123xyz"},
		{"ShouldJoinStrings", []string{"abc", "123"}, "xyz", "abcxyz123"},
		{"ShouldJoinInts", []int{1, 2, 3}, ",", "1,2,3"},
		{"ShouldJoinBooleans", []bool{true, false, true}, ".", "true.false.true"},
		{"ShouldJoinBytes", [][]byte{[]byte("abc"), []byte("123"), []byte("a")}, "$", "abc$123$a"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, FuncElemsJoin(tc.sep, tc.have))
		})
	}
}

func TestFuncIterate(t *testing.T) {
	uintptr := func(in uint) *uint {
		return &in
	}

	testCases := []struct {
		name     string
		have     *uint
		expected []uint
	}{
		{"ShouldGiveZeroResults", uintptr(0), nil},
		{"ShouldGive10Results", uintptr(10), []uint{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, FuncIterate(tc.have))
		})
	}
}

func TestFuncStringsSplit(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		sep      string
		expected map[string]string
	}{
		{"ShouldSplit", "abc,123,456", ",", map[string]string{"_0": "abc", "_1": "123", "_2": "456"}},
		{"ShouldNotSplit", "abc,123,456", "$", map[string]string{"_0": "abc,123,456"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, FuncStringSplit(tc.sep, tc.have))
		})
	}
}

func TestFuncStringSplitList(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		sep      string
		expected []string
	}{
		{"ShouldSplit", "abc,123,456", ",", []string{"abc", "123", "456"}},
		{"ShouldNotSplit", "abc,123,456", "$", []string{"abc,123,456"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, FuncStringSplitList(tc.sep, tc.have))
		})
	}
}

func TestFuncKeys(t *testing.T) {
	testCases := []struct {
		name     string
		have     []map[string]any
		expected []string
	}{
		{"ShouldProvideKeysSingle", []map[string]any{{"a": "v", "b": "v", "z": "v"}}, []string{"a", "b", "z"}},
		{"ShouldProvideKeysMultiple", []map[string]any{{"a": "v", "b": "v", "z": "v"}, {"h": "v"}}, []string{"a", "b", "z", "h"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			keys := FuncKeys(tc.have...)

			assert.Len(t, keys, len(tc.expected))

			for _, expected := range tc.expected {
				assert.Contains(t, keys, expected)
			}
		})
	}
}

func TestFuncSortAlpha(t *testing.T) {
	testCases := []struct {
		name     string
		have     any
		expected []string
	}{
		{"ShouldSortStrings", []string{"a", "c", "b"}, []string{"a", "b", "c"}},
		{"ShouldSortIntegers", []int{2, 3, 1}, []string{"1", "2", "3"}},
		{"ShouldSortSingleValue", 1, []string{"1"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, FuncSortAlpha(tc.have))
		})
	}
}

func TestFuncBEnc(t *testing.T) {
	testCases := []struct {
		name       string
		have       string
		expected32 string
		expected64 string
	}{
		{"ShouldEncodeEmptyString", "", "", ""},
		{"ShouldEncodeString", "abc", "MFRGG===", "YWJj"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("Base32", func(t *testing.T) {
				assert.Equal(t, tc.expected32, FuncB32Enc(tc.have))
			})

			t.Run("Base64", func(t *testing.T) {
				assert.Equal(t, tc.expected64, FuncB64Enc(tc.have))
			})
		})
	}
}

func TestFuncBDec(t *testing.T) {
	testCases := []struct {
		name              string
		have              string
		err32, expected32 string
		err64, expected64 string
	}{
		{"ShouldDecodeEmptyString", "", "", "", "", ""},
		{"ShouldDecodeBase32", "MFRGG===", "", "abc", "illegal base64 data at input byte 5", ""},
		{"ShouldDecodeBase64", "YWJj", "illegal base32 data at input byte 3", "", "", "abc"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				actual string
				err    error
			)

			t.Run("Base32", func(t *testing.T) {
				actual, err = FuncB32Dec(tc.have)

				if tc.err32 != "" {
					assert.Equal(t, "", actual)
					assert.EqualError(t, err, tc.err32)
				} else {
					assert.Equal(t, tc.expected32, actual)
					assert.NoError(t, err)
				}
			})

			t.Run("Base64", func(t *testing.T) {
				actual, err = FuncB64Dec(tc.have)

				if tc.err64 != "" {
					assert.Equal(t, "", actual)
					assert.EqualError(t, err, tc.err64)
				} else {
					assert.Equal(t, tc.expected64, actual)
					assert.NoError(t, err)
				}
			})
		})
	}
}

func TestFuncStringQuote(t *testing.T) {
	testCases := []struct {
		name     string
		have     []any
		expected string
	}{
		{"ShouldQuoteSingleValue", []any{"abc"}, `"abc"`},
		{"ShouldQuoteMultiValue", []any{"abc", 123}, `"abc" "123"`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, FuncStringQuote(tc.have...))
		})
	}
}

func TestFuncStringSQuote(t *testing.T) {
	testCases := []struct {
		name     string
		have     []any
		expected string
	}{
		{"ShouldQuoteSingleValue", []any{"abc"}, `'abc'`},
		{"ShouldQuoteMultiValue", []any{"abc", 123}, `'abc' '123'`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, FuncStringSQuote(tc.have...))
		})
	}
}

func TestFuncTypeOf(t *testing.T) {
	astring := "typeOfExample"
	anint := 5
	astringslice := []string{astring}
	anintslice := []int{anint}

	testCases := []struct {
		name         string
		have         any
		expected     string
		expectedKind string
	}{
		{"String", astring, "string", "string"},
		{"StringPtr", &astring, "*string", "ptr"},
		{"StringSlice", astringslice, "[]string", "slice"},
		{"StringSlicePtr", &astringslice, "*[]string", "ptr"},
		{"Integer", anint, "int", "int"},
		{"IntegerPtr", &anint, "*int", "ptr"},
		{"IntegerSlice", anintslice, "[]int", "slice"},
		{"IntegerSlicePtr", &anintslice, "*[]int", "ptr"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, FuncTypeOf(tc.have))
			assert.Equal(t, tc.expectedKind, FuncKindOf(tc.have))
		})
	}
}

func TestFuncTypeIs(t *testing.T) {
	astring := "typeIsExample"
	anint := 10
	astringslice := []string{astring}
	anintslice := []int{anint}

	testCases := []struct {
		name         string
		is           string
		have         any
		expected     bool
		expectedLike bool
		expectedKind bool
	}{
		{"ShouldMatchStringAsString", "string", astring, true, true, true},
		{"ShouldMatchStringPtrAsString", "string", &astring, false, true, false},
		{"ShouldNotMatchStringAsInt", "int", astring, false, false, false},
		{"ShouldNotMatchStringSliceAsStringSlice", "[]string", astringslice, true, true, false},
		{"ShouldNotMatchStringSlicePtrAsStringSlice", "[]string", &astringslice, false, true, false},
		{"ShouldNotMatchStringSlicePtrAsStringSlicePtr", "*[]string", &astringslice, true, true, false},
		{"ShouldNotMatchStringSliceAsString", "string", astringslice, false, false, false},
		{"ShouldMatchIntAsInt", "int", anint, true, true, true},
		{"ShouldMatchIntPtrAsInt", "int", &anint, false, true, false},
		{"ShouldNotMatchIntAsString", "string", anint, false, false, false},
		{"ShouldMatchIntegerSliceAsIntSlice", "[]int", anintslice, true, true, false},
		{"ShouldMatchIntegerSlicePtrAsIntSlice", "[]int", &anintslice, false, true, false},
		{"ShouldMatchIntegerSlicePtrAsIntSlicePtr", "*[]int", &anintslice, true, true, false},
		{"ShouldNotMatchIntegerSliceAsInt", "int", anintslice, false, false, false},
		{"ShouldMatchKindSlice", "slice", anintslice, false, false, true},
		{"ShouldMatchKindPtr", "ptr", &anintslice, false, false, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, FuncTypeIs(tc.is, tc.have))
			assert.Equal(t, tc.expectedLike, FuncTypeIsLike(tc.is, tc.have))
			assert.Equal(t, tc.expectedKind, FuncKindIs(tc.is, tc.have))
		})
	}
}

func TestFuncList(t *testing.T) {
	assert.Equal(t, []any{"a", "b", "c"}, FuncList("a", "b", "c"))
	assert.Equal(t, []any{1, 2, 3}, FuncList(1, 2, 3))
}

func TestFuncDict(t *testing.T) {
	assert.Equal(t, map[string]any{"a": 1}, FuncDict("a", 1))
	assert.Equal(t, map[string]any{"a": 1, "b": ""}, FuncDict("a", 1, "b"))
	assert.Equal(t, map[string]any{"1": 1, "b": 2}, FuncDict(1, 1, "b", 2))
	assert.Equal(t, map[string]any{"true": 1, "b": 2}, FuncDict(true, 1, "b", 2))
	assert.Equal(t, map[string]any{"a": 2, "b": 3}, FuncDict("a", 1, "a", 2, "b", 3))
}

func TestFuncGet(t *testing.T) {
	assert.Equal(t, 123, FuncGet(map[string]any{"abc": 123}, "abc"))
	assert.Equal(t, "", FuncGet(map[string]any{"abc": 123}, "123"))
}

func TestFuncSet(t *testing.T) {
	assert.Equal(t, map[string]any{"abc": 123, "123": true}, FuncSet(map[string]any{"abc": 123}, "123", true))
	assert.Equal(t, map[string]any{"abc": true}, FuncSet(map[string]any{"abc": 123}, "abc", true))
}

func TestFuncDefault(t *testing.T) {
	testCases := []struct {
		name     string
		value    []any
		have     any
		expected any
	}{
		{"ShouldDefaultEmptyString", []any{""}, "default", "default"},
		{"ShouldNotDefaultString", []any{"not default"}, "default", "not default"},
		{"ShouldDefaultEmptyInteger", []any{0}, 1, 1},
		{"ShouldNotDefaultInteger", []any{20}, 1, 20},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, FuncDefault(tc.have, tc.value...))
		})
	}
}

func TestFuncEmpty(t *testing.T) {
	var nilv *string

	testCases := []struct {
		name     string
		value    any
		expected bool
	}{
		{"ShouldBeEmptyNil", nilv, true},
		{"ShouldBeEmptyNilNil", nil, true},
		{"ShouldBeEmptyString", "", true},
		{"ShouldNotBeEmptyString", "abc", false},
		{"ShouldBeEmptyArray", []string{}, true},
		{"ShouldNotBeEmptyArray", []string{"abc"}, false},
		{"ShouldBeEmptyInteger", 0, true},
		{"ShouldNotBeEmptyInteger", 1, false},
		{"ShouldBeEmptyInteger8", int8(0), true},
		{"ShouldNotBeEmptyInteger8", int8(1), false},
		{"ShouldBeEmptyInteger16", int16(0), true},
		{"ShouldNotBeEmptyInteger16", int16(1), false},
		{"ShouldBeEmptyInteger32", int32(0), true},
		{"ShouldNotBeEmptyInteger32", int32(1), false},
		{"ShouldBeEmptyInteger64", int64(0), true},
		{"ShouldNotBeEmptyInteger64", int64(1), false},
		{"ShouldBeEmptyUnsignedInteger", uint(0), true},
		{"ShouldNotBeEmptyUnsignedInteger", uint(1), false},
		{"ShouldBeEmptyUnsignedInteger8", uint8(0), true},
		{"ShouldNotBeEmptyUnsignedInteger8", uint8(1), false},
		{"ShouldBeEmptyUnsignedInteger16", uint16(0), true},
		{"ShouldNotBeEmptyUnsignedInteger16", uint16(1), false},
		{"ShouldBeEmptyUnsignedInteger32", uint32(0), true},
		{"ShouldNotBeEmptyUnsignedInteger32", uint32(1), false},
		{"ShouldBeEmptyUnsignedInteger64", uint64(0), true},
		{"ShouldNotBeEmptyUnsignedInteger64", uint64(1), false},
		{"ShouldBeEmptyComplex64", complex64(complex(0, 0)), true},
		{"ShouldNotBeEmptyComplex64", complex64(complex(100000, 7.5)), false},
		{"ShouldBeEmptyComplex128", complex128(complex(0, 0)), true},
		{"ShouldNotBeEmptyComplex128", complex128(complex(100000, 7.5)), false},
		{"ShouldBeEmptyFloat32", float32(0), true},
		{"ShouldNotBeEmptyFloat32", float32(1), false},
		{"ShouldBeEmptyFloat64", float64(0), true},
		{"ShouldNotBeEmptyFloat64", float64(1), false},
		{"ShouldBeEmptyBoolean", false, true},
		{"ShouldNotBeEmptyBoolean", true, false},
		{"ShouldNotBeEmptyStruct", struct{}{}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, FuncEmpty(tc.value))
		})
	}
}

func TestFuncIndent(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		indent   int
		expected []string
	}{
		{"ShouldIndentZeroMultiLine", "abc\n123", 0, []string{"abc\n123", "\nabc\n123"}},
		{"ShouldIndentOneMultiLine", "abc\n123", 1, []string{" abc\n 123", "\n abc\n 123"}},
		{"ShouldIndentOneSingleLine", "abc", 1, []string{" abc", "\n abc"}},
		{"ShouldIndentZeroSingleLine", "abc", 0, []string{"abc", "\nabc"}},
	}

	for _, tc := range testCases {
		for i, f := range []func(i int, v string) string{FuncIndent, FuncNewlineIndent} {
			assert.Equal(t, tc.expected[i], f(tc.indent, tc.have))
		}
	}
}

func TestFuncMultiLineIndent(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		indent   int
		expected string
	}{
		{"ShouldIndentZeroMultiLine", "abc\n123", 0, "|\nabc\n123"},
		{"ShouldIndentOneMultiLine", "abc\n123", 1, "|\n abc\n 123"},
		{"ShouldIndentOneSingleLine", "abc", 1, "abc"},
		{"ShouldIndentZeroSingleLine", "abc", 0, "abc"},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, FuncMultilineIndent(tc.indent, "|", tc.have))
	}
}

func TestMultiLineQuote(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		char     rune
		expected string
	}{
		{
			"ShouldSQuoteSingleLine",
			"abc",
			rune(39),
			`'abc'`,
		},
		{
			"ShouldQuoteSingleLine",
			"abc",
			rune(34),
			`"abc"`,
		},
		{
			"ShouldNotQuoteLine",
			"abc\n123",
			rune(39),
			"abc\n123",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			quote := FuncStringQuoteMultiLine(tc.char)

			actual := quote(tc.have)

			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestFuncUUIDv4(t *testing.T) {
	assert.Len(t, FuncUUIDv4(), 36)
}

func TestFuncFileContent(t *testing.T) {
	testCases := []struct {
		name           string
		path           string
		expected       string
		expectedSecret string
		expectedErr    string
	}{
		{
			"ShouldReadFile",
			"../configuration/test_resources/example_secret",
			"example_secret value\n",
			"example_secret value",
			"",
		},
		{
			"ShouldNotReadBadFile",
			"../configuration/test_resources/example_secretx",
			"",
			"",
			"open ../configuration/test_resources/example_secretx: no such file or directory",
		},
	}

	for _, tc := range testCases {
		actual, theErr := FuncFileContent(tc.path)

		assert.Equal(t, tc.expected, actual)

		if tc.expectedErr != "" {
			assert.EqualError(t, theErr, tc.expectedErr)
		} else {
			assert.NoError(t, theErr)
		}

		actual, theErr = FuncSecret(tc.path)

		assert.Equal(t, tc.expectedSecret, actual)

		if tc.expectedErr != "" {
			assert.EqualError(t, theErr, tc.expectedErr)
		} else {
			assert.NoError(t, theErr)
		}
	}
}
