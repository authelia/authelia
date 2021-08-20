package utils

import (
	"fmt"
	"math/rand"
	"net/url"
	"reflect"
	"strings"
	"time"
	"unicode"
	"unsafe"
)

// IsStringAbsURL checks a string can be parsed as a URL and that is IsAbs and if it can't it returns an error
// describing why.
func IsStringAbsURL(input string) (err error) {
	parsedURL, err := url.Parse(input)
	if err != nil {
		return fmt.Errorf("could not parse '%s' as a URL", input)
	}

	if !parsedURL.IsAbs() {
		return fmt.Errorf("the url '%s' is not absolute because it doesn't start with a scheme like 'http://' or 'https://'", input)
	}

	return nil
}

// IsStringAlphaNumeric returns false if any rune in the string is not alpha-numeric.
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

// IsStringInSliceSuffix checks if the needle string has one of the suffixes in the haystack.
func IsStringInSliceSuffix(needle string, haystack []string) (hasSuffix bool) {
	for _, straw := range haystack {
		if strings.HasSuffix(needle, straw) {
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

// RandomString generate a random string of n characters.
func RandomString(n int, characters []rune) (randomString string) {
	rand.Seed(time.Now().UnixNano())

	b := make([]rune, n)
	for i := range b {
		b[i] = characters[rand.Intn(len(characters))] //nolint:gosec // Likely isn't necessary to use the more expensive crypto/rand for this utility func.
	}

	return string(b)
}

// StringHTMLEscape escapes chars for a HTML body.
func StringHTMLEscape(input string) (output string) {
	return htmlEscaper.Replace(input)
}

// StringToByteSlice converts a string into a slice of bytes without allocation of memory.
// LICENSE: Apache-2.0 https://github.com/savsgio/gotils/blob/master/LICENSE
// SOURCE: https://github.com/savsgio/gotils/blob/97865ed5a873d8bd3299e740aded793b0cfefc3a/strconv/strconv.go#L14-L26
// REFERENCE: https://groups.google.com/g/Golang-Nuts/c/ENgbUzYvCuU/m/90yGx7GUAgAJ
// Changes to the original code prior to the first commit were only aesthetic. All other changes are logged via git SCM.
func StringToByteSlice(input string) (output []byte) {
	strHeader := (*reflect.StringHeader)(unsafe.Pointer(&input))
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&output))
	sliceHeader.Data = strHeader.Data
	sliceHeader.Cap = strHeader.Len
	sliceHeader.Len = strHeader.Len

	return output
}
