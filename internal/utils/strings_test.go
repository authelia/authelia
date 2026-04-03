package utils

import (
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsStringAbsURL(t *testing.T) {
	testCases := []struct {
		name string
		have string
		err  string
	}{
		{
			"ShouldBeAbs",
			"https://google.com",
			"",
		},
		{
			"ShouldNotBeAbs",
			"google.com",
			"could not parse 'google.com' as a URL",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			theError := IsStringAbsURL(tc.have)

			if tc.err == "" {
				assert.NoError(t, theError)
			} else {
				assert.EqualError(t, theError, tc.err)
			}
		})
	}
}

func FuzzIsStringAbsURL(f *testing.F) {
	f.Add("https://google.com")
	f.Add("https://example.com")
	f.Add("https://abc.com")
	f.Fuzz(func(t *testing.T, s string) {
		assert.NoError(t, IsStringAbsURL(s))
	})
}

func TestIsStringInSliceF(t *testing.T) {
	testCases := []struct {
		name     string
		needle   string
		haystack []string
		isEqual  func(needle, item string) bool
		expected bool
	}{
		{
			"ShouldBePresent",
			"good",
			[]string{"good"},
			func(needle, item string) bool {
				return needle == item
			},
			true,
		},
		{
			"ShouldNotBePresent",
			"bad",
			[]string{"good"},
			func(needle, item string) bool {
				return needle == item
			},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, IsStringInSliceF(tc.needle, tc.haystack, tc.isEqual))
		})
	}
}

func FuzzIsStringInSliceF(f *testing.F) {
	a := func(needle, item string) bool {
		return needle == item
	}

	f.Add("abc", "abc,123,456")
	f.Add("abc", "123,abc,456")
	f.Add("456", "123,abc,456")
	f.Fuzz(func(t *testing.T, n, h string) {
		haystack := strings.Split(h, ",")
		assert.True(t, IsStringInSliceF(n, haystack, a))
		assert.True(t, IsStringInSliceF(n, haystack, strings.EqualFold))
	})
}

func TestStringHTMLEscape(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		expected string
	}{
		{
			"ShouldNotAlterAlphaNum",
			"abc123",
			"abc123",
		},
		{
			"ShouldEscapeSpecial",
			"abc123><@#@",
			"abc123&gt;&lt;@#@",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, StringHTMLEscape(tc.have))
		})
	}
}

func TestStringSplitDelimitedEscaped(t *testing.T) {
	testCases := []struct {
		desc, have string
		delimiter  rune
		want       []string
	}{
		{desc: "ShouldSplitNormalString", have: "abc,123,456", delimiter: ',', want: []string{"abc", "123", "456"}},
		{desc: "ShouldSplitEscapedString", have: "a\\,bc,123,456", delimiter: ',', want: []string{"a,bc", "123", "456"}},
		{desc: "ShouldSplitEscapedStringPipe", have: "a\\|bc|123|456", delimiter: '|', want: []string{"a|bc", "123", "456"}},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			actual := StringSplitDelimitedEscaped(tc.have, tc.delimiter)

			assert.Equal(t, tc.want, actual)
		})
	}
}

func TestStringJoinDelimitedEscaped(t *testing.T) {
	testCases := []struct {
		desc, want string
		delimiter  rune
		have       []string
	}{
		{desc: "ShouldJoinNormalStringSlice", have: []string{"abc", "123", "456"}, delimiter: ',', want: "abc,123,456"},
		{desc: "ShouldJoinEscapeNeededStringSlice", have: []string{"abc", "1,23", "456"}, delimiter: ',', want: "abc,1\\,23,456"},
		{desc: "ShouldJoinEscapeNeededStringSlicePipe", have: []string{"abc", "1|23", "456"}, delimiter: '|', want: "abc|1\\|23|456"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			actual := StringJoinDelimitedEscaped(tc.have, tc.delimiter)

			assert.Equal(t, tc.want, actual)

			// Ensure splitting again also works fine.
			split := StringSplitDelimitedEscaped(actual, tc.delimiter)

			assert.Equal(t, tc.have, split)
		})
	}
}

func TestShouldDetectAlphaNumericString(t *testing.T) {
	assert.True(t, IsStringAlphaNumeric("abc"))
	assert.True(t, IsStringAlphaNumeric("abc123"))
	assert.False(t, IsStringAlphaNumeric("abc123@"))
}

func TestShouldSplitIntoEvenStringsOfFour(t *testing.T) {
	input := testStringInput

	arrayOfStrings := SliceString(input, 4)

	assert.Equal(t, len(arrayOfStrings), 3)
	assert.Equal(t, "abcd", arrayOfStrings[0])
	assert.Equal(t, "efgh", arrayOfStrings[1])
	assert.Equal(t, "ijkl", arrayOfStrings[2])
}

func TestShouldSplitIntoEvenStringsOfOne(t *testing.T) {
	input := testStringInput

	arrayOfStrings := SliceString(input, 1)

	assert.Equal(t, 12, len(arrayOfStrings))
	assert.Equal(t, "a", arrayOfStrings[0])
	assert.Equal(t, "b", arrayOfStrings[1])
	assert.Equal(t, "c", arrayOfStrings[2])
	assert.Equal(t, "d", arrayOfStrings[3])
	assert.Equal(t, "l", arrayOfStrings[11])
}

func TestShouldSplitIntoUnevenStringsOfFour(t *testing.T) {
	input := testStringInput + "m"

	arrayOfStrings := SliceString(input, 4)

	assert.Equal(t, len(arrayOfStrings), 4)
	assert.Equal(t, "abcd", arrayOfStrings[0])
	assert.Equal(t, "efgh", arrayOfStrings[1])
	assert.Equal(t, "ijkl", arrayOfStrings[2])
	assert.Equal(t, "m", arrayOfStrings[3])
}

func TestShouldFindSliceDifferencesDelta(t *testing.T) {
	before := []string{"abc", "onetwothree"}
	after := []string{"abc", "xyz"}

	added, removed := StringSlicesDelta(before, after)

	require.Len(t, added, 1)
	require.Len(t, removed, 1)
	assert.Equal(t, "onetwothree", removed[0])
	assert.Equal(t, "xyz", added[0])
}

func TestShouldNotFindSliceDifferencesDelta(t *testing.T) {
	before := []string{"abc", "onetwothree"}
	after := []string{"abc", "onetwothree"}

	added, removed := StringSlicesDelta(before, after)

	require.Len(t, added, 0)
	require.Len(t, removed, 0)
}

func TestShouldFindSliceDifferences(t *testing.T) {
	a := []string{"abc", "onetwothree"}
	b := []string{"abc", "xyz"}

	assert.True(t, IsStringSlicesDifferent(a, b))
	assert.True(t, IsStringSlicesDifferentFold(a, b))

	c := []string{"Abc", "xyz"}

	assert.True(t, IsStringSlicesDifferent(b, c))
	assert.False(t, IsStringSlicesDifferentFold(b, c))
}

func TestShouldNotFindSliceDifferences(t *testing.T) {
	a := []string{"abc", "onetwothree"}
	b := []string{"abc", "onetwothree"}

	assert.False(t, IsStringSlicesDifferent(a, b))
	assert.False(t, IsStringSlicesDifferentFold(a, b))
}

func TestShouldFindSliceDifferenceWhenDifferentLength(t *testing.T) {
	a := []string{"abc", "onetwothree"}
	b := []string{"abc", "onetwothree", "more"}

	assert.True(t, IsStringSlicesDifferent(a, b))
	assert.True(t, IsStringSlicesDifferentFold(a, b))
}

func TestShouldFindStringInSliceContains(t *testing.T) {
	a := "abc"
	slice := []string{"abc", "onetwothree"}

	assert.True(t, IsStringInSliceContains(a, slice))
}

func TestShouldNotFindStringInSliceContains(t *testing.T) {
	a := "xyz"
	slice := []string{"abc", "onetwothree"}

	assert.False(t, IsStringInSliceContains(a, slice))
}

func TestShouldFindStringInSliceFold(t *testing.T) {
	a := "xYz"
	b := "AbC"
	slice := []string{"XYz", "abc"}

	assert.True(t, IsStringInSliceFold(a, slice))
	assert.True(t, IsStringInSliceFold(b, slice))
}

func TestShouldNotFindStringInSliceFold(t *testing.T) {
	a := "xyZ"
	b := "ABc"
	slice := []string{"cba", "zyx"}

	assert.False(t, IsStringInSliceFold(a, slice))
	assert.False(t, IsStringInSliceFold(b, slice))
}

func TestIsStringSliceContainsAll(t *testing.T) {
	needles := []string{"abc", "123", "xyz"}
	haystackOne := []string{"abc", "tvu", "123", "456", "xyz"}
	haystackTwo := []string{"tvu", "123", "456", "xyz"}

	assert.True(t, IsStringSliceContainsAll(needles, haystackOne))
	assert.False(t, IsStringSliceContainsAll(needles, haystackTwo))
}

func TestIsStringSliceContainsAny(t *testing.T) {
	needles := []string{"abc", "123", "xyz"}
	haystackOne := []string{"tvu", "456", "hij"}
	haystackTwo := []string{"tvu", "123", "456", "xyz"}

	assert.False(t, IsStringSliceContainsAny(needles, haystackOne))
	assert.True(t, IsStringSliceContainsAny(needles, haystackTwo))
}

func TestStringSliceURLConversionFuncs(t *testing.T) {
	urls := URLsFromStringSlice([]string{"https://google.com", "abc", "%*()@#$J(@*#$J@#($H"})

	require.Len(t, urls, 2)
	assert.Equal(t, "https://google.com", urls[0].String())
	assert.Equal(t, "abc", urls[1].String())

	strs := StringSliceFromURLs(urls)

	require.Len(t, strs, 2)
	assert.Equal(t, "https://google.com", strs[0])
	assert.Equal(t, "abc", strs[1])
}

func TestOriginFromURL(t *testing.T) {
	google, err := url.Parse("https://google.com/abc?a=123#five")
	assert.NoError(t, err)

	origin := OriginFromURL(google)
	assert.Equal(t, "https://google.com", origin.String())
}

func TestJoinAndCanonicalizeHeaders(t *testing.T) {
	result := JoinAndCanonicalizeHeaders([]byte(", "), "x-example-ONE", "X-EGG-Two")

	assert.Equal(t, []byte("X-Example-One, X-Egg-Two"), result)
}

func TestBuildStringFuncsMissingTests(t *testing.T) {
	assert.Equal(t, "", StringJoinBuild(".", ":", "'", nil))
	assert.Equal(t, "'abc', '123'", StringJoinComma("", []string{"abc", "123"}))
}

func TestStringJoinOr(t *testing.T) {
	testCases := []struct {
		name     string
		items    []string
		expected string
	}{
		{"Multiple items", []string{"apple", "banana", "cherry"}, "'apple', 'banana', or 'cherry'"},
		{"Single item", []string{"apple"}, "'apple'"},
		{"Empty slice", []string{}, ""},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			result := StringJoinOr(test.items)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestStringJoinAnd(t *testing.T) {
	tests := []struct {
		name     string
		items    []string
		expected string
	}{
		{"Multiple items", []string{"apple", "banana", "cherry"}, "'apple', 'banana', and 'cherry'"},
		{"Single item", []string{"apple"}, "'apple'"},
		{"Empty slice", []string{}, ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := StringJoinAnd(test.items)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestStringJoinComma(t *testing.T) {
	tests := []struct {
		name     string
		word     string
		items    []string
		expected string
	}{
		{"Multiple items with 'or'", "or", []string{"apple", "banana", "cherry"}, "'apple', 'banana', or 'cherry'"},
		{"Multiple items with 'and'", "and", []string{"apple", "banana", "cherry"}, "'apple', 'banana', and 'cherry'"},
		{"Single item", "", []string{"apple"}, "'apple'"},
		{"Empty slice", "", []string{}, ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := StringJoinComma(test.word, test.items)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestStringJoinBuild(t *testing.T) {
	tests := []struct {
		name     string
		sep      string
		sepFinal string
		quote    string
		items    []string
		expected string
	}{
		{"Multiple items with comma", ",", "and", "'", []string{"apple", "banana", "cherry"}, "'apple', 'banana', and 'cherry'"},
		{"Multiple items with semicolon", ";", "or", "\"", []string{"apple", "banana", "cherry"}, "\"apple\"; \"banana\"; or \"cherry\""},
		{"Single item", ",", "", "'", []string{"apple"}, "'apple'"},
		{"Empty slice", ",", "and", "'", []string{}, ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := StringJoinBuild(test.sep, test.sepFinal, test.quote, test.items)
			assert.Equal(t, test.expected, result)
		})
	}
}
