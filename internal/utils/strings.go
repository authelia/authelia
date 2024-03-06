package utils

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"unicode"

	"github.com/valyala/fasthttp"
)

// IsStringAbsURL checks a string can be parsed as a URL and that is IsAbs and if it can't it returns an error
// describing why.
func IsStringAbsURL(input string) (err error) {
	parsedURL, err := url.ParseRequestURI(input)
	if err != nil {
		return fmt.Errorf("could not parse '%s' as a URL", input)
	}

	if !parsedURL.IsAbs() {
		return fmt.Errorf("the url '%s' is not absolute because it doesn't start with a scheme like 'http://' or 'https://'", input)
	}

	return nil
}

// IsStringAlphaNumeric returns false if any rune in the string is not alphanumeric.
func IsStringAlphaNumeric(input string) bool {
	for _, r := range input {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			return false
		}
	}

	return true
}

// IsStringInSlice checks if a single string is in a slice of strings.
func IsStringInSlice(needle string, haystack []string) (inSlice bool) {
	for _, b := range haystack {
		if b == needle {
			return true
		}
	}

	return false
}

// IsStringInSliceF checks if a single string is in a slice of strings using the provided isEqual func.
func IsStringInSliceF(needle string, haystack []string, isEqual func(needle, item string) bool) (inSlice bool) {
	for _, b := range haystack {
		if isEqual(needle, b) {
			return true
		}
	}

	return false
}

// IsStringInSliceFold checks if a single string is in a slice of strings but uses strings.EqualFold to compare them.
func IsStringInSliceFold(needle string, haystack []string) (inSlice bool) {
	for _, b := range haystack {
		if strings.EqualFold(b, needle) {
			return true
		}
	}

	return false
}

// IsStringInSliceContains checks if a single string is in an array of strings.
func IsStringInSliceContains(needle string, haystack []string) (inSlice bool) {
	for _, b := range haystack {
		if strings.Contains(needle, b) {
			return true
		}
	}

	return false
}

// IsStringSliceContainsAll checks if the haystack contains all strings in the needles.
func IsStringSliceContainsAll(needles []string, haystack []string) (inSlice bool) {
	for _, n := range needles {
		if !IsStringInSlice(n, haystack) {
			return false
		}
	}

	return true
}

// IsStringSliceContainsAny checks if the haystack contains any of the strings in the needles.
func IsStringSliceContainsAny(needles []string, haystack []string) (inSlice bool) {
	return IsStringSliceContainsAnyF(needles, haystack, IsStringInSlice)
}

// IsStringSliceContainsAnyF checks if the haystack contains any of the strings in the needles using the isInSlice func.
func IsStringSliceContainsAnyF(needles []string, haystack []string, isInSlice func(needle string, haystack []string) bool) (inSlice bool) {
	for _, n := range needles {
		if isInSlice(n, haystack) {
			return true
		}
	}

	return false
}

// SliceString splits a string s into an array with each item being a max of int d
// d = denominator, n = numerator, q = quotient, r = remainder.
func SliceString(s string, d int) (array []string) {
	n := len(s)
	q := n / d
	r := n % d

	for i := 0; i < q; i++ {
		array = append(array, s[i*d:i*d+d])
		if i+1 == q && r != 0 {
			array = append(array, s[i*d+d:])
		}
	}

	return
}

func isStringSlicesDifferent(a, b []string, method func(s string, b []string) bool) (different bool) {
	if len(a) != len(b) {
		return true
	}

	for _, s := range a {
		if !method(s, b) {
			return true
		}
	}

	return false
}

// IsStringSlicesDifferent checks two slices of strings and on the first occurrence of a string item not existing in the
// other slice returns true, otherwise returns false.
func IsStringSlicesDifferent(a, b []string) (different bool) {
	return isStringSlicesDifferent(a, b, IsStringInSlice)
}

// IsStringSlicesDifferentFold checks two slices of strings and on the first occurrence of a string item not existing in
// the other slice (case insensitive) returns true, otherwise returns false.
func IsStringSlicesDifferentFold(a, b []string) (different bool) {
	return isStringSlicesDifferent(a, b, IsStringInSliceFold)
}

// IsURLInSlice returns true if the needle url.URL is in the []url.URL haystack.
func IsURLInSlice(needle *url.URL, haystack []*url.URL) (has bool) {
	for i := 0; i < len(haystack); i++ {
		if strings.EqualFold(needle.String(), haystack[i].String()) {
			return true
		}
	}

	return false
}

// StringSliceFromURLs returns a []string from a []url.URL.
func StringSliceFromURLs(urls []*url.URL) []string {
	result := make([]string, len(urls))

	for i := 0; i < len(urls); i++ {
		result[i] = urls[i].String()
	}

	return result
}

// URLsFromStringSlice returns a []url.URL from a []string.
func URLsFromStringSlice(urls []string) []*url.URL {
	var result []*url.URL

	for i := 0; i < len(urls); i++ {
		u, err := url.Parse(urls[i])
		if err != nil {
			continue
		}

		result = append(result, u)
	}

	return result
}

// OriginFromURL returns an origin url.URL given another url.URL.
func OriginFromURL(u *url.URL) (origin *url.URL) {
	return &url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
	}
}

// StringSlicesDelta takes a before and after []string and compares them returning a added and removed []string.
func StringSlicesDelta(before, after []string) (added, removed []string) {
	for _, s := range before {
		if !IsStringInSlice(s, after) {
			removed = append(removed, s)
		}
	}

	for _, s := range after {
		if !IsStringInSlice(s, before) {
			added = append(added, s)
		}
	}

	return added, removed
}

// StringHTMLEscape escapes chars for a HTML body.
func StringHTMLEscape(input string) (output string) {
	return htmlEscaper.Replace(input)
}

// StringJoinDelimitedEscaped joins a string with a specified rune delimiter after escaping any instance of that string
// in the string slice. Used with StringSplitDelimitedEscaped.
func StringJoinDelimitedEscaped(value []string, delimiter rune) string {
	escaped := make([]string, len(value))
	for k, v := range value {
		escaped[k] = strings.ReplaceAll(v, string(delimiter), "\\"+string(delimiter))
	}

	return strings.Join(escaped, string(delimiter))
}

// StringSplitDelimitedEscaped splits a string with a specified rune delimiter after unescaping any instance of that
// string in the string slice that has been escaped. Used with StringJoinDelimitedEscaped.
func StringSplitDelimitedEscaped(value string, delimiter rune) (out []string) {
	var escape bool

	split := strings.FieldsFunc(value, func(r rune) bool {
		if r == '\\' {
			escape = !escape
		} else if escape && r != delimiter {
			escape = false
		}

		return !escape && r == delimiter
	})

	for k, v := range split {
		split[k] = strings.ReplaceAll(v, "\\"+string(delimiter), string(delimiter))
	}

	return split
}

// JoinAndCanonicalizeHeaders join header strings by a given sep.
func JoinAndCanonicalizeHeaders(sep []byte, headers ...string) (joined []byte) {
	for i, header := range headers {
		if i != 0 {
			joined = append(joined, sep...)
		}

		joined = fasthttp.AppendNormalizedHeaderKey(joined, header)
	}

	return joined
}

// IsURLHostComponent returns true if the provided url.URL that was parsed from a string to a url.URL via url.Parse is
// just a hostname. This is needed because of the way this function parses such strings.
func IsURLHostComponent(u *url.URL) (isHostComponent bool) {
	return u != nil && u.Path != "" && u.Scheme == "" && u.Host == "" && u.RawPath == "" && u.Opaque == "" &&
		u.RawQuery == "" && u.Fragment == "" && u.RawFragment == ""
}

// IsURLHostComponentWithPort returns true if the provided url.URL that was parsed from a string to a url.URL via
// url.Parse is just a hostname with a port. This is needed because of the way this function parses such strings.
func IsURLHostComponentWithPort(u *url.URL) (isHostComponentWithPort bool) {
	if u == nil {
		return false
	}

	if u.Opaque != "" && u.Scheme != "" && u.Host == "" && u.Path == "" && u.RawPath == "" &&
		u.RawQuery == "" && u.Fragment == "" && u.RawFragment == "" {
		_, err := strconv.Atoi(u.Opaque)

		return err == nil
	}

	return false
}
