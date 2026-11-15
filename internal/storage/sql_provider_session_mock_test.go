// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package storage_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	"github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/storage"
)

func TestSQLProviderSessionGetIDsByUsernameShouldReturnNilWhenThereAreNoRows(t *testing.T) {
	ctrl := gomock.NewController(t)

	db := mocks.NewMockSQLXDB(ctrl)

	db.EXPECT().SelectContext(gomock.Any(), gomock.Any(), gomock.Any(), "an-issuer", "john", gomock.Any()).Return(sql.ErrNoRows)

	p := storage.NewSQLProviderForTesting(db)

	ids, err := p.SessionGetIDsByUsername(t.Context(), "an-issuer", "john")

	assert.NoError(t, err)
	assert.Nil(t, ids)
}

func TestSQLProviderSessionSaveShouldCheckTheAffectedRows(t *testing.T) {
	testCases := []struct {
		name   string
		result sql.Result
		err    string
	}{
		{"ShouldSucceedWhenARowWasWritten", driver.RowsAffected(1), ""},
		{"ShouldReturnSupersededWhenNoRowWasWritten", driver.RowsAffected(0), session.ErrSessionSuperseded.Error()},
		{"ShouldReturnErrorWhenTheAffectedRowsAreUnavailable", driver.ResultNoRows, "error determining whether the session was saved: no RowsAffected available after DDL statement"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			db := mocks.NewMockSQLXDB(ctrl)

			db.EXPECT().ExecContext(gomock.Any(), gomock.Any(), "an-issuer", "a-signature", "a-public-id", "john", gomock.Any(), []byte("data")).Return(tc.result, nil)

			p := storage.NewSQLProviderForTesting(db)

			err := p.SessionSave(t.Context(), "an-issuer", "a-signature", "a-public-id", "john", time.Hour, []byte("data"))

			if tc.err == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestSQLProviderLoadHMACKeyShouldAdoptTheExistingKeyOnUniqueViolation(t *testing.T) {
	testCases := []struct {
		name      string
		setup     func(db *mocks.MockSQLXDB, tx *mocks.MockSQLXTx, stored *[]byte)
		expectErr string
	}{
		{
			name: "ShouldReturnErrBeginningTransaction",
			setup: func(db *mocks.MockSQLXDB, tx *mocks.MockSQLXTx, stored *[]byte) {
				db.EXPECT().BeginTxx(gomock.Any(), gomock.Any()).Return(nil, errors.New("boom"))
			},
			expectErr: "error beginning transaction to get hmac key: boom",
		},
		{
			name: "ShouldReturnErrCommittingTransaction",
			setup: func(db *mocks.MockSQLXDB, tx *mocks.MockSQLXTx, stored *[]byte) {
				gomock.InOrder(
					db.EXPECT().BeginTxx(gomock.Any(), gomock.Any()).Return(tx, nil),
					tx.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), "hmac_key_a-name").DoAndReturn(func(_ context.Context, dest any, _ string, _ ...any) error {
						value, ok := dest.(*[]byte)
						require.True(t, ok)

						*value = *stored

						return nil
					}),
					tx.EXPECT().Commit().Return(errors.New("boom")),
				)
			},
			expectErr: "error occurred committing transaction to get hmac key: boom",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			db := mocks.NewMockSQLXDB(ctrl)
			first := mocks.NewMockSQLXTx(ctrl)
			second := mocks.NewMockSQLXTx(ctrl)

			var stored []byte

			gomock.InOrder(
				db.EXPECT().BeginTxx(gomock.Any(), gomock.Any()).Return(first, nil),
				first.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), "hmac_key_a-name").Return(sql.ErrNoRows),
				first.EXPECT().ExecContext(gomock.Any(), gomock.Any(), "hmac_key_a-name", gomock.Any()).DoAndReturn(func(_ context.Context, _ string, args ...any) (sql.Result, error) {
					value, ok := args[1].([]byte)
					require.True(t, ok)

					stored = value

					return nil, sqlite3.Error{Code: sqlite3.ErrConstraint, ExtendedCode: sqlite3.ErrConstraintUnique}
				}),
				first.EXPECT().Rollback().Return(nil),
			)

			tc.setup(db, second, &stored)

			p := storage.NewSQLProviderForTesting(db)

			key, err := p.LoadHMACKey(t.Context(), "a-name", 64)

			assert.EqualError(t, err, tc.expectErr)
			assert.Nil(t, key)
		})
	}
}
