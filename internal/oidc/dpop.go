// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"net/http"
	"net/url"

	"github.com/valyala/fasthttp"

	oauthelia2 "authelia.com/provider/oauth2"
	"authelia.com/provider/oauth2/x/errorsx"
)

// NewDPoPResourceRequest builds the [http.Request] that a DPoP proof presented to a protected resource behind Authelia
// must bind to. The proof commits to the target URI and method of the request the client made to the resource, which
// for a proxied request is not the authorization request Authelia itself received.
func NewDPoPResourceRequest(method string, target *url.URL, token, proof string) (r *http.Request) {
	r = &http.Request{
		Method: method,
		URL:    target,
		Host:   target.Host,
		Header: http.Header{},
	}

	r.Header.Set(fasthttp.HeaderXForwardedProto, target.Scheme)
	r.Header.Set(fasthttp.HeaderAuthorization, SchemeDPoP+" "+token)
	r.Header.Set(HeaderDPoP, proof)

	return r
}

// ValidateDPoPResourceAccess performs the RFC9449 Section 7.1 and 7.2 checks for a DPoP bound Access Token presented
// to a protected resource. The proof 'htu' claim may match the request URL or any of the alternates, which are the
// configured URLs of the protected resource.
func ValidateDPoPResourceAccess(ctx context.Context, strategy oauthelia2.DPoPStrategy, r *http.Request, token, jkt string, requireNonce bool, alternates ...string) (err error) {
	if jkt == "" {
		return errorsx.WithStack(oauthelia2.ErrInvalidTokenFormat.WithHint("The access token is not bound to a DPoP proof-of-possession key."))
	}

	resource, ok := strategy.(oauthelia2.DPoPResourceStrategy)
	if !ok {
		return errorsx.WithStack(oauthelia2.ErrInvalidTokenFormat.WithHint("The access token is bound to a DPoP proof-of-possession key but this issuer is not configured to validate DPoP proofs."))
	}

	_, err = resource.ValidateResourceAccess(ctx, r, token, jkt, requireNonce, alternates...)

	return err
}

// NewDPoPNonceChallenge returns a fresh nonce when err is the RFC9449 Section 9 'use_dpop_nonce' error, which the
// caller must supply in the DPoP-Nonce response header alongside the error, and an empty nonce otherwise.
func NewDPoPNonceChallenge(ctx context.Context, strategy oauthelia2.DPoPStrategy, err error) (nonce string, nerr error) {
	if err == nil || strategy == nil || oauthelia2.ErrorToRFC6749Error(err).ErrorField != oauthelia2.ErrUseDPoPNonce.ErrorField {
		return "", nil
	}

	return strategy.NewDPoPNonce(ctx)
}

type dpopDisabledTokenBindingHandler struct{}

func (h *dpopDisabledTokenBindingHandler) BindAccessRequest(ctx context.Context, request oauthelia2.AccessRequester) (err error) {
	session, ok := request.GetSession().(oauthelia2.DPoPBoundSession)
	if !ok || (session.GetDPoPJWKThumbprint() == "" && session.GetRequestedDPoPJWKThumbprint() == "") {
		return nil
	}

	return errorsx.WithStack(oauthelia2.ErrInvalidGrant.
		WithHint("The grant is bound to a DPoP proof-of-possession key which can't be verified as DPoP is disabled.").
		WithDebug("The grant records a DPoP binding but DPoP is disabled, so possession of the bound key can't be proven and the grant must not be redeemed."))
}

func (h *dpopDisabledTokenBindingHandler) PopulateBoundTokenEndpointResponse(ctx context.Context, request oauthelia2.AccessRequester, response oauthelia2.AccessResponder) (err error) {
	return nil
}

var (
	_ oauthelia2.TokenEndpointBindingHandler = (*dpopDisabledTokenBindingHandler)(nil)
)
