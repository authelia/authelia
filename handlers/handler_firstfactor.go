package handlers

import (
	"fmt"
	"net/url"
	"time"

	"github.com/clems4ever/authelia/regulation"

	"github.com/clems4ever/authelia/session"

	"github.com/clems4ever/authelia/authentication"
	"github.com/clems4ever/authelia/authorization"
	"github.com/clems4ever/authelia/middlewares"
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

	if err == regulation.ErrUserIsBanned {
		ctx.Error(fmt.Errorf("User %s is banned until %s", bodyJSON.Username, bannedUntil), userBannedMessage)
		return
	} else if err != nil {
		ctx.Error(fmt.Errorf("Unable to regulate authentication: %s", err), authenticationFailedMessage)
		return
	}

	userPasswordOk, err := ctx.Providers.UserProvider.CheckUserPassword(bodyJSON.Username, bodyJSON.Password)

	if err != nil {
		ctx.Error(fmt.Errorf("Error while checking password for user %s: %s", bodyJSON.Username, err.Error()), authenticationFailedMessage)
		return
	}

	ctx.Logger.Debugf("Mark authentication attempt made by user %s", bodyJSON.Username)
	// Mark the authentication attempt and whether it was successful.
	err = ctx.Providers.Regulator.Mark(bodyJSON.Username, userPasswordOk)

	if err != nil {
		ctx.Error(fmt.Errorf("Unable to mark authentication: %s", err), authenticationFailedMessage)
		return
	}

	if !userPasswordOk {
		ctx.Error(fmt.Errorf("Credentials are wrong for user %s", bodyJSON.Username), authenticationFailedMessage)
		return
	}

	ctx.Logger.Debugf("Credentials validation of user %s is ok", bodyJSON.Username)

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

	// and avoid the cookie to expire if "Remember me" was ticked.
	if *bodyJSON.KeepMeLoggedIn {
		err = ctx.Providers.SessionProvider.UpdateExpiration(ctx.RequestCtx, time.Duration(0))
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

	ctx.Logger.Debugf("Details for user %s => groups: %s, emails %s", bodyJSON.Username, userDetails.Groups, userDetails.Emails)

	// And set those information in the new session.
	userSession := ctx.GetSession()
	userSession.Username = bodyJSON.Username
	userSession.Groups = userDetails.Groups
	userSession.Emails = userDetails.Emails
	userSession.AuthenticationLevel = authentication.OneFactor
	userSession.LastActivity = time.Now().Unix()
	err = ctx.SaveSession(userSession)

	if err != nil {
		ctx.Error(fmt.Errorf("Unable to save session of user %s", bodyJSON.Username), authenticationFailedMessage)
		return
	}

	if bodyJSON.TargetURL != "" {
		targetURL, err := url.ParseRequestURI(bodyJSON.TargetURL)
		if err != nil {
			ctx.Error(fmt.Errorf("Unable to parse target URL %s: %s", bodyJSON.TargetURL, err), authenticationFailedMessage)
			return
		}
		requiredLevel := ctx.Providers.Authorizer.GetRequiredLevel(authorization.Subject{
			Username: userSession.Username,
			Groups:   userSession.Groups,
			IP:       ctx.RemoteIP(),
		}, *targetURL)

		ctx.Logger.Debugf("Required level for the URL %s is %d", targetURL.String(), requiredLevel)

		safeRedirection := isRedirectionSafe(*targetURL, ctx.Configuration.Session.Domain)

		if safeRedirection && requiredLevel <= authorization.OneFactor {
			response := redirectResponse{bodyJSON.TargetURL}
			ctx.SetJSONBody(response)
		} else {
			ctx.ReplyOK()
		}
	} else {
		ctx.ReplyOK()
	}
}
