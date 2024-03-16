package validator

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ValidateKeys determines if all provided keys are valid.
func ValidateKeys(keys, remapped []string, prefix string, validator *schema.StructValidator) {
	var errStrings []string

	var patterns []*regexp.Regexp

	for _, key := range schema.Keys {
		pattern, _ := NewKeyPattern(key)

		switch {
		case pattern == nil:
			continue
		default:
			patterns = append(patterns, pattern)
		}
	}

KEYS:
	for _, key := range keys {
		expectedKey := reKeyReplacer.ReplaceAllString(key, "[]")

		if utils.IsStringInSlice(expectedKey, schema.Keys) {
			continue
		}

		if newKey, ok := replacedKeys[expectedKey]; ok {
			validator.Push(fmt.Errorf(errFmtReplacedConfigurationKey, key, newKey))
			continue
		}

		for _, p := range patterns {
			if p.MatchString(expectedKey) {
				continue KEYS
			}
		}

		if utils.IsStringInSlice(expectedKey, remapped) {
			continue
		}

		if err, ok := specificErrorKeys[expectedKey]; ok {
			if !utils.IsStringInSlice(err, errStrings) {
				errStrings = append(errStrings, err)
			}
		} else {
			if strings.HasPrefix(key, prefix) {
				validator.PushWarning(fmt.Errorf("configuration environment variable not expected: %s", key))
			} else {
				validator.Push(fmt.Errorf("configuration key not expected: %s", key))
			}
		}
	}

	for _, err := range errStrings {
		validator.Push(errors.New(err))
	}
}

// NewKeyPattern returns patterns which are required to match key patterns.
func NewKeyPattern(key string) (pattern *regexp.Regexp, err error) {
	switch {
	case reIsMapKey.MatchString(key):
		return NewKeyMapPattern(key)
	default:
		return nil, nil
	}
}

var reIsMapKey = regexp.MustCompile(`\.\*(\[]|\.)`)

// NewKeyMapPattern returns a pattern required to match map keys.
func NewKeyMapPattern(key string) (pattern *regexp.Regexp, err error) {
	parts := strings.Split(key, ".*")

	buf := &strings.Builder{}

	buf.WriteString("^")

	n := len(parts) - 1

	for i, part := range parts {
		if i != 0 && !strings.HasPrefix(part, "[]") {
			buf.WriteString("\\.")
		}

		for j, r := range part {
			// Skip prefixed period.
			if j == 0 && r == '.' {
				continue
			}

			switch r {
			case '[', ']', '.', '{', '}':
				buf.WriteRune('\\')
				fallthrough
			default:
				buf.WriteRune(r)
			}
		}

		if i < n {
			buf.WriteString("\\.[a-z0-9](([a-z0-9-_]+)?[a-z0-9])?")
		}
	}

	buf.WriteString("$")

	return regexp.Compile(buf.String())
}
