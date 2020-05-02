package handlers

import (
	"fmt"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/internal/authentication"
	"github.com/authelia/authelia/internal/middlewares"
	"github.com/authelia/authelia/internal/regulation"
	"github.com/authelia/authelia/internal/session"
)

// FirstFactorPost is the handler performing the first factory.
func FirstFactorPost(ctx *middlewares.AutheliaCtx) {
	bodyJSON := firstFactorRequestBody{}
	err := ctx.ParseBody(&bodyJSON)

	if err != nil {
		handleFirstFactorError(ctx, err, authenticationFailedMessage, false)
		return
	}

	bannedUntil, err := ctx.Providers.Regulator.Regulate(bodyJSON.Username)

	if err != nil {
		if err == regulation.ErrUserIsBanned {
			handleFirstFactorError(ctx, fmt.Errorf("User %s is banned until %s", bodyJSON.Username, bannedUntil), userBannedMessage, false)
			return
		}
		handleFirstFactorError(ctx, fmt.Errorf("Unable to regulate authentication: %s", err.Error()), authenticationFailedMessage, false)
		return
	}

	userPasswordOk, err := ctx.Providers.UserProvider.CheckUserPassword(bodyJSON.Username, bodyJSON.Password)

	if err != nil {
		ctx.Logger.Debugf("Mark authentication attempt made by user %s", bodyJSON.Username)
		ctx.Providers.Regulator.Mark(bodyJSON.Username, false) //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
		handleFirstFactorError(ctx, fmt.Errorf("Error while checking password for user %s: %s", bodyJSON.Username, err.Error()), authenticationFailedMessage, false)
		return
	}

	if !userPasswordOk {
		ctx.Logger.Debugf("Mark authentication attempt made by user %s", bodyJSON.Username)
		ctx.Providers.Regulator.Mark(bodyJSON.Username, false) //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
		handleFirstFactorError(ctx, fmt.Errorf("Credentials are wrong for user %s", bodyJSON.Username), authenticationFailedMessage, true)
		return
	}

	ctx.Logger.Debugf("Credentials validation of user %s is ok", bodyJSON.Username)

	ctx.Logger.Debugf("Mark authentication attempt made by user %s", bodyJSON.Username)
	err = ctx.Providers.Regulator.Mark(bodyJSON.Username, true)

	if err != nil {
		handleFirstFactorError(ctx, fmt.Errorf("Unable to mark authentication: %s", err.Error()), authenticationFailedMessage, false)
		return
	}

	// Reset all values from previous session before regenerating the cookie.
	err = ctx.SaveSession(session.NewDefaultUserSession())

	if err != nil {
		handleFirstFactorError(ctx, fmt.Errorf("Unable to reset the session for user %s: %s", bodyJSON.Username, err.Error()), authenticationFailedMessage, false)
		return
	}

	err = ctx.Providers.SessionProvider.RegenerateSession(ctx.RequestCtx)

	if err != nil {
		handleFirstFactorError(ctx, fmt.Errorf("Unable to regenerate session for user %s: %s", bodyJSON.Username, err.Error()), authenticationFailedMessage, false)
		return
	}

	// Check if bodyJSON.KeepMeLoggedIn can be deref'd and derive the value based on the configuration and JSON data
	keepMeLoggedIn := ctx.Providers.SessionProvider.RememberMe != 0 && bodyJSON.KeepMeLoggedIn != nil && *bodyJSON.KeepMeLoggedIn

	// Set the cookie to expire if remember me is enabled and the user has asked us to
	if keepMeLoggedIn {
		err = ctx.Providers.SessionProvider.UpdateExpiration(ctx.RequestCtx, ctx.Providers.SessionProvider.RememberMe)
		if err != nil {
			handleFirstFactorError(ctx, fmt.Errorf("Unable to update expiration timer for user %s: %s", bodyJSON.Username, err.Error()), authenticationFailedMessage, false)
			return
		}
	}

	// Get the details of the given user from the user provider.
	userDetails, err := ctx.Providers.UserProvider.GetDetails(bodyJSON.Username)

	if err != nil {
		handleFirstFactorError(ctx, fmt.Errorf("Error while retrieving details from user %s: %s", bodyJSON.Username, err.Error()), authenticationFailedMessage, false)
		return
	}

	ctx.Logger.Tracef("Details for user %s => groups: %s, emails %s", bodyJSON.Username, userDetails.Groups, userDetails.Emails)

	// And set those information in the new session.
	userSession := ctx.GetSession()
	userSession.Username = userDetails.Username
	userSession.Groups = userDetails.Groups
	userSession.Emails = userDetails.Emails
	userSession.AuthenticationLevel = authentication.OneFactor
	userSession.LastActivity = time.Now().Unix()
	userSession.KeepMeLoggedIn = keepMeLoggedIn
	err = ctx.SaveSession(userSession)

	if err != nil {
		handleFirstFactorError(ctx, fmt.Errorf("Unable to save session of user %s", bodyJSON.Username), authenticationFailedMessage, false)
		return
	}

	Handle1FAResponse(ctx, bodyJSON.TargetURL, userSession.Username, userSession.Groups)
}

// handleFirstFactorError provides harmonized response codes for 1FA.
func handleFirstFactorError(ctx *middlewares.AutheliaCtx, err error, message string, reply bool) {
	ctx.SetStatusCode(fasthttp.StatusUnauthorized)
	if reply {
		ctx.ReplyError(err, message)
	} else {
		ctx.Error(err, message)
	}
}
