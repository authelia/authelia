// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"net"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/storage"
)

func TestNewStorageLogsAuthPruneCmd(t *testing.T) {
	cmd := newStorageLogsAuthPruneCmd(&CmdCtx{})

	require.NotNil(t, cmd)

	flag := cmd.Flags().Lookup(cmdFlagLogsOlderThan)

	require.NotNil(t, flag)
	assert.Equal(t, []string{"true"}, flag.Annotations["cobra_annotation_bash_completion_one_required_flag"])
}

func TestRunStorageLogsAuthPrune(t *testing.T) {
	testCases := []struct {
		name      string
		olderThan time.Duration
		batchSize int
		dryRun    bool
		setup     func(store *mocks.MockStorage)
		expected  *regexp.Regexp
		err       string
	}{
		{
			name:      "ShouldErrorOnNonPositiveDuration",
			olderThan: 0,
			batchSize: 10,
			err:       "duration must be positive, got: 0s",
		},
		{
			name:      "ShouldErrorOnNonPositiveBatchSize",
			olderThan: time.Hour,
			batchSize: 0,
			err:       "batch size must be positive, got: 0",
		},
		{
			name:      "ShouldReportNothingToDeleteOnDryRun",
			olderThan: time.Hour,
			batchSize: 10,
			dryRun:    true,
			setup: func(store *mocks.MockStorage) {
				store.EXPECT().GetAuthenticationLogsPrunePreview(gomock.Any(), time.Hour).Return(&storage.DeleteAuthLogStats{}, nil)
			},
			expected: regexp.MustCompile(`^DRY RUN: Would delete logs older than 1h0m0s\.\.\.\nNo records found to delete\.\n$`),
		},
		{
			name:      "ShouldReportPreviewOnDryRun",
			olderThan: time.Hour,
			batchSize: 10,
			dryRun:    true,
			setup: func(store *mocks.MockStorage) {
				store.EXPECT().GetAuthenticationLogsPrunePreview(gomock.Any(), time.Hour).Return(&storage.DeleteAuthLogStats{TotalCount: 25, MinID: 4, MaxID: 90}, nil)
			},
			expected: regexp.MustCompile(`^DRY RUN: Would delete logs older than 1h0m0s\.\.\.\nWould delete 25 records in 3 batches\nRecord ID range: 4 to 90\nBatch size: 10\n$`),
		},
		{
			name:      "ShouldErrorOnDryRunPreviewError",
			olderThan: time.Hour,
			batchSize: 10,
			dryRun:    true,
			setup: func(store *mocks.MockStorage) {
				store.EXPECT().GetAuthenticationLogsPrunePreview(gomock.Any(), time.Hour).Return(nil, errors.New("bad conn"))
			},
			err: "failed to get preview: bad conn",
		},
		{
			name:      "ShouldReportNothingToDelete",
			olderThan: time.Hour,
			batchSize: 10,
			setup: func(store *mocks.MockStorage) {
				store.EXPECT().PruneAuthenticationLogs(gomock.Any(), time.Hour, 10).Return(&storage.DeleteAuthLogResults{}, nil)
			},
			expected: regexp.MustCompile(`^Deleting logs older than 1h0m0s\.\.\.\nNo records found to delete\.\n$`),
		},
		{
			name:      "ShouldReportDeleted",
			olderThan: time.Hour,
			batchSize: 10,
			setup: func(store *mocks.MockStorage) {
				store.EXPECT().PruneAuthenticationLogs(gomock.Any(), time.Hour, 10).Return(&storage.DeleteAuthLogResults{TotalBatches: 2, TotalDeleted: 15, TotalTime: time.Second * 3, AverageTime: time.Millisecond * 1500, RecordsPerSecond: 5}, nil)
			},
			expected: regexp.MustCompile(`^Deleting logs older than 1h0m0s\.\.\.\nSuccessfully deleted 15 records in 2 batches\.\nTotal duration: 3\.00s\nAverage time per batch: 1\.50s\nRecords per second: 5\n$`),
		},
		{
			name:      "ShouldReportPartialDeletionOnError",
			olderThan: time.Hour,
			batchSize: 10,
			setup: func(store *mocks.MockStorage) {
				store.EXPECT().PruneAuthenticationLogs(gomock.Any(), time.Hour, 10).Return(&storage.DeleteAuthLogResults{TotalBatches: 1, TotalDeleted: 10}, errors.New("bad conn"))
			},
			expected: regexp.MustCompile(`^Deleting logs older than 1h0m0s\.\.\.\nPartially completed: deleted 10 records in 1 batches before error\n$`),
			err:      "pruning failed: bad conn",
		},
		{
			name:      "ShouldNotReportPartialDeletionWhenNothingDeleted",
			olderThan: time.Hour,
			batchSize: 10,
			setup: func(store *mocks.MockStorage) {
				store.EXPECT().PruneAuthenticationLogs(gomock.Any(), time.Hour, 10).Return(&storage.DeleteAuthLogResults{}, errors.New("bad conn"))
			},
			expected: regexp.MustCompile(`^Deleting logs older than 1h0m0s\.\.\.\n$`),
			err:      "pruning failed: bad conn",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			store := mocks.NewMockStorage(ctrl)

			if tc.setup != nil {
				tc.setup(store)
			}

			buf := new(bytes.Buffer)

			err := runStorageLogsAuthPrune(context.Background(), buf, store, tc.olderThan, tc.batchSize, tc.dryRun)

			if tc.err == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.err)
			}

			if tc.expected != nil {
				assert.Regexp(t, tc.expected, buf.String())
			}
		})
	}
}

func TestRunStorageLogsAuthPruneSQLite(t *testing.T) {
	store := newTestSQLiteStore(t)
	ctx := context.Background()
	now := time.Now()

	for i := 0; i < 5; i++ {
		require.NoError(t, store.AppendAuthenticationLog(ctx, testStorageLogsAttempt(now.Add(-time.Hour*24*time.Duration(30+i)), false, false)))
	}

	require.NoError(t, store.AppendAuthenticationLog(ctx, testStorageLogsAttempt(now, true, false)))

	buf := new(bytes.Buffer)

	require.NoError(t, runStorageLogsAuthPrune(ctx, buf, store, time.Hour*24*7, 2, false))
	assert.Regexp(t, regexp.MustCompile(`^Deleting logs older than 168h0m0s\.\.\.\nSuccessfully deleted 5 records in 3 batches\.\n`), buf.String())

	stats, err := store.GetAuthenticationLogsStats(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), stats.Total)
}

func TestRunStorageLogsAuthStats(t *testing.T) {
	oldest := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	newest := time.Date(2026, 6, 7, 8, 9, 10, 0, time.UTC)

	testCases := []struct {
		name     string
		setup    func(store *mocks.MockStorage)
		expected string
		err      string
	}{
		{
			name: "ShouldReportEmptyTable",
			setup: func(store *mocks.MockStorage) {
				store.EXPECT().GetAuthenticationLogsStats(gomock.Any()).Return(&storage.AuthLogStats{}, nil)
			},
			expected: "Authentication Log Statistics:\nTotal records: 0\n",
		},
		{
			name: "ShouldReportStatistics",
			setup: func(store *mocks.MockStorage) {
				store.EXPECT().GetAuthenticationLogsStats(gomock.Any()).Return(&storage.AuthLogStats{
					Total:        4,
					SuccessCount: 1,
					FailureCount: 3,
					BannedCount:  2,
					Oldest:       sql.NullTime{Time: oldest, Valid: true},
					Newest:       sql.NullTime{Time: newest, Valid: true},
				}, nil)
			},
			expected: "Authentication Log Statistics:\nTotal records: 4\nSuccessful logins: 1 (25.0%)\nFailed logins: 3 (75.0%)\nBanned attempts: 2 (50.0%)\nOldest record: 2026-01-02T03:04:05Z\nNewest record: 2026-06-07T08:09:10Z\n",
		},
		{
			name: "ShouldErrorOnStatsError",
			setup: func(store *mocks.MockStorage) {
				store.EXPECT().GetAuthenticationLogsStats(gomock.Any()).Return(nil, errors.New("bad conn"))
			},
			err: "failed to get authentication log statistics: bad conn",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			store := mocks.NewMockStorage(ctrl)

			if tc.setup != nil {
				tc.setup(store)
			}

			buf := new(bytes.Buffer)

			err := runStorageLogsAuthStats(context.Background(), buf, store)

			if tc.err == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.err)
			}

			assert.Equal(t, tc.expected, buf.String())
		})
	}
}

func TestRunStorageLogsAuthStatsSQLite(t *testing.T) {
	store := newTestSQLiteStore(t)
	ctx := context.Background()

	buf := new(bytes.Buffer)

	require.NoError(t, runStorageLogsAuthStats(ctx, buf, store))
	assert.Equal(t, "Authentication Log Statistics:\nTotal records: 0\n", buf.String())

	require.NoError(t, store.AppendAuthenticationLog(ctx, testStorageLogsAttempt(time.Now(), true, false)))

	buf.Reset()

	require.NoError(t, runStorageLogsAuthStats(ctx, buf, store))
	assert.Regexp(t, regexp.MustCompile(`^Authentication Log Statistics:\nTotal records: 1\nSuccessful logins: 1 \(100\.0%\)\nFailed logins: 0 \(0\.0%\)\nBanned attempts: 0 \(0\.0%\)\nOldest record: \S+\nNewest record: \S+\n$`), buf.String())
}

func testStorageLogsAttempt(at time.Time, successful, banned bool) model.AuthenticationAttempt {
	return model.AuthenticationAttempt{
		Time:       at,
		Successful: successful,
		Banned:     banned,
		Username:   "john",
		Type:       "1fa",
		RemoteIP:   model.NewNullIP(net.ParseIP("127.0.0.1")),
	}
}
