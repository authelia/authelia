// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCookies(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		skip     []string
		expected Cookies
		encoded  string
	}{
		{
			"ShouldHandleEmpty",
			"",
			nil,
			nil,
			"",
		},
		{
			"ShouldParseSingle",
			"authelia_session=abc",
			nil,
			Cookies{{Name: "authelia_session", Value: "abc"}},
			"authelia_session=abc",
		},
		{
			"ShouldSkipNamed",
			"a=1; authelia_session=abc; b=2",
			[]string{"authelia_session"},
			Cookies{{Name: "a", Value: "1"}, {Name: "authelia_session", Value: "abc"}, {Name: "b", Value: "2"}},
			"a=1; b=2",
		},
		{
			"ShouldPreserveOrder",
			"z=1; y=2; x=3; w=4; v=5; u=6",
			nil,
			Cookies{{Name: "z", Value: "1"}, {Name: "y", Value: "2"}, {Name: "x", Value: "3"}, {Name: "w", Value: "4"}, {Name: "v", Value: "5"}, {Name: "u", Value: "6"}},
			"z=1; y=2; x=3; w=4; v=5; u=6",
		},
		{
			"ShouldPreserveDuplicateNames",
			"x=1; x=2",
			nil,
			Cookies{{Name: "x", Value: "1"}, {Name: "x", Value: "2"}},
			"x=1; x=2",
		},
		{
			"ShouldSkipAllDuplicateNames",
			"x=1; authelia_session=a; x=2; authelia_session=b",
			[]string{"authelia_session"},
			Cookies{{Name: "x", Value: "1"}, {Name: "authelia_session", Value: "a"}, {Name: "x", Value: "2"}, {Name: "authelia_session", Value: "b"}},
			"x=1; x=2",
		},
		{
			"ShouldHandleValuesContainingEquals",
			"token=YWJjZA==; authelia_session=abc",
			[]string{"authelia_session"},
			Cookies{{Name: "token", Value: "YWJjZA=="}, {Name: "authelia_session", Value: "abc"}},
			"token=YWJjZA==",
		},
		{
			"ShouldHandleMissingValue",
			"flag; a=1",
			nil,
			Cookies{{Name: "flag", Value: ""}, {Name: "a", Value: "1"}},
			"flag=; a=1",
		},
		{
			"ShouldHandleMissingSpaceSeparator",
			"a=1;b=2",
			nil,
			Cookies{{Name: "a", Value: "1"}, {Name: "b", Value: "2"}},
			"a=1; b=2",
		},
		{
			"ShouldDiscardEmptyPairs",
			"a=1;; ;b=2;",
			nil,
			Cookies{{Name: "a", Value: "1"}, {Name: "b", Value: "2"}},
			"a=1; b=2",
		},
		{
			"ShouldEncodeEmptyWhenOnlyCookieIsSkipped",
			"authelia_session=abc",
			[]string{"authelia_session"},
			Cookies{{Name: "authelia_session", Value: "abc"}},
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := NewCookies([]byte(tc.have))

			assert.Equal(t, tc.expected, actual)
			assert.Equal(t, tc.encoded, actual.Encode(tc.skip...))
		})
	}
}
