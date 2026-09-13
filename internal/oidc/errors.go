// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/valyala/fasthttp"

	oauthelia2 "authelia.com/provider/oauth2"
)

var (
	errClientSecretMismatch = errors.New("The provided client secret did not match the registered client secret.") //nolint:staticcheck // Log error message.
)

var (
	// ErrSubjectCouldNotLookup is sent when the Subject Identifier for a user couldn't be generated or obtained from the database.
	ErrSubjectCouldNotLookup = oauthelia2.ErrServerError.WithHint("Could not lookup user subject.")

	// ErrConsentCouldNotPerform is sent when the Consent Session couldn't be performed for varying reasons.
	ErrConsentCouldNotPerform = oauthelia2.ErrInvalidRequest.WithHint("Could not perform consent. The consent session has already been granted, has expired, or otherwise does not appear to be valid for the authorization request.")

	// ErrConsentCouldNotGenerate is sent when the Consent Session failed to be generated for some reason, usually a failed UUIDv4 generation.
	ErrConsentCouldNotGenerate = oauthelia2.ErrServerError.WithHint("Could not generate the consent session.")

	// ErrConsentCouldNotSave is sent when the Consent Session couldn't be saved to the database.
	ErrConsentCouldNotSave = oauthelia2.ErrServerError.WithHint("Could not save the consent session.")

	// ErrConsentCouldNotLookup is sent when the Consent ID is not a known UUID.
	ErrConsentCouldNotLookup = oauthelia2.ErrServerError.WithHint("Failed to lookup the consent session.")

	// ErrConsentMalformedChallengeID is sent when the Consent ID is not a UUID.
	ErrConsentMalformedChallengeID = oauthelia2.ErrServerError.WithHint("Malformed consent session challenge ID.")

	// ErrFlowCouldNotContinue is shown to the user when a login flow can't be continued. It deliberately describes the
	// range of problems which lead here rather than which one occurred, so that it can't be used to determine whether a
	// particular flow exists, has expired, or has already been responded to. The specific cause is recorded in the log
	// and is only included in the debug field of the response when enable_client_debug_messages is enabled.
	ErrFlowCouldNotContinue = oauthelia2.ErrInvalidRequest.WithHint("Could not continue the login flow. The request may have already been completed, may have expired, may have been superseded by a newer request, or may otherwise no longer be valid. Return to the application you are signing in to and try again.")

	// ErrConsentMalformedForm is sent when the original authorization request form of a Consent Session cannot be decoded.
	ErrConsentMalformedForm = oauthelia2.ErrServerError.WithHint("Malformed consent session request form data.")

	// ErrDeviceCodeMalformedUserCode is sent when the user code of a Device Code Session is not a plausible user code.
	ErrDeviceCodeMalformedUserCode = oauthelia2.ErrInvalidRequest.WithHint("Malformed device code session user code.")

	// ErrDeviceCodeCouldNotLookup is sent when the Device Code Session for a user code can't be retrieved.
	ErrDeviceCodeCouldNotLookup = oauthelia2.ErrInvalidRequest.WithHint("Failed to lookup the device code session. The user code is unknown or has expired.")

	// ErrDeviceCodeCouldNotDetermineSignature is sent when the signature of a user code can't be determined.
	ErrDeviceCodeCouldNotDetermineSignature = oauthelia2.ErrServerError.WithHint("Could not determine the signature of the device code session user code.")

	// ErrDeviceCodeCouldNotPerform is sent when the Device Code Session can't be performed for varying reasons.
	ErrDeviceCodeCouldNotPerform = oauthelia2.ErrInvalidRequest.WithHint("Could not perform the device authorization. The device code session has already been responded to, has expired, or otherwise does not appear to be valid for the authorization request.")

	// ErrClientAuthorizationUserAccessDenied is sent when the user is denied access to a client.
	ErrClientAuthorizationUserAccessDenied = oauthelia2.ErrAccessDenied.WithHint("The user was denied access to this client.")

	// errClientRegistrationNotSupported is returned by the client registration management methods of the Store. This
	// Authorization Server registers clients from its configuration and serves neither the RFC 7591 registration
	// endpoint nor the RFC 7592 management endpoint, so these methods are unreachable in practice.
	errClientRegistrationNotSupported = oauthelia2.ErrNotFound.WithHint("Dynamic Client Registration is not supported by this Authorization Server.")
)

// RedirectAuthorizeErrorFieldResponseStrategyConfig is the configuration used by the RedirectAuthorizeErrorFieldResponseStrategy.
type RedirectAuthorizeErrorFieldResponseStrategyConfig interface {
	oauthelia2.SendDebugMessagesToClientsProvider
	GetContext(ctx context.Context) (octx Context)
}

// RedirectAuthorizeErrorFieldResponseStrategy is a strategy which writes authorization errors to the Authelia error page.
type RedirectAuthorizeErrorFieldResponseStrategy struct {
	Config RedirectAuthorizeErrorFieldResponseStrategyConfig
}

// WriteErrorFieldResponse writes the error response for the given requester.
func (s *RedirectAuthorizeErrorFieldResponseStrategy) WriteErrorFieldResponse(ctx context.Context, rw http.ResponseWriter, requester oauthelia2.AuthorizeRequester, rfc *oauthelia2.RFC6749Error) {
	var (
		issuer *url.URL
		err    error
	)

	ctxx := s.Config.GetContext(ctx)

	if issuer, err = ctxx.IssuerURL(); err != nil {
		return
	}

	location := ConsentCompletionURL(issuer, rfc, s.Config.GetSendDebugMessagesToClients(ctx))

	rw.Header().Set(fasthttp.HeaderCacheControl, "no-store")
	rw.Header().Set(fasthttp.HeaderPragma, "no-cache")
	rw.Header().Set(fasthttp.HeaderLocation, location.String())
	rw.WriteHeader(http.StatusFound)
}

// ConsentCompletionURL returns the frontend consent completion URL for the given issuer with the fields of the given
// error encoded into the query. A nil error is treated as a generic server error and the debug field is only included
// when debug is true.
func ConsentCompletionURL(issuer *url.URL, rfc *oauthelia2.RFC6749Error, debug bool) (location *url.URL) {
	if rfc == nil {
		rfc = oauthelia2.ErrServerError
	}

	location = issuer.JoinPath(FrontendEndpointPathConsentCompletion)

	query := location.Query()

	if len(rfc.ErrorField) != 0 {
		query.Set(FrontendQueryArgError, rfc.ErrorField)
	}

	if len(rfc.DescriptionField) != 0 {
		query.Set(FrontendQueryArgErrorDescription, rfc.DescriptionField)
	}

	if rfc.CodeField != 0 {
		query.Set(FrontendQueryArgErrorStatusCode, strconv.Itoa(rfc.CodeField))
	}

	if len(rfc.HintField) != 0 {
		query.Set(FrontendQueryArgErrorHint, rfc.HintField)
	}

	if debug && len(rfc.DebugField) != 0 {
		query.Set(FrontendQueryArgErrorDebug, rfc.DebugField)
	}

	location.RawQuery = query.Encode()

	return location
}
