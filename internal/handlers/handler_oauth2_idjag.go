// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"encoding/json"
	"time"

	oauthelia2 "authelia.com/provider/oauth2"
	fjwt "authelia.com/provider/oauth2/token/jwt"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func handleOAuth2IDJAGIssuePolicy(ctx *middlewares.AutheliaCtx, requester oauthelia2.AccessRequester, session *oidc.Session, details *authentication.UserDetailsExtended) (err error) {
	if details == nil {
		return oauthelia2.ErrInvalidGrant.WithHint("The subject of the 'subject_token' could not be resolved.")
	}

	ras, ok := lookupOAuth2IDJAGResourceAuthorizationServer(ctx, requester)
	if !ok {
		return oauthelia2.ErrServerError.WithDebugf("The granted audience '%v' does not identify a configured Resource Authorization Server.", requester.GetGrantedAudience())
	}

	if !ras.JWTDPoPSupported && ctx.Providers.OpenIDConnect.GetDPoPEnabled(ctx) && session.DefaultSession != nil && session.GetDPoPJWKThumbprint() != "" {
		return oauthelia2.ErrInvalidRequest.WithHint("The requested audience does not support DPoP bound grants.")
	}

	var client oidc.Client

	if client, err = ctx.Providers.OpenIDConnect.GetRegisteredClient(ctx, ras.SubjectClientID); err != nil {
		return oauthelia2.ErrServerError.WithWrap(err).WithDebugf("The subject client '%s' of the Resource Authorization Server '%s' could not be loaded.", ras.SubjectClientID, ras.Issuer)
	}

	return handleOAuth2SubjectTokenAuthorizationPolicy(ctx, client, session, details,
		"The authorization policy of the requested audience does not permit this user.",
		"The authentication level of the 'subject_token' is insufficient for the authorization policy of the requested audience.",
	)
}

func lookupOAuth2IDJAGResourceAuthorizationServer(ctx *middlewares.AutheliaCtx, requester oauthelia2.AccessRequester) (ras *oidc.IDJAGResourceAuthorizationServer, ok bool) {
	audience := requester.GetGrantedAudience()

	if len(audience) != 1 {
		return nil, false
	}

	return ctx.Providers.OpenIDConnect.IDJAG.LookupResourceAuthorizationServer(audience[0])
}

func addOAuth2IDJAGLogFields(ctx *middlewares.AutheliaCtx, fields map[string]any, requester oauthelia2.AccessRequester) {
	ras, ok := lookupOAuth2IDJAGResourceAuthorizationServer(ctx, requester)
	if !ok {
		return
	}

	fields["id_jag_audience"] = ras.Issuer

	client, ok := requester.GetClient().(oidc.Client)
	if !ok {
		return
	}

	if relationship, ok := client.GetIDJAGRelationship(ras.Issuer); ok {
		setOAuth2TokenExchangeLogField(fields, "id_jag_client_id", relationship.ClientID)
	}
}

func handleOAuth2IDJAGRedeemPolicy(ctx *middlewares.AutheliaCtx, requester oauthelia2.AccessRequester, client oidc.Client, session *oidc.Session) (details *authentication.UserDetailsExtended, err error) {
	if session.DefaultSession == nil || session.Claims == nil {
		return nil, oauthelia2.ErrServerError.WithDebug("The session of the Identity Assertion JWT Authorization Grant is missing required values.")
	}

	if details, err = handleOAuth2ResolveSessionSubject(ctx, requester, client, session,
		"Access Request encountered an error resolving the user of the Identity Assertion JWT Authorization Grant",
		"The subject of the Identity Assertion JWT Authorization Grant could not be resolved.",
	); err != nil {
		return nil, err
	}

	if err = handleOAuth2IDJAGRedeemAuthentication(ctx, requester, client, session); err != nil {
		return nil, err
	}

	if err = handleOAuth2SubjectTokenAuthorizationPolicy(ctx, client, session, details,
		"The client's authorization policy does not permit this user.",
		"The authentication level of the Identity Assertion JWT Authorization Grant is insufficient for the client's authorization policy.",
	); err != nil {
		return nil, err
	}

	if err = handleOAuth2ClientConsentModePolicy(ctx, requester, client, details, "The client requires explicit user consent which cannot be obtained when redeeming an Identity Assertion JWT Authorization Grant."); err != nil {
		return nil, err
	}

	session.Username = details.Username

	return details, nil
}

func handleOAuth2IDJAGRedeemAuthentication(ctx *middlewares.AutheliaCtx, requester oauthelia2.AccessRequester, client oidc.Client, session *oidc.Session) (err error) {
	claims := session.GetIDJAGClaims()

	issuer, _ := claims[oidc.ClaimIssuer].(string)

	trusted, ok := ctx.Providers.OpenIDConnect.IDJAG.LookupTrustedIssuer(issuer)
	if !ok {
		return oauthelia2.ErrServerError.WithDebugf("The issuer '%s' of the Identity Assertion JWT Authorization Grant is not a trusted issuer.", issuer)
	}

	var amr []string

	if amr, err = trusted.AMR.Evaluate(claims); err != nil {
		ctx.GetLogger().
			WithFields(map[string]any{"oauth2_access_request_id": requester.GetID(), "client_id": client.GetID(), "iss": issuer}).
			WithError(err).
			Error("Access Request encountered an error mapping the authentication methods of the Identity Assertion JWT Authorization Grant")

		return oauthelia2.ErrInvalidGrant.WithHint("The authentication methods of the Identity Assertion JWT Authorization Grant could not be mapped.")
	}

	session.ClientID = client.GetID()
	session.Claims.AuthorizedParty = client.GetID()
	session.Claims.Subject = session.Subject
	session.Claims.AuthenticationMethodsReferences = amr

	if authTime, ok := getOAuth2IDJAGNumericDate(claims[oidc.ClaimAuthenticationTime]); ok {
		session.Claims.AuthTime = fjwt.NewNumericDate(authTime)
	}

	return nil
}

func getOAuth2IDJAGNumericDate(value any) (date time.Time, ok bool) {
	switch v := value.(type) {
	case float64:
		return time.Unix(int64(v), 0).UTC(), true
	case int64:
		return time.Unix(v, 0).UTC(), true
	case json.Number:
		n, err := v.Int64()

		return time.Unix(n, 0).UTC(), err == nil
	default:
		return time.Time{}, false
	}
}

func isOAuth2IDJAGRedeemGrantType(grantType string) bool {
	return grantType == oidc.GrantTypeJWTBearer || grantType == oidc.GrantTypeJWTDPoP
}

//nolint:unparam // The responder keeps the signature in line with logOAuth2TokenExchange so both are called at the same points.
func logOAuth2IDJAGRedeem(ctx *middlewares.AutheliaCtx, requester oauthelia2.AccessRequester, responder oauthelia2.AccessResponder, err error) {
	if requester == nil || requester.GetClient() == nil || requester.GetClient().GetID() == "" {
		return
	}

	form := requester.GetRequestForm()

	if form == nil || !isOAuth2IDJAGRedeemGrantType(form.Get(oidc.FormParameterGrantType)) {
		return
	}

	session, _ := requester.GetSession().(*oidc.Session)

	fields := map[string]any{
		"oauth2_access_request_id": requester.GetID(),
		"client_id":                requester.GetClient().GetID(),
		"grant_type":               form.Get(oidc.FormParameterGrantType),
		"remote_ip":                ctx.RemoteIP().String(),
		"dpop_bound":               false,
	}

	for key, value := range map[string][]string{
		"granted_scope":    requester.GetGrantedScopes(),
		"granted_resource": requester.GetGrantedResource(),
	} {
		if len(value) != 0 {
			fields[key] = value
		}
	}

	if session != nil {
		claims := session.GetIDJAGClaims()

		for _, key := range []string{oidc.ClaimIssuer, oidc.ClaimJWTID, oidc.ClaimSubject} {
			if value, ok := claims[key].(string); ok {
				setOAuth2TokenExchangeLogField(fields, key, value)
			}
		}

		setOAuth2TokenExchangeLogField(fields, "username", session.Username)

		if session.DefaultSession != nil {
			setOAuth2TokenExchangeLogField(fields, "subject", session.Subject)

			fields["dpop_bound"] = session.GetDPoPJWKThumbprint() != ""
		}
	}

	entry := ctx.GetLogger().WithFields(fields)

	if err != nil {
		entry.WithError(oauthelia2.ErrorToDebugRFC6749Error(err)).Warn("Identity Assertion JWT Authorization Grant denied")

		return
	}

	entry.Info("Identity Assertion JWT Authorization Grant redeemed")
}
