package oidc

import (
	"fmt"
	"net/http"

	"github.com/ory/fosite/compose"
	"github.com/ory/herodot"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// NewOpenIDConnectProvider new-ups a OpenIDConnectProvider.
func NewOpenIDConnectProvider(configuration *schema.OpenIDConnectConfiguration) (provider OpenIDConnectProvider, err error) {
	provider = OpenIDConnectProvider{
		Fosite: nil,
	}

	if configuration == nil {
		return provider, nil
	}

	provider.Store = NewOpenIDConnectStore(configuration)

	composeConfiguration := &compose.Config{
		AccessTokenLifespan:            configuration.AccessTokenLifespan,
		AuthorizeCodeLifespan:          configuration.AuthorizeCodeLifespan,
		IDTokenLifespan:                configuration.IDTokenLifespan,
		RefreshTokenLifespan:           configuration.RefreshTokenLifespan,
		SendDebugMessagesToClients:     configuration.EnableClientDebugMessages,
		MinParameterEntropy:            configuration.MinimumParameterEntropy,
		EnforcePKCE:                    configuration.EnforcePKCE == "always",
		EnforcePKCEForPublicClients:    configuration.EnforcePKCE != "never",
		EnablePKCEPlainChallengeMethod: configuration.EnablePKCEPlainChallenge,
	}

	keyManager, err := NewKeyManagerWithConfiguration(configuration)
	if err != nil {
		return provider, err
	}

	provider.KeyManager = keyManager

	key, err := provider.KeyManager.GetActivePrivateKey()
	if err != nil {
		return provider, err
	}

	strategy := &compose.CommonStrategy{
		CoreStrategy: compose.NewOAuth2HMACStrategy(
			composeConfiguration,
			[]byte(utils.HashSHA256FromString(configuration.HMACSecret)),
			nil,
		),
		OpenIDConnectTokenStrategy: compose.NewOpenIDConnectStrategy(
			composeConfiguration,
			key,
		),
		JWTStrategy: provider.KeyManager.Strategy(),
	}

	provider.Fosite = compose.Compose(
		composeConfiguration,
		provider.Store,
		strategy,
		AutheliaHasher{},

		/*
			These are the OAuth2 and OpenIDConnect factories. Order is important (the OAuth2 factories at the top must
			be before the OpenIDConnect factories) and taken directly from fosite.compose.ComposeAllEnabled. The
			commented factories are not enabled as we don't yet use them but are still here for reference purposes.
		*/
		compose.OAuth2AuthorizeExplicitFactory,
		compose.OAuth2AuthorizeImplicitFactory,
		compose.OAuth2ClientCredentialsGrantFactory,
		compose.OAuth2RefreshTokenGrantFactory,
		compose.OAuth2ResourceOwnerPasswordCredentialsFactory,
		// compose.RFC7523AssertionGrantFactory,.

		compose.OpenIDConnectExplicitFactory,
		compose.OpenIDConnectImplicitFactory,
		compose.OpenIDConnectHybridFactory,
		compose.OpenIDConnectRefreshFactory,

		compose.OAuth2TokenIntrospectionFactory,
		compose.OAuth2TokenRevocationFactory,

		compose.OAuth2PKCEFactory,
	)

	provider.discovery = OpenIDConnectWellKnownConfiguration{
		CommonDiscoveryOptions: CommonDiscoveryOptions{
			SubjectTypesSupported: []string{
				"public",
			},
			ResponseTypesSupported: []string{
				"code",
				"token",
				"id_token",
				"code token",
				"code id_token",
				"token id_token",
				"code token id_token",
				"none",
			},
			ResponseModesSupported: []string{
				"form_post",
				"query",
				"fragment",
			},
			ScopesSupported: []string{
				ScopeOfflineAccess,
				ScopeOpenID,
				ScopeProfile,
				ScopeGroups,
				ScopeEmail,
			},
			ClaimsSupported: []string{
				"aud",
				"exp",
				"iat",
				"iss",
				"jti",
				"rat",
				"sub",
				"auth_time",
				"nonce",
				ClaimEmail,
				ClaimEmailVerified,
				ClaimEmailAlts,
				ClaimGroups,
				ClaimPreferredUsername,
				ClaimDisplayName,
			},
		},
		OAuth2DiscoveryOptions: OAuth2DiscoveryOptions{
			CodeChallengeMethodsSupported: []string{
				"S256",
			},
		},
		OpenIDConnectDiscoveryOptions: OpenIDConnectDiscoveryOptions{
			IDTokenSigningAlgValuesSupported: []string{
				"RS256",
			},
			UserinfoSigningAlgValuesSupported: []string{
				"none",
				"RS256",
			},
			RequestObjectSigningAlgValuesSupported: []string{
				"none",
				"RS256",
			},
		},
	}

	if configuration.EnablePKCEPlainChallenge {
		provider.discovery.CodeChallengeMethodsSupported = append(provider.discovery.CodeChallengeMethodsSupported, "plain")
	}

	provider.herodot = herodot.NewJSONWriter(nil)

	return provider, nil
}

// Write writes data with herodot.JSONWriter.
func (p OpenIDConnectProvider) Write(w http.ResponseWriter, r *http.Request, e interface{}, opts ...herodot.EncoderOptions) {
	p.herodot.Write(w, r, e, opts...)
}

// WriteError writes an error with herodot.JSONWriter.
func (p OpenIDConnectProvider) WriteError(w http.ResponseWriter, r *http.Request, err error, opts ...herodot.Option) {
	p.herodot.WriteError(w, r, err, opts...)
}

// WriteErrorCode writes an error with an error code with herodot.JSONWriter.
func (p OpenIDConnectProvider) WriteErrorCode(w http.ResponseWriter, r *http.Request, code int, err error, opts ...herodot.Option) {
	p.herodot.WriteErrorCode(w, r, code, err, opts...)
}

// GetOAuth2WellKnownConfiguration returns the discovery document for the OAuth Configuration.
func (p OpenIDConnectProvider) GetOAuth2WellKnownConfiguration(issuer string) OAuth2WellKnownConfiguration {
	options := OAuth2WellKnownConfiguration{
		CommonDiscoveryOptions: p.discovery.CommonDiscoveryOptions,
		OAuth2DiscoveryOptions: p.discovery.OAuth2DiscoveryOptions,
	}

	options.Issuer = issuer
	options.JWKSURI = fmt.Sprintf("%s%s", issuer, JWKsPath)

	options.IntrospectionEndpoint = fmt.Sprintf("%s%s", issuer, IntrospectionPath)
	options.TokenEndpoint = fmt.Sprintf("%s%s", issuer, TokenPath)

	options.AuthorizationEndpoint = fmt.Sprintf("%s%s", issuer, AuthorizationPath)
	options.RevocationEndpoint = fmt.Sprintf("%s%s", issuer, RevocationPath)

	return options
}

// GetOpenIDConnectWellKnownConfiguration returns the discovery document for the OpenID Configuration.
func (p OpenIDConnectProvider) GetOpenIDConnectWellKnownConfiguration(issuer string) OpenIDConnectWellKnownConfiguration {
	options := OpenIDConnectWellKnownConfiguration{
		CommonDiscoveryOptions:                          p.discovery.CommonDiscoveryOptions,
		OAuth2DiscoveryOptions:                          p.discovery.OAuth2DiscoveryOptions,
		OpenIDConnectDiscoveryOptions:                   p.discovery.OpenIDConnectDiscoveryOptions,
		OpenIDConnectFrontChannelLogoutDiscoveryOptions: p.discovery.OpenIDConnectFrontChannelLogoutDiscoveryOptions,
		OpenIDConnectBackChannelLogoutDiscoveryOptions:  p.discovery.OpenIDConnectBackChannelLogoutDiscoveryOptions,
	}

	options.Issuer = issuer
	options.JWKSURI = fmt.Sprintf("%s%s", issuer, JWKsPath)

	options.IntrospectionEndpoint = fmt.Sprintf("%s%s", issuer, IntrospectionPath)
	options.TokenEndpoint = fmt.Sprintf("%s%s", issuer, TokenPath)

	options.AuthorizationEndpoint = fmt.Sprintf("%s%s", issuer, AuthorizationPath)
	options.RevocationEndpoint = fmt.Sprintf("%s%s", issuer, RevocationPath)
	options.UserinfoEndpoint = fmt.Sprintf("%s%s", issuer, UserinfoPath)

	return options
}
