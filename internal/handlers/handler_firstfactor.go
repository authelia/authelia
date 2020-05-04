package handlers

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/authelia/authelia/internal/authentication"
	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/logging"
	"github.com/authelia/authelia/internal/middlewares"
	"github.com/authelia/authelia/internal/regulation"
	"github.com/authelia/authelia/internal/session"
	"github.com/authelia/authelia/internal/utils"
)

func getDelayAuthSettings(config schema.AuthenticationBackendConfiguration) (bool, time.Duration) {
	log := logging.Logger()

	if !config.DisableDelayAuth {
		rand.Seed(time.Now().UnixNano())
		var duration time.Duration
		if config.File != nil {
			algorithm, err := authentication.ConfigAlgoToCryptoAlgo(config.File.Password.Algorithm)
			if err != nil {
				panic(err)
			}
			password := utils.RandomString(20, authentication.HashingPossibleSaltCharacters)
			start := time.Now()
			_, _ = authentication.HashPassword(password, "",
				algorithm, config.File.Password.Iterations,
				config.File.Password.Memory*1024, config.File.Password.Parallelism,
				config.File.Password.KeyLength, config.File.Password.SaltLength)
			duration = time.Since(start) + (time.Duration(rand.Int31n(50) + 150))
		}
		if duration < 450*time.Millisecond {
			duration = time.Duration(rand.Intn(50)+450) * time.Millisecond
		}
		log.Debugf("1FA authentication requests will not return to clients until %dms after the request was received "+
			"to prevent username enumeration.", duration/time.Millisecond)
		return true, duration
	}
	log.Warn("1FA authentication requests will not be delayed as it has been disabled by configuration option " +
		"authentication_backend.disable_delay_auth, this reduces security and is not recommended.")
	return false, 0 * time.Millisecond
}

func doDelayAuth(ctx *middlewares.AutheliaCtx, username string, receivedTime time.Time, enabled bool, duration time.Duration) {
	if !enabled {
		ctx.Logger.Warnf("Skipping the authentication delay of user %s since authentication delay is disabled by configuration", username)
		return
	}
	delayTime := receivedTime.Add(duration)
	if time.Now().Before(delayTime) {
		sleepFor := time.Until(delayTime)
		ctx.Logger.Debugf("Starting the authentication delay of user %s for %dms", username, sleepFor/time.Millisecond)
		time.Sleep(sleepFor)
	}
	ctx.Logger.Warnf("Skipping the authentication delay of user %s since authentication took longer than the expected delay", username)
}

// FirstFactorPost generates the handler performing the first factor authentication factory.
func FirstFactorPost(config schema.Configuration) middlewares.RequestHandler {
	delayAuth, delayAuthDuration := getDelayAuthSettings(config.AuthenticationBackend)

	return func(ctx *middlewares.AutheliaCtx) {
		receivedTime := time.Now()
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
			ctx.Providers.Regulator.Mark(bodyJSON.Username, false) //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.

			doDelayAuth(ctx, bodyJSON.Username, receivedTime, delayAuth, delayAuthDuration)
			ctx.Error(fmt.Errorf("Error while checking password for user %s: %s", bodyJSON.Username, err.Error()), authenticationFailedMessage)
			return
		}
		doDelayAuth(ctx, bodyJSON.Username, receivedTime, delayAuth, delayAuthDuration)
		if !userPasswordOk {
			ctx.Logger.Debugf("Mark authentication attempt made by user %s", bodyJSON.Username)
			ctx.Providers.Regulator.Mark(bodyJSON.Username, false) //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.

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

		// Check if bodyJSON.KeepMeLoggedIn can be deref'd and derive the value based on the configuration and JSON data
		keepMeLoggedIn := ctx.Providers.SessionProvider.RememberMe != 0 && bodyJSON.KeepMeLoggedIn != nil && *bodyJSON.KeepMeLoggedIn

		// Set the cookie to expire if remember me is enabled and the user has asked us to
		if keepMeLoggedIn {
			err = ctx.Providers.SessionProvider.UpdateExpiration(ctx.RequestCtx, ctx.Providers.SessionProvider.RememberMe)
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
		userSession.KeepMeLoggedIn = keepMeLoggedIn
		err = ctx.SaveSession(userSession)

		if err != nil {
			ctx.Error(fmt.Errorf("Unable to save session of user %s", bodyJSON.Username), authenticationFailedMessage)
			return
		}

		Handle1FAResponse(ctx, bodyJSON.TargetURL, userSession.Username, userSession.Groups)
	}
}
