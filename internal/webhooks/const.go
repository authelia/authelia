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

	goneRecheckInterval = time.Hour

	grantNamePrefix = "wh-"
	grantNameLength = 20
	grantTimeout    = time.Second * 5
)

var errHandshakePending = errors.New("the destination withheld its confirmation and may grant it using the callback")

const (
	outcomeDelivered = "delivered"
	outcomeRetried   = "retried"
	outcomeDropped   = "dropped"
)

const (
	drainTimeout    = time.Second * 5
	validateTimeout = time.Second * 30
)
