// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"context"
	"sync"
	"time"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
)

var (
	oidcBackChannelLogoutDeliveryTimeout = time.Minute

	oidcBackChannelLogoutDeliveries sync.WaitGroup
)

func oidcLogoutSchedule(ctx *middlewares.AutheliaCtx) (schedule model.OAuth2SessionLogoutTaskSchedule) {
	now := ctx.GetClock().Now()

	schedule = oidc.NewLogoutRevokeSchedule(ctx.Configuration.IdentityProviders.OIDC, now)

	if ctx.Providers.OpenIDConnect == nil {
		return schedule
	}

	issuer, err := ctx.IssuerURL()
	if err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred scheduling the Back-Channel Logout requests: could not determine the issuer")

		return schedule
	}

	schedule.BackChannel = oidc.NewLogoutBackChannelSchedule(issuer, now)

	return schedule
}

func oidcBackChannelLogoutDeliver(ctx *middlewares.AutheliaCtx, sids []string) {
	if ctx.Providers.OpenIDConnect == nil || len(sids) == 0 {
		return
	}

	collector := oidc.NewLogoutTaskCollector(ctx.Providers.StorageProvider, ctx.GetClock(), oidc.NewBackChannelLogoutDeliverer(ctx.Providers.OpenIDConnect, oidc.NewDetachedContextProviders(ctx)))

	logger := ctx.GetLogger()

	oidcBackChannelLogoutDeliveries.Go(func() {
		c, cancel := context.WithTimeout(context.Background(), oidcBackChannelLogoutDeliveryTimeout)

		defer cancel()

		if err := collector.DeliverSessions(c, sids); err != nil {
			logger.WithError(err).Warn("Error occurred delivering Back-Channel Logout requests: they'll be retried")
		}
	})
}
