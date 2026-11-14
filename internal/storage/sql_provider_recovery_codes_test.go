// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/model"
)

func TestSQLProviderReplaceRecoveryCodesByUsernameShouldRollbackOnInsertFailure(t *testing.T) {
	provider := newTestSQLiteProviderWithEncryption(t)

	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()
	createdAt := time.Unix(1700000000, 0)

	require.NoError(t, provider.ReplaceRecoveryCodesByUsername(ctx, "john", []*model.RecoveryCode{
		{Username: "john", Plaintext: "AAAAA-BBBBB", CreatedAt: createdAt},
	}, model.NullIP{}))

	// The second code has no plaintext so it fails after the first insert has succeeded.
	assert.Error(t, provider.ReplaceRecoveryCodesByUsername(ctx, "john", []*model.RecoveryCode{
		{Username: "john", Plaintext: "CCCCC-DDDDD", CreatedAt: createdAt.Add(time.Hour)},
		{Username: "john", CreatedAt: createdAt.Add(time.Hour)},
	}, model.NullIP{}))

	codes, err := provider.LoadRecoveryCodesByUsername(ctx, "john")
	require.NoError(t, err)
	require.Len(t, codes, 1)

	assert.False(t, codes[0].RevokedAt.Valid)

	code, err := provider.LoadRecoveryCode(ctx, "john", "AAAAA-BBBBB")
	require.NoError(t, err)
	assert.NotNil(t, code)
}

func TestSQLProviderConsumeRecoveryCodeShouldOnlyConsumeActiveCodes(t *testing.T) {
	provider := newTestSQLiteProviderWithEncryption(t)

	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()

	require.NoError(t, provider.ReplaceRecoveryCodesByUsername(ctx, "john", []*model.RecoveryCode{
		{Username: "john", Plaintext: "AAAAA-BBBBB", CreatedAt: time.Now()},
		{Username: "john", Plaintext: "CCCCC-DDDDD", CreatedAt: time.Now()},
	}, model.NullIP{}))

	consumed, err := provider.LoadRecoveryCode(ctx, "john", "AAAAA-BBBBB")
	require.NoError(t, err)
	require.NotNil(t, consumed)

	require.NoError(t, provider.ConsumeRecoveryCode(ctx, consumed.ID, model.NullIP{}))
	assert.ErrorIs(t, provider.ConsumeRecoveryCode(ctx, consumed.ID, model.NullIP{}), ErrRecoveryCodeNotActive)

	revoked, err := provider.LoadRecoveryCode(ctx, "john", "CCCCC-DDDDD")
	require.NoError(t, err)
	require.NotNil(t, revoked)

	require.NoError(t, provider.RevokeRecoveryCodesByUsername(ctx, "john", model.NullIP{}))
	assert.ErrorIs(t, provider.ConsumeRecoveryCode(ctx, revoked.ID, model.NullIP{}), ErrRecoveryCodeNotActive)
}

func TestSQLProviderReplaceRecoveryCodesByUsernameShouldSerializeConcurrentReplacements(t *testing.T) {
	provider := newTestSQLiteProviderWithEncryption(t)

	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()

	const workers, size = 8, 3

	var wg sync.WaitGroup

	errs := make(chan error, workers)

	for i := 0; i < workers; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			batch := make([]*model.RecoveryCode, 0, size)

			for j := 0; j < size; j++ {
				batch = append(batch, &model.RecoveryCode{Username: "john", Plaintext: fmt.Sprintf("CODE-%d-%d", i, j), CreatedAt: time.Now()})
			}

			errs <- provider.ReplaceRecoveryCodesByUsername(ctx, "john", batch, model.NullIP{})
		}(i)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		require.NoError(t, err)
	}

	count, err := provider.CountUnusedRecoveryCodesByUsername(ctx, "john")
	require.NoError(t, err)
	assert.Equal(t, size, count)
}
