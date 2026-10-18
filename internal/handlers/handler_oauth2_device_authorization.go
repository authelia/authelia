// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	oauthelia2 "authelia.com/provider/oauth2"
	"authelia.com/provider/oauth2/x/errorsx"

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

	if _, err = ctx.IssuerURL(); err != nil {
		rfc := oidc.ErrEffectiveIssuer.WithWrap(err)

		ctx.GetLogger().WithError(err).Errorf("Device Authorization Request could not be processed: %s", oauthelia2.ErrorToDebugRFC6749Error(rfc))

		errorsx.WriteJSONError(rw, r, rfc)

		return
	}

	if requester, err = ctx.Providers.OpenIDConnect.NewRFC862DeviceAuthorizeRequest(ctx, r); err != nil {
		ctx.GetLogger().
			WithError(oauthelia2.ErrorToDebugRFC6749Error(err)).
			Error("Device Authorization Request failed with error during the Device Authorization Flow")

		errorsx.WriteJSONError(rw, r, err)

		return
	}

	oidc.PopulateRequestedResourceImplicit(requester)

	log := ctx.GetLogger().WithFields(map[string]any{logging.FieldRequestID: requester.GetID(), logging.FieldClientID: requester.GetClient().GetID(), logging.FieldScope: strings.Join(requester.GetRequestedScopes(), " ")})

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
//
//nolint:gocyclo
func OAuth2DeviceAuthorizationPUT(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, r *http.Request) {
	var (
		issuer    *url.URL
		requester oauthelia2.DeviceAuthorizeRequester
		responder oauthelia2.DeviceUserAuthorizeResponder
		flowID    uuid.UUID
		client    oidc.Client
		consent   *model.OAuth2ConsentSession

		err error
	)

	if issuer, err = ctx.IssuerURL(); err != nil {
		rfc := oidc.ErrEffectiveIssuer.WithWrap(err)

		ctx.GetLogger().WithError(err).Errorf("Device Authorization Request during the User Authorization Flow could not be processed: %s", oauthelia2.ErrorToDebugRFC6749Error(rfc))

		ctx.Providers.OpenIDConnect.WriteRFC8628UserAuthorizeError(ctx, rw, requester, rfc)

		return
	}

	if requester, err = ctx.Providers.OpenIDConnect.NewRFC8628UserAuthorizeRequest(ctx, r, oidc.NewSession()); err != nil {
		ctx.GetLogger().
			WithError(oauthelia2.ErrorToDebugRFC6749Error(err)).
			Error("Device Authorization Request failed with error during the User Authorization Flow")

		ctx.Providers.OpenIDConnect.WriteRFC8628UserAuthorizeError(ctx, rw, requester, err)

		return
	}

	log := ctx.GetLogger().WithFields(map[string]any{logging.FieldRequestID: requester.GetID(), logging.FieldClientID: requester.GetClient().GetID()})

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

	if userSession.IsAnonymous() {
		log.Error("Device Authorization Request failed as the user is anonymous during the User Authorization Flow")

		ctx.Providers.OpenIDConnect.WriteRFC8628UserAuthorizeError(ctx, rw, requester, oidc.ErrClientAuthorizationUserAccessDenied)

		return
	}

	log = log.WithField(logging.FieldUsername, userSession.Username)

	if handled = handleOAuth2DeviceAuthorizationSessionRefresh(ctx, rw, requester, &userSession, log); handled {
		return
	}

	level := userSession.AuthenticationLevel(ctx.Configuration.WebAuthn.EnablePasskey2FA)
	subjectAuthz := authorization.Subject{Username: userSession.Username, Groups: userSession.Groups, IP: ctx.RemoteIP()}

	if client.GetAuthorizationPolicyRequiredLevel(subjectAuthz) == authorization.Denied {
		log.
			WithFields(map[string]any{logging.FieldGroups: userSession.Groups, logging.FieldAuthorizationPolicy: client.GetAuthorizationPolicy().Name}).
			Error("Device Authorization Request failed as the user is not authorized to use the client during the User Authorization Flow")

		ctx.Providers.OpenIDConnect.WriteRFC8628UserAuthorizeError(ctx, rw, requester, oidc.ErrClientAuthorizationUserAccessDenied)

		return
	}

	if !client.IsAuthenticationLevelSufficient(level, subjectAuthz) {
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

	if handled = handleOAuth2DeviceAuthorizationConsent(ctx, rw, requester, consent, flowID, subject, log); handled {
		return
	}

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
		session.Claims.Audience = append([]string{client.GetID()}, oauthelia2.JoinGrantedAudienceAndResource(requester.GetGrantedAudience(), requester.GetGrantedResource())...)
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

func handleOAuth2DeviceAuthorizationConsent(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, requester oauthelia2.DeviceAuthorizeRequester, consent *model.OAuth2ConsentSession, flowID, subject uuid.UUID, log *logrus.Entry) (handled bool) {
	var (
		device *model.OAuth2DeviceCodeSession
		err    error
	)

	log = log.WithField(logging.FieldSessionID, consent.ID)

	if !consent.Subject.Valid || subject != consent.Subject.UUID {
		log.Error("Device Authorization Request failed to match the session to the subject during the User Authorization Flow")

		ctx.Providers.OpenIDConnect.WriteRFC8628UserAuthorizeError(ctx, rw, requester, oauthelia2.ErrServerError.WithHint("Could not match the consent session to the subject."))

		return true
	}

	if device, err = ctx.Providers.StorageProvider.LoadOAuth2DeviceCodeSessionByUserCode(ctx, requester.GetUserCodeSignature()); err != nil {
		log.
			WithError(err).
			Error("Device Authorization Request failed with error to load the device code session during the User Authorization Flow")

		ctx.Providers.OpenIDConnect.WriteRFC8628UserAuthorizeError(ctx, rw, requester, oauthelia2.ErrServerError)

		return true
	}

	if consent.ChallengeID != flowID || !device.ChallengeID.Valid || device.ChallengeID.UUID != consent.ChallengeID {
		log.
			WithField("device_flow_id", device.ChallengeID.UUID.String()).
			Error("Device Authorization Request failed to match the consent session to the flow of the device code session during the User Authorization Flow")

		ctx.Providers.OpenIDConnect.WriteRFC8628UserAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotPerform)

		return true
	}

	if !device.Subject.Valid || device.Subject.String != consent.Subject.UUID.String() {
		log.Error("Device Authorization Request failed to match the consent session to the subject of the device code session during the User Authorization Flow")

		ctx.Providers.OpenIDConnect.WriteRFC8628UserAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotPerform)

		return true
	}

	if err = consent.MatchesRequester(requester, ""); err != nil {
		log.
			WithError(oauthelia2.ErrorToDebugRFC6749Error(err)).
			Error("Device Authorization Request failed to match the consent session to the request during the User Authorization Flow")

		ctx.Providers.OpenIDConnect.WriteRFC8628UserAuthorizeError(ctx, rw, requester, err)

		return true
	}

	if !consent.CanGrant(ctx.GetClock().Now()) {
		log.
			WithFields(map[string]any{logging.FieldExpiration: consent.ExpiresAt.Unix(), "granted": consent.Granted}).
			Error("Device Authorization Request failed as the consent session has either expired or already been granted during the User Authorization Flow")

		ctx.Providers.OpenIDConnect.WriteRFC8628UserAuthorizeError(ctx, rw, requester, oidc.ErrConsentCouldNotPerform)

		return true
	}

	if !consent.IsAuthorized() {
		log.
			WithField("responded", consent.Responded()).
			Error("Device Authorization Request failed as the consent session was not authorized by the user during the User Authorization Flow")

		ctx.Providers.OpenIDConnect.WriteRFC8628UserAuthorizeError(ctx, rw, requester, oauthelia2.ErrAccessDenied)

		return true
	}

	return false
}

// handleOAuth2DeviceAuthorizationSessionRefresh refreshes the user session from the authentication backend when the
// refresh interval requires it, so a user who was deleted or disabled can't approve the request, and the client
// authorization policy is evaluated against their current groups.
func handleOAuth2DeviceAuthorizationSessionRefresh(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, requester oauthelia2.DeviceAuthorizeRequester, userSession *session.UserSession, log *logrus.Entry) (handled bool) {
	var err error

	modified, invalid := handleSessionValidateRefresh(ctx, userSession, ctx.Configuration.AuthenticationBackend.RefreshInterval)

	switch {
	case invalid:
		if err = ctx.DestroySession(); err != nil {
			log.
				WithError(err).
				Error("Device Authorization Request had an error while destroying the user session during the User Authorization Flow")
		}

		log.Error("Device Authorization Request failed as the user is no longer authorized to login during the User Authorization Flow")

		ctx.Providers.OpenIDConnect.WriteRFC8628UserAuthorizeError(ctx, rw, requester, oidc.ErrClientAuthorizationUserAccessDenied)

		return true
	case modified:
		if err = ctx.SaveSession(*userSession); err != nil {
			log.
				WithError(err).
				Error("Device Authorization Request had an error while saving the updated user session during the User Authorization Flow")

			ctx.Providers.OpenIDConnect.WriteRFC8628UserAuthorizeError(ctx, rw, requester, oauthelia2.ErrServerError.WithHint("Could not save the user session."))

			return true
		}
	}

	return false
}
