// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/model"
)

func TestSQLProviderGetAuthenticationLogsStats(t *testing.T) {
	now := time.Now().Truncate(time.Second)

	testCases := []struct {
		name                            string
		attempts                        []model.AuthenticationAttempt
		total, success, failure, banned int64
		oldest, newest                  time.Time
	}{
		{
			name: "ShouldReturnZeroValuesForEmptyTable",
		},
		{
			name: "ShouldReturnStatistics",
			attempts: []model.AuthenticationAttempt{
				testAuthenticationAttempt(now.Add(-time.Hour*48), true, false),
				testAuthenticationAttempt(now.Add(-time.Hour), false, false),
				testAuthenticationAttempt(now.Add(-time.Minute), false, true),
			},
			total:   3,
			success: 1,
			failure: 2,
			banned:  1,
			oldest:  now.Add(-time.Hour * 48),
			newest:  now.Add(-time.Minute),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := newTestSQLiteProvider(t)
			require.NoError(t, provider.StartupCheck())

			ctx := context.Background()

			for _, attempt := range tc.attempts {
				require.NoError(t, provider.AppendAuthenticationLog(ctx, attempt))
			}

			stats, err := provider.GetAuthenticationLogsStats(ctx)
			require.NoError(t, err)

			assert.Equal(t, tc.total, stats.Total)
			assert.Equal(t, tc.success, stats.SuccessCount)
			assert.Equal(t, tc.failure, stats.FailureCount)
			assert.Equal(t, tc.banned, stats.BannedCount)
			assert.Equal(t, !tc.oldest.IsZero(), stats.Oldest.Valid)
			assert.Equal(t, !tc.newest.IsZero(), stats.Newest.Valid)
			assert.True(t, tc.oldest.Equal(stats.Oldest.Time))
			assert.True(t, tc.newest.Equal(stats.Newest.Time))
		})
	}
}

func TestSQLProviderAuthenticationLogsPrune(t *testing.T) {
	now := time.Now().Truncate(time.Second)

	testCases := []struct {
		name            string
		old, recent     int
		batchSize       int
		expectedBatches int
		err             string
	}{
		{
			name:      "ShouldDeleteNothingForEmptyTable",
			batchSize: 10,
		},
		{
			name:      "ShouldDeleteNothingWhenAllRecent",
			recent:    5,
			batchSize: 10,
		},
		{
			name:            "ShouldDeleteInSingleBatch",
			old:             5,
			recent:          3,
			batchSize:       10,
			expectedBatches: 1,
		},
		{
			name:            "ShouldDeleteInMultipleBatches",
			old:             7,
			recent:          3,
			batchSize:       3,
			expectedBatches: 3,
		},
		{
			name:            "ShouldDeleteInExactMultipleOfBatchSize",
			old:             6,
			recent:          2,
			batchSize:       3,
			expectedBatches: 2,
		},
		{
			name:      "ShouldErrorOnInvalidBatchSize",
			old:       2,
			batchSize: 0,
			err:       "error pruning table 'authentication_logs': batch size must be greater than 0 but is 0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := newTestSQLiteProvider(t)
			require.NoError(t, provider.StartupCheck())

			ctx := context.Background()

			for i := 0; i < tc.recent; i++ {
				require.NoError(t, provider.AppendAuthenticationLog(ctx, testAuthenticationAttempt(now.Add(-time.Minute*time.Duration(i)), true, false)))
			}

			for i := 0; i < tc.old; i++ {
				require.NoError(t, provider.AppendAuthenticationLog(ctx, testAuthenticationAttempt(now.Add(-time.Hour*24*time.Duration(10+i)), false, false)))
			}

			preview, err := provider.GetAuthenticationLogsPrunePreview(ctx, time.Hour*24*7)
			require.NoError(t, err)
			assert.Equal(t, int64(tc.old), preview.TotalCount)

			results, err := provider.PruneAuthenticationLogs(ctx, time.Hour*24*7, tc.batchSize)
			require.NotNil(t, results)

			if tc.err != "" {
				assert.EqualError(t, err, tc.err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, int64(tc.old), results.TotalDeleted)
			assert.Equal(t, tc.expectedBatches, results.TotalBatches)

			stats, err := provider.GetAuthenticationLogsStats(ctx)
			require.NoError(t, err)
			assert.Equal(t, int64(tc.recent), stats.Total)

			preview, err = provider.GetAuthenticationLogsPrunePreview(ctx, time.Hour*24*7)
			require.NoError(t, err)
			assert.Equal(t, &DeleteAuthLogStats{}, preview)
		})
	}
}

func testAuthenticationAttempt(at time.Time, successful, banned bool) model.AuthenticationAttempt {
	return model.AuthenticationAttempt{
		Time:       at,
		Successful: successful,
		Banned:     banned,
		Username:   "john",
		Type:       "1fa",
		RemoteIP:   model.NewNullIP(net.ParseIP("127.0.0.1")),
	}
}
