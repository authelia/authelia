// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/mocks"
)

func TestStorageCleanBefore(t *testing.T) {
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)

	testCases := []struct {
		name     string
		have     string
		expected time.Time
		err      string
	}{
		{
			"ShouldUseNowWhenTheFlagIsAbsent",
			"",
			now,
			"",
		},
		{
			"ShouldSubtractADurationInTheCommonSyntax",
			"30 days",
			now.Add(-time.Hour * 24 * 30),
			"",
		},
		{
			"ShouldSubtractADurationInTheGoSyntax",
			"1h",
			now.Add(-time.Hour),
			"",
		},
		{
			"ShouldRaiseErrorOnAnUnparsableDuration",
			"not a duration",
			time.Time{},
			"error parsing the value of the --before flag: could not parse 'not a duration' as a duration",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String(cmdFlagNameBefore, "", "")

			if tc.have != "" {
				require.NoError(t, cmd.Flags().Set(cmdFlagNameBefore, tc.have))
			}

			before, err := storageCleanBefore(cmd, now)

			if tc.err == "" {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, before)
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestStorageCleanOAuth2RunE(t *testing.T) {
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)

	testCases := []struct {
		name     string
		flags    map[string]string
		setup    func(store *mocks.MockStorage)
		expected string
		err      string
	}{
		{
			"ShouldRemoveStaleSessions",
			nil,
			func(store *mocks.MockStorage) {
				store.EXPECT().DeleteStaleOAuth2ConsentSessions(gomock.Any(), now).Return(3, nil)
			},
			"Removed 3 stale OpenID Connect 1.0 consent session(s) responded to and expired before 2026-09-10T12:00:00Z, along with the sessions which reference them.\n",
			"",
		},
		{
			"ShouldRemoveStaleSessionsBeforeTheDuration",
			map[string]string{cmdFlagNameBefore: "30 days"},
			func(store *mocks.MockStorage) {
				store.EXPECT().DeleteStaleOAuth2ConsentSessions(gomock.Any(), now.Add(-time.Hour*24*30)).Return(1, nil)
			},
			"Removed 1 stale OpenID Connect 1.0 consent session(s) responded to and expired before 2026-08-11T12:00:00Z, along with the sessions which reference them.\n",
			"",
		},
		{
			"ShouldCountStaleSessionsOnDryRun",
			map[string]string{cmdFlagNameDryRun: "true"},
			func(store *mocks.MockStorage) {
				store.EXPECT().CountStaleOAuth2ConsentSessions(gomock.Any(), now).Return(5, nil)
			},
			"Dry run: 5 stale OpenID Connect 1.0 consent session(s) responded to and expired before 2026-09-10T12:00:00Z would be removed, along with the sessions which reference them.\n",
			"",
		},
		{
			"ShouldRaiseErrorOnRemoveError",
			nil,
			func(store *mocks.MockStorage) {
				store.EXPECT().DeleteStaleOAuth2ConsentSessions(gomock.Any(), now).Return(0, errors.New("delete failed"))
			},
			"",
			"delete failed",
		},
		{
			"ShouldRaiseErrorOnCountError",
			map[string]string{cmdFlagNameDryRun: "true"},
			func(store *mocks.MockStorage) {
				store.EXPECT().CountStaleOAuth2ConsentSessions(gomock.Any(), now).Return(0, errors.New("count failed"))
			},
			"",
			"count failed",
		},
		{
			"ShouldRaiseErrorOnAnUnparsableDuration",
			map[string]string{cmdFlagNameBefore: "not a duration"},
			nil,
			"",
			"error parsing the value of the --before flag: could not parse 'not a duration' as a duration",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			store := mocks.NewMockStorage(ctrl)

			if tc.setup != nil {
				tc.setup(store)
				store.EXPECT().Close().Return(nil)
			}

			ctx := NewCmdCtx()
			ctx.providers.StorageProvider = store
			ctx.providers.Clock = clock.NewFixed(now)

			cmd := newStorageCleanOAuth2Cmd(ctx)

			buf := new(bytes.Buffer)
			cmd.SetOut(buf)

			for name, value := range tc.flags {
				require.NoError(t, cmd.Flags().Set(name, value))
			}

			err := cmd.RunE(cmd, nil)

			if tc.err == "" {
				require.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.err)
			}

			assert.Equal(t, tc.expected, buf.String())
		})
	}
}
