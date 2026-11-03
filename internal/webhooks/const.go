// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"errors"
	"time"
)

const (
	responseBodyLimit = 4096
)

const (
	// PathConfirm is the path of the endpoint a destination requests to confirm it accepts deliveries.
	PathConfirm = "/api/webhooks/confirm"

	// QueryConfirmID is the query parameter of the confirmation endpoint which names the destination.
	QueryConfirmID = "id"

	// QueryConfirmKey is the query parameter of the confirmation endpoint which carries the destination's key.
	QueryConfirmKey = "key"

	callbackSecretLength = 32
	callbackKeyContext   = "authelia webhooks callback v1"

	pendingRecheckInterval = time.Minute

	unansweredRecheckInterval = time.Minute * 5

	goneRecheckInterval = time.Hour

	// retryAfterLimit is the longest wait a Retry-After header can impose, so a single response cannot silence a
	// destination for longer than this.
	retryAfterLimit = time.Hour

	grantNamePrefix = "wh-"
	grantNameLength = 20
	grantTimeout    = time.Second * 5
)

var errHandshakePending = errors.New("the destination withheld its confirmation and may grant it using the callback")

// errHandshakeRetired is the handshake answered with 410 Gone, the one status code the webhook specification gives a
// receiver to withdraw from deliveries. Unlike every other answer without a WebHook-Allowed-Origin header, it cancels a
// confirmation recorded from the callback.
var errHandshakeRetired = errors.New("received status code 410")

const (
	outcomeDelivered = "delivered"
	outcomeRetried   = "retried"
	outcomeDropped   = "dropped"
)

const (
	drainTimeout    = time.Second * 5
	validateTimeout = time.Second * 30
)
