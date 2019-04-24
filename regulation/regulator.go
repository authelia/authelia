package regulation

import (
	"fmt"
	"time"

	"github.com/clems4ever/authelia/configuration/schema"
	"github.com/clems4ever/authelia/models"
	"github.com/clems4ever/authelia/storage"
)

// NewRegulator create a regulator instance.
func NewRegulator(configuration *schema.RegulationConfiguration, provider storage.Provider) *Regulator {
	regulator := &Regulator{storageProvider: provider}
	if configuration != nil {
		if configuration.FindTime > configuration.BanTime {
			panic(fmt.Errorf("find_time cannot be greater than ban_time"))
		}
		regulator.enabled = true
		regulator.maxRetries = configuration.MaxRetries
		regulator.findTime = time.Duration(configuration.FindTime) * time.Second
		regulator.banTime = time.Duration(configuration.BanTime) * time.Second
	}
	return regulator
}

// Mark mark an authentication attempt.
// We split Mark and Regulate in order to avoid timing attacks since if
func (r *Regulator) Mark(username string, successful bool) error {
	return r.storageProvider.AppendAuthenticationLog(models.AuthenticationAttempt{
		Username:   username,
		Successful: successful,
		Time:       time.Now(),
	})
}

// Regulate regulate the authentication attempts for a given user.
// This method returns ErrUserIsBanned if the user is banned along with the time until when
// the user is banned.
func (r *Regulator) Regulate(username string) (time.Time, error) {
	// If there is regulation configuration, no regulation applies.
	if !r.enabled {
		return time.Time{}, nil
	}
	now := time.Now()

	// TODO(c.michaud): make sure FindTime < BanTime.
	attempts, err := r.storageProvider.LoadLatestAuthenticationLogs(username, now.Add(-r.banTime))

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
