package schema

// AuthenticationBackendOpenIDConnect represents the OpenID Connect 1.0 Relying Party configuration.
type AuthenticationBackendOpenIDConnect struct {
	Providers []AuthenticationBackendOpenIDConnectProvider `koanf:"providers" yaml:"providers,omitempty" toml:"providers,omitempty" json:"providers,omitempty" jsonschema:"title=Providers" jsonschema_description:"The list of external OpenID Connect 1.0 Providers users may authenticate with."`
}

// AuthenticationBackendOpenIDConnectProvider represents a single external OpenID Connect 1.0 Provider.
type AuthenticationBackendOpenIDConnectProvider struct {
	ID     string `koanf:"id" yaml:"id,omitempty" toml:"id,omitempty" json:"id,omitempty" jsonschema:"required,title=ID" jsonschema_description:"The unique identifier for this provider which appears in URLs and stored links."`
	Name   string `koanf:"name" yaml:"name,omitempty" toml:"name,omitempty" json:"name,omitempty" jsonschema:"required,title=Name" jsonschema_description:"The display name for this provider shown on the login page."`
	Issuer string `koanf:"issuer" yaml:"issuer,omitempty" toml:"issuer,omitempty" json:"issuer,omitempty" jsonschema:"required,format=uri,title=Issuer" jsonschema_description:"The issuer identifier of the external provider."`

	ClientID     string `koanf:"client_id" yaml:"client_id,omitempty" toml:"client_id,omitempty" json:"client_id,omitempty" jsonschema:"required,title=Client ID" jsonschema_description:"The client identifier issued to Authelia by the external provider."`
	ClientSecret string `koanf:"client_secret" yaml:"client_secret,omitempty" toml:"client_secret,omitempty" json:"client_secret,omitempty" jsonschema:"title=Client Secret" jsonschema_description:"The client secret issued to Authelia by the external provider."`

	Scopes []string `koanf:"scopes" yaml:"scopes,omitempty" toml:"scopes,omitempty" json:"scopes,omitempty" jsonschema:"default=openid,default=profile,default=email,title=Scopes" jsonschema_description:"The scopes requested from the external provider."`

	TokenEndpointAuthMethod  string `koanf:"token_endpoint_auth_method" yaml:"token_endpoint_auth_method,omitempty" toml:"token_endpoint_auth_method,omitempty" json:"token_endpoint_auth_method,omitempty" jsonschema:"default=client_secret_basic,enum=client_secret_basic,enum=client_secret_post,enum=none,title=Token Endpoint Auth Method" jsonschema_description:"The client authentication method used at the token endpoint."`
	IDTokenSignedResponseAlg string `koanf:"id_token_signed_response_alg" yaml:"id_token_signed_response_alg,omitempty" toml:"id_token_signed_response_alg,omitempty" json:"id_token_signed_response_alg,omitempty" jsonschema:"default=RS256,title=ID Token Signed Response Algorithm" jsonschema_description:"The JWS algorithm the ID Token must be signed with."`

	PKCE                           AuthenticationBackendOpenIDConnectProviderPKCE      `koanf:"pkce" yaml:"pkce,omitempty" toml:"pkce,omitempty" json:"pkce,omitempty" jsonschema:"title=PKCE" jsonschema_description:"Proof Key for Code Exchange configuration."`
	AuthenticationMethodsReference AuthenticationBackendOpenIDConnectProviderAMR       `koanf:"authentication_methods_reference" yaml:"authentication_methods_reference,omitempty" toml:"authentication_methods_reference,omitempty" json:"authentication_methods_reference,omitempty" jsonschema:"title=Authentication Methods Reference" jsonschema_description:"Configures the handling of the amr claim from this provider."`
	Discovery                      AuthenticationBackendOpenIDConnectProviderDiscovery `koanf:"discovery" yaml:"discovery,omitempty" toml:"discovery,omitempty" json:"discovery,omitempty" jsonschema:"title=Discovery" jsonschema_description:"Configures OpenID Connect Discovery 1.0 behavior for this provider."`
	Endpoints                      AuthenticationBackendOpenIDConnectProviderEndpoints `koanf:"endpoints" yaml:"endpoints,omitempty" toml:"endpoints,omitempty" json:"endpoints,omitempty" jsonschema:"title=Endpoints" jsonschema_description:"Explicit endpoints which override or replace discovery."`

	JSONWebKeys []JWK `koanf:"jwks" yaml:"jwks,omitempty" toml:"jwks,omitempty" json:"jwks,omitempty" jsonschema:"title=JSON Web Keys" jsonschema_description:"Inline JSON Web Keys used to verify ID Tokens instead of fetching them."`
}

// AuthenticationBackendOpenIDConnectProviderPKCE represents the PKCE configuration for a provider.
type AuthenticationBackendOpenIDConnectProviderPKCE struct {
	ChallengeMethod string `koanf:"challenge_method" yaml:"challenge_method,omitempty" toml:"challenge_method,omitempty" json:"challenge_method,omitempty" jsonschema:"default=S256,enum=S256,title=Challenge Method" jsonschema_description:"The PKCE code challenge method."`
}

// AuthenticationBackendOpenIDConnectProviderAMR represents the amr claim handling for a provider.
type AuthenticationBackendOpenIDConnectProviderAMR struct {
	Trust bool `koanf:"trust" yaml:"trust" toml:"trust" json:"trust" jsonschema:"default=false,title=Trust" jsonschema_description:"Trusts the amr claim from this provider and merges it into the session authentication method references."`
}

// AuthenticationBackendOpenIDConnectProviderDiscovery represents the discovery configuration for a provider.
type AuthenticationBackendOpenIDConnectProviderDiscovery struct {
	Disable bool `koanf:"disable" yaml:"disable" toml:"disable" json:"disable" jsonschema:"default=false,title=Disable" jsonschema_description:"Disables OpenID Connect Discovery 1.0 requiring all endpoints to be explicitly configured."`
}

// AuthenticationBackendOpenIDConnectProviderEndpoints represents explicit endpoints for a provider.
type AuthenticationBackendOpenIDConnectProviderEndpoints struct {
	Authorization string `koanf:"authorization" yaml:"authorization,omitempty" toml:"authorization,omitempty" json:"authorization,omitempty" jsonschema:"format=uri,title=Authorization Endpoint" jsonschema_description:"The authorization endpoint of the external provider."`
	Token         string `koanf:"token" yaml:"token,omitempty" toml:"token,omitempty" json:"token,omitempty" jsonschema:"format=uri,title=Token Endpoint" jsonschema_description:"The token endpoint of the external provider."`
	JSONWebKeys   string `koanf:"jwks" yaml:"jwks,omitempty" toml:"jwks,omitempty" json:"jwks,omitempty" jsonschema:"format=uri,title=JSON Web Key Set URI" jsonschema_description:"The JSON Web Key Set URI of the external provider."`
}
