// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"fmt"

	"github.com/google/uuid"

	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/storage"
)

func handleOAuth2TokenExchangePolicy(ctx *middlewares.AutheliaCtx, requester oauthelia2.AccessRequester, client oidc.Client, session *oidc.Session) (details *authentication.UserDetailsExtended, err error) {
	own := isOAuth2TokenExchangeOwnToken(requester, client, session)
	cc := oidc.TokenExchangeSubjectClientCredentials(session)
	replacement := oidc.IsTokenExchangeTxnTokenReplacement(requester)

	if !own && cc {
		return nil, oauthelia2.ErrInvalidGrant.WithHint("A token issued via the client credentials grant may only be exchanged by the client it was issued to.")
	}

	if !own && !replacement && requester.GetRequestForm().Get("actor_token") == "" && !client.GetImpersonationAllowed() {
		return nil, oauthelia2.ErrInvalidRequest.WithHint("The client is not permitted to perform impersonation; an 'actor_token' is required.")
	}

	if !own {
		if details, err = handleOAuth2TokenExchangeConsent(ctx, requester, client, session); err != nil {
			return nil, err
		}
	}

	if cc {
		return nil, nil
	}

	if details == nil {
		if details, err = handleOAuth2TokenExchangeResolveSubject(ctx, requester, client, session); err != nil {
			return nil, err
		}
	}

	if err = handleOAuth2TokenExchangeAuthorizationPolicy(ctx, client, session, details); err != nil {
		return nil, err
	}

	session.Username = details.Username

	if replacement {
		return details, nil
	}

	if err = handleOAuth2TokenExchangeClientSubject(ctx, requester, client, session, details); err != nil {
		return nil, err
	}

	return details, nil
}

func handleOAuth2TokenExchangeClientSubject(ctx *middlewares.AutheliaCtx, requester oauthelia2.AccessRequester, client oidc.Client, session *oidc.Session, details *authentication.UserDetailsExtended) (err error) {
	var subject uuid.UUID

	if subject, err = ctx.Providers.OpenIDConnect.GetSubject(ctx, client.GetSectorIdentifierURI(), details.Username); err != nil {
		ctx.GetLogger().
			WithFields(map[string]any{"oauth2_access_request_id": requester.GetID(), "client_id": client.GetID()}).
			WithError(err).
			Error("Access Request encountered an error resolving the subject of the client")

		return oauthelia2.ErrServerError.WithWrap(err).WithDebugf("Error occurred resolving the subject of the client: %+v", err)
	}

	session.Subject = subject.String()

	if session.Claims != nil {
		session.Claims.Subject = session.Subject
	}

	return nil
}

func logOAuth2TokenExchange(ctx *middlewares.AutheliaCtx, requester oauthelia2.AccessRequester, responder oauthelia2.AccessResponder, err error) {
	if requester == nil || requester.GetClient() == nil || requester.GetClient().GetID() == "" {
		return
	}

	form := requester.GetRequestForm()

	if form == nil || form.Get(oidc.FormParameterGrantType) != oidc.GrantTypeTokenExchange {
		return
	}

	session, _ := requester.GetSession().(*oidc.Session)

	fields := map[string]any{
		"oauth2_access_request_id": requester.GetID(),
		"client_id":                requester.GetClient().GetID(),
		"impersonation":            form.Get(oidc.FormParameterActorToken) == "" && !isOAuth2TokenExchangeOwnToken(requester, requester.GetClient(), session) && !oidc.IsTokenExchangeTxnTokenReplacement(requester),
		"remote_ip":                ctx.RemoteIP().String(),
	}

	for key, value := range map[string]string{
		oidc.FormParameterSubjectTokenType:   form.Get(oidc.FormParameterSubjectTokenType),
		oidc.FormParameterActorTokenType:     form.Get(oidc.FormParameterActorTokenType),
		oidc.FormParameterRequestedTokenType: form.Get(oidc.FormParameterRequestedTokenType),
	} {
		setOAuth2TokenExchangeLogField(fields, key, value)
	}

	for key, value := range map[string][]string{
		"granted_scope":    requester.GetGrantedScopes(),
		"granted_audience": requester.GetGrantedAudience(),
		"granted_resource": requester.GetGrantedResource(),
	} {
		if len(value) != 0 {
			fields[key] = value
		}
	}

	if session != nil {
		addOAuth2TokenExchangeSessionLogFields(fields, session)
	}

	if responder != nil {
		issued, _ := responder.GetExtra(oidc.FormParameterIssuedTokenType).(string)

		setOAuth2TokenExchangeLogField(fields, oidc.FormParameterIssuedTokenType, issued)

		if issued == oidc.TokenTypeIDJAG {
			addOAuth2IDJAGLogFields(ctx, fields, requester)
		}
	}

	entry := ctx.GetLogger().WithFields(fields)

	if err != nil {
		entry.WithError(oauthelia2.ErrorToDebugRFC6749Error(err)).Warn("Token Exchange denied")

		return
	}

	entry.Info("Token Exchange completed")
}

func addOAuth2TokenExchangeSessionLogFields(fields map[string]any, session *oidc.Session) {
	setOAuth2TokenExchangeLogField(fields, "subject", session.Subject)
	setOAuth2TokenExchangeLogField(fields, "username", session.Username)

	if owner, stored := oidc.TokenExchangeSubjectClientID(session); stored {
		setOAuth2TokenExchangeLogField(fields, "subject_token_client_id", owner)
	}

	act, _ := session.Extra[oidc.ClaimActor].(map[string]any)

	for _, key := range []string{oidc.ClaimSubject, oidc.ClaimClientIdentifier} {
		if value, ok := act[key].(string); ok && value != "" {
			fields["actor"] = value

			return
		}
	}
}

func setOAuth2TokenExchangeLogField(fields map[string]any, key, value string) {
	if value != "" {
		fields[key] = value
	}
}

func isOAuth2TokenExchangeOwnToken(requester oauthelia2.AccessRequester, client oauthelia2.Client, session *oidc.Session) (own bool) {
	if clientID, stored := oidc.TokenExchangeSubjectClientID(session); stored {
		return clientID == client.GetID()
	}

	if session != nil && session.GetSubjectRequester() != nil {
		return false
	}

	return requester.GetRequestForm().Get(oidc.FormParameterSubjectTokenType) == oidc.TokenTypeIDToken
}

func handleOAuth2TokenExchangeConsent(ctx *middlewares.AutheliaCtx, requester oauthelia2.AccessRequester, client oidc.Client, session *oidc.Session) (details *authentication.UserDetailsExtended, err error) {
	if client.GetConsentPolicy().Mode == oidc.ClientConsentModePreConfigured {
		if details, err = handleOAuth2TokenExchangeResolveSubject(ctx, requester, client, session); err != nil {
			return nil, err
		}
	}

	if err = handleOAuth2ClientConsentModePolicy(ctx, requester, client, details, "The client requires explicit user consent which cannot be obtained during a token exchange."); err != nil {
		return nil, err
	}

	return details, nil
}

func handleOAuth2ClientConsentModePolicy(ctx *middlewares.AutheliaCtx, requester oauthelia2.AccessRequester, client oidc.Client, details *authentication.UserDetailsExtended, hintExplicit string) (err error) {
	switch client.GetConsentPolicy().Mode {
	case oidc.ClientConsentModeImplicit:
		return nil
	case oidc.ClientConsentModePreConfigured:
		var consented bool

		if consented, err = handleOAuth2TokenExchangePreConfiguredConsent(ctx, requester, client, details); err != nil {
			return oauthelia2.ErrServerError.WithWrap(err).WithDebugf("Error occurred looking up pre-configured consent: %+v", err)
		}

		if !consented {
			return oauthelia2.ErrInvalidGrant.WithHint("The user has not consented to this client for the requested scopes.")
		}

		return nil
	default:
		return oauthelia2.ErrInvalidGrant.WithHint(hintExplicit)
	}
}

func handleOAuth2TokenExchangePreConfiguredConsent(ctx *middlewares.AutheliaCtx, requester oauthelia2.AccessRequester, client oidc.Client, details *authentication.UserDetailsExtended) (consented bool, err error) {
	var subject uuid.UUID

	if subject, err = ctx.Providers.OpenIDConnect.GetSubject(ctx, client.GetSectorIdentifierURI(), details.Username); err != nil {
		return false, fmt.Errorf("error resolving the subject of the client: %w", err)
	}

	var rows *storage.ConsentPreConfigRows

	if rows, err = ctx.Providers.StorageProvider.LoadOAuth2ConsentPreConfigurations(ctx, client.GetID(), subject, ctx.GetClock().Now()); err != nil {
		return false, fmt.Errorf("error loading rows: %w", err)
	}

	defer func() {
		if errc := rows.Close(); errc != nil {
			ctx.GetLogger().
				WithFields(map[string]any{"oauth2_access_request_id": requester.GetID(), "client_id": client.GetID()}).
				WithError(errc).
				Error("Access Request encountered an error closing the pre-configured consent rows")
		}
	}()

	var config *model.OAuth2ConsentPreConfig

	now := ctx.GetClock().Now()

	for rows.Next() {
		if config, err = rows.Get(); err != nil {
			return false, fmt.Errorf("error iterating rows: %w", err)
		}

		if config.CanConsentAt(now) && config.HasGrantsCovering(requester.GetGrantedScopes(), requester.GetGrantedAudience(), requester.GetGrantedResource()) {
			return true, nil
		}
	}

	return false, nil
}

func handleOAuth2TokenExchangeResolveSubject(ctx *middlewares.AutheliaCtx, requester oauthelia2.AccessRequester, client oidc.Client, session *oidc.Session) (details *authentication.UserDetailsExtended, err error) {
	return handleOAuth2ResolveSessionSubject(ctx, requester, client, session,
		"Access Request encountered an error resolving the user of the 'subject_token'",
		"The subject of the 'subject_token' could not be resolved.",
	)
}

func handleOAuth2ResolveSessionSubject(ctx *middlewares.AutheliaCtx, requester oauthelia2.AccessRequester, client oidc.Client, session *oidc.Session, message, hint string) (details *authentication.UserDetailsExtended, err error) {
	var (
		detailer oidc.UserDetailer
		ok       bool
	)

	if detailer, err = oidc.UserDetailerFromSubjectString(ctx, session.Subject); err == nil {
		details, ok = detailer.(*authentication.UserDetailsExtended)
	}

	if err != nil || !ok || details == nil || details.UserDetails == nil {
		ctx.GetLogger().
			WithFields(map[string]any{"oauth2_access_request_id": requester.GetID(), "client_id": client.GetID(), "subject": session.Subject}).
			WithError(err).
			Error(message)

		return nil, oauthelia2.ErrInvalidGrant.WithHint(hint)
	}

	return details, nil
}

func handleOAuth2TokenExchangeAuthorizationPolicy(ctx *middlewares.AutheliaCtx, client oidc.Client, session *oidc.Session, details *authentication.UserDetailsExtended) (err error) {
	return handleOAuth2SubjectTokenAuthorizationPolicy(ctx, client, session, details,
		"The client's authorization policy does not permit this user.",
		"The authentication level of the 'subject_token' is insufficient for the client's authorization policy.",
	)
}

func handleOAuth2SubjectTokenAuthorizationPolicy(ctx *middlewares.AutheliaCtx, client oidc.Client, session *oidc.Session, details *authentication.UserDetailsExtended, hintDenied, hintInsufficient string) (err error) {
	have := authentication.OneFactor

	if session.DefaultSession != nil && session.Claims != nil && authorization.NewAuthenticationMethodsReferencesFromClaim(session.Claims.AuthenticationMethodsReferences).MultiFactorAuthentication() {
		have = authentication.TwoFactor
	}

	required := client.GetAuthorizationPolicyRequiredLevel(authorization.Subject{Username: details.Username, Groups: details.Groups, IP: ctx.RemoteIP()})

	if required == authorization.Denied {
		return oauthelia2.ErrInvalidGrant.WithHint(hintDenied)
	}

	if !authorization.IsAuthLevelSufficient(have, required) {
		return oauthelia2.ErrInvalidGrant.WithHint(hintInsufficient)
	}

	return nil
}
