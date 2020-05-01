package regulation

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/models"
	"github.com/authelia/authelia/internal/storage"
	"github.com/authelia/authelia/internal/utils"
)

// NewRegulator create a regulator instance.
func NewRegulator(configuration *schema.RegulationConfiguration, provider storage.Provider, clock utils.Clock, firstFactorDelay time.Duration) *Regulator {
	regulator := &Regulator{storageProvider: provider}
	regulator.clock = clock
	if firstFactorDelay < 0 {
		regulator.firstFactorDelay = 0 * time.Millisecond
	} else if firstFactorDelay > 500*time.Millisecond {
		rand.Seed(time.Now().UnixNano())
		ms := rand.Int31n(50) + 150
		regulator.firstFactorDelay = firstFactorDelay + (time.Millisecond * time.Duration(ms))
	} else {
		regulator.firstFactorDelay = 500 * time.Millisecond
	}
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

// Mark mark an authentication attempt.
// We split Mark and Regulate in order to avoid timing attacks since if
func (r *Regulator) Mark(username string, successful bool) error {
	return r.storageProvider.AppendAuthenticationLog(models.AuthenticationAttempt{
		Username:   username,
		Successful: successful,
		Time:       r.clock.Now(),
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
	now := r.clock.Now()

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

// FirstFactorDelay delays the first factor return to prevent username enumeration.
func (r *Regulator) FirstFactorDelay(postTime time.Time, real bool) {
	/*
		// This section contains a basic adaptive algorithm.
		// It's possible the algorithm could cause massive delays due to various reasons.
		// IT's questionable if we want this.

		if real {
			r.firstFactorDelays = append(r.firstFactorDelays, time.Now().Sub(postTime))
		}
		if len(r.firstFactorDelays) <= 10 {
			// Copy r.firstFactorDelays so we can reset it ASAP.
			delays := r.firstFactorDelays
			r.firstFactorDelays = []time.Duration{}

			// Find the highest duration in the list of real logins.
			var max time.Duration
			for _, delay := range delays {
				if delay > max {
					max = delay
				}
			}
			r.firstFactorDelay = max
		}
	*/
	// For test purposes.
	if r.firstFactorDelay < 0 {
		return
	}
	delayTime := postTime.Add(r.firstFactorDelay)
	if time.Now().Before(delayTime) {
		time.Sleep(time.Until(delayTime))
	}
}
