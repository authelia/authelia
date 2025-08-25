package authorization

import (
	"net"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestLevelToString(t *testing.T) {
	testCases := []struct {
		have     Level
		expected string
	}{
		{Bypass, "bypass"},
		{OneFactor, "one_factor"},
		{TwoFactor, "two_factor"},
		{Denied, "deny"},
		{99, "deny"},
	}

	for _, tc := range testCases {
		t.Run("Expected_"+tc.expected, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.have.String())
		})
	}
}

func TestShouldNotParseInvalidSubjects(t *testing.T) {
	subjectsSchema := [][]string{{"groups:z"}, {"group:z", "users:b"}}
	subjectsACL := schemaSubjectsToACL(subjectsSchema)

	require.Len(t, subjectsACL, 1)

	require.Len(t, subjectsACL[0].Subjects, 1)

	assert.True(t, subjectsACL[0].IsMatch(Subject{Username: "a", Groups: []string{"z"}}))

	assert.Equal(t, subjectsACL, NewSubjects(subjectsSchema))
}

func TestShouldSplitDomainCorrectly(t *testing.T) {
	prefix, suffix := domainToPrefixSuffix("apple.example.com")

	assert.Equal(t, "apple", prefix)
	assert.Equal(t, "example.com", suffix)

	prefix, suffix = domainToPrefixSuffix("example")

	assert.Equal(t, "", prefix)
	assert.Equal(t, "example", suffix)

	prefix, suffix = domainToPrefixSuffix("example.com")

	assert.Equal(t, "example", prefix)
	assert.Equal(t, "com", suffix)
}

func TestIsAuthLevelSufficient(t *testing.T) {
	assert.False(t, IsAuthLevelSufficient(authentication.NotAuthenticated, Denied))
	assert.False(t, IsAuthLevelSufficient(authentication.OneFactor, Denied))
	assert.False(t, IsAuthLevelSufficient(authentication.TwoFactor, Denied))
	assert.True(t, IsAuthLevelSufficient(authentication.NotAuthenticated, Bypass))
	assert.True(t, IsAuthLevelSufficient(authentication.OneFactor, Bypass))
	assert.True(t, IsAuthLevelSufficient(authentication.TwoFactor, Bypass))
	assert.False(t, IsAuthLevelSufficient(authentication.NotAuthenticated, OneFactor))
	assert.True(t, IsAuthLevelSufficient(authentication.OneFactor, OneFactor))
	assert.True(t, IsAuthLevelSufficient(authentication.TwoFactor, OneFactor))
	assert.False(t, IsAuthLevelSufficient(authentication.NotAuthenticated, TwoFactor))
	assert.False(t, IsAuthLevelSufficient(authentication.OneFactor, TwoFactor))
	assert.True(t, IsAuthLevelSufficient(authentication.TwoFactor, TwoFactor))
}

func TestStringSliceToRegexpSlice(t *testing.T) {
	testCases := []struct {
		name     string
		have     []string
		expected []regexp.Regexp
		err      string
	}{
		{
			"ShouldNotParseBadRegex",
			[]string{`\q`},
			[]regexp.Regexp(nil),
			"error parsing regexp: invalid escape sequence: `\\q`",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, theError := stringSliceToRegexpSlice(tc.have)

			assert.Equal(t, tc.expected, actual)

			if tc.err == "" {
				assert.NoError(t, theError)
			} else {
				assert.EqualError(t, theError, tc.err)
			}
		})
	}
}

func TestIsOpenIDConnectMFA(t *testing.T) {
	testCases := []struct {
		name     string
		have     *schema.Configuration
		expected bool
	}{
		{
			"ShouldReturnFalseNilConfig",
			nil,
			false,
		},
		{
			"ShouldReturnFalseNilOIDC",
			&schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: nil,
				},
			},
			false,
		},
		{
			"ShouldReturnFalseNoClients",
			&schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						Clients: nil,
					},
				},
			},
			false,
		},
		{
			"ShouldReturnFalseNoClients2FA",
			&schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						Clients: []schema.IdentityProvidersOpenIDConnectClient{
							{
								ID:                  "one",
								AuthorizationPolicy: "one_factor",
							},
						},
					},
				},
			},
			false,
		},
		{
			"ShouldReturnTrueClientsDirect2FA",
			&schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						Clients: []schema.IdentityProvidersOpenIDConnectClient{
							{
								ID:                  "one",
								AuthorizationPolicy: "two_factor",
							},
						},
					},
				},
			},
			true,
		},
		{
			"ShouldReturnTrueClientsIndirect2FADefault",
			&schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						AuthorizationPolicies: map[string]schema.IdentityProvidersOpenIDConnectPolicy{
							"example": {
								DefaultPolicy: "two_factor",
							},
						},
						Clients: []schema.IdentityProvidersOpenIDConnectClient{
							{
								ID:                  "one",
								AuthorizationPolicy: "example",
							},
						},
					},
				},
			},
			true,
		},
		{
			"ShouldReturnTrueClientsIndirect2FADefault",
			&schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						AuthorizationPolicies: map[string]schema.IdentityProvidersOpenIDConnectPolicy{
							"example": {
								DefaultPolicy: "deny",
								Rules: []schema.IdentityProvidersOpenIDConnectPolicyRule{
									{
										Policy: "two_factor",
									},
								},
							},
						},
						Clients: []schema.IdentityProvidersOpenIDConnectClient{
							{
								ID:                  "one",
								AuthorizationPolicy: "example",
							},
						},
					},
				},
			},
			true,
		},
		{
			"ShouldReturnTrueClientsIndirect2FADefault",
			&schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						AuthorizationPolicies: map[string]schema.IdentityProvidersOpenIDConnectPolicy{
							"example": {
								DefaultPolicy: "deny",
								Rules: []schema.IdentityProvidersOpenIDConnectPolicyRule{
									{
										Policy: "two_factor",
									},
								},
							},
						},
						Clients: []schema.IdentityProvidersOpenIDConnectClient{
							{
								ID:                  "skip",
								AuthorizationPolicy: "skip",
							},
							{
								ID:                  "one",
								AuthorizationPolicy: "example",
							},
						},
					},
				},
			},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, isOpenIDConnectMFA(tc.have))
		})
	}
}

func MustParseCIDR(input string) *net.IPNet {
	_, out, err := net.ParseCIDR(input)
	if err != nil {
		panic(err)
	}

	return out
}
