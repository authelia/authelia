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
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/oidcrp"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/storage"
)

// FirstFactorOpenIDConnectCallbackGET is the redirect URI endpoint for external OpenID Connect 1.0 logins. It always
// responds with a redirect and never renders content controlled by the external provider.
//
//nolint:gocyclo // The function is a linear sequence of validations which is clearer than the sum of its parts.
func FirstFactorOpenIDConnectCallbackGET(ctx *middlewares.AutheliaCtx) {
	userSession, err := ctx.GetSession()
	if err != nil {
		ctx.Logger.WithError(err).Errorf("%s: %s", logFmtErrOpenIDConnectCallback, errStrUserSessionData)

		ctx.SpecialRedirect(pathRoot, fasthttp.StatusFound)

		return
	}

	// The flow state is single use. It is taken from the session and persisted as absent before any validation of the
	// callback occurs so a replayed callback cannot reuse it.
	flow := userSession.OpenIDConnect

	userSession.OpenIDConnect = nil

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.WithError(err).Errorf("%s: %s", logFmtErrOpenIDConnectCallback, errStrUserSessionDataSave)

		ctx.SpecialRedirect(pathRoot, fasthttp.StatusFound)

		return
	}

	id, _ := ctx.UserValue("provider").(string)

	log := ctx.Logger.WithField("provider", id)

	provider, ok := ctx.Providers.OpenIDConnectRelyingParty.Get(id)

	switch {
	case flow == nil:
		log.Errorf("%s: there is no flow in progress", logFmtErrOpenIDConnectCallback)

		ctx.SpecialRedirect(pathRoot, fasthttp.StatusFound)

		return
	case !ok:
		log.Errorf("%s: %s", logFmtErrOpenIDConnectCallback, errStrOpenIDConnectProviderUnknown)

		ctx.SpecialRedirect(pathRoot, fasthttp.StatusFound)

		return
	case flow.Provider != id:
		log.Errorf("%s: the flow in progress is for the provider '%s'", logFmtErrOpenIDConnectCallback, flow.Provider)

		ctx.SpecialRedirect(pathRoot, fasthttp.StatusFound)

		return
	case flow.State == "":
		// A flow state which carries no state value must never be compared against the callback: the comparison below
		// would match a callback which also omits the parameter.
		log.Errorf("%s: the flow in progress has no state", logFmtErrOpenIDConnectCallback)

		ctx.SpecialRedirect(pathRoot, fasthttp.StatusFound)

		return
	case ctx.GetClock().Now().After(flow.Expires):
		log.Errorf("%s: the flow has expired", logFmtErrOpenIDConnectCallback)

		ctx.SpecialRedirect(pathRoot, fasthttp.StatusFound)

		return
	}

	if subtle.ConstantTimeCompare(ctx.QueryArgs().Peek(queryArgState), []byte(flow.State)) != 1 {
		log.Errorf("%s: the state does not match", logFmtErrOpenIDConnectCallback)

		ctx.SpecialRedirect(pathRoot, fasthttp.StatusFound)

		return
	}

	if issuer := string(ctx.QueryArgs().Peek(queryArgISS)); issuer != "" && issuer != provider.Issuer {
		log.Errorf("%s: the 'iss' parameter does not match the configured issuer", logFmtErrOpenIDConnectCallback)

		ctx.SpecialRedirect(pathRoot, fasthttp.StatusFound)

		return
	}

	if e := string(ctx.QueryArgs().Peek(queryArgError)); e != "" {
		log.WithField("error", e).Errorf("%s: the provider returned an error", logFmtErrOpenIDConnectCallback)

		ctx.SpecialRedirect(pathRoot, fasthttp.StatusFound)

		return
	}

	code := string(ctx.QueryArgs().Peek(queryArgCode))

	if code == "" {
		log.Errorf("%s: the authorization code is absent", logFmtErrOpenIDConnectCallback)

		ctx.SpecialRedirect(pathRoot, fasthttp.StatusFound)

		return
	}

	var redirectURI string

	if redirectURI, err = oidcRelyingPartyRedirectURI(ctx, provider.ID); err != nil {
		log.WithError(err).Error(logFmtErrOpenIDConnectCallback)

		ctx.SpecialRedirect(pathRoot, fasthttp.StatusFound)

		return
	}

	var raw string

	if raw, err = provider.Exchange(ctx, code, flow.CodeVerifier, redirectURI); err != nil {
		log.WithError(err).Error(logFmtErrOpenIDConnectCallback)

		ctx.SpecialRedirect(pathRoot, fasthttp.StatusFound)

		return
	}

	var claims *oidcrp.IdentityClaims

	if claims, err = provider.ValidateIDToken(ctx, raw, flow.Nonce, ctx.GetClock().Now()); err != nil {
		log.WithError(err).Error(logFmtErrOpenIDConnectCallback)

		ctx.SpecialRedirect(pathRoot, fasthttp.StatusFound)

		return
	}

	// The account link is looked up by the issuer and subject pair alone. The 'email' claim is not verified by Authelia
	// and must never be used to find, match, or link an account.
	link, err := ctx.Providers.StorageProvider.LoadOpenIDConnectLinkBySubject(ctx, claims.Issuer, claims.Subject)

	switch {
	case err == nil:
		handleOpenIDConnectCallbackLinked(ctx, provider, flow, claims, link.ID, link.Username)
	case errors.Is(err, storage.ErrNoOpenIDConnectLink):
		handleOpenIDConnectCallbackUnlinked(ctx, provider, claims)
	default:
		log.WithError(err).Error(logFmtErrOpenIDConnectCallback)

		ctx.SpecialRedirect(pathRoot, fasthttp.StatusFound)
	}
}

// handleOpenIDConnectCallbackLinked completes the first factor authentication for an external identity which is
// already linked to a local account.
func handleOpenIDConnectCallbackLinked(ctx *middlewares.AutheliaCtx, provider *oidcrp.Provider, flow *session.OpenIDConnectFlow, claims *oidcrp.IdentityClaims, id int, username string) {
	log := ctx.Logger.WithFields(map[string]any{"provider": provider.ID, "username": username})

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

		ctx.SpecialRedirect(pathRoot, fasthttp.StatusFound)

		return
	}

	var details *authentication.UserDetails

	if details, err = ctx.Providers.UserProvider.GetDetails(username); err != nil || details == nil {
		doMarkAuthenticationAttemptWithRequest(ctx, false, regulation.NewBan(regulation.BanTypeNone, username, nil), regulation.AuthType1FA, flow.TargetURL, flow.RequestMethod, err)

		ctx.SpecialRedirect(pathRoot, fasthttp.StatusFound)

		return
	}

	var sessionProvider *session.Session

	if sessionProvider, err = oidcRelyingPartyResetSession(ctx, log, username); err != nil {
		ctx.SpecialRedirect(pathRoot, fasthttp.StatusFound)

		return
	}

	// The remember me policy is applied here rather than when the flow was started as the configuration in force when
	// the session is actually created is the one which must be honored. The value stored in the flow state is only
	// the users request; it is not authoritative.
	keepMeLoggedIn := !sessionProvider.Config.DisableRememberMe && flow.KeepMeLoggedIn

	if keepMeLoggedIn {
		if err = sessionProvider.UpdateExpiration(ctx.RequestCtx, sessionProvider.Config.RememberMe); err != nil {
			log.WithError(err).Errorf(logFmtErrSessionSave, "updated expiration", regulation.AuthType1FA, logFmtActionAuthentication, username)

			ctx.SpecialRedirect(pathRoot, fasthttp.StatusFound)

			return
		}
	}

	var userSession session.UserSession

	if userSession, err = ctx.GetSession(); err != nil {
		log.WithError(err).Errorf("%s: %s", logFmtErrOpenIDConnectCallback, errStrUserSessionData)

		ctx.SpecialRedirect(pathRoot, fasthttp.StatusFound)

		return
	}

	userSession.SetOneFactorOpenIDConnect(ctx.GetClock().Now(), details, keepMeLoggedIn)

	// Authelia does not observe the authentication performed at the external provider. The provider's asserted
	// Authentication Method Reference values are only adopted when the administrator has configured Authelia to trust
	// them for this provider.
	if provider.TrustAuthenticationMethodsReference {
		userSession.AuthenticationMethodRefs = userSession.AuthenticationMethodRefs.Merge(authorization.NewAuthenticationMethodsReferencesFromClaim(claims.AuthenticationMethodsReference))
	}

	if ctx.Configuration.AuthenticationBackend.RefreshInterval.Update() {
		userSession.RefreshTTL = ctx.GetClock().Now().Add(ctx.Configuration.AuthenticationBackend.RefreshInterval.Value())
	}

	if err = ctx.SaveSession(userSession); err != nil {
		log.WithError(err).Errorf("%s: %s", logFmtErrOpenIDConnectCallback, errStrUserSessionDataSave)

		ctx.SpecialRedirect(pathRoot, fasthttp.StatusFound)

		return
	}

	if err = ctx.Providers.StorageProvider.UpdateOpenIDConnectLinkSignIn(ctx, id, ctx.GetClock().Now()); err != nil {
		log.WithError(err).Warnf("%s: the sign in could not be recorded against the link", logFmtErrOpenIDConnectCallback)
	}

	doMarkAuthenticationAttemptWithRequest(ctx, true, regulation.NewBan(regulation.BanTypeNone, username, nil), regulation.AuthType1FA, flow.TargetURL, flow.RequestMethod, nil)

	// The target URL is attacker influenced as it originates from the request which started the flow. It must be
	// validated by the safe redirection rules before it is used as a redirect location.
	Handle1FARedirect(ctx, flow.TargetURL, flow.RequestMethod, userSession.Username, userSession.Groups)
}

// oidcRelyingPartyResetSession discards every value held by the current session, then regenerates the session cookie.
// The username resolved from the external identity is not necessarily the username the session currently holds, so
// nothing may survive this change of identity: in particular the Authentication Method Reference values, the
// elevation, and any pending external identity proposal must not be inherited by the account being authenticated.
func oidcRelyingPartyResetSession(ctx *middlewares.AutheliaCtx, log *logrus.Entry, username string) (provider *session.Session, err error) {
	if provider, err = ctx.GetSessionProvider(); err != nil {
		log.WithError(err).Errorf("%s: %s", logFmtErrOpenIDConnectCallback, errStrSessionProvider)

		return nil, err
	}

	if err = provider.DestroySession(ctx.RequestCtx); err != nil {
		// This failure is not likely to be critical as the session is reset and the cookie regenerated below.
		log.WithError(err).Trace("Failed to destroy session during the external OpenID Connect 1.0 callback")
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

// handleOpenIDConnectCallbackUnlinked stores the validated external identity as a proposal the user may accept from
// their settings. It does not authenticate the user.
func handleOpenIDConnectCallbackUnlinked(ctx *middlewares.AutheliaCtx, provider *oidcrp.Provider, claims *oidcrp.IdentityClaims) {
	userSession, err := ctx.GetSession()
	if err != nil {
		ctx.Logger.WithError(err).Errorf("%s: %s", logFmtErrOpenIDConnectCallback, errStrUserSessionData)

		ctx.SpecialRedirect(pathRoot, fasthttp.StatusFound)

		return
	}

	// No identity bearing value is carried across the authentication boundary. When the session is anonymous the
	// validated identity is discarded entirely and only the provider identifier is handed back to the portal so the
	// flow can be performed again by the user once they have authenticated. An attacker who can plant a session cookie
	// on the Authelia domain therefore cannot plant a proposal for their own external identity: the only value they
	// can influence is the provider identifier, which merely offers the victim a link to the victims own account.
	if userSession.IsAnonymous() {
		ctx.Logger.WithField("provider", provider.ID).Info("The external OpenID Connect 1.0 identity is not linked to a local account and the session is not authenticated: the identity has been discarded and the user must authenticate before the link can be proposed")

		userSession.OpenIDConnectPending = nil

		if err = ctx.SaveSession(userSession); err != nil {
			ctx.Logger.WithError(err).Errorf("%s: %s", logFmtErrOpenIDConnectCallback, errStrUserSessionDataSave)

			ctx.SpecialRedirect(pathRoot, fasthttp.StatusFound)

			return
		}

		ctx.SpecialRedirect(pathRoot+"?"+queryArgLinkProvider+"="+url.QueryEscape(provider.ID), fasthttp.StatusFound)

		return
	}

	// The 'email' claim is stored for display purposes only. Authelia does not consult the 'email_verified' claim and
	// must never use this value to find, match, or link an account.
	userSession.OpenIDConnectPending = &session.OpenIDConnectPending{
		Provider:       provider.ID,
		Issuer:         claims.Issuer,
		Subject:        claims.Subject,
		RemoteUsername: claims.PreferredUsername,
		DisplayName:    claims.Name,
		Email:          claims.Email,
		Expires:        ctx.GetClock().Now().Add(timeoutOpenIDConnectPending),
	}

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.WithError(err).Errorf("%s: %s", logFmtErrOpenIDConnectCallback, errStrUserSessionDataSave)

		ctx.SpecialRedirect(pathRoot, fasthttp.StatusFound)

		return
	}

	ctx.SpecialRedirect(pathOpenIDConnectLinkedAccounts, fasthttp.StatusFound)
}
