package templates

import (
	"fmt"
	"strings"
)

// StringMapLookupDefaultEmptyFunc is function which takes a map[string]string and returns a template function which
// takes a string which is used as a key lookup for the map[string]string. If the value isn't found it returns an empty
// string.
func StringMapLookupDefaultEmptyFunc(m map[string]string) func(key string) (value string) {
	return func(key string) (value string) {
		var ok bool

		if value, ok = m[key]; !ok {
			return ""
		}

		return value
	}
}

// StringMapLookupFunc is function which takes a map[string]string and returns a template function which
// takes a string which is used as a key lookup for the map[string]string. If the value isn't found it returns an error.
func StringMapLookupFunc(m map[string]string) func(key string) (value string, err error) {
	return func(key string) (value string, err error) {
		var ok bool

		if value, ok = m[key]; !ok {
			return value, fmt.Errorf("failed to lookup key '%s' from map", key)
		}

		return value, nil
	}
}

// IterateFunc is a template function which takes a single uint returning a slice of units from 0 up to that number.
func IterateFunc(count *uint) (out []uint) {
	var i uint

	for i = 0; i < (*count); i++ {
		out = append(out, i)
	}

	return
}

// StringsSplitFunc is a template function which takes sep and value, splitting the value by the sep into a slice.
func StringsSplitFunc(sep, value string) []string {
	return strings.Split(value, sep)
}

// StringJoinXFunc takes a list of string elements, joins them by the sep string, before every int n characters are
// written it writes string p. This is useful for line breaks mostly.
func StringJoinXFunc(elems []string, sep string, n int, p string) string {
	buf := strings.Builder{}

	c := 0
	e := len(elems) - 1

	for i := 0; i <= e; i++ {
		if c+len(elems[i])+1 > n {
			c = 0

			buf.WriteString(p)
		}

		c += len(elems[i]) + 1

		buf.WriteString(elems[i])

		if i < e {
			buf.WriteString(sep)
		}
	}

	return buf.String()
}
