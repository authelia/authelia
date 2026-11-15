// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedactEmail(t *testing.T) {
	testCases := []struct {
		testName string
		input    string
		expected string
	}{
		{"ShouldRedactEmail", "james.dean@authelia.com", "j********n@authelia.com"},
		{"ShouldRedactShortEmail", "me@authelia.com", "**@authelia.com"},
		{"ShouldRedactInvalidEmail", "invalidEmail.com", ""},
		{"ShouldRedactUnicode", "søren@example.com", "s***n@example.com"},
		{"ShouldReturnUnicodeInRedactedEmail", "øpenme@example.com", "ø****e@example.com"},
	}
	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			require.Equal(t, tc.expected, redactEmail(tc.input))
		})
	}
}

func TestValidatePasswordLength(t *testing.T) {
	testCases := []struct {
		name     string
		have     []string
		expected string
	}{
		{"ShouldPassNone", nil, ""},
		{"ShouldPassEmpty", []string{""}, ""},
		{"ShouldPassMaximum", []string{strings.Repeat("a", passwordMaxLength)}, ""},
		{"ShouldPassMultipleMaximum", []string{"abc", strings.Repeat("a", passwordMaxLength)}, ""},
		{"ShouldFailExceedsMaximum", []string{strings.Repeat("a", passwordMaxLength+1)}, "password exceeds the maximum length of 256 bytes"},
		{"ShouldFailMultipleExceedsMaximum", []string{"abc", strings.Repeat("a", passwordMaxLength+1)}, "password exceeds the maximum length of 256 bytes"},
		{"ShouldFailExceedsMaximumMultiByte", []string{strings.Repeat("é", passwordMaxLength/2+1)}, "password exceeds the maximum length of 256 bytes"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validatePasswordLength(tc.have...)

			if tc.expected == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expected)
			}
		})
	}
}

func TestPasswordResetFailureReason(t *testing.T) {
	testCases := []struct {
		name     string
		have     error
		expected string
	}{
		{
			"ShouldReturnNothingForNilError",
			nil,
			"",
		},
		{
			"ShouldReturnNothingForUnknownError",
			errors.New("failed to update"),
			"",
		},
		{
			"ShouldNotDiscloseBackendConnectionErrors",
			errors.New("LDAP Result Code 200 \"Network Error\": dial tcp 10.0.0.1:389: connect: connection refused"),
			"",
		},
		{
			"ShouldReturnReuseForUnchangedPassword",
			errors.New("LDAP Result Code 19 \"Constraint Violation\": Password is not being changed from existing value"),
			eventEmailReasonPasswordReuse,
		},
		{
			"ShouldReturnReuseForPasswordHistory",
			errors.New("LDAP Result Code 19 \"Constraint Violation\": Password is in history of old passwords"),
			eventEmailReasonPasswordReuse,
		},
		{
			"ShouldReturnReuseForFreeIPAPasswordHistory",
			errors.New("unable to update password. Cause: LDAP Result Code 19 \"Constraint Violation\": password in history"),
			eventEmailReasonPasswordReuse,
		},
		{
			"ShouldReturnTooYoungForOpenLDAPMinimumAge",
			errors.New("LDAP Result Code 19 \"Constraint Violation\": Password is too young to change"),
			eventEmailReasonPasswordTooYoung,
		},
		{
			"ShouldReturnTooYoungForFreeIPAMinimumAge",
			errors.New("LDAP Result Code 19 \"Constraint Violation\": within password minimum age"),
			eventEmailReasonPasswordTooYoung,
		},
		{
			"ShouldReturnBackendForQualityPolicy",
			errors.New("LDAP Result Code 19 \"Constraint Violation\": Password fails quality checking policy"),
			eventEmailReasonPasswordBackend,
		},
		{
			"ShouldReturnBackendForActiveDirectoryCode",
			errors.New("LDAP Result Code 53 \"Unwilling To Perform\": 0000052D: SvcErr: DSID-031A120C"),
			eventEmailReasonPasswordBackend,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, passwordResetFailureReason(tc.have))
		})
	}
}
