// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"sync"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
)

// HealthVerboseResponse is the body of the verbose health check endpoint.
type HealthVerboseResponse struct {
	Status    string                        `json:"status"`
	CheckedAt time.Time                     `json:"checked_at"`
	Cached    bool                          `json:"cached"`
	Providers map[string]HealthVerboseCheck `json:"providers"`
}

// HealthVerboseCheck is the outcome for a single provider. The error is omitted unless the endpoint is configured
// as detailed, as a provider error carries connection strings and host names an anonymous caller has no business
// reading.
type HealthVerboseCheck struct {
	Status string `json:"status"`
	Took   string `json:"took"`
	Error  string `json:"error,omitempty"`
}

type healthVerboseCache struct {
	mu       sync.Mutex
	at       time.Time
	response HealthVerboseResponse
	status   int
}

// HealthVerboseGET returns a handler which probes the configured providers and reports the outcome of each.
//
// The result is cached for the configured duration so that a readiness probe polling every second does not open a
// connection to every provider every second. The response says whether it was served from that cache.
func HealthVerboseGET(config schema.ServerEndpointHealth) middlewares.RequestHandler {
	cache := &healthVerboseCache{}

	return func(ctx *middlewares.AutheliaCtx) {
		response, status, cached := cache.get(ctx, config)

		response.Cached = cached

		// deliberately not SetJSONBody: that wraps the value in the OKResponse envelope, which would report
		// "status": "OK" alongside a 503 and a body saying otherwise.
		if err := ctx.ReplyJSON(response, status); err != nil {
			ctx.Logger.WithError(err).Error("Error occurred writing the verbose health check response")
		}
	}
}

func (c *healthVerboseCache) get(ctx *middlewares.AutheliaCtx, config schema.ServerEndpointHealth) (response HealthVerboseResponse, status int, cached bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := ctx.GetClock().Now()

	if config.Cache > 0 && !c.at.IsZero() && now.Sub(c.at) < config.Cache {
		return c.response, c.status, true
	}

	response, status = healthVerboseProbe(ctx, config, now)

	c.at, c.response, c.status = now, response, status

	return response, status, false
}

func healthVerboseProbe(ctx *middlewares.AutheliaCtx, config schema.ServerEndpointHealth, now time.Time) (response HealthVerboseResponse, status int) {
	checks := ctx.Providers.HealthChecks(ctx.GetClock(), config.Providers)

	response = HealthVerboseResponse{
		Status:    "ok",
		CheckedAt: now,
		Providers: make(map[string]HealthVerboseCheck, len(checks)),
	}

	status = fasthttp.StatusOK

	for _, check := range checks {
		result := HealthVerboseCheck{Status: "ok", Took: check.Took.String()}

		if check.Err != nil {
			result.Status = "error"

			if config.Detailed {
				result.Error = check.Err.Error()
			}

			ctx.Logger.WithError(check.Err).WithField("provider", check.Name).
				Error("Error occurred performing a health check")
		}

		response.Providers[check.Name] = result
	}

	if !middlewares.HealthChecksOK(checks) {
		response.Status = "error"
		status = fasthttp.StatusServiceUnavailable
	}

	return response, status
}
