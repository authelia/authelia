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

	for i := range workers {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			batch := make([]*model.RecoveryCode, 0, size)

			for j := range size {
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

func TestSQLProviderLoadRecoveryCodeHMACKey(t *testing.T) {
	provider := newTestSQLiteProviderWithEncryption(t)

	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()

	expected := provider.keys.recoveryCodeHMAC
	require.Len(t, expected, 64)

	provider.keys.recoveryCodeHMAC = nil

	require.NoError(t, provider.LoadRecoveryCodeHMACKey(ctx))
	assert.Equal(t, expected, provider.keys.recoveryCodeHMAC)
}

func TestSQLProviderLoadRecoveryCodeShouldReturnNilWhenNotFound(t *testing.T) {
	provider := newTestSQLiteProviderWithEncryption(t)

	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()

	require.NoError(t, provider.ReplaceRecoveryCodesByUsername(ctx, "john", []*model.RecoveryCode{
		{Username: "john", Plaintext: "AAAAA-BBBBB", CreatedAt: time.Now()},
	}, model.NullIP{}))

	testCases := []struct {
		name     string
		username string
		raw      string
		found    bool
	}{
		{"ShouldFindExactCode", "john", "AAAAA-BBBBB", true},
		{"ShouldFindNormalizedCode", "john", " aaaaa bbbbb ", true},
		{"ShouldNotFindUnknownCode", "john", "ZZZZZ-ZZZZZ", false},
		{"ShouldNotFindCodeForOtherUser", "harry", "AAAAA-BBBBB", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			code, err := provider.LoadRecoveryCode(ctx, tc.username, tc.raw)

			require.NoError(t, err)

			if tc.found {
				require.NotNil(t, code)
				assert.Equal(t, tc.username, code.Username)
			} else {
				assert.Nil(t, code)
			}
		})
	}
}

func TestSQLProviderCountRecoveryCodes(t *testing.T) {
	provider := newTestSQLiteProviderWithEncryption(t)

	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()

	serial := 0

	seed := func(username string, n int) []model.RecoveryCode {
		batch := make([]*model.RecoveryCode, 0, n)

		for range n {
			serial++

			batch = append(batch, &model.RecoveryCode{Username: username, Plaintext: fmt.Sprintf("CODE-%d", serial), CreatedAt: time.Now()})
		}

		require.NoError(t, provider.ReplaceRecoveryCodesByUsername(ctx, username, batch, model.NullIP{}))

		codes, err := provider.LoadRecoveryCodesByUsername(ctx, username)
		require.NoError(t, err)

		return codes
	}

	seed("alice", 5)

	bob := seed("bob", 3)
	require.NoError(t, provider.ConsumeRecoveryCode(ctx, bob[0].ID, model.NullIP{}))

	seed("carol", 1)

	dave := seed("dave", 2)
	require.NoError(t, provider.ConsumeRecoveryCode(ctx, dave[0].ID, model.NullIP{}))
	require.NoError(t, provider.RevokeRecoveryCodesByUsername(ctx, "dave", model.NullIP{}))

	seed("erin", 1)
	require.NoError(t, provider.RevokeRecoveryCodesByUsername(ctx, "erin", model.NullIP{}))

	frank := seed("frank", 1)
	require.NoError(t, provider.ConsumeRecoveryCode(ctx, frank[0].ID, model.NullIP{}))
	seed("frank", 4)

	unused := map[string]int{"alice": 5, "bob": 2, "carol": 1, "dave": 0, "erin": 0, "frank": 4, "nobody": 0}

	for username, expected := range unused {
		count, err := provider.CountUnusedRecoveryCodesByUsername(ctx, username)
		require.NoError(t, err)
		assert.Equal(t, expected, count, username)
	}

	users, err := provider.CountUsersWithRecoveryCodes(ctx)
	require.NoError(t, err)
	assert.Equal(t, 4, users)

	low, err := provider.CountUsersWithLowRecoveryCodes(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, low)

	depleted, err := provider.CountUsersWithDepletedRecoveryCodes(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, depleted)
}

func TestSQLProviderCountRecoveryCodesShouldReturnZeroWhenEmpty(t *testing.T) {
	provider := newTestSQLiteProviderWithEncryption(t)

	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()

	users, err := provider.CountUsersWithRecoveryCodes(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, users)

	low, err := provider.CountUsersWithLowRecoveryCodes(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, low)

	depleted, err := provider.CountUsersWithDepletedRecoveryCodes(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, depleted)
}

func TestSQLProviderStartupCheckShouldErrOnInvalidRecoveryCodeHMACKey(t *testing.T) {
	provider := newTestSQLiteProviderWithEncryption(t)

	require.NoError(t, provider.StartupCheck())

	_, err := provider.db.ExecContext(context.Background(), "UPDATE encryption SET value = ? WHERE name = ?;", []byte("not-a-valid-ciphertext"), "hmac_key_rc")
	require.NoError(t, err)

	assert.ErrorContains(t, provider.StartupCheck(), "failed to initialize the hmac recovery code signature key during startup: ")
}
