// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/model"
)

func TestShouldRotateTheWebhookCallbackKeyAndDeleteOnlyTheConfirmations(t *testing.T) {
	provider := newTestSQLiteProviderWithEncryption(t)

	ctx := context.Background()

	require.NoError(t, provider.StartupCheck())

	before := provider.WebhookCallbackSignature([]byte("destination"))

	require.NotEmpty(t, before)

	require.NoError(t, provider.SaveCachedData(ctx, model.CachedData{Name: CachedDataNamePrefixWebhookConfirmation + "abc", Encrypted: true, Value: []byte(`{}`)}))
	require.NoError(t, provider.SaveCachedData(ctx, model.CachedData{Name: "other", Value: []byte(`{}`)}))

	require.NoError(t, provider.SchemaEncryptionRotateHMACKey(ctx, hmacNameWebhookCallback))

	confirmation, err := provider.LoadCachedData(ctx, CachedDataNamePrefixWebhookConfirmation+"abc")

	require.NoError(t, err)
	assert.Nil(t, confirmation, "the confirmations given using the old callback addresses must be withdrawn")

	other, err := provider.LoadCachedData(ctx, "other")

	require.NoError(t, err)
	assert.NotNil(t, other, "cached data which is not a webhook confirmation must be kept")

	require.NoError(t, provider.StartupCheck())

	assert.NotEqual(t, before, provider.WebhookCallbackSignature([]byte("destination")), "the callback key must change")
}
