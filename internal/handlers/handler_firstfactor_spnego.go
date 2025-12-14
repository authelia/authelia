package handlers

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/jcmturner/gokrb5/v8/credentials"
	"github.com/jcmturner/gokrb5/v8/gssapi"
	"github.com/jcmturner/gokrb5/v8/spnego"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/jcmturner/gofork/encoding/asn1"
)

const (
	// spnegoNegTokenRespKRBAcceptCompleted - The response on successful authentication always has this header. Capturing as const so we don't have marshaling and encoding overhead.
	spnegoNegTokenRespKRBAcceptCompleted = "Negotiate oRQwEqADCgEAoQsGCSqGSIb3EgECAg=="
	// spnegoNegTokenRespReject - The response on a failed authentication always has this rejection header. Capturing as const so we don't have marshaling and encoding overhead.
	spnegoNegTokenRespReject = "Negotiate oQcwBaADCgEC"
	// spnegoNegTokenRespIncompleteKRB5 - Response token specifying incomplete context and KRB5 as the supported mechtype.
	spnegoNegTokenRespIncompleteKRB5 = "Negotiate oRQwEqADCgEBoQsGCSqGSIb3EgECAg=="

	// HTTPHeaderAuthRequest is the header that will hold authn/z information.
	HTTPHeaderAuthRequest = "Authorization"
	// HTTPHeaderAuthResponse is the header that will hold SPNEGO data from the server.
	HTTPHeaderAuthResponse = "WWW-Authenticate"
	// HTTPHeaderAuthResponseValueKey is the key in the auth header for SPNEGO.
	HTTPHeaderAuthResponseValueKey = "Negotiate"
	ctxCredentials                 = "github.com/jcmturner/gokrb5/v8/ctxCredentials"
)

// FirstFactorPasskeyPOST handler completes the assertion ceremony after verifying the challenge.
//
//nolint:gocyclo
func FirstFactorSPNEGOPOST(ctx *middlewares.AutheliaCtx) {
	var (
		provider    *session.Session
		userSession session.UserSession
		err         error

		bodyJSON bodySignKerberosRequest
	)
	if provider, err = ctx.GetSessionProvider(); err != nil {
		ctx.Logger.WithError(err).Errorf(logFmtErrKerberosAuthenticationChallengeGenerate, errStrUserSessionData)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, "", nil), regulation.AuthTypePasskey, nil)

		return
	}

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		ctx.Logger.WithError(err).Errorf(logFmtErrKerberosAuthenticationChallengeGenerate, errStrUserSessionData)

		return
	}

	if !userSession.IsAnonymous() {
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		ctx.Logger.WithError(errUserIsAlreadyAuthenticated).Error("Error occurred validating a WebAuthn passkey authentication challenge")

		doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, "", nil), regulation.AuthTypePasskey, nil)

		return
	}

	if err = ctx.ParseBody(&bodyJSON); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(messageMFAValidationFailed)

		ctx.Logger.WithError(err).Errorf(logFmtErrKerberosAuthenticationChallengeGenerate, errStrReqBodyParse)

		doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, "", nil), regulation.AuthTypePasskey, nil)

		return
	}

	s := strings.SplitN(string(ctx.Request.Header.Peek(HTTPHeaderAuthRequest)), " ", 2)
	if len(s) != 2 || s[0] != HTTPHeaderAuthResponseValueKey {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.Response.Header.Set(HTTPHeaderAuthResponse, HTTPHeaderAuthResponseValueKey)

		ctx.Logger.WithError(fmt.Errorf("missing or malformed %s header", HTTPHeaderAuthRequest)).Errorf(logFmtErrKerberosAuthenticationChallengeGenerate, errStrReqHeaderParse)

		doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, "", nil), regulation.AuthTypePasskey, nil)

		return
	}

	b, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.Response.Header.Set(HTTPHeaderAuthResponse, HTTPHeaderAuthResponseValueKey)

		ctx.Logger.WithError(fmt.Errorf("missing or malformed %s header", HTTPHeaderAuthRequest)).Errorf(logFmtErrKerberosAuthenticationChallengeGenerate, errStrReqHeaderParse)

		doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, "", nil), regulation.AuthTypePasskey, nil)

		return
	}

	var st spnego.SPNEGOToken
	err = st.Unmarshal(b)
	if err != nil {
		ctx.Logger.WithError(err).Errorf(logFmtErrKerberosAuthenticationChallengeGenerate, "error performing the login validation")
		// Check if this is a raw KRB5 context token - issue #347
		var k5t spnego.KRB5Token
		err = k5t.Unmarshal(b)
		if err != nil {
			ctx.SetStatusCode(fasthttp.StatusForbidden)
			ctx.SetJSONError(messageMFAValidationFailed)

			ctx.Logger.WithError(err).Errorf(logFmtErrKerberosAuthenticationChallengeGenerate, "error performing the login validation")

			doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, "", nil), regulation.AuthTypeKerberos, nil)

			return
		}

		// Wrap it into an SPNEGO context token
		st.Init = true
		st.NegTokenInit = spnego.NegTokenInit{
			MechTypes:      []asn1.ObjectIdentifier{k5t.OID},
			MechTokenBytes: b,
		}
	}

	kerberosProvider, err := ctx.GetSPNEGOProvider()
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		ctx.Logger.WithError(err).Errorf(logFmtErrKerberosAuthenticationChallengeGenerate, "error occurred provisioning the configuration")

		doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, "", nil), regulation.AuthTypePasskey, nil)

		return
	}

	authed, context, status := kerberosProvider.AcceptSecContext(&st)
	if status.Code != gssapi.StatusComplete && status.Code != gssapi.StatusContinueNeeded {
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.Response.Header.Set(HTTPHeaderAuthResponse, spnegoNegTokenRespReject)
		ctx.SetJSONError(messageMFAValidationFailed)

		ctx.Logger.WithError(err).Errorf(logFmtErrKerberosAuthenticationChallengeGenerate, "error performing the login validation")

		doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, "", nil), regulation.AuthTypePasskey, nil)

		return
	}

	if status.Code == gssapi.StatusContinueNeeded {
		ctx.Response.Header.Set(HTTPHeaderAuthResponse, spnegoNegTokenRespIncompleteKRB5)
		ctx.SetStatusCode(fasthttp.StatusUnauthorized)

		ctx.Logger.WithError(err).Errorf(logFmtErrKerberosAuthenticationChallengeGenerate, "error performing the login validation")

		doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, "", nil), regulation.AuthTypePasskey, nil)

		return
	}

	if !authed {
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.Response.Header.Set(HTTPHeaderAuthResponse, spnegoNegTokenRespReject)
		ctx.SetJSONError(messageMFAValidationFailed)

		ctx.Logger.WithError(err).Errorf(logFmtErrKerberosAuthenticationChallengeGenerate, "error performing the login validation")

		doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, "", nil), regulation.AuthTypePasskey, nil)

		return
	}

	id := context.Value(ctxCredentials).(*credentials.Credentials)

	var (
		details *authentication.UserDetails
	)

	userName := id.UserName()

	if details, err = ctx.Providers.UserProvider.GetDetails(userName); err != nil || details == nil {
		doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, "", nil), regulation.AuthType1FA, err)

		ctx.Logger.WithError(err).Errorf("Error occurred getting details for user with username input '%s' which usually indicates they do not exist", userName)

		respondUnauthorized(ctx, messageAuthenticationFailed)

		return
	}

	if ban, _, expires, err := ctx.Providers.Regulator.BanCheck(ctx, details.Username); err != nil {
		if errors.Is(err, regulation.ErrUserIsBanned) {
			doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(ban, details.Username, expires), regulation.AuthType1FA, nil)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		ctx.Logger.WithError(err).Errorf(logFmtErrRegulationFail, regulation.AuthType1FA, details.Username)

		respondUnauthorized(ctx, messageAuthenticationFailed)

		return
	}

	if err = ctx.RegenerateSession(); err != nil {
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageMFAValidationFailed)

		ctx.Logger.WithError(err).Errorf(logFmtErrKerberosAuthenticationChallengeValidateUser, details.Username, "error regenerating the user session")

		doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, details.Username, nil), regulation.AuthTypePasskey, nil)

		return
	}

	doMarkAuthenticationAttempt(ctx, true, regulation.NewBan(regulation.BanTypeNone, details.Username, nil), regulation.AuthType1FA, nil)

	keepMeLoggedIn := !provider.Config.DisableRememberMe && bodyJSON.KeepMeLoggedIn != nil && *bodyJSON.KeepMeLoggedIn

	// Set the cookie to expire if remember me is enabled and the user has asked us to.
	if keepMeLoggedIn {
		if err = provider.UpdateExpiration(ctx.RequestCtx, provider.Config.RememberMe); err != nil {
			ctx.SetStatusCode(fasthttp.StatusForbidden)
			ctx.SetJSONError(messageMFAValidationFailed)

			ctx.Logger.WithError(err).Errorf(logFmtErrSessionSave, "updated expiration", regulation.AuthTypePasskey, logFmtActionAuthentication, details.Username)

			return
		}
	}

	userSession.SetOneFactorKerberos(
		ctx.Clock.Now(),
		details,
		keepMeLoggedIn,
	)

	ctx.Response.Header.Set(HTTPHeaderAuthResponse, spnegoNegTokenRespKRBAcceptCompleted)

	if ctx.Configuration.AuthenticationBackend.RefreshInterval.Update() {
		userSession.RefreshTTL = id.ValidUntil()
	}

	if len(bodyJSON.Flow) > 0 {
		handleFlowResponse(ctx, &userSession, bodyJSON.FlowID, bodyJSON.Flow, bodyJSON.SubFlow, bodyJSON.UserCode)
	} else {
		HandlePasskeyResponse(ctx, bodyJSON.TargetURL, bodyJSON.RequestMethod, userSession.Username, userSession.Groups, userSession.AuthenticationLevel(ctx.Configuration.WebAuthn.EnablePasskey2FA) == authentication.TwoFactor)
	}

}
