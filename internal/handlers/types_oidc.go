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
