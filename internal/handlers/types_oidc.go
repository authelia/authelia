package handlers

// ConsentPostRequestBody schema of the request body of the consent POST endpoint.
type ConsentPostRequestBody struct {
	ClientID       string `json:"client_id"`
	AcceptOrReject string `json:"accept_or_reject"`
}

// ConsentPostResponseBody schema of the response body of the consent POST endpoint.
type ConsentPostResponseBody struct {
	RedirectURI string `json:"redirect_uri"`
}

// WellKnownConfigurationJSON is the OIDC well known config struct.
type WellKnownConfigurationJSON struct {
	Issuer                             string   `json:"issuer"`
	AuthURL                            string   `json:"authorization_endpoint"`
	TokenURL                           string   `json:"token_endpoint"`
	RevocationEndpoint                 string   `json:"revocation_endpoint"`
	JWKSURL                            string   `json:"jwks_uri"`
	Algorithms                         []string `json:"id_token_signing_alg_values_supported"`
	ResponseTypesSupported             []string `json:"response_types_supported"`
	ScopesSupported                    []string `json:"scopes_supported"`
	ClaimsSupported                    []string `json:"claims_supported"`
	BackChannelLogoutSupported         bool     `json:"backchannel_logout_supported"`
	BackChannelLogoutSessionSupported  bool     `json:"backchannel_logout_session_supported"`
	FrontChannelLogoutSupported        bool     `json:"frontchannel_logout_supported"`
	FrontChannelLogoutSessionSupported bool     `json:"frontchannel_logout_session_supported"`
}
