package regulation

import (
	"context"
	"strings"
	"time"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/storage"
)

// NewRegulator create a regulator instance.
func NewRegulator(config schema.Regulation, store storage.RegulatorProvider, clock clock.Provider) *Regulator {
	return &Regulator{
		enabled: config.MaxRetries > 0,
		store:   store,
		clock:   clock,
		config:  config,
	}
}

// Mark an authentication attempt.
// We split Mark and Regulate in order to avoid timing attacks.
func (r *Regulator) Mark(ctx Context, successful, banned bool, username, requestURI, requestMethod, authType string) error {
	ctx.RecordAuthn(successful, banned, strings.ToLower(authType))

	return r.store.AppendAuthenticationLog(ctx, model.AuthenticationAttempt{
		Time:          r.clock.Now(),
		Successful:    successful,
		Banned:        banned,
		Username:      username,
		Type:          authType,
		RemoteIP:      model.NewNullIP(ctx.RemoteIP()),
		RequestURI:    requestURI,
		RequestMethod: requestMethod,
	})
}

// Regulate the authentication attempts for a given user.
// This method returns ErrUserIsBanned if the user is banned along with the time until when the user is banned.
func (r *Regulator) Regulate(ctx context.Context, username string) (time.Time, error) {
	// If there is regulation configuration, no regulation applies.
	if !r.enabled {
		return time.Time{}, nil
	}

	attempts, err := r.store.LoadAuthenticationLogs(ctx, username, r.clock.Now().Add(-r.config.BanTime), 10, 0)
	if err != nil {
		return time.Time{}, nil
	}

	latestFailedAttempts := make([]model.AuthenticationAttempt, 0, r.config.MaxRetries)

	for _, attempt := range attempts {
		if attempt.Successful || len(latestFailedAttempts) >= r.config.MaxRetries {
			// We stop appending failed attempts once we find the first successful attempts or we reach
			// the configured number of retries, meaning the user is already banned.
			break
		} else {
			latestFailedAttempts = append(latestFailedAttempts, attempt)
		}
	}

	// If the number of failed attempts within the ban time is less than the max number of retries
	// then the user is not banned.
	if len(latestFailedAttempts) < r.config.MaxRetries {
		return time.Time{}, nil
	}

	// Now we compute the time between the latest attempt and the MaxRetry-th one. If it's
	// within the FindTime then it means that the user has been banned.
	durationBetweenLatestAttempts := latestFailedAttempts[0].Time.Sub(
		latestFailedAttempts[r.config.MaxRetries-1].Time)

	if durationBetweenLatestAttempts < r.config.FindTime {
		bannedUntil := latestFailedAttempts[0].Time.Add(r.config.BanTime)
		return bannedUntil, ErrUserIsBanned
	}

	return time.Time{}, nil
}
