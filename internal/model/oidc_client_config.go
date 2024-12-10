package model

import (
	"encoding/json"
	"net/url"
	"time"
)

// OpenIDConnectClient represents an OpenIDConnect Client configuration in a yaml file.
type OpenIDConnectClient struct {
	ID   string `koanf:"client_id" json:"client_id" jsonschema:"required,minLength=1,title=Client ID" jsonschema_description:"The Client ID."`
	Name string `koanf:"client_name" json:"client_name" jsonschema:"title=Client Name" jsonschema_description:"The Client Name displayed to End-Users."`
	// TODO (Crowley723): fix converting *url.URL to string and serializing it causing a json error.
	SectorIdentifierURI *url.URL `koanf:"sector_identifier_uri" json:"sector_identifier_uri" jsonschema:"title=Sector Identifier URI" jsonschema_description:"The Client Sector Identifier URI for Privacy Isolation via Pairwise subject types."`
	Public              bool     `koanf:"public" json:"public" jsonschema:"default=false,title=Public" jsonschema_description:"Enables the Public Client Type."`

	RedirectURIs []string `koanf:"redirect_uris" json:"redirect_uris" jsonschema:"title=Redirect URIs" jsonschema_description:"List of whitelisted redirect URIs."`
	RequestURIs  []string `koanf:"request_uris" json:"request_uris" jsonschema:"title=Request URIs" jsonschema_description:"List of whitelisted request URIs."`

	Audience      []string `koanf:"audience" json:"audience" jsonschema:"uniqueItems,title=Audience" jsonschema_description:"List of authorized audiences."`
	Scopes        []string `koanf:"scopes" json:"scopes" jsonschema:"required,enum=openid,enum=offline_access,enum=groups,enum=email,enum=profile,enum=authelia.bearer.authz,uniqueItems,title=Scopes" jsonschema_description:"The Scopes this client is allowed request and be granted."`
	GrantTypes    []string `koanf:"grant_types" json:"grant_types" jsonschema:"enum=authorization_code,enum=implicit,enum=refresh_token,enum=client_credentials,uniqueItems,title=Grant Types" jsonschema_description:"The Grant Types this client is allowed to use for the protected endpoints."`
	ResponseTypes []string `koanf:"response_types" json:"response_types" jsonschema:"enum=code,enum=id_token token,enum=id_token,enum=token,enum=code token,enum=code id_token,enum=code id_token token,uniqueItems,title=Response Types" jsonschema_description:"The Response Types the client is authorized to request."`
	ResponseModes []string `koanf:"response_modes" json:"response_modes" jsonschema:"enum=form_post,enum=form_post.jwt,enum=query,enum=query.jwt,enum=fragment,enum=fragment.jwt,enum=jwt,uniqueItems,title=Response Modes" jsonschema_description:"The Response Modes this client is authorized request."`

	AuthorizationPolicy          string         `koanf:"authorization_policy" json:"authorization_policy" jsonschema:"title=Authorization Policy" jsonschema_description:"The Authorization Policy to apply to this client."`
	Lifespan                     string         `koanf:"lifespan" json:"lifespan" jsonschema:"title=Lifespan Name" jsonschema_description:"The name of the custom lifespan to utilize for this client."`
	RequestedAudienceMode        string         `koanf:"requested_audience_mode" json:"requested_audience_mode" jsonschema:"enum=explicit,enum=implicit,title=Requested Audience Mode" jsonschema_description:"The Requested Audience Mode used for this client."`
	ConsentMode                  string         `koanf:"consent_mode" json:"consent_mode" jsonschema:"enum=auto,enum=explicit,enum=implicit,enum=pre-configured,title=Consent Mode" jsonschema_description:"The Consent Mode used for this client."`
	ConsentPreConfiguredDuration *time.Duration `koanf:"pre_configured_consent_duration" json:"pre_configured_consent_duration" jsonschema:"default=7 days,title=Pre-Configured Consent Duration" jsonschema_description:"The Pre-Configured Consent Duration when using Consent Mode pre-configured for this client."`

	RequirePushedAuthorizationRequests bool `koanf:"require_pushed_authorization_requests" json:"require_pushed_authorization_requests" jsonschema:"default=false,title=Require Pushed Authorization Requests" jsonschema_description:"Requires Pushed Authorization Requests for this client to perform an authorization."`
	RequirePKCE                        bool `koanf:"require_pkce" json:"require_pkce" jsonschema:"default=false,title=Require PKCE" jsonschema_description:"Requires a Proof Key for this client to perform Code Exchange."`

	PKCEChallengeMethod string `koanf:"pkce_challenge_method" json:"pkce_challenge_method" jsonschema:"enum=plain,enum=S256,title=PKCE Challenge Method" jsonschema_description:"The PKCE Challenge Method enforced on this client."`

	TokenEndpointAuthMethod            string `koanf:"token_endpoint_auth_method" json:"token_endpoint_auth_method" jsonschema:"enum=none,enum=client_secret_post,enum=client_secret_basic,enum=private_key_jwt,enum=client_secret_jwt,title=Token Endpoint Auth Method" jsonschema_description:"The Token Endpoint Auth Method enforced by the provider for this client."`
	AllowMultipleAuthenticationMethods bool   `koanf:"allow_multiple_auth_methods" json:"allow_multiple_auth_methods" jsonschema:"title=Allow Multiple Authentication Methods" jsonschema_description:"Permits this registered client to accept misbehaving clients which use a broad authentication approach. This is not standards complaint, use at your own security risk."`
}

// Allow JSONMarshal for OpenIDConnectClient to OpenIDConnectClientData.
func (c *OpenIDConnectClient) ToData() OpenIDConnectClientData {
	o := OpenIDConnectClientData{}

	if c.ID != "" {
		o.ID = c.ID
	}

	if c.Name != "" {
		o.Name = c.Name
	}

	o.Public = c.Public

	if c.RedirectURIs != nil {
		o.RedirectURIs = c.RedirectURIs
	}

	if c.RequestURIs != nil {
		o.RequestURIs = c.RequestURIs
	}

	if c.Audience != nil {
		o.Audience = c.Audience
	}

	if c.Scopes != nil {
		o.Scopes = c.Scopes
	}

	return o
}

func (c *OpenIDConnectClient) MarshalJSON() (data []byte, err error) {
	return json.Marshal(c.ToData())
}

// OpenIDConnectClientData is the JSON representation for an OpenIDConnect Client.
type OpenIDConnectClientData struct {
	ID   string
	Name string
	// SectorIdentifierURI string.
	Public                             bool
	RedirectURIs                       []string
	RequestURIs                        []string
	Audience                           []string
	Scopes                             []string
	GrantTypes                         []string
	ResponseTypes                      []string
	ResponseModes                      []string
	AuthorizationPolicy                string
	Lifespan                           string
	RequestedAudienceMode              string
	ConsentMode                        string
	RequirePushedAuthorizationRequests bool
	RequirePKCE                        bool
}
