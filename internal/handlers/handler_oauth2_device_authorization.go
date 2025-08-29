package handlers

import (
	"errors"
	"net/http"
	"strings"

	oauthelia2 "authelia.com/provider/oauth2"
	"authelia.com/provider/oauth2/x/errorsx"
	"github.com/google/uuid"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
)

// OAuth2DeviceAuthorizationPOST handles the Device Code Flow of the the Device Authorization Flow.
func OAuth2DeviceAuthorizationPOST(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, r *http.Request) {
	var (
		requester oauthelia2.DeviceAuthorizeRequester
		response  oauthelia2.DeviceAuthorizeResponder

		err error
	)
	if requester, err = ctx.Providers.OpenIDConnect.NewRFC862DeviceAuthorizeRequest(ctx, r); err != nil {
		ctx.Logger.
			WithError(oauthelia2.ErrorToDebugRFC6749Error(err)).
			Error("Device Authorization Request failed with error during the Device Authorization Flow")

		errorsx.WriteJSONError(rw, r, err)

		return
	}

	log := ctx.Logger.WithFields(map[string]any{logging.FieldRequestID: requester.GetID(), logging.FieldClientID: requester.GetClient().GetID(), logging.FieldScope: strings.Join(requester.GetRequestedScopes(), " ")})

	log.Debug("Device Authorization Request is processing the Device Authorization Flow")

	if response, err = ctx.Providers.OpenIDConnect.NewRFC862DeviceAuthorizeResponse(ctx, requester, oidc.NewSessionWithRequestedAt(ctx.GetClock().Now())); err != nil {
		log.WithError(oauthelia2.ErrorToDebugRFC6749Error(err)).Error("Device Authorization Request had an error while trying to create a response during the Device Authorization Flow")

		errorsx.WriteJSONError(rw, r, err)

		return
	}

	log.Debug("Device Authorization Request has successfully processed the Device Authorization Flow")

	ctx.Providers.OpenIDConnect.WriteRFC862DeviceAuthorizeResponse(ctx, rw, requester, response)
}

// OAuth2DeviceAuthorizationPUT handles the User Code Flow of the the Device Authorization Flow.
func OAuth2DeviceAuthorizationPUT(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, r *http.Request) {
	var (
		requester oauthelia2.DeviceAuthorizeRequester
		responder oauthelia2.DeviceUserAuthorizeResponder
		flowID    uuid.UUID
		client    oidc.Client
		consent   *model.OAuth2ConsentSession

		err error
	)
	if requester, err = ctx.Providers.OpenIDConnect.NewRFC8628UserAuthorizeRequest(ctx, r); err != nil {
		ctx.Logger.
			WithError(oauthelia2.ErrorToDebugRFC6749Error(err)).
			Error("Device Authorization Request failed with error during the User Authorization Flow")

		ctx.Providers.OpenIDConnect.WriteRFC8628UserAuthorizeError(ctx, rw, requester, err)

		return
	}

	log := ctx.Logger.WithFields(map[string]any{logging.FieldRequestID: requester.GetID(), logging.FieldClientID: requester.GetClient().GetID()})

	log.Debug("Device Authorization Request is processing the User Authorization Flow")

	if flowID, err = uuid.Parse(requester.GetRequestForm().Get(oidc.FormParameterFlowID)); err != nil {
		log.
			WithError(oauthelia2.ErrorToDebugRFC6749Error(err)).
			Error("Device Authorization Request failed with error to parse the flow ID during the User Authorization Flow")

		ctx.Providers.OpenIDConnect.WriteRFC8628UserAuthorizeError(ctx, rw, requester, oauthelia2.ErrInvalidRequest)

		return
	}

	log = log.WithField(logging.FieldFlowID, flowID)

	if consent, err = ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionByChallengeID(ctx, flowID); err != nil {
		log.
			WithError(oauthelia2.ErrorToDebugRFC6749Error(err)).
			Error("Device Authorization Request failed with error to load the consent session during the User Authorization Flow")

		ctx.Providers.OpenIDConnect.WriteRFC8628UserAuthorizeError(ctx, rw, requester, oauthelia2.ErrServerError)

		return
	}

	if client, err = ctx.Providers.OpenIDConnect.GetRegisteredClient(ctx, requester.GetClient().GetID()); err != nil {
		if errors.Is(err, oauthelia2.ErrNotFound) {
			log.
				WithError(oauthelia2.ErrorToDebugRFC6749Error(err)).
				Error("Device Authorization Request failed to find client during the User Authorization Flow")
		} else {
			log.
				WithError(oauthelia2.ErrorToDebugRFC6749Error(err)).
				Error("Device Authorization Request failed to find client due to an unknown error during the User Authorization Flow")
		}

		ctx.Providers.OpenIDConnect.WriteRFC8628UserAuthorizeError(ctx, rw, requester, err)

		return
	}

	var (
		userSession session.UserSession
		handled     bool
	)

	if userSession, err = ctx.GetSession(); err != nil {
		log.
			WithError(oauthelia2.ErrorToDebugRFC6749Error(err)).
			Error("Device Authorization Request failed to obtain the user session during the User Authorization Flow")

		ctx.Providers.OpenIDConnect.WriteRFC8628UserAuthorizeError(ctx, rw, requester, oauthelia2.ErrServerError.WithHint("Could not obtain the user session."))

		return
	}

	log = log.WithField(logging.FieldUsername, userSession.Username)

	level := userSession.AuthenticationLevel(ctx.Configuration.WebAuthn.EnablePasskey2FA)

	if !client.IsAuthenticationLevelSufficient(level, authorization.Subject{Username: userSession.Username, Groups: userSession.Groups, IP: ctx.RemoteIP()}) {
		log.
			WithFields(map[string]any{logging.FieldAuthenticationLevel: level.String(), logging.FieldGroups: userSession.Groups, logging.FieldAuthorizationPolicy: client.GetAuthorizationPolicy().Name}).
			Error("Device Authorization Request failed as the user did not satisfy the client authorization policy during the User Authorization Flow")

		ctx.Providers.OpenIDConnect.WriteRFC8628UserAuthorizeError(ctx, rw, requester, oauthelia2.ErrServerError.WithHint("Could not authorize the user."))

		return
	}

	var (
		subject uuid.UUID
	)

	if subject, err = ctx.Providers.OpenIDConnect.GetSubject(ctx, client.GetSectorIdentifierURI(), userSession.Username); err != nil {
		log.
			WithError(oauthelia2.ErrorToDebugRFC6749Error(err)).
			Error("Device Authorization Request failed to obtain the user subject value for the user during the User Authorization Flow")

		ctx.Providers.OpenIDConnect.WriteRFC8628UserAuthorizeError(ctx, rw, requester, oauthelia2.ErrServerError.WithHint("Could not determine the subject value for the user."))

		return
	}

	if subject != consent.Subject.UUID {
		log.Error("Device Authorization Request failed to match the session to the subject during the User Authorization Flow")

		ctx.Providers.OpenIDConnect.WriteRFC8628UserAuthorizeError(ctx, rw, requester, oauthelia2.ErrServerError.WithHint("Could not match the consent session to the subject."))

		return
	}

	issuer := ctx.RootURL()

	var details *authentication.UserDetailsExtended

	if details, err = ctx.Providers.UserProvider.GetDetailsExtended(userSession.Username); err != nil {
		log.
			WithError(oauthelia2.ErrorToDebugRFC6749Error(err)).
			Error("Device Authorization Request failed to obtain the user details during the User Authorization Flow")

		ctx.Providers.OpenIDConnect.WriteRFC8628UserAuthorizeError(ctx, rw, requester, oauthelia2.ErrServerError.WithHint("Could not obtain the users details."))

		return
	}

	var requests *oidc.ClaimsRequests

	extra := map[string]any{}

	if requests, handled = handleOAuth2AuthorizationClaims(ctx, rw, r, "Device Authorization", userSession, details, client, requester, issuer, consent, extra); handled {
		return
	}

	session := oidc.NewSessionWithRequester(ctx, issuer, ctx.Providers.OpenIDConnect.Issuer.GetKeyID(ctx, client.GetIDTokenSignedResponseKeyID(), client.GetIDTokenSignedResponseAlg()), details.Username, userSession.AuthenticationMethodRefs.MarshalRFC8176(), extra, userSession.LastAuthenticatedTime(), consent, requester, requests)

	if client.GetClaimsStrategy().MergeAccessTokenAudienceWithIDTokenAudience() {
		session.Claims.Audience = append([]string{client.GetID()}, requester.GetGrantedAudience()...)
	}

	requester.SetStatus(oauthelia2.DeviceAuthorizeStatusApproved)

	if responder, err = ctx.Providers.OpenIDConnect.NewRFC8628UserAuthorizeResponse(ctx, requester, session); err != nil {
		log.
			WithError(oauthelia2.ErrorToDebugRFC6749Error(err)).
			Error("Device Authorization Request had an error while attempting to generate the responder during the User Authorization Flow")

		ctx.Providers.OpenIDConnect.WriteRFC8628UserAuthorizeError(ctx, rw, requester, err)

		return
	}

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSessionGranted(ctx, consent.ID); err != nil {
		log.
			WithError(oauthelia2.ErrorToDebugRFC6749Error(err)).
			Error("Device Authorization Request had an error while saving the session during the User Authorization Flow")

		ctx.Providers.OpenIDConnect.WriteRFC8628UserAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotSave)

		return
	}

	log.Debug("Device Authorization Request was successfully processed during the User Authorization Flow")

	ctx.Providers.OpenIDConnect.WriteRFC8628UserAuthorizeResponse(ctx, rw, requester, responder)
}
