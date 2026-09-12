// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/events"
)

func originOf(config schema.WebhookValidation, derived string) (origin string, err error) {
	if config.ForceOrigin {
		if config.Origin == "" {
			return "", fmt.Errorf("option 'force_origin' is enabled but option 'origin' is not configured")
		}

		return config.Origin, nil
	}

	if derived != "" {
		return derived, nil
	}

	if config.Origin == "" {
		return "", fmt.Errorf("no origin could be derived from the configuration and option 'origin' is not configured")
	}

	return config.Origin, nil
}

func originFromSource(source string) (origin string) {
	if source == "" {
		return ""
	}

	parsed, err := url.Parse(source)
	if err != nil {
		return ""
	}

	return parsed.Hostname()
}

func requestedRate(config schema.WebhookValidation) (rate string, send bool) {
	if config.Rate <= 0 {
		return "", false
	}

	return strconv.Itoa(config.Rate), true
}

func (d *destination) validate(origin string) (rate int, err error) {
	deadline := time.Now().Add(validateTimeout)
	interval := d.config.Retry.InitialInterval

	for attempt := 1; attempt <= d.config.Retry.Attempts; attempt++ {
		if rate, err = d.validateAttempt(origin, attempt); err == nil {
			return rate, nil
		}

		var permanent *permanentError

		if errors.As(err, &permanent) {
			return 0, err
		}

		if attempt == d.config.Retry.Attempts {
			return 0, err
		}

		wait := interval

		if after, ok := retryAfterOf(err); ok {
			wait = after
		}

		if wait > d.config.Retry.MaximumInterval {
			wait = d.config.Retry.MaximumInterval
		}

		wait = jitter(wait)

		if time.Now().Add(wait).After(deadline) {
			return 0, fmt.Errorf("%w: the handshake budget of %s elapsed before it could be retried", err, validateTimeout)
		}

		d.log.WithError(err).WithFields(map[string]any{"attempt": attempt, "retry_in": wait.String()}).
			Warn("Webhook destination handshake failed and will be retried")

		timer := time.NewTimer(wait)

		select {
		case <-timer.C:
		case <-d.quit:
			timer.Stop()

			return 0, fmt.Errorf("%w: the handshake was abandoned during shutdown", err)
		}

		timer.Stop()

		if interval *= 2; interval > d.config.Retry.MaximumInterval {
			interval = d.config.Retry.MaximumInterval
		}
	}

	return 0, err
}

func (d *destination) validateAttempt(origin string, attempt int) (rate int, err error) {
	req, err := http.NewRequestWithContext(d.ctx, http.MethodOptions, d.config.Address.String(), nil)
	if err != nil {
		return 0, &permanentError{err: fmt.Errorf("error building request: %w", err)}
	}

	req.Header.Set("User-Agent", "Authelia/"+d.version)
	req.Header.Set(events.HeaderWebhookRequestOrigin, origin)
	req.Header.Set(events.HeaderDeliveryAttempt, strconv.Itoa(attempt))

	requested, send := requestedRate(d.config.Validation)

	if send {
		req.Header.Set(events.HeaderWebhookRequestRate, requested)
	}

	d.setHeaders(req)

	resp, err := d.client.Do(req)
	if err != nil {
		// No status code was produced, so this is a transport failure and worth another attempt.
		return 0, fmt.Errorf("error performing request: %w", err)
	}

	defer resp.Body.Close()

	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, responseBodyLimit))

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		break
	case resp.StatusCode == http.StatusTooManyRequests:
		return 0, &retryableError{status: resp.StatusCode, after: parseRetryAfter(resp.Header.Get("Retry-After"))}
	case resp.StatusCode >= 500:
		return 0, &retryableError{status: resp.StatusCode}
	default:
		return 0, &permanentError{err: fmt.Errorf("received status code %d", resp.StatusCode)}
	}

	allowed := resp.Header.Get(events.HeaderWebhookAllowedOrigin)

	switch {
	case allowed == "":
		return 0, &permanentError{err: fmt.Errorf("the response did not include a %s header", events.HeaderWebhookAllowedOrigin)}
	case allowed == events.WebhookRateUnlimited:
		break
	case !strings.EqualFold(allowed, origin):
		return 0, &permanentError{err: fmt.Errorf("the response permitted the origin '%s' rather than '%s'", allowed, origin)}
	}

	permitted := resp.Header.Get(events.HeaderWebhookAllowedRate)

	if send && permitted == "" {
		return 0, &permanentError{err: fmt.Errorf("the response did not include a %s header for the requested rate of %s", events.HeaderWebhookAllowedRate, requested)}
	}

	if rate, err = allowedRate(permitted); err != nil {
		return 0, &permanentError{err: err}
	}

	return rate, nil
}

func allowedRate(value string) (rate int, err error) {
	if value == "" || value == events.WebhookRateUnlimited {
		return 0, nil
	}

	if rate, err = strconv.Atoi(value); err != nil || rate <= 0 {
		return 0, fmt.Errorf("the response included an invalid %s header with value '%s'", events.HeaderWebhookAllowedRate, value)
	}

	return rate, nil
}

func interval(rate int) time.Duration {
	if rate <= 0 {
		return 0
	}

	return time.Minute / time.Duration(rate)
}
