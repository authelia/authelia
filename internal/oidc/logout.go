// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"net/url"
	"time"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
)

const (
	// LogoutRevokeGrace is how long after an OpenID Connect session ends that its online refresh tokens are revoked
	// again, which revokes those issued by refresh token grants which were in progress when it ended.
	LogoutRevokeGrace = 2 * time.Minute

	// LogoutBackChannelLifespan is how long the delivery of a Logout Token is retried for.
	LogoutBackChannelLifespan = 24 * time.Hour

	logoutRevokeMargin = time.Minute
)

// NewLogoutBackChannelSchedule returns the schedule of the deliveries of the Logout Tokens for the OpenID Connect
// sessions of the issuer which ended at the time. Each is delivered when the session ends and is retried by the
// LogoutTaskCollector from one minute later until LogoutBackChannelLifespan elapses.
func NewLogoutBackChannelSchedule(issuer *url.URL, now time.Time) *model.OAuth2SessionLogoutTaskScheduleBackChannel {
	return &model.OAuth2SessionLogoutTaskScheduleBackChannel{
		IssuerURL:     issuer.String(),
		NextAttemptAt: now.Add(logoutTaskBackoffMin),
		ExpiresAt:     now.Add(LogoutBackChannelLifespan),
	}
}

// NewLogoutRevokeSchedule returns the schedule of the record of an OpenID Connect session which ended at the time.
// The record exists until every authorization code and device code which may have been issued to the session has
// expired, and its online refresh tokens are revoked again after LogoutRevokeGrace and once more when it expires.
func NewLogoutRevokeSchedule(config *schema.IdentityProvidersOpenIDConnect, now time.Time) model.OAuth2SessionLogoutTaskSchedule {
	return model.OAuth2SessionLogoutTaskSchedule{
		CreatedAt:     now,
		NextAttemptAt: now.Add(LogoutRevokeGrace),
		ExpiresAt:     now.Add(LogoutRevokeHorizon(config)),
	}
}

// LogoutRevokeHorizon returns how long the record of an ended OpenID Connect session is kept, which is the longest
// lifespan of an authorization code or device code plus a margin, and never less than LogoutRevokeGrace plus the
// margin.
func LogoutRevokeHorizon(config *schema.IdentityProvidersOpenIDConnect) (horizon time.Duration) {
	horizon = max(schema.DefaultOpenIDConnectConfiguration.Lifespans.AuthorizeCode, schema.DefaultOpenIDConnectConfiguration.Lifespans.DeviceCode, LogoutRevokeGrace)

	if config != nil {
		lifespans := config.Lifespans

		horizon = max(horizon, lifespans.AuthorizeCode, lifespans.DeviceCode)

		for _, custom := range lifespans.Custom {
			horizon = max(horizon, custom.AuthorizeCode, custom.DeviceCode, custom.Grants.AuthorizeCode.AuthorizeCode, custom.Grants.DeviceCode.AuthorizeCode)
		}
	}

	return horizon + logoutRevokeMargin
}
