// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package authorization

import (
	"net/url"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestNewAccessControlQuery(t *testing.T) {
	testCases := []struct {
		name     string
		have     [][]schema.AccessControlRuleQuery
		expected []AccessControlQuery
		matches  [][]Object
	}{
		{
			"ShouldSkipInvalidTypeEqual",
			[][]schema.AccessControlRuleQuery{
				{
					{Operator: operatorEqual, Key: "example", Value: 1},
				},
			},
			[]AccessControlQuery{{Rules: []ObjectMatcher(nil)}},
			[][]Object{{{}}},
		},
		{
			"ShouldSkipInvalidTypePattern",
			[][]schema.AccessControlRuleQuery{
				{
					{Operator: operatorPattern, Key: "example", Value: 1},
				},
			},
			[]AccessControlQuery{{Rules: []ObjectMatcher(nil)}},
			[][]Object{{{}}},
		},
		{
			"ShouldSkipInvalidOperator",
			[][]schema.AccessControlRuleQuery{
				{
					{Operator: "nop", Key: "example", Value: 1},
				},
			},
			[]AccessControlQuery{{Rules: []ObjectMatcher(nil)}},
			[][]Object{{{}}},
		},
		{
			"ShouldNormalizeOperatorCase",
			[][]schema.AccessControlRuleQuery{
				{
					{Operator: "Present", Key: "Present"},
					{Operator: "ABSENT", Key: "Absent"},
					{Operator: "Equal", Key: "Equal", Value: "Value"},
					{Operator: "Not Equal", Key: "NotEqual", Value: "Value"},
					{Operator: "Pattern", Key: "Pattern", Value: regexp.MustCompile(`^Value$`)},
					{Operator: "NOT PATTERN", Key: "NotPattern", Value: regexp.MustCompile(`^Value$`)},
				},
			},
			[]AccessControlQuery{
				{
					Rules: []ObjectMatcher{
						&AccessControlQueryMatcherPresent{key: "Present", present: true},
						&AccessControlQueryMatcherPresent{key: "Absent", present: false},
						&AccessControlQueryMatcherEqual{key: "Equal", value: "Value", equal: true},
						&AccessControlQueryMatcherEqual{key: "NotEqual", value: "Value", equal: false},
						&AccessControlQueryMatcherPattern{key: "Pattern", pattern: regexp.MustCompile(`^Value$`), match: true},
						&AccessControlQueryMatcherPattern{key: "NotPattern", pattern: regexp.MustCompile(`^Value$`), match: false},
					},
				},
			},
			[][]Object{{{URL: &url.URL{RawQuery: "Present=&Equal=Value&NotEqual=value&Pattern=Value&NotPattern=value"}}}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := NewAccessControlQuery(tc.have)
			assert.Equal(t, tc.expected, actual)

			for i, rule := range actual {
				for _, object := range tc.matches[i] {
					assert.True(t, rule.IsMatch(object))
				}
			}
		})
	}
}

func TestNewObjectQueryInterpretations(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		expected []string
	}{
		{"ShouldReturnNilForNoQuery", "https://app.example.com/", nil},
		{"ShouldReturnNilForUnambiguousQuery", "https://app.example.com/?a=1&b=2", nil},
		{"ShouldReturnNilForEscapedQuery", "https://app.example.com/?a=%31&b=x+y", nil},
		{"ShouldReturnFirstAndLastForRepeatedKey", "https://app.example.com/?a=1&a=2", []string{"a=1", "a=2"}},
		{"ShouldReturnSingleForRepeatedKeySameValue", "https://app.example.com/?a=1&a=1", []string{"a=1"}},
		{"ShouldReturnEachParseForSemicolon", "https://app.example.com/?x=1;admin=1", []string{"", "x=1%3Badmin%3D1", "admin=1&x=1"}},
		{"ShouldReturnEachParseForInvalidEscape", "https://app.example.com/?a=%zz&b=1", []string{"b=1", "a=%25zz&b=1"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u, err := url.Parse(tc.have)
			require.NoError(t, err)

			object := NewObject(u, "GET")

			actual := NewObjectQueryInterpretations(object)

			if tc.expected == nil {
				assert.Nil(t, actual)

				return
			}

			queries := make([]string, len(actual))

			for i, o := range actual {
				queries[i] = o.URL.RawQuery

				assert.Equal(t, object.Domain, o.Domain)
				assert.Equal(t, object.Path, o.Path)
				assert.Equal(t, object.Method, o.Method)
			}

			assert.Equal(t, tc.expected, queries)
			assert.Equal(t, u.RawQuery, object.URL.RawQuery)
		})
	}

	t.Run("ShouldReturnNilForNilURL", func(t *testing.T) {
		assert.Nil(t, NewObjectQueryInterpretations(Object{}))
	})
}
