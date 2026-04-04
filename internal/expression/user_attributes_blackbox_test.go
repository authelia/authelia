package expression_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	. "github.com/authelia/authelia/v4/internal/expression"
)

func TestResolve(t *testing.T) {
	testCases := []struct {
		name      string
		have      func(t *testing.T) UserAttributeResolver
		attribute string
		detailer  UserDetailer
		updated   time.Time
		expected  any
		found     bool
	}{
		{
			name: "ShouldHandleBasicResolver",
			have: func(t *testing.T) UserAttributeResolver {
				return &UserAttributes{}
			},
			attribute: "example",
			detailer:  &authentication.UserDetailsExtended{Extra: map[string]any{"example": 1}},
			updated:   time.Now(),
			expected:  1,
			found:     true,
		},
		{
			name: "ShouldHandleBasicResolverNotExpression",
			have: func(t *testing.T) UserAttributeResolver {
				resolver := NewUserAttributes(&schema.Configuration{
					AuthenticationBackend: schema.AuthenticationBackend{
						File: &schema.AuthenticationBackendFile{
							ExtraAttributes: map[string]schema.AuthenticationBackendExtraAttribute{
								"example_1": {
									ValueType: "string",
								},
								"example_2": {
									ValueType: "string",
								},
							},
						},
					},
					Definitions: schema.Definitions{
						UserAttributes: map[string]schema.UserAttribute{
							"example": {
								Expression: "example_1 + example_2",
							},
						},
					},
				})

				require.NoError(t, resolver.StartupCheck())

				return resolver
			},
			attribute: "example_1",
			detailer:  &authentication.UserDetailsExtended{Extra: map[string]any{"example_1": "abc", "example_2": "xyz"}},
			updated:   time.Now(),
			expected:  "abc",
			found:     true,
		},
		{
			name: "ShouldHandleBasicResolverAdvanced",
			have: func(t *testing.T) UserAttributeResolver {
				resolver := NewUserAttributes(&schema.Configuration{
					AuthenticationBackend: schema.AuthenticationBackend{
						File: &schema.AuthenticationBackendFile{
							ExtraAttributes: map[string]schema.AuthenticationBackendExtraAttribute{
								"example_1": {
									ValueType: "string",
								},
								"example_2": {
									ValueType: "string",
								},
							},
						},
					},
					Definitions: schema.Definitions{
						UserAttributes: map[string]schema.UserAttribute{
							"example": {
								Expression: "example_1 + example_2",
							},
						},
					},
				})

				require.NoError(t, resolver.StartupCheck())

				return resolver
			},
			attribute: "example",
			detailer:  &authentication.UserDetailsExtended{Extra: map[string]any{"example_1": "abc", "example_2": "xyz"}},
			updated:   time.Now(),
			expected:  "abcxyz",
			found:     true,
		},
		{
			name: "ShouldHandleBasicResolverAdvancedNoValue",
			have: func(t *testing.T) UserAttributeResolver {
				resolver := NewUserAttributes(&schema.Configuration{
					AuthenticationBackend: schema.AuthenticationBackend{
						File: &schema.AuthenticationBackendFile{
							ExtraAttributes: map[string]schema.AuthenticationBackendExtraAttribute{
								"example_1": {
									ValueType: "string",
								},
								"example_2": {
									ValueType: "string",
								},
							},
						},
					},
					Definitions: schema.Definitions{
						UserAttributes: map[string]schema.UserAttribute{
							"example": {
								Expression: "example_1 + example_2",
							},
						},
					},
				})

				require.NoError(t, resolver.StartupCheck())

				return resolver
			},
			attribute: "example",
			detailer:  &authentication.UserDetailsExtended{Extra: map[string]any{}},
			updated:   time.Now(),
			expected:  nil,
			found:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resolver := tc.have(t)

			actual, found := resolver.Resolve(tc.attribute, tc.detailer, tc.updated)
			assert.Equal(t, tc.expected, actual)
			assert.Equal(t, tc.found, found)
		})
	}
}

func TestResolveWithExtra(t *testing.T) {
	testCases := []struct {
		name      string
		have      func(t *testing.T) UserAttributeResolver
		attribute string
		detailer  UserDetailer
		updated   time.Time
		extra     map[string]any
		expected  any
		found     bool
	}{
		{
			name: "ShouldResolveExtraFromUserAttributes",
			have: func(t *testing.T) UserAttributeResolver {
				return &UserAttributes{}
			},
			attribute: AttributeUserUsername,
			detailer: &authentication.UserDetailsExtended{
				UserDetails: &authentication.UserDetails{
					Username: "jsmith",
				},
			},
			updated:  time.Now(),
			extra:    map[string]any{"custom": "extra_value"},
			expected: "jsmith",
			found:    true,
		},
		{
			name: "ShouldResolveExtraAttributeFromUserAttributes",
			have: func(t *testing.T) UserAttributeResolver {
				return &UserAttributes{}
			},
			attribute: "custom",
			detailer: &authentication.UserDetailsExtended{
				UserDetails: &authentication.UserDetails{
					Username: "jsmith",
				},
			},
			updated:  time.Now(),
			extra:    map[string]any{"custom": "extra_value"},
			expected: "extra_value",
			found:    true,
		},
		{
			name: "ShouldResolveWithNilExtra",
			have: func(t *testing.T) UserAttributeResolver {
				return &UserAttributes{}
			},
			attribute: AttributeUserUsername,
			detailer: &authentication.UserDetailsExtended{
				UserDetails: &authentication.UserDetails{
					Username: "jsmith",
				},
			},
			updated:  time.Now(),
			extra:    nil,
			expected: "jsmith",
			found:    true,
		},
		{
			name: "ShouldResolveExtraFromExpressions",
			have: func(t *testing.T) UserAttributeResolver {
				resolver := NewUserAttributes(&schema.Configuration{
					AuthenticationBackend: schema.AuthenticationBackend{
						File: &schema.AuthenticationBackendFile{
							ExtraAttributes: map[string]schema.AuthenticationBackendExtraAttribute{
								"example_1": {ValueType: "string"},
							},
						},
					},
					Definitions: schema.Definitions{
						UserAttributes: map[string]schema.UserAttribute{
							"computed": {Expression: "example_1 + '_suffix'"},
						},
					},
				})

				require.NoError(t, resolver.StartupCheck())

				return resolver
			},
			attribute: "computed",
			detailer:  &authentication.UserDetailsExtended{Extra: map[string]any{"example_1": "abc"}},
			updated:   time.Now(),
			extra:     map[string]any{"other": "value"},
			expected:  "abc_suffix",
			found:     true,
		},
		{
			name: "ShouldResolveExtraAttributeFromExpressionsParent",
			have: func(t *testing.T) UserAttributeResolver {
				resolver := NewUserAttributes(&schema.Configuration{
					AuthenticationBackend: schema.AuthenticationBackend{
						File: &schema.AuthenticationBackendFile{},
					},
					Definitions: schema.Definitions{
						UserAttributes: map[string]schema.UserAttribute{
							"computed": {Expression: "username"},
						},
					},
				})

				require.NoError(t, resolver.StartupCheck())

				return resolver
			},
			attribute: "custom_extra",
			detailer: &authentication.UserDetailsExtended{
				UserDetails: &authentication.UserDetails{
					Username: "jsmith",
				},
			},
			updated:  time.Now(),
			extra:    map[string]any{"custom_extra": "from_extra"},
			expected: "from_extra",
			found:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resolver := tc.have(t)

			actual, found := resolver.ResolveWithExtra(tc.attribute, tc.detailer, tc.updated, tc.extra)
			assert.Equal(t, tc.expected, actual)
			assert.Equal(t, tc.found, found)
		})
	}
}
