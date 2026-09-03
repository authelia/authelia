// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package schema

// AuthenticationBackendExternalIdentity represents the external identity configuration.
type AuthenticationBackendExternalIdentity struct {
	Providers []AuthenticationBackendExternalIdentityProvider `koanf:"providers" yaml:"providers,omitempty" toml:"providers,omitempty" json:"providers,omitempty" jsonschema:"title=Providers" jsonschema_description:"The list of external identity providers users may authenticate with."`
}

// AuthenticationBackendExternalIdentityProvider represents a single external identity provider.
type AuthenticationBackendExternalIdentityProvider struct {
	ID     string `koanf:"id" yaml:"id,omitempty" toml:"id,omitempty" json:"id,omitempty" jsonschema:"required,title=ID" jsonschema_description:"The unique identifier for this provider which appears in URLs and stored links."`
	Type   string `koanf:"type" yaml:"type,omitempty" toml:"type,omitempty" json:"type,omitempty" jsonschema:"default=openid_connect,enum=openid_connect,enum=discord,enum=plex,enum=github,title=Type" jsonschema_description:"The type of the external identity provider."`
	Name   string `koanf:"name" yaml:"name,omitempty" toml:"name,omitempty" json:"name,omitempty" jsonschema:"required,title=Name" jsonschema_description:"The display name for this provider shown on the login page."`
	Issuer string `koanf:"issuer" yaml:"issuer,omitempty" toml:"issuer,omitempty" json:"issuer,omitempty" jsonschema:"required,format=uri,title=Issuer" jsonschema_description:"The issuer identifier of the external provider."`

	ClientID     string `koanf:"client_id" yaml:"client_id,omitempty" toml:"client_id,omitempty" json:"client_id,omitempty" jsonschema:"required,title=Client ID" jsonschema_description:"The client identifier issued to Authelia by the external provider."`
	ClientSecret string `koanf:"client_secret" yaml:"client_secret,omitempty" toml:"client_secret,omitempty" json:"client_secret,omitempty" jsonschema:"title=Client Secret" jsonschema_description:"The client secret issued to Authelia by the external provider."`

	Scopes []string `koanf:"scopes" yaml:"scopes,omitempty" toml:"scopes,omitempty" json:"scopes,omitempty" jsonschema:"default=openid,default=profile,default=email,title=Scopes" jsonschema_description:"The scopes requested from the external provider."`

	AuthorizationResponseIssParameterSupported bool `koanf:"authorization_response_iss_parameter_supported" yaml:"authorization_response_iss_parameter_supported" toml:"authorization_response_iss_parameter_supported" json:"authorization_response_iss_parameter_supported" jsonschema:"default=false,title=Authorization Response Issuer Parameter Supported" jsonschema_description:"Requires the external provider to send the iss parameter in the authorization response as described by RFC9207. This is also required when the discovery document advertises support for it."`
	RequirePushedAuthorizationRequests         bool `koanf:"require_pushed_authorization_requests" yaml:"require_pushed_authorization_requests" toml:"require_pushed_authorization_requests" json:"require_pushed_authorization_requests" jsonschema:"default=false,title=Require Pushed Authorization Requests" jsonschema_description:"Sends every authorization request to the external provider as a Pushed Authorization Request as described by RFC9126. Pushed Authorization Requests are also used when the discovery document requires them."`

	ResponseMode              string `koanf:"response_mode" yaml:"response_mode,omitempty" toml:"response_mode,omitempty" json:"response_mode,omitempty" jsonschema:"default=query,enum=query,enum=form_post,title=Response Mode" jsonschema_description:"The response mode the external provider is asked to deliver the authorization response with."`
	TokenEndpointAuthMethod   string `koanf:"token_endpoint_auth_method" yaml:"token_endpoint_auth_method,omitempty" toml:"token_endpoint_auth_method,omitempty" json:"token_endpoint_auth_method,omitempty" jsonschema:"default=client_secret_basic,enum=client_secret_basic,enum=client_secret_post,enum=none,title=Token Endpoint Auth Method" jsonschema_description:"The client authentication method used at the token endpoint."`
	IDTokenSignedResponseAlg  string `koanf:"id_token_signed_response_alg" yaml:"id_token_signed_response_alg,omitempty" toml:"id_token_signed_response_alg,omitempty" json:"id_token_signed_response_alg,omitempty" jsonschema:"enum=ES256,enum=ES384,enum=ES512,enum=PS256,enum=PS384,enum=PS512,enum=RS256,enum=RS384,enum=RS512,title=ID Token Signed Response Algorithm" jsonschema_description:"The JWS algorithm the ID Token must be signed with. When not configured it is selected from the algorithms the discovery document advertises, preferring RS256, and is RS256 when discovery is disabled."`
	UserInfoSignedResponseAlg string `koanf:"userinfo_signed_response_alg" yaml:"userinfo_signed_response_alg,omitempty" toml:"userinfo_signed_response_alg,omitempty" json:"userinfo_signed_response_alg,omitempty" jsonschema:"enum=ES256,enum=ES384,enum=ES512,enum=PS256,enum=PS384,enum=PS512,enum=RS256,enum=RS384,enum=RS512,title=UserInfo Signed Response Algorithm" jsonschema_description:"The JWS algorithm the UserInfo Response must be signed with. When not configured the UserInfo Response must not be signed."`

	PKCE                           AuthenticationBackendExternalIdentityProviderPKCE      `koanf:"pkce" yaml:"pkce,omitempty" toml:"pkce,omitempty" json:"pkce,omitempty" jsonschema:"title=PKCE" jsonschema_description:"Proof Key for Code Exchange configuration."`
	AuthenticationMethodsReference AuthenticationBackendExternalIdentityProviderAMR       `koanf:"authentication_methods_reference" yaml:"authentication_methods_reference,omitempty" toml:"authentication_methods_reference,omitempty" json:"authentication_methods_reference,omitempty" jsonschema:"title=Authentication Methods Reference" jsonschema_description:"Configures the handling of the amr claim from this provider."`
	Discovery                      AuthenticationBackendExternalIdentityProviderDiscovery `koanf:"discovery" yaml:"discovery,omitempty" toml:"discovery,omitempty" json:"discovery,omitempty" jsonschema:"title=Discovery" jsonschema_description:"Configures OpenID Connect Discovery 1.0 behavior for this provider."`
	Endpoints                      AuthenticationBackendExternalIdentityProviderEndpoints `koanf:"endpoints" yaml:"endpoints,omitempty" toml:"endpoints,omitempty" json:"endpoints,omitempty" jsonschema:"title=Endpoints" jsonschema_description:"Explicit endpoints which override or replace discovery."`

	JSONWebKeys []JWK `koanf:"jwks" yaml:"jwks,omitempty" toml:"jwks,omitempty" json:"jwks,omitempty" jsonschema:"title=JSON Web Keys" jsonschema_description:"Inline JSON Web Keys used to verify ID Tokens instead of fetching them."`
}

// AuthenticationBackendExternalIdentityProviderPKCE represents the PKCE configuration for a provider.
type AuthenticationBackendExternalIdentityProviderPKCE struct {
	ChallengeMethod string `koanf:"challenge_method" yaml:"challenge_method,omitempty" toml:"challenge_method,omitempty" json:"challenge_method,omitempty" jsonschema:"default=S256,enum=S256,title=Challenge Method" jsonschema_description:"The PKCE code challenge method."`
}

// AuthenticationBackendExternalIdentityProviderAMR represents the amr claim handling for a provider.
type AuthenticationBackendExternalIdentityProviderAMR struct {
	Trust    bool     `koanf:"trust" yaml:"trust" toml:"trust" json:"trust" jsonschema:"default=false,title=Trust" jsonschema_description:"Trusts the amr claim from this provider and merges it into the session authentication method references."`
	Default  []string `koanf:"default" yaml:"default,omitempty" toml:"default,omitempty" json:"default,omitempty" jsonschema:"title=Default" jsonschema_description:"The amr values merged into the session authentication method references when this provider does not assert any, or in place of the values it asserts when override is enabled."`
	Override bool     `koanf:"override" yaml:"override" toml:"override" json:"override" jsonschema:"default=false,title=Override" jsonschema_description:"Merges the default amr values into the session authentication method references in place of any values this provider asserts."`
}

// AuthenticationBackendExternalIdentityProviderDiscovery represents the discovery configuration for a provider.
type AuthenticationBackendExternalIdentityProviderDiscovery struct {
	Disable bool `koanf:"disable" yaml:"disable" toml:"disable" json:"disable" jsonschema:"default=false,title=Disable" jsonschema_description:"Disables OpenID Connect Discovery 1.0 requiring all endpoints to be explicitly configured."`
}

// AuthenticationBackendExternalIdentityProviderEndpoints represents explicit endpoints for a provider.
type AuthenticationBackendExternalIdentityProviderEndpoints struct {
	Authorization string `koanf:"authorization" yaml:"authorization,omitempty" toml:"authorization,omitempty" json:"authorization,omitempty" jsonschema:"format=uri,title=Authorization Endpoint" jsonschema_description:"The authorization endpoint of the external provider."`
	Token         string `koanf:"token" yaml:"token,omitempty" toml:"token,omitempty" json:"token,omitempty" jsonschema:"format=uri,title=Token Endpoint" jsonschema_description:"The token endpoint of the external provider."`
	UserInfo      string `koanf:"userinfo" yaml:"userinfo,omitempty" toml:"userinfo,omitempty" json:"userinfo,omitempty" jsonschema:"format=uri,title=UserInfo Endpoint" jsonschema_description:"The UserInfo endpoint of the external provider."`
	JSONWebKeys   string `koanf:"jwks" yaml:"jwks,omitempty" toml:"jwks,omitempty" json:"jwks,omitempty" jsonschema:"format=uri,title=JSON Web Key Set URI" jsonschema_description:"The JSON Web Key Set URI of the external provider."`

	PushedAuthorizationRequest string `koanf:"pushed_authorization_request" yaml:"pushed_authorization_request,omitempty" toml:"pushed_authorization_request,omitempty" json:"pushed_authorization_request,omitempty" jsonschema:"format=uri,title=Pushed Authorization Request Endpoint" jsonschema_description:"The Pushed Authorization Request endpoint of the external provider."`
}
