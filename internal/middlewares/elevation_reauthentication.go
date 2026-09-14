// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"time"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/session"
)

const (
	// ReauthenticationMethodPassword indicates the user may reauthenticate with their password.
	ReauthenticationMethodPassword = "password"

	// ReauthenticationMethodSecondFactor indicates the user may reauthenticate with a second factor method.
	ReauthenticationMethodSecondFactor = "second_factor"
)

// Enrollment describes the known second factor enrollment of a user.
type Enrollment int

const (
	// EnrollmentUnknown indicates the enrollment could not be determined.
	EnrollmentUnknown Enrollment = iota

	// EnrollmentNone indicates the user has no second factor methods.
	EnrollmentNone

	// EnrollmentSecondFactor indicates the user has at least one second factor method.
	EnrollmentSecondFactor
)

// ReauthenticationState describes whether the elevated session reauthentication requirement is unmet and which methods
// the user may use to satisfy it.
type ReauthenticationState struct {
	Required bool
	Methods  []string
}

// NewReauthenticationState computes the ReauthenticationState from the configuration, user session, current time, and
// second factor enrollment.
func NewReauthenticationState(config schema.IdentityValidationElevatedSession, userSession *session.UserSession, now time.Time, enrollment Enrollment) ReauthenticationState {
	password := isAuthenticationFresh(userSession.FirstFactorAuthnTimestamp, now, config.ReauthenticationLifespan)

	// Only possession based second factors satisfy the second factor requirement, as a password used as a second factor
	// must never satisfy it.
	secondFactor := isAuthenticationFresh(userSession.SecondFactorPossessionAuthnTimestamp, now, config.ReauthenticationLifespan)

	switch config.RequireReauthentication {
	case schema.ElevatedSessionReauthenticationPassword:
		return newReauthenticationState(password, ReauthenticationMethodPassword)
	case schema.ElevatedSessionReauthenticationSecondFactor:
		switch enrollment {
		case EnrollmentSecondFactor:
			return newReauthenticationState(secondFactor, ReauthenticationMethodSecondFactor)
		case EnrollmentNone:
			return newReauthenticationState(password, ReauthenticationMethodPassword)
		default:
			return newReauthenticationState(false, ReauthenticationMethodSecondFactor)
		}
	case schema.ElevatedSessionReauthenticationAny:
		switch enrollment {
		case EnrollmentSecondFactor:
			return newReauthenticationState(password || secondFactor, ReauthenticationMethodPassword, ReauthenticationMethodSecondFactor)
		case EnrollmentNone:
			return newReauthenticationState(password, ReauthenticationMethodPassword)
		default:
			return newReauthenticationState(false, ReauthenticationMethodPassword)
		}
	default:
		return ReauthenticationState{Methods: []string{}}
	}
}

// GetReauthenticationState computes the ReauthenticationState for the user session, loading the user second factor
// enrollment from the storage provider only when the configured mode requires it.
func GetReauthenticationState(ctx *AutheliaCtx, userSession *session.UserSession) ReauthenticationState {
	config := ctx.Configuration.IdentityValidation.ElevatedSession

	enrollment := EnrollmentUnknown

	switch config.RequireReauthentication {
	case schema.ElevatedSessionReauthenticationSecondFactor, schema.ElevatedSessionReauthenticationAny:
		info, err := ctx.Providers.StorageProvider.LoadUserInfo(ctx, userSession.Username)
		if err != nil {
			ctx.Logger.WithError(err).Error("Error occurred attempting to lookup user information during a reauthentication check.")

			break
		}

		if ctx.Configuration.TOTP.Disable {
			info.HasTOTP = false
		}

		if ctx.Configuration.WebAuthn.Disable {
			info.HasWebAuthn = false
		}

		if ctx.Configuration.DuoAPI.Disable {
			info.HasDuo = false
		}

		if info.HasTOTP || info.HasWebAuthn || info.HasDuo {
			enrollment = EnrollmentSecondFactor
		} else {
			enrollment = EnrollmentNone
		}
	}

	return NewReauthenticationState(config, userSession, ctx.GetClock().Now(), enrollment)
}

func newReauthenticationState(satisfied bool, methods ...string) ReauthenticationState {
	if satisfied {
		return ReauthenticationState{Methods: []string{}}
	}

	return ReauthenticationState{Required: true, Methods: methods}
}

func isAuthenticationFresh(timestamp int64, now time.Time, lifespan time.Duration) bool {
	if timestamp <= 0 {
		return false
	}

	return !now.After(time.Unix(timestamp, 0).Add(lifespan))
}
