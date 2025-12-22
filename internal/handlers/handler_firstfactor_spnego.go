package handlers

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/jcmturner/gofork/encoding/asn1"
	"github.com/jcmturner/gokrb5/v8/credentials"
	"github.com/jcmturner/gokrb5/v8/gssapi"
	"github.com/jcmturner/gokrb5/v8/spnego"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/session"
)

const (
	// spnegoNegTokenRespKRBAcceptCompleted - The response on successful authentication always has this header. Capturing as const so we don't have marshaling and encoding overhead.
	spnegoNegTokenRespKRBAcceptCompleted = "Negotiate oRQwEqADCgEAoQsGCSqGSIb3EgECAg==" // #nosec
	// spnegoNegTokenRespReject - The response on a failed authentication always has this rejection header. Capturing as const so we don't have marshaling and encoding overhead.
	spnegoNegTokenRespReject = "Negotiate oQcwBaADCgEC" // #nosec
	// spnegoNegTokenRespIncompleteKRB5 - Response token specifying incomplete context and KRB5 as the supported mechtype.
	spnegoNegTokenRespIncompleteKRB5 = "Negotiate oRQwEqADCgEBoQsGCSqGSIb3EgECAg==" // #nosec

	// HTTPHeaderAuthRequest is the header that will hold authn/z information.
	HTTPHeaderAuthRequest = "Authorization"
	// HTTPHeaderAuthResponse is the header that will hold SPNEGO data from the server.
	HTTPHeaderAuthResponse = "WWW-Authenticate"
	// HTTPHeaderAuthResponseValueKey is the key in the auth header for SPNEGO.
	HTTPHeaderAuthResponseValueKey = "Negotiate"
	ctxCredentials                 = "github.com/jcmturner/gokrb5/v8/ctxCredentials" // #nosec
)

func parseAuthorizationHeader(header []byte) (string, error) {
	s := strings.SplitN(string(header), " ", 2)
	if len(s) != 2 || s[0] != HTTPHeaderAuthResponseValueKey {
		return "", fmt.Errorf("missing or malformed %s header", HTTPHeaderAuthRequest)
	}

	return s[1], nil
}

func parseSPNEGOToken(header string) (*spnego.SPNEGOToken, error) {
	b, err := base64.StdEncoding.DecodeString(header)
	if err != nil {
		return nil, fmt.Errorf("error decoding base64 SPNEGO token: %w", err)
	}

	var st spnego.SPNEGOToken

	err = st.Unmarshal(b)
	if err != nil {
		// Check if this is a raw KRB5 context token - issue jcmturner/gokrb5 #347.
		var k5t spnego.KRB5Token

		err = k5t.Unmarshal(b)
		if err != nil {
			return nil, err
		}

		// Wrap it into an SPNEGO context token.
		st.Init = true
		st.NegTokenInit = spnego.NegTokenInit{
			MechTypes:      []asn1.ObjectIdentifier{k5t.OID},
			MechTokenBytes: b,
		}
	}

	return &st, nil
}

// FirstFactorPasskeyPOST handler completes the assertion ceremony after verifying the challenge.
//
//nolint:gocyclo
func FirstFactorSPNEGOPOST(ctx *middlewares.AutheliaCtx) {
	var (
		err      error
		bodyJSON = bodyFirstFactorKerberosRequest{}
	)

	if err = ctx.ParseBody(&bodyJSON); err != nil {
		respondUnauthorized(ctx, messageAuthenticationFailed)

		ctx.Logger.WithError(err).Errorf(logFmtErrParseRequestBody, regulation.AuthTypeKerberos)

		return
	}

	header, err := parseAuthorizationHeader(ctx.Request.Header.Peek(HTTPHeaderAuthRequest))
	if err != nil {
		// This NEEDS to be a 401 to trigger the client to send the ticket.
		ctx.SetStatusCode(fasthttp.StatusUnauthorized)
		ctx.Response.Header.Set(HTTPHeaderAuthResponse, HTTPHeaderAuthResponseValueKey)

		ctx.Logger.WithError(fmt.Errorf("missing or malformed %s header", HTTPHeaderAuthRequest)).Errorf(logFmtErrKerberosAuthenticationChallengeGenerate, errStrReqHeaderParse)

		doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, "", nil), regulation.AuthTypeKerberos, nil)

		return
	}

	token, err := parseSPNEGOToken(header)
	if err != nil {
		// This NEEDS to be a 401 to trigger the client to send the ticket.
		ctx.SetStatusCode(fasthttp.StatusUnauthorized)
		ctx.Response.Header.Set(HTTPHeaderAuthResponse, spnegoNegTokenRespIncompleteKRB5)

		ctx.Logger.WithError(fmt.Errorf("error parsing SPNEGO token: %w", err)).Errorf(logFmtErrKerberosAuthenticationChallengeGenerate, errStrReqHeaderParse)

		doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, "", nil), regulation.AuthTypeKerberos, nil)

		return
	}

	kerberosProvider, err := ctx.GetSPNEGOProvider()
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.Response.Header.Set(HTTPHeaderAuthResponse, spnegoNegTokenRespReject)
		ctx.SetJSONError(messageMFAValidationFailed)

		ctx.Logger.WithError(err).Errorf(logFmtErrKerberosAuthenticationChallengeGenerate, "error obtaining SPNEGO service")

		doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, "", nil), regulation.AuthTypeKerberos, nil)

		return
	}

	authed, context, status := kerberosProvider.AcceptSecContext(token)

	var (
		userDetails  *authentication.UserDetails
		kerberosUser = context.Value(ctxCredentials).(*credentials.Credentials)
		username     = kerberosUser.UserName()
	)

	if userDetails, err = ctx.Providers.UserProvider.GetDetails(username); err != nil || userDetails == nil {
		doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, "", nil), regulation.AuthTypeKerberos, err)

		ctx.Logger.WithError(err).Errorf("Error occurred getting details for user with username input '%s' which usually indicates they do not exist", username)

		respondUnauthorized(ctx, messageAuthenticationFailed)

		return
	}

	if ban, _, expires, err := ctx.Providers.Regulator.BanCheck(ctx, userDetails.Username); err != nil {
		if errors.Is(err, regulation.ErrUserIsBanned) {
			doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(ban, userDetails.Username, expires), regulation.AuthTypeKerberos, nil)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		ctx.Logger.WithError(err).Errorf(logFmtErrRegulationFail, regulation.AuthTypeKerberos, userDetails.Username)

		respondUnauthorized(ctx, messageAuthenticationFailed)

		return
	}

	if status.Code == gssapi.StatusContinueNeeded {
		// we need to continue the negotiation.
		ctx.Response.Header.Set(HTTPHeaderAuthResponse, spnegoNegTokenRespIncompleteKRB5)
		respondUnauthorized(ctx, messageAuthenticationFailed)

		return
	}

	if status.Code != gssapi.StatusComplete && status.Code != gssapi.StatusContinueNeeded || !authed {
		if isRegulatorSkippedErr(err) {
			ctx.Logger.WithError(err).Errorf("Unsuccessful %s authentication attempt by user '%s'", regulation.AuthTypeKerberos, userDetails.Username)
		} else {
			doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, userDetails.Username, nil), regulation.AuthTypeKerberos, err)
		}

		ctx.Response.Header.Set(HTTPHeaderAuthResponse, spnegoNegTokenRespReject)
		respondUnauthorized(ctx, messageAuthenticationFailed)

		return
	}

	doMarkAuthenticationAttempt(ctx, true, regulation.NewBan(regulation.BanTypeNone, userDetails.Username, nil), regulation.AuthTypeKerberos, nil)

	var provider *session.Session

	if provider, err = ctx.GetSessionProvider(); err != nil {
		ctx.Logger.WithError(err).Error("Failed to get session provider during 1FA attempt")

		respondUnauthorized(ctx, messageAuthenticationFailed)

		return
	}

	if err = provider.DestroySession(ctx.RequestCtx); err != nil {
		// This failure is not likely to be critical as we ensure to regenerate the session below.
		ctx.Logger.WithError(err).Trace("Failed to destroy session during 1FA attempt")
	}

	userSession := provider.NewDefaultUserSession()

	// Reset all values from previous session except OIDC workflow before regenerating the cookie.
	if err = provider.SaveSession(ctx.RequestCtx, userSession); err != nil {
		ctx.Logger.WithError(err).Errorf(logFmtErrSessionReset, regulation.AuthTypeKerberos, userDetails.Username)

		respondUnauthorized(ctx, messageAuthenticationFailed)

		return
	}

	if err = provider.RegenerateSession(ctx.RequestCtx); err != nil {
		ctx.Logger.WithError(err).Errorf(logFmtErrSessionRegenerate, regulation.AuthTypeKerberos, userDetails.Username)

		respondUnauthorized(ctx, messageAuthenticationFailed)

		return
	}

	keepMeLoggedIn := !provider.Config.DisableRememberMe && bodyJSON.KeepMeLoggedIn != nil && *bodyJSON.KeepMeLoggedIn

	// Set the cookie to expire if remember me is enabled and the user has asked us to.
	if keepMeLoggedIn {
		if err = provider.UpdateExpiration(ctx.RequestCtx, provider.Config.RememberMe); err != nil {
			ctx.SetStatusCode(fasthttp.StatusForbidden)
			ctx.SetJSONError(messageMFAValidationFailed)

			ctx.Logger.WithError(err).Errorf(logFmtErrSessionSave, "updated expiration", regulation.AuthTypePasskey, logFmtActionAuthentication, userDetails.Username)

			return
		}
	}

	ctx.Logger.Tracef(logFmtTraceProfileDetails, userDetails.Username, userDetails.Groups, userDetails.Emails)
	userSession.SetOneFactorKerberos(
		ctx.Clock.Now(),
		userDetails,
		keepMeLoggedIn,
	)

	if ctx.Configuration.AuthenticationBackend.RefreshInterval.Update() {
		userSession.RefreshTTL = ctx.Clock.Now().Add(ctx.Configuration.AuthenticationBackend.RefreshInterval.Value())

		if userSession.RefreshTTL.After(kerberosUser.ValidUntil()) {
			userSession.RefreshTTL = kerberosUser.ValidUntil()
		}
	}

	if err = provider.SaveSession(ctx.RequestCtx, userSession); err != nil {
		ctx.Logger.WithError(err).Errorf(logFmtErrSessionSave, "updated profile", regulation.AuthTypeKerberos, logFmtActionAuthentication, userDetails.Username)

		respondUnauthorized(ctx, messageAuthenticationFailed)

		return
	}

	ctx.Response.Header.Set(HTTPHeaderAuthResponse, spnegoNegTokenRespKRBAcceptCompleted)

	if len(bodyJSON.Flow) > 0 {
		handleFlowResponse(ctx, &userSession, bodyJSON.FlowID, bodyJSON.Flow, bodyJSON.SubFlow, bodyJSON.UserCode)
	} else {
		Handle1FAResponse(ctx, bodyJSON.TargetURL, bodyJSON.RequestMethod, userSession.Username, userSession.Groups)
	}
}
