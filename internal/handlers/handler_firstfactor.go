package handlers

import (
	"fmt"
	"time"

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
		ctx.Error(err, authenticationFailedMessage)
		return
	}

	bannedUntil, err := ctx.Providers.Regulator.Regulate(bodyJSON.Username)

	if err != nil {
		if err == regulation.ErrUserIsBanned {
			ctx.Error(fmt.Errorf("User %s is banned until %s", bodyJSON.Username, bannedUntil), userBannedMessage)
			return
		}
		ctx.Error(fmt.Errorf("Unable to regulate authentication: %s", err), authenticationFailedMessage)
		return
	}

	userPasswordOk, err := ctx.Providers.UserProvider.CheckUserPassword(bodyJSON.Username, bodyJSON.Password)

	if err != nil {
		ctx.Logger.Debugf("Mark authentication attempt made by user %s", bodyJSON.Username)
		ctx.Providers.Regulator.Mark(bodyJSON.Username, false)

		ctx.Error(fmt.Errorf("Error while checking password for user %s: %s", bodyJSON.Username, err.Error()), authenticationFailedMessage)
		return
	}

	if !userPasswordOk {
		ctx.Logger.Debugf("Mark authentication attempt made by user %s", bodyJSON.Username)
		ctx.Providers.Regulator.Mark(bodyJSON.Username, false)

		ctx.ReplyError(fmt.Errorf("Credentials are wrong for user %s", bodyJSON.Username), authenticationFailedMessage)
		return
	}

	ctx.Logger.Debugf("Credentials validation of user %s is ok", bodyJSON.Username)

	ctx.Logger.Debugf("Mark authentication attempt made by user %s", bodyJSON.Username)
	err = ctx.Providers.Regulator.Mark(bodyJSON.Username, true)

	if err != nil {
		ctx.Error(fmt.Errorf("Unable to mark authentication: %s", err), authenticationFailedMessage)
		return
	}

	// Reset all values from previous session before regenerating the cookie.
	err = ctx.SaveSession(session.NewDefaultUserSession())

	if err != nil {
		ctx.Error(fmt.Errorf("Unable to reset the session for user %s: %s", bodyJSON.Username, err), authenticationFailedMessage)
		return
	}

	err = ctx.Providers.SessionProvider.RegenerateSession(ctx.RequestCtx)

	if err != nil {
		ctx.Error(fmt.Errorf("Unable to regenerate session for user %s: %s", bodyJSON.Username, err), authenticationFailedMessage)
		return
	}

	// set the cookie to expire in 1 year if "Remember me" was ticked.
	if *bodyJSON.KeepMeLoggedIn {
		err = ctx.Providers.SessionProvider.UpdateExpiration(ctx.RequestCtx, time.Duration(31556952*time.Second))
		if err != nil {
			ctx.Error(fmt.Errorf("Unable to update expiration timer for user %s: %s", bodyJSON.Username, err), authenticationFailedMessage)
			return
		}
	}

	// Get the details of the given user from the user provider.
	userDetails, err := ctx.Providers.UserProvider.GetDetails(bodyJSON.Username)

	if err != nil {
		ctx.Error(fmt.Errorf("Error while retrieving details from user %s: %s", bodyJSON.Username, err.Error()), authenticationFailedMessage)
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
	userSession.KeepMeLoggedIn = *bodyJSON.KeepMeLoggedIn
	err = ctx.SaveSession(userSession)

	if err != nil {
		ctx.Error(fmt.Errorf("Unable to save session of user %s", bodyJSON.Username), authenticationFailedMessage)
		return
	}

	Handle1FAResponse(ctx, bodyJSON.TargetURL, userSession.Username, userSession.Groups)
}
