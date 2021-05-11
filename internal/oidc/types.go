package oidc

import "github.com/ory/fosite/handler/openid"

// ConsentGetResponseBody schema of the response body of the consent GET endpoint.
type ConsentGetResponseBody struct {
	ClientID          string     `json:"client_id"`
	ClientDescription string     `json:"client_description"`
	Scopes            []Scope    `json:"scopes"`
	Audience          []Audience `json:"audience"`
}

// Scope represents the scope information.
type Scope struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Audience represents the audience information.
type Audience struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// WellKnownConfiguration is the OIDC well known config struct.
//
// See https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderMetadata
type WellKnownConfiguration struct {
	Issuer             string   `json:"issuer"`
	AuthURL            string   `json:"authorization_endpoint"`
	TokenURL           string   `json:"token_endpoint"`
	RevocationURL      string   `json:"revocation_endpoint"`
	UserinfoURL        string   `json:"userinfo_endpoint"`
	JWKSURL            string   `json:"jwks_uri"`
	IDTokenAlgorithms  []string `json:"id_token_signing_alg_values_supported"`
	UserinfoAlgorithms []string `json:"userinfo_signing_alg_values_supported"`
	GrantTypes         []string `json:"grant_types_supported"`
	SubjectTypes       []string `json:"subject_types_supported"`
	ResponseTypes      []string `json:"response_types_supported"`
	ResponseModes      []string `json:"response_modes_supported"`
	Scopes             []string `json:"scopes_supported"`
	Claims             []string `json:"claims_supported"`
	//RequireRequestURIRegistration bool     `json:"require_request_uri_registration"`
	RequestURIParameter       bool `json:"request_uri_parameter_supported"`
	RequestParameter          bool `json:"request_parameter_supported"`
	BackChannelLogout         bool `json:"backchannel_logout_supported"`
	BackChannelLogoutSession  bool `json:"backchannel_logout_session_supported"`
	FrontChannelLogout        bool `json:"frontchannel_logout_supported"`
	FrontChannelLogoutSession bool `json:"frontchannel_logout_session_supported"`
}

// OpenIDSession holds OIDC Session information.
type OpenIDSession struct {
	*openid.DefaultSession `json:"idToken"`

	Extra    map[string]interface{} `json:"extra"`
	ClientID string
}
