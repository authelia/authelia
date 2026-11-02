// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package expression_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/expression"
)

func TestNewClaimsMappingShouldRejectInvalid(t *testing.T) {
	testCases := []struct {
		name     string
		mapping  map[string]string
		expected string
	}{
		{
			"ShouldRejectSubject",
			map[string]string{"sub": "claims.id"},
			"the claim 'sub' can't be mapped: it must be one of 'amr', 'disabled', 'email', 'name', or 'preferred_username'",
		},
		{
			"ShouldRejectIssuer",
			map[string]string{"iss": "claims.id"},
			"the claim 'iss' can't be mapped: it must be one of 'amr', 'disabled', 'email', 'name', or 'preferred_username'",
		},
		{
			"ShouldRejectSyntax",
			map[string]string{"email": "claims."},
			"the claim 'email' has an invalid expression: ",
		},
		{
			"ShouldRejectUnknownVariable",
			map[string]string{"email": "other.email"},
			"the claim 'email' has an invalid expression: ",
		},
		{
			"ShouldRejectStringClaimWithListResult",
			map[string]string{"email": `["a"]`},
			"the claim 'email' has an invalid expression: the result type must be 'string' but it is 'list(string)'",
		},
		{
			"ShouldRejectDisabledWithStringResult",
			map[string]string{"disabled": `"true"`},
			"the claim 'disabled' has an invalid expression: the result type must be 'bool' but it is 'string'",
		},
		{
			"ShouldRejectAMRWithStringResult",
			map[string]string{"amr": `"mfa"`},
			"the claim 'amr' has an invalid expression: the result type must be 'list(string)' but it is 'string'",
		},
		{
			"ShouldRejectAMRWithIntegerListResult",
			map[string]string{"amr": `[1]`},
			"the claim 'amr' has an invalid expression: the result type must be 'list(string)' but it is 'list(int)'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := expression.NewClaimsMapping(tc.mapping)

			assert.ErrorContains(t, err, tc.expected)
		})
	}
}

func TestClaimsMappingEvaluate(t *testing.T) {
	testCases := []struct {
		name        string
		mapping     map[string]string
		claims      map[string]any
		amr         []string
		strings     map[string]string
		expectedAMR []string
	}{
		{
			"ShouldMapOnlyConfiguredStringClaims",
			map[string]string{"email": "claims.contact_email"},
			map[string]any{"contact_email": "john@example.com", "name": "ignored"},
			[]string{"pwd"},
			map[string]string{"email": "john@example.com"},
			nil,
		},
		{
			"ShouldAddToAMR",
			map[string]string{"amr": `amr + ["mfa"]`},
			nil,
			[]string{"pwd"},
			map[string]string{},
			[]string{"pwd", "mfa"},
		},
		{
			"ShouldAddToEmptyAMR",
			map[string]string{"amr": `amr + ["mfa"]`},
			nil,
			nil,
			map[string]string{},
			[]string{"mfa"},
		},
		{
			"ShouldReplaceAMR",
			map[string]string{"amr": `["hwk", "mfa"]`},
			nil,
			[]string{"pwd"},
			map[string]string{},
			[]string{"hwk", "mfa"},
		},
		{
			"ShouldAddToAMRWhenClaimIsTrue",
			map[string]string{"amr": `claims.mfa_enabled == true ? amr + ["mfa"] : amr`},
			map[string]any{"mfa_enabled": true},
			[]string{"pwd"},
			map[string]string{},
			[]string{"pwd", "mfa"},
		},
		{
			"ShouldNotAddToAMRWhenClaimIsFalse",
			map[string]string{"amr": `claims.mfa_enabled == true ? amr + ["mfa"] : amr`},
			map[string]any{"mfa_enabled": false},
			[]string{"pwd"},
			map[string]string{},
			[]string{"pwd"},
		},
		{
			"ShouldHandleAbsentClaimWithHas",
			map[string]string{"email": `has(claims.alt) ? claims.alt : "fallback@example.com"`},
			map[string]any{},
			nil,
			map[string]string{"email": "fallback@example.com"},
			nil,
		},
		{
			"ShouldAcceptDynamicListFromClaim",
			map[string]string{"amr": `claims.methods`},
			map[string]any{"methods": []any{"otp", "pwd"}},
			nil,
			map[string]string{},
			[]string{"otp", "pwd"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mapping, err := expression.NewClaimsMapping(tc.mapping)
			require.NoError(t, err)

			values, amr, err := mapping.Evaluate(tc.claims, tc.amr)
			require.NoError(t, err)

			assert.Equal(t, tc.strings, values)
			assert.Equal(t, tc.expectedAMR, amr)
		})
	}
}

func TestClaimsMappingEvaluateShouldFailClosed(t *testing.T) {
	testCases := []struct {
		name     string
		mapping  map[string]string
		claims   map[string]any
		expected string
	}{
		{
			"ShouldFailOnAbsentClaim",
			map[string]string{"email": "claims.absent"},
			map[string]any{},
			"error evaluating the expression for the claim 'email': ",
		},
		{
			"ShouldFailOnDynamicStringClaimOfWrongType",
			map[string]string{"email": "claims.count"},
			map[string]any{"count": float64(1)},
			"error evaluating the expression for the claim 'email': the result is not a string",
		},
		{
			"ShouldFailOnDynamicAMRClaimOfWrongType",
			map[string]string{"amr": "claims.count"},
			map[string]any{"count": float64(1)},
			"error evaluating the expression for the claim 'amr': the result is not a list of strings",
		},
		{
			"ShouldFailOnDynamicAMRClaimWithNonStringValue",
			map[string]string{"amr": "claims.methods"},
			map[string]any{"methods": []any{"otp", float64(1)}},
			"error evaluating the expression for the claim 'amr': the result is not a list of strings",
		},
		{
			"ShouldFailOnEmptyAMRValue",
			map[string]string{"amr": `[""]`},
			nil,
			"error evaluating the expression for the claim 'amr': the result contains an empty value",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mapping, err := expression.NewClaimsMapping(tc.mapping)
			require.NoError(t, err)

			_, _, err = mapping.Evaluate(tc.claims, nil)

			assert.ErrorContains(t, err, tc.expected)
		})
	}
}

func TestClaimsMappingMapsAMR(t *testing.T) {
	mapping, err := expression.NewClaimsMapping(map[string]string{"email": "claims.alt"})
	require.NoError(t, err)
	assert.False(t, mapping.MapsAMR())

	mapping, err = expression.NewClaimsMapping(map[string]string{"amr": "amr"})
	require.NoError(t, err)
	assert.True(t, mapping.MapsAMR())
}

func TestClaimsMappingDisabled(t *testing.T) {
	testCases := []struct {
		name     string
		mapping  map[string]string
		claims   map[string]any
		expected bool
		err      string
	}{
		{"ShouldNotDisableWhenNotConfigured", map[string]string{"email": `"a"`}, nil, false, ""},
		{"ShouldDisableWhenTrue", map[string]string{"disabled": "claims.locked"}, map[string]any{"locked": true}, true, ""},
		{"ShouldNotDisableWhenFalse", map[string]string{"disabled": "claims.locked"}, map[string]any{"locked": false}, false, ""},
		{"ShouldUseAMR", map[string]string{"disabled": `!("otp" in amr)`}, nil, true, ""},
		{"ShouldFailOnAbsentClaim", map[string]string{"disabled": "claims.locked"}, nil, false, "error evaluating the expression for the claim 'disabled': "},
		{"ShouldFailOnWrongDynamicType", map[string]string{"disabled": "claims.locked"}, map[string]any{"locked": "true"}, false, "error evaluating the expression for the claim 'disabled': the result is not a boolean"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mapping, err := expression.NewClaimsMapping(tc.mapping)
			require.NoError(t, err)

			disabled, err := mapping.Disabled(tc.claims, nil)

			if tc.err != "" {
				assert.True(t, disabled)
				assert.ErrorContains(t, err, tc.err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expected, disabled)
		})
	}
}
