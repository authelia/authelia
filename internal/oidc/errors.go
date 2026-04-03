package oidc

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	oauthelia2 "authelia.com/provider/oauth2"
	"github.com/valyala/fasthttp"
)

var (
	errClientSecretMismatch = errors.New("The provided client secret did not match the registered client secret.") //nolint:staticcheck // Log error message.
)

var (
	// ErrSubjectCouldNotLookup is sent when the Subject Identifier for a user couldn't be generated or obtained from the database.
	ErrSubjectCouldNotLookup = oauthelia2.ErrServerError.WithHint("Could not lookup user subject.")

	// ErrConsentCouldNotPerform is sent when the Consent Session couldn't be performed for varying reasons.
	ErrConsentCouldNotPerform = oauthelia2.ErrServerError.WithHint("Could not perform consent.")

	// ErrConsentCouldNotGenerate is sent when the Consent Session failed to be generated for some reason, usually a failed UUIDv4 generation.
	ErrConsentCouldNotGenerate = oauthelia2.ErrServerError.WithHint("Could not generate the consent session.")

	// ErrConsentCouldNotSave is sent when the Consent Session couldn't be saved to the database.
	ErrConsentCouldNotSave = oauthelia2.ErrServerError.WithHint("Could not save the consent session.")

	// ErrConsentCouldNotLookup is sent when the Consent ID is not a known UUID.
	ErrConsentCouldNotLookup = oauthelia2.ErrServerError.WithHint("Failed to lookup the consent session.")

	// ErrConsentMalformedChallengeID is sent when the Consent ID is not a UUID.
	ErrConsentMalformedChallengeID = oauthelia2.ErrServerError.WithHint("Malformed consent session challenge ID.")

	ErrClientAuthorizationUserAccessDenied = oauthelia2.ErrAccessDenied.WithHint("The user was denied access to this client.")
)

type RedirectAuthorizeErrorFieldResponseStrategyConfig interface {
	oauthelia2.SendDebugMessagesToClientsProvider
	GetContext(ctx context.Context) (octx Context)
}

type RedirectAuthorizeErrorFieldResponseStrategy struct {
	Config RedirectAuthorizeErrorFieldResponseStrategyConfig
}

func (s *RedirectAuthorizeErrorFieldResponseStrategy) WriteErrorFieldResponse(ctx context.Context, rw http.ResponseWriter, requester oauthelia2.AuthorizeRequester, rfc *oauthelia2.RFC6749Error) {
	if rfc == nil {
		rfc = oauthelia2.ErrServerError
	}

	query := url.Values{}

	if len(rfc.ErrorField) != 0 {
		query.Set("error", rfc.ErrorField)
	}

	if len(rfc.DescriptionField) != 0 {
		query.Set("error_description", rfc.DescriptionField)
	}

	if rfc.CodeField != 0 {
		query.Set("error_status_code", strconv.Itoa(rfc.CodeField))
	}

	if len(rfc.HintField) != 0 {
		query.Set("error_hint", rfc.HintField)
	}

	if s.Config.GetSendDebugMessagesToClients(ctx) && len(rfc.DebugField) != 0 {
		query.Set("error_debug", rfc.DebugField)
	}

	ctxx := s.Config.GetContext(ctx)

	location := ctxx.RootURL().JoinPath(FrontendEndpointPathConsentCompletion)

	location.RawQuery = query.Encode()

	rw.Header().Set(fasthttp.HeaderCacheControl, "no-store")
	rw.Header().Set(fasthttp.HeaderPragma, "no-cache")
	rw.Header().Set(fasthttp.HeaderLocation, location.String())
	rw.WriteHeader(http.StatusFound)
}
