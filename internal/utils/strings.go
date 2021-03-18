package utils

import (
	"math/rand"
	"strings"
	"time"
	"unicode"
)

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
func IsStringInSlice(a string, slice []string) (inSlice bool) {
	for _, b := range slice {
		if b == a {
			return true
		}
	}

	return false
}

// IsStringInSliceFold checks if a single string is in a slice of strings but uses strings.EqualFold to compare them.
func IsStringInSliceFold(a string, slice []string) (inSlice bool) {
	for _, b := range slice {
		if strings.EqualFold(b, a) {
			return true
		}
	}

	return false
}

// IsStringInSliceContains checks if a single string is in an array of strings.
func IsStringInSliceContains(a string, list []string) (inSlice bool) {
	for _, b := range list {
		if strings.Contains(a, b) {
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
