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

func TestShouldParseRuleNetworks(t *testing.T) {
	schemaNetworks := []schema.AccessControlNetwork{
		{
			Name: "desktop",
			Networks: []string{
				"10.0.0.1",
			},
		},
		{
			Name: "lan",
			Networks: []string{
				"10.0.0.0/8",
				"172.16.0.0/12",
				"192.168.0.0/16",
			},
		},
	}

	_, firstNetwork, err := net.ParseCIDR("192.168.1.20/32")
	require.NoError(t, err)

	networksMap, networksCacheMap := parseSchemaNetworks(schemaNetworks)

	assert.Len(t, networksCacheMap, 5)

	networks := []string{"192.168.1.20", "lan"}

	acl := schemaNetworksToACL(networks, networksMap, networksCacheMap)

	assert.Len(t, networksCacheMap, 7)

	require.Len(t, acl, 4)
	assert.Equal(t, firstNetwork, acl[0])
	assert.Equal(t, networksMap["lan"][0], acl[1])
	assert.Equal(t, networksMap["lan"][1], acl[2])
	assert.Equal(t, networksMap["lan"][2], acl[3])

	// Check they are the same memory address.
	assert.True(t, networksMap["lan"][0] == acl[1])
	assert.True(t, networksMap["lan"][1] == acl[2])
	assert.True(t, networksMap["lan"][2] == acl[3])

	assert.False(t, firstNetwork == acl[0])
}

func TestShouldParseACLNetworks(t *testing.T) {
	schemaNetworks := []schema.AccessControlNetwork{
		{
			Name: "test",
			Networks: []string{
				"10.0.0.1",
			},
		},
		{
			Name: "second",
			Networks: []string{
				"10.0.0.1",
			},
		},
		{
			Name: "duplicate",
			Networks: []string{
				"10.0.0.1",
			},
		},
		{
			Name: "duplicate",
			Networks: []string{
				"10.0.0.1",
			},
		},
		{
			Name: "ipv6",
			Networks: []string{
				"fec0::1",
			},
		},
		{
			Name: "ipv6net",
			Networks: []string{
				"fec0::1/64",
			},
		},
		{
			Name: "net",
			Networks: []string{
				"10.0.0.0/8",
			},
		},
		{
			Name: "badnet",
			Networks: []string{
				"bad/8",
			},
		},
	}

	_, firstNetwork, err := net.ParseCIDR("10.0.0.1/32")
	require.NoError(t, err)

	_, secondNetwork, err := net.ParseCIDR("10.0.0.0/8")
	require.NoError(t, err)

	_, thirdNetwork, err := net.ParseCIDR("fec0::1/64")
	require.NoError(t, err)

	_, fourthNetwork, err := net.ParseCIDR("fec0::1/128")
	require.NoError(t, err)

	networksMap, networksCacheMap := parseSchemaNetworks(schemaNetworks)

	require.Len(t, networksMap, 6)
	require.Contains(t, networksMap, "test")
	require.Contains(t, networksMap, "second")
	require.Contains(t, networksMap, "duplicate")
	require.Contains(t, networksMap, "ipv6")
	require.Contains(t, networksMap, "ipv6net")
	require.Contains(t, networksMap, "net")
	require.Len(t, networksMap["test"], 1)

	require.Len(t, networksCacheMap, 7)
	require.Contains(t, networksCacheMap, "10.0.0.1")
	require.Contains(t, networksCacheMap, "10.0.0.1/32")
	require.Contains(t, networksCacheMap, "10.0.0.1/32")
	require.Contains(t, networksCacheMap, "10.0.0.0/8")
	require.Contains(t, networksCacheMap, "fec0::1")
	require.Contains(t, networksCacheMap, "fec0::1/128")
	require.Contains(t, networksCacheMap, "fec0::1/64")

	assert.Equal(t, firstNetwork, networksMap["test"][0])
	assert.Equal(t, secondNetwork, networksMap["net"][0])
	assert.Equal(t, thirdNetwork, networksMap["ipv6net"][0])
	assert.Equal(t, fourthNetwork, networksMap["ipv6"][0])

	assert.Equal(t, firstNetwork, networksCacheMap["10.0.0.1"])
	assert.Equal(t, firstNetwork, networksCacheMap["10.0.0.1/32"])

	assert.Equal(t, secondNetwork, networksCacheMap["10.0.0.0/8"])

	assert.Equal(t, thirdNetwork, networksCacheMap["fec0::1/64"])

	assert.Equal(t, fourthNetwork, networksCacheMap["fec0::1"])
	assert.Equal(t, fourthNetwork, networksCacheMap["fec0::1/128"])
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

func TestSchemaNetworksToACL(t *testing.T) {
	testCases := []struct {
		name     string
		have     []string
		globals  map[string][]*net.IPNet
		cache    map[string]*net.IPNet
		expected []*net.IPNet
	}{
		{
			"ShouldLoadFromCache",
			[]string{"192.168.0.0/24"},
			nil,
			map[string]*net.IPNet{"192.168.0.0/24": MustParseCIDR("192.168.0.0/24")},
			[]*net.IPNet{MustParseCIDR("192.168.0.0/24")},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.globals == nil {
				tc.globals = map[string][]*net.IPNet{}
			}

			if tc.cache == nil {
				tc.cache = map[string]*net.IPNet{}
			}

			actual := schemaNetworksToACL(tc.have, tc.globals, tc.cache)

			assert.Equal(t, tc.expected, actual)
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
