package regulation

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/models"
	"github.com/authelia/authelia/v4/internal/storage"
	"github.com/authelia/authelia/v4/internal/utils"
)

// NewRegulator create a regulator instance.
func NewRegulator(configuration *schema.RegulationConfiguration, provider storage.RegulatorProvider, clock utils.Clock) *Regulator {
	regulator := &Regulator{storageProvider: provider}
	regulator.clock = clock

	if configuration != nil {
		findTime, err := utils.ParseDurationString(configuration.FindTime)
		if err != nil {
			panic(err)
		}

		banTime, err := utils.ParseDurationString(configuration.BanTime)
		if err != nil {
			panic(err)
		}

		if findTime > banTime {
			panic(fmt.Errorf("find_time cannot be greater than ban_time"))
		}

		// Set regulator enabled only if MaxRetries is not 0.
		regulator.enabled = configuration.MaxRetries > 0
		regulator.maxRetries = configuration.MaxRetries
		regulator.findTime = findTime
		regulator.banTime = banTime
	}

	return regulator
}

// Mark an authentication attempt.
// We split Mark and Regulate in order to avoid timing attacks.
func (r *Regulator) Mark(ctx context.Context, successful bool, username, requestURI, requestMethod string, remoteIP net.IP) error {
	return r.storageProvider.AppendAuthenticationLog(ctx, models.AuthenticationAttempt{
		Time:          r.clock.Now(),
		Successful:    successful,
		Username:      username,
		RemoteIP:      models.IPAddress{IP: &remoteIP},
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

	attempts, err := r.storageProvider.LoadAuthenticationLogs(ctx, username, r.clock.Now().Add(-r.banTime), 10, 0)
	if err != nil {
		return time.Time{}, nil
	}

	latestFailedAttempts := make([]models.AuthenticationAttempt, 0, r.maxRetries)

	for _, attempt := range attempts {
		if attempt.Successful || len(latestFailedAttempts) >= r.maxRetries {
			// We stop appending failed attempts once we find the first successful attempts or we reach
			// the configured number of retries, meaning the user is already banned.
			break
		} else {
			latestFailedAttempts = append(latestFailedAttempts, attempt)
		}
	}

	// If the number of failed attempts within the ban time is less than the max number of retries
	// then the user is not banned.
	if len(latestFailedAttempts) < r.maxRetries {
		return time.Time{}, nil
	}

	// Now we compute the time between the latest attempt and the MaxRetry-th one. If it's
	// within the FindTime then it means that the user has been banned.
	durationBetweenLatestAttempts := latestFailedAttempts[0].Time.Sub(
		latestFailedAttempts[r.maxRetries-1].Time)

	if durationBetweenLatestAttempts < r.findTime {
		bannedUntil := latestFailedAttempts[0].Time.Add(r.banTime)
		return bannedUntil, ErrUserIsBanned
	}

	return time.Time{}, nil
}
