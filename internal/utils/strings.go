package utils

import (
	"math/rand"
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

// IsStringInSlice checks if a single string is in an array of strings.
func IsStringInSlice(a string, list []string) (inSlice bool) {
	for _, b := range list {
		if b == a {
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

// IsStringSlicesDifferent checks two slices of strings and on the first occurrence of a string item not existing in the
// other slice returns true, otherwise returns false.
func IsStringSlicesDifferent(a, b []string) (different bool) {
	for _, s := range a {
		if !IsStringInSlice(s, b) {
			return true
		}
	}

	for _, s := range b {
		if !IsStringInSlice(s, a) {
			return true
		}
	}

	return false
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
