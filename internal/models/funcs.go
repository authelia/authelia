package models

import (
	"strings"
)

func valueStringSlice(delimiter rune, value []string) string {
	escaped := make([]string, len(value))
	for k, v := range value {
		escaped[k] = strings.ReplaceAll(v, string(delimiter), "\\"+string(delimiter))
	}

	return strings.Join(escaped, string(delimiter))
}

func scanStringSlice(delimiter rune, value string) (out []string) {
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
