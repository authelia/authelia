// Copyright © 2023 Ory Corp.
// SPDX-License-Identifier: Apache-2.0.

package oidc

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"regexp"

	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/oauth2"
	"github.com/ory/fosite/handler/pkce"
	"github.com/ory/x/errorsx"
	"github.com/pkg/errors"
)

var _ fosite.TokenEndpointHandler = (*PKCEHandler)(nil)

// PKCEHandler is a fork of pkce.Handler with modifications to rectify bugs. It implements the
// fosite.TokenEndpointHandler.
type PKCEHandler struct {
	AuthorizeCodeStrategy oauth2.AuthorizeCodeStrategy
	Storage               pkce.PKCERequestStorage
	Config                interface {
		fosite.EnforcePKCEProvider
		fosite.EnforcePKCEForPublicClientsProvider
		fosite.EnablePKCEPlainChallengeMethodProvider
	}
}

var (
	rePKCEVerifier = regexp.MustCompile(`[^\w.\-~]`)
)

// HandleAuthorizeEndpointRequest implements fosite.TokenEndpointHandler partially.
func (c *PKCEHandler) HandleAuthorizeEndpointRequest(ctx context.Context, requester fosite.AuthorizeRequester, responder fosite.AuthorizeResponder) error {
	// This let's us define multiple response types, for example open id connect's id_token.
	if !requester.GetResponseTypes().Has(FormParameterAuthorizationCode) {
		return nil
	}

	challenge := requester.GetRequestForm().Get(FormParameterCodeChallenge)
	method := requester.GetRequestForm().Get(FormParameterCodeChallengeMethod)
	client := requester.GetClient()

	if err := c.validate(ctx, challenge, method, client); err != nil {
		return err
	}

	// We don't need a session if it's not enforced and the PKCE parameters are not provided by the client.
	if challenge == "" && method == "" {
		return nil
	}

	code := responder.GetCode()
	if len(code) == 0 {
		return errorsx.WithStack(fosite.ErrServerError.WithDebug("The PKCE handler must be loaded after the authorize code handler."))
	}

	signature := c.AuthorizeCodeStrategy.AuthorizeCodeSignature(ctx, code)
	if err := c.Storage.CreatePKCERequestSession(ctx, signature, requester.Sanitize([]string{
		FormParameterCodeChallenge,
		FormParameterCodeChallengeMethod,
	})); err != nil {
		return errorsx.WithStack(fosite.ErrServerError.WithWrap(err).WithDebugf("The recorded error is: %s.", err.Error()))
	}

	return nil
}

func (c *PKCEHandler) validate(ctx context.Context, challenge, method string, client fosite.Client) error {
	if len(challenge) == 0 {
		// If the server requires Proof Key for Code Exchange (PKCE) by OAuth
		// clients and the client does not send the "code_challenge" in
		// the request, the authorization endpoint MUST return the authorization
		// error response with the "error" value set to "invalid_request".  The
		// "error_description" or the response of "error_uri" SHOULD explain the
		// nature of error, e.g., code challenge required.
		return c.validateNoPKCE(ctx, client)
	}

	// If the server supporting PKCE does not support the requested
	// transformation, the authorization endpoint MUST return the
	// authorization error response with "error" value set to
	// "invalid_request".  The "error_description" or the response of
	// "error_uri" SHOULD explain the nature of error, e.g., transform
	// algorithm not supported.
	switch method {
	case PKCEChallengeMethodSHA256:
		break
	case PKCEChallengeMethodPlain:
		fallthrough
	case "":
		if !c.Config.GetEnablePKCEPlainChallengeMethod(ctx) {
			return errorsx.WithStack(fosite.ErrInvalidRequest.
				WithHint("Clients must use the 'S256' PKCE 'code_challenge_method' but the 'plain' method was requested.").
				WithDebug("The server is configured in a way that enforces PKCE 'S256' as challenge method for clients."))
		}
	default:
		return errorsx.WithStack(fosite.ErrInvalidRequest.
			WithHint("The code_challenge_method is not supported, use S256 instead."))
	}

	return nil
}

func (c *PKCEHandler) validateNoPKCE(ctx context.Context, client fosite.Client) error {
	var enforce bool

	enforce = c.Config.GetEnforcePKCE(ctx)

	if enforce {
		return errorsx.WithStack(fosite.ErrInvalidRequest.
			WithHint("Clients must include a code_challenge when performing the authorize code flow, but it is missing.").
			WithDebug("The server is configured in a way that enforces PKCE for clients."))
	}

	enforce = c.Config.GetEnforcePKCEForPublicClients(ctx)
	public := client.IsPublic()

	if enforce && public {
		return errorsx.WithStack(fosite.ErrInvalidRequest.
			WithHint("This client must include a code_challenge when performing the authorize code flow, but it is missing.").
			WithDebug("The server is configured in a way that enforces PKCE for this client."))
	}

	return nil
}

// HandleTokenEndpointRequest implements fosite.TokenEndpointHandler partially.
//
//nolint:gocyclo
func (c *PKCEHandler) HandleTokenEndpointRequest(ctx context.Context, requester fosite.AccessRequester) error {
	if !c.CanHandleTokenEndpointRequest(ctx, requester) {
		return errorsx.WithStack(fosite.ErrUnknownRequest)
	}

	// code_verifier
	// REQUIRED.  Code verifier
	//
	// The "code_challenge_method" is bound to the Authorization Code when
	// the Authorization Code is issued.  That is the method that the token
	// endpoint MUST use to verify the "code_verifier".
	verifier := requester.GetRequestForm().Get(FormParameterCodeVerifier)

	code := requester.GetRequestForm().Get(FormParameterAuthorizationCode)
	signature := c.AuthorizeCodeStrategy.AuthorizeCodeSignature(ctx, code)
	pkceRequest, err := c.Storage.GetPKCERequestSession(ctx, signature, requester.GetSession())

	nv := len(verifier)

	if errors.Is(err, fosite.ErrNotFound) {
		if nv == 0 {
			return c.validateNoPKCE(ctx, requester.GetClient())
		}

		return errorsx.WithStack(fosite.ErrInvalidGrant.WithHint("Unable to find initial PKCE data tied to this request").WithWrap(err).WithDebugf("The recorded error is: %s.", err.Error()))
	} else if err != nil {
		return errorsx.WithStack(fosite.ErrServerError.WithWrap(err).WithDebugf("The recorded error is: %s.", err.Error()))
	}

	if err = c.Storage.DeletePKCERequestSession(ctx, signature); err != nil {
		return errorsx.WithStack(fosite.ErrServerError.WithWrap(err).WithDebugf("The recorded error is: %s.", err.Error()))
	}

	challenge := pkceRequest.GetRequestForm().Get(FormParameterCodeChallenge)
	method := pkceRequest.GetRequestForm().Get(FormParameterCodeChallengeMethod)
	client := pkceRequest.GetClient()

	if err = c.validate(ctx, challenge, method, client); err != nil {
		return err
	}

	nc := len(challenge)

	if !c.Config.GetEnforcePKCE(ctx) && nc == 0 && nv == 0 {
		return nil
	}

	// NOTE: The code verifier SHOULD have enough entropy to make it
	// 	impractical to guess the value.  It is RECOMMENDED that the output of
	// 	a suitable random number generator be used to create a 32-octet
	// 	sequence.  The octet sequence is then base64url-encoded to produce a
	// 	43-octet URL safe string to use as the code verifier.

	// Miscellaneous validations.
	switch {
	case nv < 43:
		return errorsx.WithStack(fosite.ErrInvalidGrant.
			WithHint("The PKCE code verifier must be at least 43 characters."))
	case nv > 128:
		return errorsx.WithStack(fosite.ErrInvalidGrant.
			WithHint("The PKCE code verifier can not be longer than 128 characters."))
	case rePKCEVerifier.MatchString(verifier):
		return errorsx.WithStack(fosite.ErrInvalidGrant.
			WithHint("The PKCE code verifier must only contain [a-Z], [0-9], '-', '.', '_', '~'."))
	case nc == 0:
		return errorsx.WithStack(fosite.ErrInvalidGrant.
			WithHint("The PKCE code verifier was provided but the code challenge was absent from the authorization request."))
	}

	// Upon receipt of the request at the token endpoint, the server
	// verifies it by calculating the code challenge from the received
	// "code_verifier" and comparing it with the previously associated
	// "code_challenge", after first transforming it according to the
	// "code_challenge_method" method specified by the client.
	//
	// 	If the "code_challenge_method" from Section 4.3 was "S256", the
	// received "code_verifier" is hashed by SHA-256, base64url-encoded, and
	// then compared to the "code_challenge", i.e.:
	//
	// BASE64URL-ENCODE(SHA256(ASCII(code_verifier))) == code_challenge
	//
	// If the "code_challenge_method" from Section 4.3 was "plain", they are
	// compared directly, i.e.:
	//
	// code_verifier == code_challenge.
	//
	// 	If the values are equal, the token endpoint MUST continue processing
	// as normal (as defined by OAuth 2.0 [RFC6749]).  If the values are not
	// equal, an error response indicating "invalid_grant" as described in
	// Section 5.2 of [RFC6749] MUST be returned.
	switch method {
	case PKCEChallengeMethodSHA256:
		hash := sha256.New()
		if _, err = hash.Write([]byte(verifier)); err != nil {
			return errorsx.WithStack(fosite.ErrServerError.WithWrap(err).WithDebugf("The recorded error is: %s.", err.Error()))
		}

		sum := hash.Sum([]byte{})

		expected := make([]byte, base64.RawURLEncoding.EncodedLen(len(sum)))

		base64.RawURLEncoding.Strict().Encode(expected, sum)

		if subtle.ConstantTimeCompare(expected, []byte(challenge)) == 0 {
			return errorsx.WithStack(fosite.ErrInvalidGrant.
				WithHint("The PKCE code challenge did not match the code verifier."))
		}
	case PKCEChallengeMethodPlain:
		fallthrough
	default:
		if subtle.ConstantTimeCompare([]byte(verifier), []byte(challenge)) == 0 {
			return errorsx.WithStack(fosite.ErrInvalidGrant.
				WithHint("The PKCE code challenge did not match the code verifier."))
		}
	}

	return nil
}

// PopulateTokenEndpointResponse implements fosite.TokenEndpointHandler partially.
func (c *PKCEHandler) PopulateTokenEndpointResponse(ctx context.Context, requester fosite.AccessRequester, responder fosite.AccessResponder) error {
	return nil
}

// CanSkipClientAuth implements fosite.TokenEndpointHandler partially.
func (c *PKCEHandler) CanSkipClientAuth(ctx context.Context, requester fosite.AccessRequester) bool {
	return false
}

// CanHandleTokenEndpointRequest implements fosite.TokenEndpointHandler partially.
func (c *PKCEHandler) CanHandleTokenEndpointRequest(ctx context.Context, requester fosite.AccessRequester) bool {
	return requester.GetGrantTypes().ExactOne(GrantTypeAuthorizationCode)
}
