package oidc

import (
	"context"

	"github.com/ory/fosite"
)

// AuthorizationServerIssuerIdentificationHandler handles RFC9207: OAuth 2.0 Authorization Server Issuer Identification
// response parameters as per the https://datatracker.ietf.org/doc/html/rfc9207 specification document.
type AuthorizationServerIssuerIdentificationHandler struct {
	Config interface {
		AuthorizationServerIssuerIdentificationProvider
	}
}

// HandleAuthorizeEndpointRequest implements the fosite.AuthorizeEndpointHandler for RFC9207.
func (h *AuthorizationServerIssuerIdentificationHandler) HandleAuthorizeEndpointRequest(ctx context.Context, requester fosite.AuthorizeRequester, responder fosite.AuthorizeResponder) (err error) {
	switch requester.GetResponseMode() {
	case ResponseModeJWT, ResponseModeFormPostJWT, ResponseModeQueryJWT, ResponseModeFragmentJWT:
		break
	default:
		if issuer := h.Config.GetAuthorizationServerIdentificationIssuer(ctx); len(issuer) != 0 {
			responder.GetParameters().Set(FormParameterIssuer, issuer)
		}
	}

	return nil
}

var (
	_ fosite.AuthorizeEndpointHandler = (*AuthorizationServerIssuerIdentificationHandler)(nil)
)
