// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/storage"
)

func newIDJAGTestConfiguration() *schema.Configuration {
	singleUse := false

	return &schema.Configuration{
		IdentityProviders: schema.IdentityProviders{
			OIDC: &schema.IdentityProvidersOpenIDConnect{
				Scopes: map[string]schema.IdentityProvidersOpenIDConnectScope{
					"tenancy": {Claims: []string{"tenant", "aud_tenant"}},
				},
				ClaimsPolicies: map[string]schema.IdentityProvidersOpenIDConnectClaimsPolicy{
					"ras": {
						IDToken: []string{oidc.ClaimEmail, oidc.ClaimFullName, "tenant", "aud_tenant"},
						CustomClaims: schema.IdentityProvidersOpenIDConnectCustomClaims{
							"tenant":     {Name: "tenant", Attribute: "display_name"},
							"aud_tenant": {Name: "aud_tenant", Attribute: "username"},
						},
					},
				},
				IdentityAssertionAuthorizationGrant: schema.IdentityProvidersOpenIDConnectIDJAG{
					Lifespan:             time.Minute * 2,
					SingleUse:            &singleUse,
					AccessTokenAllowance: time.Minute,
					ResourceOptional:     true,
					ResourceAuthorizationServers: []schema.IdentityProvidersOpenIDConnectIDJAGResourceAuthorizationServer{
						{
							Issuer:           MustParseRequestURI("https://ras.example.com"),
							Aliases:          []string{"ras", "https://ras.example.com/alt"},
							SigningAlgorithm: oidc.SigningAlgECDSAUsingP256AndSHA256,
							MaxAge:           time.Hour,
							SubjectClientID:  "ras-sso",
							ClaimsPolicy:     "ras",
							JWTDPoPSupported: true,
						},
						{
							Issuer:           MustParseRequestURI("https://other.example.com"),
							SigningAlgorithm: oidc.SigningAlgRSAUsingSHA256,
							SubjectClientID:  "ras-sso",
						},
					},
					TrustedIssuers: []schema.IdentityProvidersOpenIDConnectIDJAGTrustedIssuer{
						{
							Issuer:                          MustParseRequestURI("https://idp.example.com"),
							JSONWebKeysURI:                  MustParseRequestURI("https://idp.example.com/jwks.json"),
							SigningAlgorithms:               []string{oidc.SigningAlgRSAPSSUsingSHA256, oidc.SigningAlgRSAUsingSHA256},
							Clients:                         []string{"app-a"},
							Resources:                       []string{"https://api.example.com"},
							UsernameClaim:                   oidc.ClaimPreferredUsername,
							AuthenticationMethodsReferences: `["pwd"]`,
						},
						{
							Issuer:            MustParseRequestURI("https://idp2.example.com"),
							JSONWebKeysURI:    MustParseRequestURI("https://idp2.example.com/jwks.json"),
							SigningAlgorithms: []string{oidc.SigningAlgECDSAUsingP256AndSHA256, oidc.SigningAlgRSAUsingSHA256},
							Clients:           []string{"app-b"},
							UsernameClaim:     oidc.ClaimSubject,
						},
					},
				},
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:                  "app-a",
						SectorIdentifierURI: MustParseRequestURI("https://a.example.com"),
						IDJAGRelationships: []schema.IdentityProvidersOpenIDConnectClientIDJAGRelationship{
							{
								Issuer:    MustParseRequestURI("https://ras.example.com"),
								ClientID:  "ras-client-a",
								Scopes:    []string{"read", "write"},
								Resources: []string{"https://api.example.com"},
							},
						},
						IDJAGResourceOptional: true,
					},
					{
						ID:                  "app-b",
						SectorIdentifierURI: MustParseRequestURI("https://b.example.com"),
						IDJAGRelationships: []schema.IdentityProvidersOpenIDConnectClientIDJAGRelationship{
							{
								Issuer:                    MustParseRequestURI("https://ras.example.com"),
								ClientID:                  "ras-client-b",
								AuthorizationDetailsTypes: []string{"payment"},
							},
							{
								Issuer:   MustParseRequestURI("https://other.example.com"),
								ClientID: "ras-client-b-other",
							},
						},
					},
					{
						ID:                  "ras-sso",
						SectorIdentifierURI: MustParseRequestURI("https://ras.example.com"),
					},
				},
			},
		},
	}
}

func newIDJAGTestConfig(t *testing.T, config *schema.IdentityProvidersOpenIDConnect) *oidc.Config {
	idjag, err := oidc.NewIDJAGConfig(config)
	require.NoError(t, err)

	c := oidc.NewConfig(config, nil, nil)
	c.IDJAG = idjag

	return c
}

func newIDJAGTestStore(t *testing.T, provider storage.Provider) *oidc.Store {
	config := newIDJAGTestConfiguration()

	return oidc.NewStore(config, provider, newIDJAGTestConfig(t, config.IdentityProviders.OIDC))
}

func newIDJAGTestRequest(t *testing.T, store *oidc.Store, clientID, username string) *oauthelia2.AccessRequest {
	client, err := store.GetRegisteredClient(context.Background(), clientID)
	require.NoError(t, err)

	session := oidc.NewSession()
	session.Username = username

	request := oauthelia2.NewAccessRequest(session)
	request.Client = client

	return request
}

func TestIDJAGConfigLookup(t *testing.T) {
	config := newIDJAGTestConfig(t, newIDJAGTestConfiguration().IdentityProviders.OIDC)

	testCases := []struct {
		name     string
		audience string
		expected string
		ok       bool
	}{
		{"ShouldMatchIssuer", "https://ras.example.com", "https://ras.example.com", true},
		{"ShouldMatchAlias", "ras", "https://ras.example.com", true},
		{"ShouldMatchURLAlias", "https://ras.example.com/alt", "https://ras.example.com", true},
		{"ShouldMatchOtherIssuer", "https://other.example.com", "https://other.example.com", true},
		{"ShouldNotMatchUnknown", "https://unknown.example.com", "", false},
		{"ShouldNotMatchEmpty", "", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ras, ok := config.IDJAG.LookupResourceAuthorizationServer(tc.audience)

			assert.Equal(t, tc.ok, ok)

			if tc.ok {
				require.NotNil(t, ras)
				assert.Equal(t, tc.expected, ras.Issuer)
			} else {
				assert.Nil(t, ras)
			}
		})
	}

	ras, ok := config.IDJAG.LookupResourceAuthorizationServer("https://ras.example.com")
	require.True(t, ok)

	assert.Equal(t, oidc.SigningAlgECDSAUsingP256AndSHA256, ras.SigningAlg)
	assert.Equal(t, time.Hour, ras.MaxAge)
	assert.Equal(t, "ras-sso", ras.SubjectClientID)
	assert.Equal(t, "ras", ras.ClaimsPolicy)
	assert.True(t, ras.JWTDPoPSupported)

	trusted, ok := config.IDJAG.TrustedIssuers["https://idp.example.com"]
	require.True(t, ok)

	assert.Equal(t, "https://idp.example.com", trusted.Issuer)
	assert.Equal(t, "https://idp.example.com/jwks.json", trusted.JSONWebKeysURI)
	assert.Nil(t, trusted.JSONWebKeys)
	assert.Equal(t, []string{oidc.SigningAlgRSAPSSUsingSHA256, oidc.SigningAlgRSAUsingSHA256}, trusted.SigningAlgs)
	assert.Equal(t, []string{"app-a"}, trusted.Clients)
	assert.False(t, trusted.MultiTenant)
	assert.Equal(t, []string{"https://api.example.com"}, trusted.Resources)
	assert.Equal(t, oidc.ClaimPreferredUsername, trusted.UsernameClaim)
	require.NotNil(t, trusted.AMR)

	amr, err := trusted.AMR.Evaluate(map[string]any{})
	require.NoError(t, err)
	assert.Equal(t, []string{"pwd"}, amr)

	trusted, ok = config.IDJAG.TrustedIssuers["https://idp2.example.com"]
	require.True(t, ok)
	assert.Nil(t, trusted.AMR)
}

func TestConfigIDJAGGetters(t *testing.T) {
	ctx := context.Background()

	config := oidc.NewConfig(&schema.IdentityProvidersOpenIDConnect{}, nil, nil)

	assert.Equal(t, time.Minute*5, config.GetIDJAGLifespan(ctx))
	assert.True(t, config.GetIDJAGSingleUse(ctx))
	assert.Equal(t, time.Duration(0), config.GetIDJAGAccessTokenAllowance(ctx))
	assert.False(t, config.GetIDJAGResourceOptional(ctx))

	config = newIDJAGTestConfig(t, newIDJAGTestConfiguration().IdentityProviders.OIDC)

	assert.Equal(t, time.Minute*2, config.GetIDJAGLifespan(ctx))
	assert.False(t, config.GetIDJAGSingleUse(ctx))
	assert.Equal(t, time.Minute, config.GetIDJAGAccessTokenAllowance(ctx))
	assert.True(t, config.GetIDJAGResourceOptional(ctx))
}

func TestConfigIDJAGSigningAlgValues(t *testing.T) {
	ctx := context.Background()

	config := oidc.NewConfig(&schema.IdentityProvidersOpenIDConnect{}, nil, nil)

	assert.Nil(t, config.GetIDJAGIssueSigningAlgValues(ctx))
	assert.Nil(t, config.GetIDJAGRedeemSigningAlgValues(ctx))

	config = newIDJAGTestConfig(t, newIDJAGTestConfiguration().IdentityProviders.OIDC)

	assert.Equal(t, []string{oidc.SigningAlgECDSAUsingP256AndSHA256, oidc.SigningAlgRSAUsingSHA256}, config.GetIDJAGIssueSigningAlgValues(ctx))
	assert.Equal(t, []string{oidc.SigningAlgECDSAUsingP256AndSHA256, oidc.SigningAlgRSAPSSUsingSHA256, oidc.SigningAlgRSAUsingSHA256}, config.GetIDJAGRedeemSigningAlgValues(ctx))

	config = newIDJAGTestConfig(t, &schema.IdentityProvidersOpenIDConnect{
		IdentityAssertionAuthorizationGrant: schema.IdentityProvidersOpenIDConnectIDJAG{
			ResourceAuthorizationServers: []schema.IdentityProvidersOpenIDConnectIDJAGResourceAuthorizationServer{
				{Issuer: MustParseRequestURI("https://ras.example.com"), SigningAlgorithm: oidc.SigningAlgRSAPSSUsingSHA512},
				{Issuer: MustParseRequestURI("https://ras2.example.com"), SigningAlgorithm: oidc.SigningAlgRSAPSSUsingSHA512},
			},
		},
	})

	assert.Equal(t, []string{oidc.SigningAlgRSAPSSUsingSHA512, oidc.SigningAlgRSAUsingSHA256}, config.GetIDJAGIssueSigningAlgValues(ctx))
	assert.Nil(t, config.GetIDJAGRedeemSigningAlgValues(ctx))
}

func TestRegisteredClientIDJAG(t *testing.T) {
	config := newIDJAGTestConfiguration()

	client := oidc.NewClient(config.IdentityProviders.OIDC.Clients[0], config.IdentityProviders.OIDC, nil)

	assert.True(t, client.GetIDJAGResourceOptional())

	relationship, ok := client.GetIDJAGRelationship("https://ras.example.com")
	require.True(t, ok)
	assert.Equal(t, &oidc.IDJAGClientRelationship{
		ClientID:  "ras-client-a",
		Scopes:    []string{"read", "write"},
		Resources: []string{"https://api.example.com"},
	}, relationship)

	relationship, ok = client.GetIDJAGRelationship("https://other.example.com")
	assert.False(t, ok)
	assert.Nil(t, relationship)

	client = oidc.NewClient(config.IdentityProviders.OIDC.Clients[1], config.IdentityProviders.OIDC, nil)

	assert.False(t, client.GetIDJAGResourceOptional())

	relationship, ok = client.GetIDJAGRelationship("https://ras.example.com")
	require.True(t, ok)
	assert.Equal(t, "ras-client-b", relationship.ClientID)
	assert.Equal(t, []string{"payment"}, relationship.AuthorizationDetailsTypes)

	client = oidc.NewClient(config.IdentityProviders.OIDC.Clients[2], config.IdentityProviders.OIDC, nil)

	relationship, ok = client.GetIDJAGRelationship("https://ras.example.com")
	assert.False(t, ok)
	assert.Nil(t, relationship)

	var nilClient *oidc.RegisteredClient

	relationship, ok = nilClient.GetIDJAGRelationship("https://ras.example.com")
	assert.False(t, ok)
	assert.Nil(t, relationship)
	assert.False(t, nilClient.GetIDJAGResourceOptional())
}

func TestStoreGetIDJAGRelationship(t *testing.T) {
	store := newIDJAGTestStore(t, nil)

	t.Run("ShouldReturnNotFoundForUnknownAudience", func(t *testing.T) {
		relationship, err := store.GetIDJAGRelationship(context.Background(), newIDJAGTestRequest(t, store, "app-a", "john"), "https://unknown.example.com")

		assert.ErrorIs(t, err, oauthelia2.ErrNotFound)
		assert.Nil(t, relationship)
	})

	t.Run("ShouldReturnAccessDeniedWithoutClientRelationship", func(t *testing.T) {
		relationship, err := store.GetIDJAGRelationship(context.Background(), newIDJAGTestRequest(t, store, "app-a", "john"), "https://other.example.com")

		assert.ErrorIs(t, err, oauthelia2.ErrAccessDenied)
		assert.Nil(t, relationship)

		relationship, err = store.GetIDJAGRelationship(context.Background(), newIDJAGTestRequest(t, store, "ras-sso", "john"), "ras")

		assert.ErrorIs(t, err, oauthelia2.ErrAccessDenied)
		assert.Nil(t, relationship)
	})

	t.Run("ShouldMergeRelationshipForIssuer", func(t *testing.T) {
		relationship, err := store.GetIDJAGRelationship(context.Background(), newIDJAGTestRequest(t, store, "app-a", "john"), "https://ras.example.com")

		require.NoError(t, err)
		assert.Equal(t, &oauthelia2.IDJAGRelationship{
			Issuer:                    "https://ras.example.com",
			ClientID:                  "ras-client-a",
			Scopes:                    []string{"read", "write"},
			Resources:                 []string{"https://api.example.com"},
			AuthorizationDetailsTypes: []string{},
			SigningAlg:                oidc.SigningAlgECDSAUsingP256AndSHA256,
			MaxAge:                    time.Hour,
		}, relationship)
		assert.NotNil(t, relationship.AuthorizationDetailsTypes)
	})

	t.Run("ShouldResolveAliasToIssuer", func(t *testing.T) {
		relationship, err := store.GetIDJAGRelationship(context.Background(), newIDJAGTestRequest(t, store, "app-b", "john"), "https://ras.example.com/alt")

		require.NoError(t, err)
		assert.Equal(t, "https://ras.example.com", relationship.Issuer)
		assert.Equal(t, "ras-client-b", relationship.ClientID)
		assert.Equal(t, []string{"payment"}, relationship.AuthorizationDetailsTypes)
		assert.Nil(t, relationship.ACRValues)
	})

	t.Run("ShouldReturnServerErrorForNonOIDCClient", func(t *testing.T) {
		request := oauthelia2.NewAccessRequest(oidc.NewSession())
		request.Client = &oauthelia2.DefaultClient{ID: "app-a"}

		relationship, err := store.GetIDJAGRelationship(context.Background(), request, "https://ras.example.com")

		assert.ErrorIs(t, err, oauthelia2.ErrServerError)
		assert.Nil(t, relationship)
	})
}

func TestStoreGetIDJAGSubjectStableAcrossClients(t *testing.T) {
	ctrl := gomock.NewController(t)

	provider := mocks.NewMockStorage(ctrl)

	store := newIDJAGTestStore(t, provider)

	ctx := &TestContext{Context: context.Background(), Storage: provider}

	expected := uuid.MustParse("8d6c7e0c-5f0e-4a4b-9b3f-0f3e9e3c1a2b")

	provider.
		EXPECT().
		LoadUserOpaqueIdentifierBySignature(ctx, "openid", "https://ras.example.com", "john").
		Return(&model.UserOpaqueIdentifier{Service: "openid", SectorID: "https://ras.example.com", Username: "john", Identifier: expected}, nil).
		Times(2)

	subjects := make([]string, 0, 2)

	for _, clientID := range []string{"app-a", "app-b"} {
		request := newIDJAGTestRequest(t, store, clientID, "john")

		relationship, err := store.GetIDJAGRelationship(ctx, request, "https://ras.example.com")
		require.NoError(t, err)

		subject, err := store.GetIDJAGSubject(ctx, request, relationship)
		require.NoError(t, err)

		subjects = append(subjects, subject)
	}

	assert.Equal(t, []string{expected.String(), expected.String()}, subjects)
}

func TestStoreGetIDJAGSubjectNoUsername(t *testing.T) {
	ctrl := gomock.NewController(t)

	provider := mocks.NewMockStorage(ctrl)

	store := newIDJAGTestStore(t, provider)

	ctx := &TestContext{Context: context.Background(), Storage: provider}

	request := newIDJAGTestRequest(t, store, "app-a", "")

	subject, err := store.GetIDJAGSubject(ctx, request, &oauthelia2.IDJAGRelationship{Issuer: "https://ras.example.com"})

	assert.ErrorIs(t, err, oauthelia2.ErrNotFound)
	assert.Equal(t, "", subject)
}

func TestStoreGetIDJAGClaims(t *testing.T) {
	ctrl := gomock.NewController(t)

	users := mocks.NewMockUserProvider(ctrl)

	store := newIDJAGTestStore(t, nil)

	ctx := &TestContext{Context: context.Background(), UserProvider: users}

	t.Run("ShouldReturnPolicyClaims", func(t *testing.T) {
		users.
			EXPECT().
			GetDetailsExtended("john").
			Return(&authentication.UserDetailsExtended{
				UserDetails: &authentication.UserDetails{
					Username:    "john",
					DisplayName: "John Smith",
					Emails:      []string{"john@example.com"},
					Groups:      []string{"admins"},
				},
			}, nil)

		claims, err := store.GetIDJAGClaims(ctx, newIDJAGTestRequest(t, store, "app-a", "john"), &oauthelia2.IDJAGRelationship{Issuer: "https://ras.example.com"})

		require.NoError(t, err)
		assert.Equal(t, map[string]any{
			oidc.ClaimEmail:    "john@example.com",
			oidc.ClaimFullName: "John Smith",
		}, claims)
		assert.NotContains(t, claims, "tenant")
		assert.NotContains(t, claims, "aud_tenant")
	})

	t.Run("ShouldReturnPolicyClaimsFromWrappedContext", func(t *testing.T) {
		users.
			EXPECT().
			GetDetailsExtended("john").
			Return(&authentication.UserDetailsExtended{
				UserDetails: &authentication.UserDetails{
					Username: "john",
					Emails:   []string{"john@example.com"},
				},
			}, nil)

		wrapped := context.WithValue(context.WithValue(context.Background(), model.CtxKeyAutheliaCtx, ctx), oauthelia2.RequestContextKey, nil)

		claims, err := store.GetIDJAGClaims(wrapped, newIDJAGTestRequest(t, store, "app-a", "john"), &oauthelia2.IDJAGRelationship{Issuer: "https://ras.example.com"})

		require.NoError(t, err)
		assert.Equal(t, "john@example.com", claims[oidc.ClaimEmail])
	})

	t.Run("ShouldReturnNilWithoutPolicy", func(t *testing.T) {
		claims, err := store.GetIDJAGClaims(ctx, newIDJAGTestRequest(t, store, "app-b", "john"), &oauthelia2.IDJAGRelationship{Issuer: "https://other.example.com"})

		require.NoError(t, err)
		assert.Nil(t, claims)
	})

	t.Run("ShouldReturnServerErrorOnUserLookupError", func(t *testing.T) {
		users.
			EXPECT().
			GetDetailsExtended("john").
			Return(nil, authentication.ErrUserNotFound)

		claims, err := store.GetIDJAGClaims(ctx, newIDJAGTestRequest(t, store, "app-a", "john"), &oauthelia2.IDJAGRelationship{Issuer: "https://ras.example.com"})

		assert.ErrorIs(t, err, oauthelia2.ErrServerError)
		assert.Nil(t, claims)
	})

	t.Run("ShouldReturnServerErrorWithoutUserProvider", func(t *testing.T) {
		claims, err := store.GetIDJAGClaims(&TestContext{Context: context.Background()}, newIDJAGTestRequest(t, store, "app-a", "john"), &oauthelia2.IDJAGRelationship{Issuer: "https://ras.example.com"})

		assert.ErrorIs(t, err, oauthelia2.ErrServerError)
		assert.Nil(t, claims)
	})

	t.Run("ShouldReturnNotFoundWithoutUsername", func(t *testing.T) {
		claims, err := store.GetIDJAGClaims(ctx, newIDJAGTestRequest(t, store, "app-a", ""), &oauthelia2.IDJAGRelationship{Issuer: "https://ras.example.com"})

		assert.ErrorIs(t, err, oauthelia2.ErrNotFound)
		assert.Nil(t, claims)
	})

	t.Run("ShouldReturnServerErrorWithoutOIDCContext", func(t *testing.T) {
		claims, err := store.GetIDJAGClaims(context.Background(), newIDJAGTestRequest(t, store, "app-a", "john"), &oauthelia2.IDJAGRelationship{Issuer: "https://ras.example.com"})

		assert.ErrorIs(t, err, oauthelia2.ErrServerError)
		assert.Nil(t, claims)
	})
}

func TestNewIDJAGConfigAuthenticationMethodsReferencesInvalid(t *testing.T) {
	config := newIDJAGTestConfiguration()
	config.IdentityProviders.OIDC.IdentityAssertionAuthorizationGrant.TrustedIssuers[0].AuthenticationMethodsReferences = "claims.amr +"

	idjag, err := oidc.NewIDJAGConfig(config.IdentityProviders.OIDC)

	assert.Nil(t, idjag)
	assert.ErrorContains(t, err, "error occurred compiling the authentication methods references expression for the trusted issuer 'https://idp.example.com'")
}

func TestNewOpenIDConnectProviderIDJAG(t *testing.T) {
	config := newIDJAGTestConfiguration()
	config.IdentityProviders.OIDC.IssuerPrivateKey = x509PrivateKeyRSA2048
	config.IdentityProviders.OIDC.HMACSecret = badhmac

	provider, err := oidc.NewOpenIDConnectProvider(config, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, provider)
	require.NotNil(t, provider.IDJAG)

	trusted, err := provider.GetIDJAGTrustedIssuer(context.Background(), "https://idp.example.com")
	require.NoError(t, err)
	assert.Same(t, &provider.IDJAG.TrustedIssuers["https://idp.example.com"].IDJAGTrustedIssuer, trusted)

	config.IdentityProviders.OIDC.IdentityAssertionAuthorizationGrant.TrustedIssuers[0].AuthenticationMethodsReferences = "claims.amr +"

	provider, err = oidc.NewOpenIDConnectProvider(config, nil, nil)
	assert.Nil(t, provider)
	assert.ErrorContains(t, err, "error occurred compiling the authentication methods references expression for the trusted issuer 'https://idp.example.com'")
}

func TestStoreGetIDJAGTrustedIssuer(t *testing.T) {
	store := newIDJAGTestStore(t, nil)

	trusted, err := store.GetIDJAGTrustedIssuer(context.Background(), "https://idp.example.com")
	require.NoError(t, err)
	assert.Equal(t, &oauthelia2.IDJAGTrustedIssuer{
		Issuer:         "https://idp.example.com",
		JSONWebKeysURI: "https://idp.example.com/jwks.json",
		SigningAlgs:    []string{oidc.SigningAlgRSAPSSUsingSHA256, oidc.SigningAlgRSAUsingSHA256},
		MultiTenant:    false,
		Clients:        []string{"app-a"},
	}, trusted)

	trusted, err = store.GetIDJAGTrustedIssuer(context.Background(), "https://unknown.example.com")
	assert.ErrorIs(t, err, oauthelia2.ErrNotFound)
	assert.Nil(t, trusted)

	trusted, err = oidc.NewStore(newIDJAGTestConfiguration(), nil, nil).GetIDJAGTrustedIssuer(context.Background(), "https://idp.example.com")
	assert.ErrorIs(t, err, oauthelia2.ErrNotFound)
	assert.Nil(t, trusted)
}

func TestStoreIsIDJAGTenantTrusted(t *testing.T) {
	store := newIDJAGTestStore(t, nil)

	trusted, err := store.IsIDJAGTenantTrusted(context.Background(), map[string]any{oidc.ClaimIssuer: "https://idp.example.com", oidc.ClaimTenant: "abc"})
	assert.NoError(t, err)
	assert.False(t, trusted)
}

func TestStoreIsIDJAGResourceGoverned(t *testing.T) {
	store := newIDJAGTestStore(t, nil)

	testCases := []struct {
		name     string
		issuer   any
		resource string
		expected bool
	}{
		{"ShouldMatchResource", "https://idp.example.com", "https://api.example.com", true},
		{"ShouldNotMatchPrefix", "https://idp.example.com", "https://api.example.com/", false},
		{"ShouldNotMatchOtherResource", "https://idp.example.com", "https://other.example.com", false},
		{"ShouldNotMatchIssuerWithoutResources", "https://idp2.example.com", "https://api.example.com", false},
		{"ShouldNotMatchUnknownIssuer", "https://unknown.example.com", "https://api.example.com", false},
		{"ShouldNotMatchNonStringIssuer", 1, "https://api.example.com", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			governed, err := store.IsIDJAGResourceGoverned(context.Background(), map[string]any{oidc.ClaimIssuer: tc.issuer}, tc.resource)

			assert.NoError(t, err)
			assert.Equal(t, tc.expected, governed)
		})
	}
}

func newIDJAGTestRedeemClient(t *testing.T, store *oidc.Store) oidc.Client {
	client, err := store.GetRegisteredClient(context.Background(), "app-a")
	require.NoError(t, err)

	return client
}

func TestStoreResolveIDJAGSubject(t *testing.T) {
	ctrl := gomock.NewController(t)

	provider := mocks.NewMockStorage(ctrl)
	users := mocks.NewMockUserProvider(ctrl)

	store := newIDJAGTestStore(t, provider)

	ctx := &TestContext{Context: context.Background(), Storage: provider, UserProvider: users}

	expected := uuid.MustParse("8d6c7e0c-5f0e-4a4b-9b3f-0f3e9e3c1a2b")

	gomock.InOrder(
		users.EXPECT().GetDetailsExtended("john").Return(&authentication.UserDetailsExtended{UserDetails: &authentication.UserDetails{Username: "john"}}, nil),
		provider.
			EXPECT().
			LoadUserOpaqueIdentifierBySignature(ctx, "openid", "https://a.example.com", "john").
			Return(&model.UserOpaqueIdentifier{Service: "openid", SectorID: "https://a.example.com", Username: "john", Identifier: expected}, nil),
	)

	subject, err := store.ResolveIDJAGSubject(ctx, newIDJAGTestRedeemClient(t, store), map[string]any{
		oidc.ClaimIssuer:            "https://idp.example.com",
		oidc.ClaimSubject:           "external-subject",
		oidc.ClaimPreferredUsername: "john",
	})

	require.NoError(t, err)
	assert.Equal(t, expected.String(), subject)
}

func TestStoreResolveIDJAGSubjectCanonicalUsername(t *testing.T) {
	ctrl := gomock.NewController(t)

	provider := mocks.NewMockStorage(ctrl)
	users := mocks.NewMockUserProvider(ctrl)

	store := newIDJAGTestStore(t, provider)

	ctx := &TestContext{Context: context.Background(), Storage: provider, UserProvider: users}

	expected := uuid.MustParse("8d6c7e0c-5f0e-4a4b-9b3f-0f3e9e3c1a2b")

	gomock.InOrder(
		users.EXPECT().GetDetailsExtended("JOHN").Return(&authentication.UserDetailsExtended{UserDetails: &authentication.UserDetails{Username: "john"}}, nil),
		provider.
			EXPECT().
			LoadUserOpaqueIdentifierBySignature(ctx, "openid", "https://a.example.com", "john").
			Return(&model.UserOpaqueIdentifier{Service: "openid", SectorID: "https://a.example.com", Username: "john", Identifier: expected}, nil),
	)

	subject, err := store.ResolveIDJAGSubject(ctx, newIDJAGTestRedeemClient(t, store), map[string]any{
		oidc.ClaimIssuer:            "https://idp.example.com",
		oidc.ClaimPreferredUsername: "JOHN",
	})

	require.NoError(t, err)
	assert.Equal(t, expected.String(), subject)
}

func TestStoreIDJAGUserDetailsSingleLookupPerRequest(t *testing.T) {
	ctrl := gomock.NewController(t)

	provider := mocks.NewMockStorage(ctrl)
	users := mocks.NewMockUserProvider(ctrl)

	store := newIDJAGTestStore(t, provider)

	ctx := &idjagUserValueContext{TestContext: &TestContext{Context: context.Background(), Storage: provider, UserProvider: users}, values: map[any]any{}}

	expected := uuid.MustParse("8d6c7e0c-5f0e-4a4b-9b3f-0f3e9e3c1a2b")

	users.EXPECT().GetDetailsExtended("JOHN").Times(1).Return(&authentication.UserDetailsExtended{UserDetails: &authentication.UserDetails{Username: "john", Emails: []string{"john@example.com"}}}, nil)

	provider.
		EXPECT().
		LoadUserOpaqueIdentifierBySignature(ctx, "openid", "https://a.example.com", "john").
		Return(&model.UserOpaqueIdentifier{Service: "openid", SectorID: "https://a.example.com", Username: "john", Identifier: expected}, nil)

	provider.
		EXPECT().
		LoadUserOpaqueIdentifier(ctx, expected).
		Return(&model.UserOpaqueIdentifier{Service: "openid", SectorID: "https://a.example.com", Username: "john", Identifier: expected}, nil)

	subject, err := store.ResolveIDJAGSubject(ctx, newIDJAGTestRedeemClient(t, store), map[string]any{
		oidc.ClaimIssuer:            "https://idp.example.com",
		oidc.ClaimPreferredUsername: "JOHN",
	})

	require.NoError(t, err)
	assert.Equal(t, expected.String(), subject)

	detailer, err := oidc.UserDetailerFromSubject(ctx, expected)

	require.NoError(t, err)
	assert.Equal(t, "john", detailer.GetUsername())

	claims, err := store.GetIDJAGClaims(ctx, newIDJAGTestRequest(t, store, "app-a", "john"), &oauthelia2.IDJAGRelationship{Issuer: "https://ras.example.com"})

	require.NoError(t, err)
	assert.Equal(t, "john@example.com", claims[oidc.ClaimEmail])
}

func TestStoreResolveIDJAGSubjectDetailsMissingUserDetails(t *testing.T) {
	ctrl := gomock.NewController(t)

	provider := mocks.NewMockStorage(ctrl)
	users := mocks.NewMockUserProvider(ctrl)

	store := newIDJAGTestStore(t, provider)

	ctx := &TestContext{Context: context.Background(), Storage: provider, UserProvider: users}

	users.EXPECT().GetDetailsExtended("john").Return(&authentication.UserDetailsExtended{}, nil)

	subject, err := store.ResolveIDJAGSubject(ctx, newIDJAGTestRedeemClient(t, store), map[string]any{
		oidc.ClaimIssuer:            "https://idp.example.com",
		oidc.ClaimPreferredUsername: "john",
	})

	assert.ErrorIs(t, err, oauthelia2.ErrServerError)
	assert.Empty(t, subject)
}

type idjagUserValueContext struct {
	*TestContext

	values map[any]any
}

func (c *idjagUserValueContext) Value(key any) any {
	if key == model.CtxKeyAutheliaCtx {
		return c
	}

	return c.TestContext.Value(key)
}

func (c *idjagUserValueContext) UserValue(key any) any {
	return c.values[key]
}

func (c *idjagUserValueContext) SetUserValue(key, value any) {
	c.values[key] = value
}

func TestStoreResolveIDJAGSubjectDetailsMissingUsername(t *testing.T) {
	ctrl := gomock.NewController(t)

	provider := mocks.NewMockStorage(ctrl)
	users := mocks.NewMockUserProvider(ctrl)

	store := newIDJAGTestStore(t, provider)

	ctx := &TestContext{Context: context.Background(), Storage: provider, UserProvider: users}

	users.EXPECT().GetDetailsExtended("john").Return(&authentication.UserDetailsExtended{UserDetails: &authentication.UserDetails{}}, nil)

	subject, err := store.ResolveIDJAGSubject(ctx, newIDJAGTestRedeemClient(t, store), map[string]any{
		oidc.ClaimIssuer:            "https://idp.example.com",
		oidc.ClaimPreferredUsername: "john",
	})

	assert.ErrorIs(t, err, oauthelia2.ErrServerError)
	assert.Empty(t, subject)
}

func TestStoreResolveIDJAGSubjectClaimInvalid(t *testing.T) {
	ctrl := gomock.NewController(t)

	users := mocks.NewMockUserProvider(ctrl)

	store := newIDJAGTestStore(t, nil)

	ctx := &TestContext{Context: context.Background(), UserProvider: users}

	testCases := []struct {
		name   string
		claims map[string]any
	}{
		{"ShouldRejectMissing", map[string]any{oidc.ClaimIssuer: "https://idp.example.com", oidc.ClaimSubject: "john"}},
		{"ShouldRejectArray", map[string]any{oidc.ClaimIssuer: "https://idp.example.com", oidc.ClaimPreferredUsername: []any{"john"}}},
		{"ShouldRejectEmpty", map[string]any{oidc.ClaimIssuer: "https://idp.example.com", oidc.ClaimPreferredUsername: ""}},
		{"ShouldRejectUnknownIssuer", map[string]any{oidc.ClaimIssuer: "https://unknown.example.com", oidc.ClaimPreferredUsername: "john"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			subject, err := store.ResolveIDJAGSubject(ctx, newIDJAGTestRedeemClient(t, store), tc.claims)

			assert.ErrorIs(t, err, oauthelia2.ErrNotFound)
			assert.Equal(t, "", subject)
		})
	}
}

func TestStoreResolveIDJAGSubjectUserMissing(t *testing.T) {
	ctrl := gomock.NewController(t)

	users := mocks.NewMockUserProvider(ctrl)

	store := newIDJAGTestStore(t, nil)

	ctx := &TestContext{Context: context.Background(), UserProvider: users}

	users.EXPECT().GetDetailsExtended("john").Return(nil, authentication.ErrUserNotFound)

	subject, err := store.ResolveIDJAGSubject(ctx, newIDJAGTestRedeemClient(t, store), map[string]any{oidc.ClaimIssuer: "https://idp.example.com", oidc.ClaimPreferredUsername: "john"})

	assert.ErrorIs(t, err, oauthelia2.ErrNotFound)
	assert.Equal(t, "", subject)
}

func TestStoreResolveIDJAGSubjectUserDisabled(t *testing.T) {
	ctrl := gomock.NewController(t)

	users := mocks.NewMockUserProvider(ctrl)

	store := newIDJAGTestStore(t, nil)

	ctx := &TestContext{Context: context.Background(), UserProvider: users}

	users.EXPECT().GetDetailsExtended("john").Return(nil, fmt.Errorf("%w : user is disabled", authentication.ErrUserNotFound))

	subject, err := store.ResolveIDJAGSubject(ctx, newIDJAGTestRedeemClient(t, store), map[string]any{oidc.ClaimIssuer: "https://idp.example.com", oidc.ClaimPreferredUsername: "john"})

	assert.ErrorIs(t, err, oauthelia2.ErrNotFound)
	assert.Equal(t, "", subject)
}

func TestStoreResolveIDJAGSubjectUserLookupError(t *testing.T) {
	ctrl := gomock.NewController(t)

	users := mocks.NewMockUserProvider(ctrl)

	store := newIDJAGTestStore(t, nil)

	ctx := &TestContext{Context: context.Background(), UserProvider: users}

	lookup := errors.New("ldap connection refused")

	users.EXPECT().GetDetailsExtended("john").Return(nil, lookup)

	subject, err := store.ResolveIDJAGSubject(ctx, newIDJAGTestRedeemClient(t, store), map[string]any{oidc.ClaimIssuer: "https://idp.example.com", oidc.ClaimPreferredUsername: "john"})

	assert.ErrorIs(t, err, lookup)
	assert.NotErrorIs(t, err, oauthelia2.ErrNotFound)
	assert.NotErrorIs(t, err, oauthelia2.ErrAccessDenied)
	assert.Equal(t, "", subject)
}

func TestStoreResolveIDJAGSubjectServerError(t *testing.T) {
	store := newIDJAGTestStore(t, nil)

	claims := map[string]any{oidc.ClaimIssuer: "https://idp.example.com", oidc.ClaimPreferredUsername: "john"}

	subject, err := store.ResolveIDJAGSubject(context.Background(), newIDJAGTestRedeemClient(t, store), claims)
	assert.ErrorIs(t, err, oauthelia2.ErrServerError)
	assert.Equal(t, "", subject)

	ctx := &TestContext{Context: context.Background(), UserProvider: mocks.NewMockUserProvider(gomock.NewController(t))}

	subject, err = store.ResolveIDJAGSubject(ctx, &oauthelia2.DefaultClient{ID: "app-a"}, claims)
	assert.ErrorIs(t, err, oauthelia2.ErrServerError)
	assert.Equal(t, "", subject)
}

func TestStoreIDJAGUsed(t *testing.T) {
	ctrl := gomock.NewController(t)

	provider := mocks.NewMockStorage(ctrl)

	store := newIDJAGTestStore(t, provider)

	ctx := context.Background()

	signature := model.NewOAuth2BlacklistedJTISignature("abc")

	t.Run("ShouldReturnFalseWhenAbsent", func(t *testing.T) {
		provider.EXPECT().LoadOAuth2BlacklistedIssuerJTI(ctx, "https://idp.example.com", signature).Return(nil, fmt.Errorf("error selecting: %w", sql.ErrNoRows))

		used, err := store.IsIDJAGUsed(ctx, "https://idp.example.com", "abc")
		assert.NoError(t, err)
		assert.False(t, used)
	})

	t.Run("ShouldReturnTrueWhenPresent", func(t *testing.T) {
		provider.EXPECT().LoadOAuth2BlacklistedIssuerJTI(ctx, "https://idp.example.com", signature).Return(&model.OAuth2BlacklistedIssuerJTI{Issuer: "https://idp.example.com", Signature: signature, ExpiresAt: time.Now().Add(time.Minute)}, nil)

		used, err := store.IsIDJAGUsed(ctx, "https://idp.example.com", "abc")
		assert.NoError(t, err)
		assert.True(t, used)
	})

	t.Run("ShouldReturnFalseWhenExpired", func(t *testing.T) {
		provider.EXPECT().LoadOAuth2BlacklistedIssuerJTI(ctx, "https://idp.example.com", signature).Return(&model.OAuth2BlacklistedIssuerJTI{Issuer: "https://idp.example.com", Signature: signature, ExpiresAt: time.Now().Add(-time.Minute)}, nil)

		used, err := store.IsIDJAGUsed(ctx, "https://idp.example.com", "abc")
		assert.NoError(t, err)
		assert.False(t, used)
	})

	t.Run("ShouldReturnError", func(t *testing.T) {
		provider.EXPECT().LoadOAuth2BlacklistedIssuerJTI(ctx, "https://idp.example.com", signature).Return(nil, errors.New("bad conn"))

		used, err := store.IsIDJAGUsed(ctx, "https://idp.example.com", "abc")
		assert.EqualError(t, err, "bad conn")
		assert.False(t, used)
	})

	t.Run("ShouldMarkUsed", func(t *testing.T) {
		exp := time.Unix(10000, 0).UTC()

		provider.EXPECT().SaveOAuth2BlacklistedIssuerJTI(ctx, model.NewOAuth2BlacklistedIssuerJTI("https://idp.example.com", "abc", exp)).Return(nil)

		assert.NoError(t, store.MarkIDJAGUsed(ctx, "https://idp.example.com", "abc", exp))
	})
}

func TestStoreMarkIDJAGUsedDuplicate(t *testing.T) {
	ctrl := gomock.NewController(t)

	provider := mocks.NewMockStorage(ctrl)

	store := newIDJAGTestStore(t, provider)

	ctx := context.Background()

	exp := time.Unix(10000, 0).UTC()

	provider.EXPECT().SaveOAuth2BlacklistedIssuerJTI(ctx, model.NewOAuth2BlacklistedIssuerJTI("https://idp.example.com", "abc", exp)).Return(fmt.Errorf("error inserting: %w", storage.ErrNoRowsAffected))

	err := store.MarkIDJAGUsed(ctx, "https://idp.example.com", "abc", exp)
	assert.True(t, errors.Is(err, oauthelia2.ErrJTIKnown))

	provider.EXPECT().SaveOAuth2BlacklistedIssuerJTI(ctx, model.NewOAuth2BlacklistedIssuerJTI("https://idp.example.com", "abc", exp)).Return(errors.New("bad conn"))

	err = store.MarkIDJAGUsed(ctx, "https://idp.example.com", "abc", exp)
	assert.EqualError(t, err, "bad conn")
	assert.False(t, errors.Is(err, oauthelia2.ErrJTIKnown))
}
