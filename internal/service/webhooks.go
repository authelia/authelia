// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"github.com/authelia/authelia/v4/internal/webhooks"
)

// ProvisionWebhooks provisions the service which delivers events to the configured webhook destinations. It returns
// nil when no destinations are configured.
func ProvisionWebhooks(ctx Context) (service Provider, err error) {
	dispatcher, ok := ctx.GetProviders().Events.(*webhooks.Dispatcher)
	if !ok || dispatcher == nil {
		return nil, nil
	}

	return webhooks.NewService("main", dispatcher, ctx.GetConfiguration().Webhooks.StartupCheck, ctx.GetLogger().WithField(logFieldService, serviceTypeWebhooks)), nil
}
