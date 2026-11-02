// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

type claimsMappingTestProvider struct {
	Provider

	claims *IdentityClaims
	err    error
	policy func(asserted []string) []string
}

func (p *claimsMappingTestProvider) Complete(_ context.Context, _ CompletionRequest) (*IdentityClaims, error) {
	return p.claims, p.err
}

func (p *claimsMappingTestProvider) AuthenticationMethodsReference(asserted []string) []string {
	if p.policy == nil {
		return asserted
	}

	return p.policy(asserted)
}

func TestClaimsMappedProviderShouldOnlyWrapWhenConfigured(t *testing.T) {
	inner := &claimsMappingTestProvider{}

	assert.Same(t, inner, newClaimsMappedProvider(inner, nil))
	assert.Same(t, inner, newClaimsMappedProvider(inner, map[string]string{}))
	assert.NotSame(t, inner, newClaimsMappedProvider(inner, map[string]string{"email": "claims.alt"}))
}

func TestClaimsMappedProviderComplete(t *testing.T) {
	newInner := func() *claimsMappingTestProvider {
		return &claimsMappingTestProvider{
			claims: &IdentityClaims{
				Issuer:                         "https://idp.example.com",
				Subject:                        "abc123",
				PreferredUsername:              "john",
				Name:                           "John",
				Email:                          "john@example.com",
				AuthenticationMethodsReference: []string{"otp"},
				RawUserInfo:                    map[string]any{"alt": "alt@example.com", "mfa_enabled": true},
			},
			policy: func(asserted []string) []string {
				return append([]string{"policy"}, asserted...)
			},
		}
	}

	testCases := []struct {
		name     string
		mapping  map[string]string
		expected func(claims *IdentityClaims)
		amr      []string
	}{
		{
			"ShouldOverrideOnlyConfiguredClaim",
			map[string]string{"email": "claims.userinfo.alt"},
			func(claims *IdentityClaims) { claims.Email = "alt@example.com" },
			[]string{"policy", "otp"},
		},
		{
			"ShouldAddToTheResolvedPolicyValues",
			map[string]string{"amr": `claims.userinfo.mfa_enabled == true ? amr + ["mfa"] : amr`},
			func(claims *IdentityClaims) { claims.AuthenticationMethodsReference = []string{"policy", "otp", "mfa"} },
			[]string{"policy", "otp", "mfa"},
		},
		{
			"ShouldReplaceTheResolvedPolicyValues",
			map[string]string{"amr": `["hwk"]`},
			func(claims *IdentityClaims) { claims.AuthenticationMethodsReference = []string{"hwk"} },
			[]string{"hwk"},
		},
		{
			"ShouldClearTheResolvedPolicyValues",
			map[string]string{"amr": `amr.filter(x, x == "none")`},
			func(claims *IdentityClaims) { claims.AuthenticationMethodsReference = []string{} },
			[]string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inner := newInner()
			provider := newClaimsMappedProvider(inner, tc.mapping)

			claims, err := provider.Complete(context.Background(), CompletionRequest{})
			require.NoError(t, err)

			expected := newInner().claims
			tc.expected(expected)

			assert.Equal(t, expected, claims)
			assert.Equal(t, tc.amr, provider.AuthenticationMethodsReference(claims.AuthenticationMethodsReference))
		})
	}
}

func TestClaimsMappedProviderShouldNotChangeIssuerOrSubject(t *testing.T) {
	inner := &claimsMappingTestProvider{claims: &IdentityClaims{Issuer: "i", Subject: "s", RawUserInfo: map[string]any{}}}

	claims, err := newClaimsMappedProvider(inner, map[string]string{"email": `"a@example.com"`}).Complete(context.Background(), CompletionRequest{})
	require.NoError(t, err)

	assert.Equal(t, "i", claims.Issuer)
	assert.Equal(t, "s", claims.Subject)
}

func TestClaimsMappedProviderShouldFailClosed(t *testing.T) {
	sentinel := errors.New("inner failure")

	testCases := []struct {
		name     string
		inner    *claimsMappingTestProvider
		mapping  map[string]string
		expected string
	}{
		{
			"ShouldReturnProviderError",
			&claimsMappingTestProvider{err: sentinel},
			map[string]string{"email": "claims.alt"},
			"inner failure",
		},
		{
			"ShouldFailOnEvaluationError",
			&claimsMappingTestProvider{claims: &IdentityClaims{Email: "unmapped@example.com", RawUserInfo: map[string]any{}}},
			map[string]string{"email": "claims.userinfo.absent"},
			"error evaluating the expression for the claim 'email': ",
		},
		{
			"ShouldFailOnInvalidMapping",
			&claimsMappingTestProvider{claims: &IdentityClaims{RawUserInfo: map[string]any{}}},
			map[string]string{"sub": "claims.id"},
			"the claim 'sub' can't be mapped",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			claims, err := newClaimsMappedProvider(tc.inner, tc.mapping).Complete(context.Background(), CompletionRequest{})

			assert.Nil(t, claims)
			assert.ErrorContains(t, err, tc.expected)
		})
	}
}

func TestNewProvidersShouldApplyClaimsMapping(t *testing.T) {
	providers := NewProviders(&schema.AuthenticationBackendExternalIdentity{
		Providers: []schema.AuthenticationBackendExternalIdentityProvider{
			{ID: "mapped", ClaimsMapping: map[string]string{"email": "claims.alt"}},
			{ID: "unmapped"},
		},
	}, nil)

	mapped, ok := providers.Get("mapped")
	require.True(t, ok)
	assert.IsType(t, &claimsMappedProvider{}, mapped)

	unmapped, ok := providers.Get("unmapped")
	require.True(t, ok)
	assert.IsType(t, &OpenIDConnectProvider{}, unmapped)
}

func TestClaimsMappedProviderShouldSeparateTheClaimSources(t *testing.T) {
	testCases := []struct {
		name     string
		claims   *IdentityClaims
		mapping  string
		expected string
	}{
		{
			"ShouldReadTheIDToken",
			&IdentityClaims{RawIDToken: map[string]any{"email": "id@example.com"}, RawUserInfo: map[string]any{"email": "userinfo@example.com"}},
			"claims.id_token.email",
			"id@example.com",
		},
		{
			"ShouldReadTheUserInfo",
			&IdentityClaims{RawIDToken: map[string]any{"email": "id@example.com"}, RawUserInfo: map[string]any{"email": "userinfo@example.com"}},
			"claims.userinfo.email",
			"userinfo@example.com",
		},
		{
			"ShouldAllowTestingForAClaimWithoutAnIDToken",
			&IdentityClaims{RawUserInfo: map[string]any{"email": "userinfo@example.com"}},
			`has(claims.id_token.email) ? claims.id_token.email : claims.userinfo.email`,
			"userinfo@example.com",
		},
		{
			"ShouldAllowTestingForAClaimWithoutAUserInfoResponse",
			&IdentityClaims{RawIDToken: map[string]any{"email": "id@example.com"}},
			`has(claims.userinfo.email) ? claims.userinfo.email : claims.id_token.email`,
			"id@example.com",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			claims, err := newClaimsMappedProvider(&claimsMappingTestProvider{claims: tc.claims}, map[string]string{"email": tc.mapping}).Complete(context.Background(), CompletionRequest{})
			require.NoError(t, err)

			assert.Equal(t, tc.expected, claims.Email)
		})
	}
}

func TestClaimsMappedProviderDisabled(t *testing.T) {
	newInner := func() *claimsMappingTestProvider {
		return &claimsMappingTestProvider{
			claims: &IdentityClaims{Issuer: "https://idp.example.com", Subject: "abc123", RawUserInfo: map[string]any{"locked": true, "open": false}},
			policy: func(_ []string) []string { return []string{"pwd"} },
		}
	}

	testCases := []struct {
		name     string
		mapping  map[string]string
		disabled bool
		err      string
	}{
		{"ShouldRejectWhenTrue", map[string]string{"disabled": "claims.userinfo.locked"}, true, ""},
		{"ShouldAllowWhenFalse", map[string]string{"disabled": "claims.userinfo.open"}, false, ""},
		{"ShouldUseTheResolvedPolicyValues", map[string]string{"disabled": `"pwd" in amr`}, true, ""},
		{"ShouldUseTheResolvedPolicyValuesNotTheMappedValues", map[string]string{"disabled": `"otp" in amr`, "amr": `["otp"]`}, false, ""},
		{"ShouldFailClosedOnEvaluationError", map[string]string{"disabled": "claims.userinfo.absent"}, false, "error evaluating the expression for the claim 'disabled': "},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			claims, err := newClaimsMappedProvider(newInner(), tc.mapping).Complete(context.Background(), CompletionRequest{})

			switch {
			case tc.disabled:
				assert.Nil(t, claims)
				require.ErrorIs(t, err, ErrIdentityDisabled)
				assert.EqualError(t, err, "the claims mapping disabled the sign in: issuer 'https://idp.example.com' subject 'abc123'")
			case tc.err != "":
				assert.Nil(t, claims)
				assert.ErrorContains(t, err, tc.err)
			default:
				require.NoError(t, err)
				assert.NotNil(t, claims)
			}
		})
	}
}

func TestClaimsMappedProviderShouldExposeProviderSpecificClaims(t *testing.T) {
	inner := &claimsMappingTestProvider{claims: &IdentityClaims{RawProviders: map[string]map[string]any{"example": {"value": "provider@example.com"}}}}

	claims, err := newClaimsMappedProvider(inner, map[string]string{"email": "claims.example.value"}).Complete(context.Background(), CompletionRequest{})
	require.NoError(t, err)

	assert.Equal(t, "provider@example.com", claims.Email)
}

func TestValidateClaimsShouldRetainRawClaims(t *testing.T) {
	now := time.Unix(1700000000, 0)

	claims, err := validateClaims(map[string]any{
		"iss": "https://idp.example.com", "sub": "abc123", "aud": "client", "nonce": "n", "custom": "value",
		"exp": float64(now.Add(time.Hour).Unix()), "iat": float64(now.Unix()),
	}, ValidateOptions{Issuer: "https://idp.example.com", ClientID: "client", Nonce: "n", Now: now})
	require.NoError(t, err)

	assert.Equal(t, "value", claims.RawIDToken["custom"])
}
