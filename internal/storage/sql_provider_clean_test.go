// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package storage_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/storage"
)

type cleanResult struct {
	affected int64
	err      error
}

func (r cleanResult) LastInsertId() (int64, error) { return 0, nil }
func (r cleanResult) RowsAffected() (int64, error) { return r.affected, r.err }

func TestSQLProviderCountStaleOAuth2ConsentSessions(t *testing.T) {
	before := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)

	testCases := []struct {
		name     string
		have     error
		expected int
		err      string
	}{
		{
			"ShouldReturnTheCount",
			nil,
			0,
			"",
		},
		{
			"ShouldWrapTheProviderError",
			errors.New("connection refused"),
			0,
			"error counting stale oauth2 consent sessions: connection refused",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			db := mocks.NewMockSQLXDB(ctrl)
			p := storage.NewSQLProviderForTesting(db)

			db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), before, before).Return(tc.have)

			count, err := p.CountStaleOAuth2ConsentSessions(context.Background(), before)

			if tc.err == "" {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, count)
			} else {
				assert.EqualError(t, err, tc.err)
				assert.Equal(t, 0, count)
			}
		})
	}
}

func TestSQLProviderDeleteStaleOAuth2ConsentSessions(t *testing.T) {
	before := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)

	testCases := []struct {
		name     string
		result   sql.Result
		have     error
		expected int
		err      string
	}{
		{
			"ShouldReturnTheNumberDeleted",
			cleanResult{affected: 42},
			nil,
			42,
			"",
		},
		{
			"ShouldReturnZeroWhenNothingWasStale",
			cleanResult{affected: 0},
			nil,
			0,
			"",
		},
		{
			"ShouldWrapTheProviderError",
			nil,
			errors.New("connection refused"),
			0,
			"error deleting stale oauth2 consent sessions: connection refused",
		},
		{
			"ShouldWrapAnUncountableResult",
			cleanResult{err: errors.New("unsupported")},
			nil,
			0,
			"error determining the number of stale oauth2 consent sessions deleted: unsupported",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			db := mocks.NewMockSQLXDB(ctrl)
			p := storage.NewSQLProviderForTesting(db)

			db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), before, before).Return(tc.result, tc.have)

			deleted, err := p.DeleteStaleOAuth2ConsentSessions(context.Background(), before)

			if tc.err == "" {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, deleted)
			} else {
				assert.EqualError(t, err, tc.err)
				assert.Equal(t, 0, deleted)
			}
		})
	}
}
