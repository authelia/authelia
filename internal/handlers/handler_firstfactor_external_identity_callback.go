// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"crypto/subtle"
	"errors"
	"net/url"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/externalidentity"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/storage"
)

// FirstFactorExternalIdentityCallbackGET is the redirect URI endpoint for external identity logins which use the
// query response mode. It always responds with a redirect and never renders content controlled by the external
// provider.
func FirstFactorExternalIdentityCallbackGET(ctx *middlewares.AutheliaCtx) {
	handleExternalIdentityCallback(ctx, ctx.QueryArgs(), externalidentity.ResponseModeQuery)
}

// FirstFactorExternalIdentityCallbackPOST is the redirect URI endpoint for external identity logins which use
// the form post response mode. It always responds with a redirect and never renders content controlled by the external
// provider.
func FirstFactorExternalIdentityCallbackPOST(ctx *middlewares.AutheliaCtx) {
	handleExternalIdentityCallback(ctx, ctx.PostArgs(), externalidentity.ResponseModeFormPost)
}

//nolint:gocyclo // The function is a linear sequence of validations which is clearer than the sum of its parts.
func handleExternalIdentityCallback(ctx *middlewares.AutheliaCtx, args *fasthttp.Args, mode string) {
	userSession, err := ctx.GetSession()
	if err != nil {
		ctx.Logger.WithError(err).Errorf("%s: %s", logFmtErrExternalIdentityCallback, errStrUserSessionData)

		externalIdentityRedirect(ctx, pathExternalIdentityError)

		return
	}

	// The flow state is single use. It is taken from the session and persisted as absent before any validation of the
	// callback occurs so a replayed callback cannot reuse it.
	flow := userSession.ExternalIdentity

	userSession.ExternalIdentity = nil

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.WithError(err).Errorf("%s: %s", logFmtErrExternalIdentityCallback, errStrUserSessionDataSave)

		externalIdentityRedirect(ctx, pathExternalIdentityError)

		return
	}

	id, _ := ctx.UserValue("provider").(string)

	log := ctx.Logger.WithField("provider", id)

	provider, ok := ctx.Providers.ExternalIdentity.Get(id)

	switch {
	case flow == nil:
		log.Errorf("%s: there is no flow in progress", logFmtErrExternalIdentityCallback)

		externalIdentityRedirect(ctx, pathExternalIdentityError)

		return
	case !ok:
		log.Errorf("%s: %s", logFmtErrExternalIdentityCallback, errStrExternalIdentityProviderUnknown)

		externalIdentityRedirect(ctx, pathExternalIdentityError)

		return
	case flow.Provider != id:
		log.Errorf("%s: the flow in progress is for the provider '%s'", logFmtErrExternalIdentityCallback, flow.Provider)

		externalIdentityRedirect(ctx, pathExternalIdentityError)

		return
	case provider.ResponseMode() != mode:
		// The response must arrive the way it was requested. Accepting it any other way would let an authorization
		// response which was never meant for this flow be replayed through the response mode the provider does not use.
		log.Errorf("%s: the authorization response was delivered with the '%s' response mode but the provider is configured to use the '%s' response mode", logFmtErrExternalIdentityCallback, mode, provider.ResponseMode())

		externalIdentityRedirect(ctx, pathExternalIdentityError)

		return
	case flow.State == "":
		log.Errorf("%s: the flow in progress has no state", logFmtErrExternalIdentityCallback)

		externalIdentityRedirect(ctx, pathExternalIdentityError)

		return
	case ctx.GetClock().Now().After(flow.Expires):
		log.Errorf("%s: the flow has expired", logFmtErrExternalIdentityCallback)

		externalIdentityRedirect(ctx, pathExternalIdentityError)

		return
	}

	if subtle.ConstantTimeCompare(args.Peek(queryArgState), []byte(flow.State)) != 1 {
		log.Errorf("%s: the state does not match", logFmtErrExternalIdentityCallback)

		externalIdentityRedirect(ctx, pathExternalIdentityError)

		return
	}

	// The 'iss' parameter is checked before the 'error' parameter as RFC9207 applies to error responses as well.
	if issuer := string(args.Peek(queryArgISS)); issuer != "" {
		if issuer != provider.Issuer() {
			log.Errorf("%s: the 'iss' parameter does not match the configured issuer", logFmtErrExternalIdentityCallback)

			externalIdentityRedirect(ctx, pathExternalIdentityError)

			return
		}
	} else {
		var required bool

		if required, err = provider.AuthorizationResponseIssuerRequired(ctx); err != nil {
			log.WithError(err).Errorf("%s: error determining if the 'iss' parameter is required", logFmtErrExternalIdentityCallback)

			externalIdentityRedirect(ctx, pathExternalIdentityError)

			return
		}

		if required {
			log.Errorf("%s: the 'iss' parameter is absent but the provider is known to send it", logFmtErrExternalIdentityCallback)

			externalIdentityRedirect(ctx, pathExternalIdentityError)

			return
		}
	}

	if e := string(args.Peek(queryArgError)); e != "" {
		// Every value is supplied by the provider, or by anyone able to direct the browser to the callback, so each is
		// sanitized before it is logged. None of them is ever rendered.
		fields := logrus.Fields{queryArgError: externalidentity.SanitizeProviderErrorValue(e)}

		for _, name := range []string{queryArgErrorDescription, queryArgErrorHint} {
			if value := args.Peek(name); len(value) != 0 {
				fields[name] = externalidentity.SanitizeProviderErrorValue(string(value))
			}
		}

		log.WithFields(fields).Errorf("%s: the provider returned an error", logFmtErrExternalIdentityCallback)

		externalIdentityRedirect(ctx, pathExternalIdentityError)

		return
	}

	code := string(args.Peek(queryArgCode))

	if code == "" {
		log.Errorf("%s: the authorization code is absent", logFmtErrExternalIdentityCallback)

		externalIdentityRedirect(ctx, pathExternalIdentityError)

		return
	}

	var redirectURI string

	if redirectURI, err = externalIdentityRedirectURI(ctx, provider.ID()); err != nil {
		log.WithError(err).Error(logFmtErrExternalIdentityCallback)

		externalIdentityRedirect(ctx, pathExternalIdentityError)

		return
	}

	var claims *externalidentity.IdentityClaims

	if claims, err = provider.Complete(ctx, externalidentity.CompletionRequest{
		Code:         code,
		Nonce:        flow.Nonce,
		CodeVerifier: flow.CodeVerifier,
		Handle:       flow.Handle,
		RedirectURI:  redirectURI,
		Language:     flow.Language,
		Now:          ctx.GetClock().Now(),
	}); err != nil {
		log.WithError(err).Error(logFmtErrExternalIdentityCallback)

		externalIdentityRedirect(ctx, pathExternalIdentityError)

		return
	}

	// The account link is looked up by the provider type, issuer, and subject alone. The 'email' claim is not verified by Authelia
	// and must never be used to find, match, or link an account.
	link, err := ctx.Providers.StorageProvider.LoadExternalIdentityLinkBySubject(ctx, provider.Type(), claims.Issuer, claims.Subject)

	switch {
	case err == nil:
		handleExternalIdentityCallbackLinked(ctx, provider, flow, claims, link.ID, link.Username)
	case errors.Is(err, storage.ErrNoExternalIdentityLink):
		handleExternalIdentityCallbackUnlinked(ctx, provider, claims)
	default:
		log.WithError(err).Error(logFmtErrExternalIdentityCallback)

		externalIdentityRedirect(ctx, pathExternalIdentityError)
	}
}

func handleExternalIdentityCallbackLinked(ctx *middlewares.AutheliaCtx, provider externalidentity.Provider, flow *session.ExternalIdentityFlow, claims *externalidentity.IdentityClaims, id int, username string) {
	log := ctx.Logger.WithFields(map[string]any{"provider": provider.ID(), "username": username})

	var (
		ban     regulation.BanType
		value   string
		expires *time.Time
		err     error
	)

	if ban, value, expires, err = ctx.Providers.Regulator.BanCheck(ctx, username); err != nil {
		if errors.Is(err, regulation.ErrUserIsBanned) {
			doMarkAuthenticationAttemptWithRequest(ctx, false, regulation.NewBan(ban, value, expires), regulation.AuthType1FA, flow.TargetURL, flow.RequestMethod, nil)
		} else {
			log.WithError(err).Errorf(logFmtErrRegulationFail, regulation.AuthType1FA, username)
		}

		externalIdentityRedirect(ctx, pathExternalIdentityError)

		return
	}

	var details *authentication.UserDetails

	if details, err = ctx.Providers.UserProvider.GetDetails(username); err != nil || details == nil {
		doMarkAuthenticationAttemptWithRequest(ctx, false, regulation.NewBan(regulation.BanTypeNone, username, nil), regulation.AuthType1FA, flow.TargetURL, flow.RequestMethod, err)

		externalIdentityRedirect(ctx, pathExternalIdentityError)

		return
	}

	var sessionProvider *session.Session

	if sessionProvider, err = externalIdentityResetSession(ctx, log, username); err != nil {
		externalIdentityRedirect(ctx, pathExternalIdentityError)

		return
	}

	// The remember me policy is applied here rather than when the flow was started as the configuration in force when
	// the session is actually created is the one which must be honored. The value stored in the flow state is only
	// the users request; it is not authoritative.
	keepMeLoggedIn := !sessionProvider.Config.DisableRememberMe && flow.KeepMeLoggedIn

	if keepMeLoggedIn {
		if err = sessionProvider.UpdateExpiration(ctx.RequestCtx, sessionProvider.Config.RememberMe); err != nil {
			log.WithError(err).Errorf(logFmtErrSessionSave, "updated expiration", regulation.AuthType1FA, logFmtActionAuthentication, username)

			externalIdentityRedirect(ctx, pathExternalIdentityError)

			return
		}
	}

	var userSession session.UserSession

	if userSession, err = ctx.GetSession(); err != nil {
		log.WithError(err).Errorf("%s: %s", logFmtErrExternalIdentityCallback, errStrUserSessionData)

		externalIdentityRedirect(ctx, pathExternalIdentityError)

		return
	}

	userSession.SetOneFactorExternalIdentity(ctx.GetClock().Now(), details, keepMeLoggedIn)

	// Authelia does not observe the authentication performed at the external provider. The provider decides which
	// Authentication Method Reference values are adopted, from what it asserted and what the administrator configured.
	if amr := provider.AuthenticationMethodsReference(claims.AuthenticationMethodsReference); len(amr) != 0 {
		userSession.AuthenticationMethodRefs = userSession.AuthenticationMethodRefs.Merge(authorization.NewAuthenticationMethodsReferencesFromClaim(amr))
	}

	if ctx.Configuration.AuthenticationBackend.RefreshInterval.Update() {
		userSession.RefreshTTL = ctx.GetClock().Now().Add(ctx.Configuration.AuthenticationBackend.RefreshInterval.Value())
	}

	if err = ctx.SaveSession(userSession); err != nil {
		log.WithError(err).Errorf("%s: %s", logFmtErrExternalIdentityCallback, errStrUserSessionDataSave)

		externalIdentityRedirect(ctx, pathExternalIdentityError)

		return
	}

	if err = ctx.Providers.StorageProvider.UpdateExternalIdentityLinkSignIn(ctx, id, ctx.GetClock().Now()); err != nil {
		log.WithError(err).Warnf("%s: the sign in could not be recorded against the link", logFmtErrExternalIdentityCallback)
	}

	doMarkAuthenticationAttemptWithRequest(ctx, true, regulation.NewBan(regulation.BanTypeNone, username, nil), regulation.AuthType1FA, flow.TargetURL, flow.RequestMethod, nil)

	Handle1FARedirect(ctx, flow.TargetURL, flow.RequestMethod, userSession.Username, userSession.Groups)
}

func externalIdentityResetSession(ctx *middlewares.AutheliaCtx, log *logrus.Entry, username string) (provider *session.Session, err error) {
	if provider, err = ctx.GetSessionProvider(); err != nil {
		log.WithError(err).Errorf("%s: %s", logFmtErrExternalIdentityCallback, errStrSessionProvider)

		return nil, err
	}

	if err = provider.DestroySession(ctx.RequestCtx); err != nil {
		// This failure is not likely to be critical as the session is reset and the cookie regenerated below.
		log.WithError(err).Trace("Failed to destroy session during the external identity callback")
	}

	if err = provider.SaveSession(ctx.RequestCtx, provider.NewDefaultUserSession()); err != nil {
		log.WithError(err).Errorf(logFmtErrSessionReset, regulation.AuthType1FA, username)

		return nil, err
	}

	if err = ctx.RegenerateSession(); err != nil {
		log.WithError(err).Errorf(logFmtErrSessionRegenerate, regulation.AuthType1FA, username)

		return nil, err
	}

	return provider, nil
}

// handleExternalIdentityCallbackUnlinked stores the validated external identity as a proposal the user may accept from
// their settings. It does not authenticate the user.
func handleExternalIdentityCallbackUnlinked(ctx *middlewares.AutheliaCtx, provider externalidentity.Provider, claims *externalidentity.IdentityClaims) {
	userSession, err := ctx.GetSession()
	if err != nil {
		ctx.Logger.WithError(err).Errorf("%s: %s", logFmtErrExternalIdentityCallback, errStrUserSessionData)

		externalIdentityRedirect(ctx, pathExternalIdentityError)

		return
	}

	// No identity bearing value is carried across the authentication boundary. When the session is anonymous the
	// validated identity is discarded entirely and only the provider identifier is handed back to the portal so the
	// flow can be performed again by the user once they have authenticated. An attacker who can plant a session cookie
	// on the Authelia domain therefore cannot plant a proposal for their own external identity: the only value they
	// can influence is the provider identifier, which merely offers the victim a link to the victims own account.
	if userSession.IsAnonymous() {
		ctx.Logger.WithField("provider", provider.ID()).Info("The external identity is not linked to a local account and the session is not authenticated: the identity has been discarded and the user must authenticate before the link can be proposed")

		userSession.ExternalIdentityPending = nil

		if err = ctx.SaveSession(userSession); err != nil {
			ctx.Logger.WithError(err).Errorf("%s: %s", logFmtErrExternalIdentityCallback, errStrUserSessionDataSave)

			externalIdentityRedirect(ctx, pathExternalIdentityError)

			return
		}

		externalIdentityRedirect(ctx, pathExternalIdentityLink+"?"+queryArgLinkProvider+"="+url.QueryEscape(provider.ID()))

		return
	}

	// The 'email' claim is stored for display purposes only. Authelia must never use this value to find, match, or link
	// an account.
	userSession.ExternalIdentityPending = &session.ExternalIdentityPending{
		Provider:       provider.ID(),
		Type:           provider.Type(),
		Issuer:         claims.Issuer,
		Subject:        claims.Subject,
		RemoteUsername: claims.PreferredUsername,
		DisplayName:    claims.Name,
		Email:          claims.Email,
		Expires:        ctx.GetClock().Now().Add(timeoutExternalIdentityPending),
	}

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.WithError(err).Errorf("%s: %s", logFmtErrExternalIdentityCallback, errStrUserSessionDataSave)

		externalIdentityRedirect(ctx, pathExternalIdentityError)

		return
	}

	externalIdentityRedirect(ctx, pathExternalIdentityLinkedAccounts)
}
