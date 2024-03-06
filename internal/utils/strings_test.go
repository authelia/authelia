package utils

import (
	"net/url"
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

func TestIsURLInSlice(t *testing.T) {
	urls := URLsFromStringSlice([]string{"https://google.com", "https://example.com"})

	google, err := url.ParseRequestURI("https://google.com")
	assert.NoError(t, err)

	microsoft, err := url.ParseRequestURI("https://microsoft.com")
	assert.NoError(t, err)

	example, err := url.ParseRequestURI("https://example.com")
	assert.NoError(t, err)

	assert.True(t, IsURLInSlice(google, urls))
	assert.False(t, IsURLInSlice(microsoft, urls))
	assert.True(t, IsURLInSlice(example, urls))
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

func TestIsURLHostComponent(t *testing.T) {
	testCases := []struct {
		desc, have           string
		expectedA, expectedB bool
	}{
		{
			desc:      "ShouldBeFalseWithScheme",
			have:      "https://google.com",
			expectedA: false, expectedB: false,
		},
		{
			desc:      "ShouldBeTrueForHostComponentButFalseForWithPort",
			have:      "google.com",
			expectedA: true, expectedB: false,
		},
		{
			desc:      "ShouldBeFalseForHostComponentButTrueForWithPort",
			have:      "google.com:8000",
			expectedA: false, expectedB: true,
		},
		{
			desc:      "ShouldBeFalseWithPath",
			have:      "google.com:8000/path",
			expectedA: false, expectedB: false,
		},
		{
			desc:      "ShouldBeFalseWithFragment",
			have:      "google.com:8000#test",
			expectedA: false, expectedB: false,
		},
		{
			desc:      "ShouldBeFalseWithQuery",
			have:      "google.com:8000?test=1",
			expectedA: false, expectedB: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			u, err := url.Parse(tc.have)

			require.NoError(t, err)
			require.NotNil(t, u)

			assert.Equal(t, tc.expectedA, IsURLHostComponent(u))
			assert.Equal(t, tc.expectedB, IsURLHostComponentWithPort(u))
		})
	}
}
