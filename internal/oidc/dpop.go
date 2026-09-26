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

// DPoPResourceStrategy is the subset of the RFC9449 strategy required to validate a request to a protected resource.
// The oauthelia2.DPoPStrategy interface does not describe ValidateResourceAccess as it is only meaningful on the
// resource server side of the specification.
type DPoPResourceStrategy interface {
	oauthelia2.DPoPStrategy

	ValidateResourceAccess(ctx context.Context, r *http.Request, accessToken, boundJKT string, requireNonce bool) (parsed *oauthelia2.DPoPProof, err error)
}

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
// to a protected resource. When the failure is a nonce challenge it returns the fresh nonce the caller must echo in
// the DPoP-Nonce response header alongside the error.
func ValidateDPoPResourceAccess(ctx context.Context, strategy oauthelia2.DPoPStrategy, r *http.Request, token, jkt string, requireNonce bool) (nonce string, err error) {
	if jkt == "" {
		return "", errorsx.WithStack(oauthelia2.ErrInvalidTokenFormat.WithHint("The access token is not bound to a DPoP proof-of-possession key."))
	}

	resource, ok := strategy.(DPoPResourceStrategy)
	if !ok {
		return "", errorsx.WithStack(oauthelia2.ErrInvalidTokenFormat.WithHint("The access token is bound to a DPoP proof-of-possession key but this issuer is not configured to validate DPoP proofs."))
	}

	if _, err = resource.ValidateResourceAccess(ctx, r, token, jkt, requireNonce); err == nil {
		return "", nil
	}

	if oauthelia2.ErrorToRFC6749Error(err).ErrorField != oauthelia2.ErrUseDPoPNonce.ErrorField {
		return "", err
	}

	var nerr error

	if nonce, nerr = resource.NewDPoPNonce(ctx); nerr != nil {
		return "", nerr
	}

	return nonce, err
}
