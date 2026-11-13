// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/model"
)

const backChannelLogoutDeliveryTimeout = 35 * time.Second

// NewBackChannelLogoutDeliverer returns a *BackChannelLogoutDeliverer which delivers Logout Tokens with the provider
// using the providers.
func NewBackChannelLogoutDeliverer(provider *OpenIDConnectProvider, providers DetachedContextProviders) *BackChannelLogoutDeliverer {
	return &BackChannelLogoutDeliverer{provider: provider, providers: providers}
}

// BackChannelLogoutDeliverer delivers the Logout Token of a task of the kind model.OAuth2SessionLogoutTaskKindBackChannel
// to its client, as described by OpenID Connect Back-Channel Logout 1.0.
type BackChannelLogoutDeliverer struct {
	provider  *OpenIDConnectProvider
	providers DetachedContextProviders
}

// AcquireBackChannelLogoutSlot blocks until a Logout Token may be delivered, or the context is done.
func (d *BackChannelLogoutDeliverer) AcquireBackChannelLogoutSlot(ctx context.Context) (release func(), err error) {
	return d.provider.AcquireBackChannelLogoutSlot(ctx)
}

// DeliverBackChannelLogout delivers a newly generated Logout Token for the task to its client. It returns true when the
// task is done, which is when the client acknowledged the Logout Token, when the client is no longer registered or no
// longer has a back-channel logout URI, or when the client rejected the Logout Token with a response which won't
// succeed if it's retried, in which case an error is also returned. It returns false with an error when the delivery
// may succeed if it's retried, which is when the client couldn't be reached, is rate limiting, or responded with a
// server error. The delivery is cancelled when the context is cancelled, but isn't bound by its deadline, so the time
// spent waiting to deliver it doesn't count towards the time allowed for the delivery.
func (d *BackChannelLogoutDeliverer) DeliverBackChannelLogout(ctx context.Context, task model.OAuth2SessionLogoutTask) (done bool, err error) {
	client, err := d.provider.GetRegisteredClient(ctx, task.ClientID)
	if err != nil || client.GetBackChannelLogoutURI() == "" {
		return true, nil
	}

	issuer, err := url.Parse(task.IssuerURL)
	if err != nil {
		return true, fmt.Errorf("the issuer could not be parsed: %w", err)
	}

	parent, stop := withoutDeadline(ctx)

	defer stop()

	detached, cancel := NewDetachedContextFromProviders(parent, min(d.provider.GetBackChannelLogoutLifespan(ctx), backChannelLogoutDeliveryTimeout), issuer, d.providers)

	defer cancel()

	results, err := d.provider.SendBackChannelLogout(detached, oauthelia2.NewBackChannelLogoutRequest(task.Subject, task.SessionID, []oauthelia2.Client{client}))
	if err != nil {
		return true, err
	}

	if len(results) != 1 {
		return true, fmt.Errorf("expected a single result but received %d", len(results))
	}

	result := results[0]

	switch {
	case result.Skipped, result.Err == nil:
		return true, nil
	case IsBackChannelLogoutRetryable(result.Status):
		return false, fmt.Errorf("the client responded with status code '%d' and the request will be retried: %w", result.Status, result.Err)
	default:
		return true, fmt.Errorf("the client responded with status code '%d': %w", result.Status, result.Err)
	}
}

// IsBackChannelLogoutRetryable returns true when a delivery which failed with the status code may succeed if it's
// retried, i.e. when the client couldn't be reached, is rate limiting, or responded with a server error.
func IsBackChannelLogoutRetryable(status int) bool {
	return status == 0 || status == http.StatusTooManyRequests || status >= http.StatusInternalServerError
}

func withoutDeadline(parent context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.WithoutCancel(parent))

	stop := context.AfterFunc(parent, func() {
		if errors.Is(parent.Err(), context.Canceled) {
			cancel()
		}
	})

	return ctx, func() {
		stop()
		cancel()
	}
}
