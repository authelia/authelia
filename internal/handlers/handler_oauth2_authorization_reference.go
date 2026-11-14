// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"net/http"
	"net/url"

	"github.com/google/uuid"

	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
)

// handleOAuth2AuthorizationConsentReferenceURL returns the Authorization Endpoint URL which refers to the Authorization
// Request stored in the consent session rather than carrying its parameters, so they never appear in a URL.
func handleOAuth2AuthorizationConsentReferenceURL(issuer *url.URL, consent *model.OAuth2ConsentSession) (location *url.URL) {
	location = issuer.JoinPath(oidc.EndpointPathAuthorization)

	query := url.Values{}

	query.Set(oidc.FormParameterClientID, consent.ClientID)
	query.Set(queryArgConsentID, consent.ChallengeID.String())

	location.RawQuery = query.Encode()

	return location
}

func isOAuth2AuthorizationConsentReference(query url.Values) bool {
	return len(query) == 2 && len(query[oidc.FormParameterClientID]) == 1 && len(query[queryArgConsentID]) == 1
}

// handleOAuth2AuthorizationConsentReference replaces the parameters of a request which refers to an Authorization
// Request stored in a consent session with the parameters of that Authorization Request, so the request is processed
// as if the user agent had sent them. The consent session must be bound to the user agent and belong to the client.
func handleOAuth2AuthorizationConsentReference(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, r *http.Request) (handled bool) {
	query := r.URL.Query()

	if !isOAuth2AuthorizationConsentReference(query) {
		return false
	}

	var (
		challengeID uuid.UUID
		consent     *model.OAuth2ConsentSession
		form        url.Values
		err         error
	)

	requester := oauthelia2.NewAuthorizeRequest()

	if challengeID, err = uuid.Parse(query.Get(queryArgConsentID)); err != nil {
		ctx.GetLogger().WithError(err).Errorf("Authorization Request with id '%s' could not be processed: error occurred parsing the consent session id", requester.GetID())

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oidc.ErrConsentMalformedChallengeID)

		return true
	}

	if consent, err = handleOAuth2ConsentSessionLoadBound(ctx, challengeID); err != nil {
		ctx.GetLogger().WithError(err).Errorf("Authorization Request with id '%s' could not be processed: error occurred loading the consent session with id '%s'", requester.GetID(), challengeID)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotLookup)

		return true
	}

	if consent.ClientID != query.Get(oidc.FormParameterClientID) {
		ctx.GetLogger().Errorf("Authorization Request with id '%s' could not be processed: the client id '%s' does not match the client id '%s' of the consent session with id '%s'", requester.GetID(), query.Get(oidc.FormParameterClientID), consent.ClientID, challengeID)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotLookup)

		return true
	}

	if form, err = consent.GetForm(); err != nil {
		ctx.GetLogger().WithError(err).Errorf("Authorization Request with id '%s' could not be processed: error occurred obtaining the request form of the consent session with id '%s'", requester.GetID(), challengeID)

		ctx.Providers.OpenIDConnect.WriteAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotLookup)

		return true
	}

	form.Set(queryArgConsentID, challengeID.String())

	r.URL.RawQuery = form.Encode()
	r.RequestURI = r.URL.RequestURI()
	r.Form, r.PostForm = nil, nil

	return false
}
