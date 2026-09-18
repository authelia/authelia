// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultWebhookDestinationShouldHaveExpectedValues(t *testing.T) {
	assert.Equal(t, time.Second*10, DefaultWebhookDestination.Timeout)
	assert.Equal(t, 256, DefaultWebhookDestination.BufferSize)
	assert.False(t, DefaultWebhookDestination.DisableRedaction)
	assert.Equal(t, "sha256", DefaultWebhookDestination.Signature.Algorithm)
	assert.Equal(t, 3, DefaultWebhookDestination.Retry.Attempts)
	assert.Equal(t, time.Second, DefaultWebhookDestination.Retry.InitialInterval)
	assert.Equal(t, time.Minute, DefaultWebhookDestination.Retry.MaximumInterval)
}

func TestWebhooksShouldBePartOfConfiguration(t *testing.T) {
	config := &Configuration{}

	config.Webhooks.Destinations = []WebhookDestination{{Name: "example"}}

	assert.Equal(t, "example", config.Webhooks.Destinations[0].Name)
}
