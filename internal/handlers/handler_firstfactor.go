package handlers

import (
	"errors"
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

		if err := ctx.ParseBody(&bodyJSON); err != nil {
			ctx.Logger.Errorf(logFmtErrParseRequestBody, regulation.AuthType1FA, err)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		if bannedUntil, err := ctx.Providers.Regulator.Regulate(ctx, bodyJSON.Username); err != nil {
			if errors.Is(err, regulation.ErrUserIsBanned) {
				_ = markAuthenticationAttempt(ctx, false, &bannedUntil, bodyJSON.Username, regulation.AuthType1FA, nil)

				respondUnauthorized(ctx, messageAuthenticationFailed)

				return
			}

			ctx.Logger.Errorf(logFmtErrRegulationFail, regulation.AuthType1FA, bodyJSON.Username, err)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		userPasswordOk, err := ctx.Providers.UserProvider.CheckUserPassword(bodyJSON.Username, bodyJSON.Password)
		if err != nil {
			_ = markAuthenticationAttempt(ctx, false, nil, bodyJSON.Username, regulation.AuthType1FA, err)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		if !userPasswordOk {
			_ = markAuthenticationAttempt(ctx, false, nil, bodyJSON.Username, regulation.AuthType1FA, nil)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		if err = markAuthenticationAttempt(ctx, true, nil, bodyJSON.Username, regulation.AuthType1FA, nil); err != nil {
			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		userSession := ctx.GetSession()
		newSession := session.NewDefaultUserSession()
		newSession.OIDCWorkflowSession = userSession.OIDCWorkflowSession

		// Reset all values from previous session except OIDC workflow before regenerating the cookie.
		if err = ctx.SaveSession(newSession); err != nil {
			ctx.Logger.Errorf(logFmtErrSessionReset, regulation.AuthType1FA, bodyJSON.Username, err)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		if err = ctx.Providers.SessionProvider.RegenerateSession(ctx.RequestCtx); err != nil {
			ctx.Logger.Errorf(logFmtErrSessionRegenerate, regulation.AuthType1FA, bodyJSON.Username, err)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		// Check if bodyJSON.KeepMeLoggedIn can be deref'd and derive the value based on the configuration and JSON data
		keepMeLoggedIn := ctx.Providers.SessionProvider.RememberMe != 0 && bodyJSON.KeepMeLoggedIn != nil && *bodyJSON.KeepMeLoggedIn

		// Set the cookie to expire if remember me is enabled and the user has asked us to
		if keepMeLoggedIn {
			err = ctx.Providers.SessionProvider.UpdateExpiration(ctx.RequestCtx, ctx.Providers.SessionProvider.RememberMe)
			if err != nil {
				ctx.Logger.Errorf(logFmtErrSessionSave, "updated expiration", regulation.AuthType1FA, bodyJSON.Username, err)

				respondUnauthorized(ctx, messageAuthenticationFailed)

				return
			}
		}

		// Get the details of the given user from the user provider.
		userDetails, err := ctx.Providers.UserProvider.GetDetails(bodyJSON.Username)
		if err != nil {
			ctx.Logger.Errorf(logFmtErrObtainProfileDetails, regulation.AuthType1FA, bodyJSON.Username, err)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		ctx.Logger.Tracef(logFmtTraceProfileDetails, bodyJSON.Username, userDetails.Groups, userDetails.Emails)

		userSession.SetOneFactor(ctx.Clock.Now(), userDetails, keepMeLoggedIn)

		if refresh, refreshInterval := getProfileRefreshSettings(ctx.Configuration.AuthenticationBackend); refresh {
			userSession.RefreshTTL = ctx.Clock.Now().Add(refreshInterval)
		}

		if err = ctx.SaveSession(userSession); err != nil {
			ctx.Logger.Errorf(logFmtErrSessionSave, "updated profile", regulation.AuthType1FA, bodyJSON.Username, err)

			respondUnauthorized(ctx, messageAuthenticationFailed)

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
