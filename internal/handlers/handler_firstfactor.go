package handlers

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/session"
)

func movingAverageIteration(value time.Duration, successful bool, movingAverageCursor *int, execDurationMovingAverage *[]time.Duration, mutex sync.Locker) float64 {
	mutex.Lock()
	if successful {
		(*execDurationMovingAverage)[*movingAverageCursor] = value
		*movingAverageCursor = (*movingAverageCursor + 1) % loginDelayMovingAverageWindow
	}

	var sum int64

	for _, v := range *execDurationMovingAverage {
		sum += v.Milliseconds()
	}
	mutex.Unlock()

	return float64(sum / loginDelayMovingAverageWindow)
}

func calculateActualDelay(ctx *middlewares.AutheliaCtx, execDuration time.Duration, avgExecDurationMs float64, successful *bool) float64 {
	randomDelayMs := float64(rand.Int63n(loginDelayMaximumRandomDelayMilliseconds)) //nolint:gosec // TODO: Consider use of crypto/rand, this should be benchmarked and measured first.
	totalDelayMs := math.Max(avgExecDurationMs, loginDelayMinimumDelayMilliseconds) + randomDelayMs
	actualDelayMs := math.Max(totalDelayMs-float64(execDuration.Milliseconds()), 1.0)
	ctx.Logger.Tracef("Attempt successful: %t, exec duration: %d, avg execution duration: %d, random delay ms: %d, total delay ms: %d, actual delay ms: %d", *successful, execDuration.Milliseconds(), int64(avgExecDurationMs), int64(randomDelayMs), int64(totalDelayMs), int64(actualDelayMs))

	return actualDelayMs
}

func delayToPreventTimingAttacks(ctx *middlewares.AutheliaCtx, requestTime time.Time, successful *bool, movingAverageCursor *int, execDurationMovingAverage *[]time.Duration, mutex sync.Locker) {
	execDuration := time.Since(requestTime)
	avgExecDurationMs := movingAverageIteration(execDuration, *successful, movingAverageCursor, execDurationMovingAverage, mutex)
	actualDelayMs := calculateActualDelay(ctx, execDuration, avgExecDurationMs, successful)
	time.Sleep(time.Duration(actualDelayMs) * time.Millisecond)
}

// FirstFactorPost is the handler performing the first factory.
//nolint:gocyclo // TODO: Consider refactoring time permitting.
func FirstFactorPost(msInitialDelay time.Duration, delayEnabled bool) middlewares.RequestHandler {
	var execDurationMovingAverage = make([]time.Duration, loginDelayMovingAverageWindow)

	var movingAverageCursor = 0

	var mutex = &sync.Mutex{}

	for i := range execDurationMovingAverage {
		execDurationMovingAverage[i] = msInitialDelay * time.Millisecond
	}

	rand.Seed(time.Now().UnixNano())

	return func(ctx *middlewares.AutheliaCtx) {
		var successful bool

		requestTime := time.Now()

		if delayEnabled {
			defer delayToPreventTimingAttacks(ctx, requestTime, &successful, &movingAverageCursor, &execDurationMovingAverage, mutex)
		}

		bodyJSON := firstFactorRequestBody{}
		err := ctx.ParseBody(&bodyJSON)

		if err != nil {
			handleAuthenticationUnauthorized(ctx, err, messageAuthenticationFailed)
			return
		}

		bannedUntil, err := ctx.Providers.Regulator.Regulate(bodyJSON.Username)

		if err != nil {
			if err == regulation.ErrUserIsBanned {
				handleAuthenticationUnauthorized(ctx, fmt.Errorf("user %s is banned until %s", bodyJSON.Username, bannedUntil), messageAuthenticationFailed)
				return
			}

			handleAuthenticationUnauthorized(ctx, fmt.Errorf("unable to regulate authentication: %s", err.Error()), messageAuthenticationFailed)

			return
		}

		userPasswordOk, err := ctx.Providers.UserProvider.CheckUserPassword(bodyJSON.Username, bodyJSON.Password)

		if err != nil {
			ctx.Logger.Debugf("Mark authentication attempt made by user %s", bodyJSON.Username)

			if err := ctx.Providers.Regulator.Mark(bodyJSON.Username, false); err != nil {
				ctx.Logger.Errorf("unable to mark authentication: %s", err.Error())
			}

			handleAuthenticationUnauthorized(ctx, fmt.Errorf("error while checking password for user %s: %s", bodyJSON.Username, err.Error()), messageAuthenticationFailed)

			return
		}

		if !userPasswordOk {
			ctx.Logger.Debugf("Mark authentication attempt made by user %s", bodyJSON.Username)

			if err := ctx.Providers.Regulator.Mark(bodyJSON.Username, false); err != nil {
				ctx.Logger.Errorf("unable to mark authentication: %s", err.Error())
			}

			handleAuthenticationUnauthorized(ctx, fmt.Errorf("credentials are wrong for user %s", bodyJSON.Username), messageAuthenticationFailed)

			return
		}

		ctx.Logger.Debugf("Mark authentication attempt made by user %s", bodyJSON.Username)
		err = ctx.Providers.Regulator.Mark(bodyJSON.Username, true)

		if err != nil {
			handleAuthenticationUnauthorized(ctx, fmt.Errorf("unable to mark authentication: %s", err.Error()), messageAuthenticationFailed)
			return
		}

		ctx.Logger.Debugf("Credentials validation of user %s is ok", bodyJSON.Username)

		userSession := ctx.GetSession()
		newSession := session.NewDefaultUserSession()
		newSession.OIDCWorkflowSession = userSession.OIDCWorkflowSession

		// Reset all values from previous session except OIDC workflow before regenerating the cookie.
		err = ctx.SaveSession(newSession)

		if err != nil {
			handleAuthenticationUnauthorized(ctx, fmt.Errorf("unable to reset the session for user %s: %s", bodyJSON.Username, err.Error()), messageAuthenticationFailed)
			return
		}

		err = ctx.Providers.SessionProvider.RegenerateSession(ctx.RequestCtx)

		if err != nil {
			handleAuthenticationUnauthorized(ctx, fmt.Errorf("unable to regenerate session for user %s: %s", bodyJSON.Username, err.Error()), messageAuthenticationFailed)
			return
		}

		// Check if bodyJSON.KeepMeLoggedIn can be deref'd and derive the value based on the configuration and JSON data
		keepMeLoggedIn := ctx.Providers.SessionProvider.RememberMe != 0 && bodyJSON.KeepMeLoggedIn != nil && *bodyJSON.KeepMeLoggedIn

		// Set the cookie to expire if remember me is enabled and the user has asked us to
		if keepMeLoggedIn {
			err = ctx.Providers.SessionProvider.UpdateExpiration(ctx.RequestCtx, ctx.Providers.SessionProvider.RememberMe)
			if err != nil {
				handleAuthenticationUnauthorized(ctx, fmt.Errorf("unable to update expiration timer for user %s: %s", bodyJSON.Username, err.Error()), messageAuthenticationFailed)
				return
			}
		}

		// Get the details of the given user from the user provider.
		userDetails, err := ctx.Providers.UserProvider.GetDetails(bodyJSON.Username)

		if err != nil {
			handleAuthenticationUnauthorized(ctx, fmt.Errorf("error while retrieving details from user %s: %s", bodyJSON.Username, err.Error()), messageAuthenticationFailed)
			return
		}

		ctx.Logger.Tracef("Details for user %s => groups: %s, emails %s", bodyJSON.Username, userDetails.Groups, userDetails.Emails)

		userSession.SetOneFactor(ctx.Clock.Now(), userDetails, keepMeLoggedIn)

		if refresh, refreshInterval := getProfileRefreshSettings(ctx.Configuration.AuthenticationBackend); refresh {
			userSession.RefreshTTL = ctx.Clock.Now().Add(refreshInterval)
		}

		err = ctx.SaveSession(userSession)
		if err != nil {
			handleAuthenticationUnauthorized(ctx, fmt.Errorf("unable to save session of user %s", bodyJSON.Username), messageAuthenticationFailed)
			return
		}

		successful = true

		if userSession.OIDCWorkflowSession != nil {
			handleOIDCWorkflowResponse(ctx)
		} else {
			Handle1FAResponse(ctx, bodyJSON.TargetURL, bodyJSON.RequestMethod, userSession.Username, userSession.Groups)
		}
	}
}
