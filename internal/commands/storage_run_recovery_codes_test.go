// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"context"
	"errors"
	"net"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/storage"
)

func TestCmdRecoveryCodeContextRemoteIP(t *testing.T) {
	assert.Equal(t, net.IP{}, cmdRecoveryCodeContext{}.RemoteIP())
}

func TestStorageUserRecoveryCodesStatusRunE(t *testing.T) {
	generatedAt := time.Date(2026, 1, 2, 3, 4, 0, 0, time.UTC)

	testCases := []struct {
		name        string
		seed        bool
		expected    []string
		notExpected []string
	}{
		{
			"ShouldShowNoCodes",
			false,
			[]string{
				"User: john\n",
				"Recovery codes total: 0 (unused: 0, consumed: 0, revoked: 0)\n",
				"Last generated: never\n",
				"Last consumed: never\n",
			},
			nil,
		},
		{
			"ShouldShowMixedCodes",
			true,
			[]string{
				"User: john\n",
				"Recovery codes total: 4 (unused: 2, consumed: 1, revoked: 1)\n",
				"Last generated: 2026-01-02 03:04 UTC\n",
				"Last consumed: ",
			},
			[]string{
				"Last generated: never",
				"Last consumed: never",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmdCtx := newTestCmdCtx(t)

			if tc.seed {
				seedRecoveryCodesMixed(t, cmdCtx.providers.StorageProvider, "john", generatedAt)
			}

			seedRecoveryCodes(t, cmdCtx.providers.StorageProvider, "harry", generatedAt.Add(time.Hour), "HHHHH-11111")

			cmd, buf := newTestCmdWithBuf()

			require.NoError(t, cmdCtx.StorageUserRecoveryCodesStatusRunE(cmd, []string{"john"}))

			for _, expected := range tc.expected {
				assert.Contains(t, buf.String(), expected)
			}

			for _, notExpected := range tc.notExpected {
				assert.NotContains(t, buf.String(), notExpected)
			}
		})
	}
}

func TestStorageUserRecoveryCodesListRunE(t *testing.T) {
	generatedAt := time.Date(2026, 1, 2, 3, 4, 0, 0, time.UTC)

	t.Run("ShouldShowNoCodes", func(t *testing.T) {
		cmdCtx := newTestCmdCtx(t)

		cmd, buf := newTestCmdWithBuf()

		require.NoError(t, cmdCtx.StorageUserRecoveryCodesListRunE(cmd, []string{"john"}))

		assert.Equal(t, "User 'john' has no recovery codes.\n", buf.String())
	})

	t.Run("ShouldListMixedCodes", func(t *testing.T) {
		cmdCtx := newTestCmdCtx(t)

		seedRecoveryCodesMixed(t, cmdCtx.providers.StorageProvider, "john", generatedAt)

		cmd, buf := newTestCmdWithBuf()

		require.NoError(t, cmdCtx.StorageUserRecoveryCodesListRunE(cmd, []string{"john"}))

		lines := regexp.MustCompile(`\r?\n`).Split(buf.String(), -1)
		require.Len(t, lines, 6)

		assert.Equal(t, "ID\tCreated\tStatus", lines[0])
		assert.Regexp(t, `^\d+\t2026-01-01 03:04 UTC\tconsumed at \d{4}-\d{2}-\d{2} \d{2}:\d{2} \S+$`, lines[1])
		assert.Regexp(t, `^\d+\t2026-01-01 03:04 UTC\trevoked at \d{4}-\d{2}-\d{2} \d{2}:\d{2} \S+$`, lines[2])
		assert.Regexp(t, `^\d+\t2026-01-02 03:04 UTC\tunused$`, lines[3])
		assert.Regexp(t, `^\d+\t2026-01-02 03:04 UTC\tunused$`, lines[4])
		assert.Equal(t, "", lines[5])
	})
}

func TestStorageUserRecoveryCodesDeleteRunE(t *testing.T) {
	cmdCtx := newTestCmdCtx(t)

	seedRecoveryCodes(t, cmdCtx.providers.StorageProvider, "john", time.Now(), "AAAAA-11111", "BBBBB-22222")
	seedRecoveryCodes(t, cmdCtx.providers.StorageProvider, "harry", time.Now(), "CCCCC-33333")

	cmd, buf := newTestCmdWithBuf()

	require.NoError(t, cmdCtx.StorageUserRecoveryCodesDeleteRunE(cmd, []string{"john"}))

	assert.Equal(t, "Successfully revoked all recovery codes for user 'john'.\n", buf.String())

	store, err := storage.NewProvider(cmdCtx.config, nil)
	require.NoError(t, err)

	defer store.Close()

	ctx := context.Background()

	count, err := store.CountUnusedRecoveryCodesByUsername(ctx, "john")
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	codes, err := store.LoadRecoveryCodesByUsername(ctx, "john")
	require.NoError(t, err)
	require.Len(t, codes, 2)

	for _, code := range codes {
		assert.True(t, code.RevokedAt.Valid)
	}

	count, err = store.CountUnusedRecoveryCodesByUsername(ctx, "harry")
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestStorageSchemaEncryptionRotateRunERecoveryCodes(t *testing.T) {
	cmdCtx := newTestCmdCtx(t)

	seedRecoveryCodes(t, cmdCtx.providers.StorageProvider, "john", time.Now(), "AAAAA-11111")

	cmd, buf := newTestCmdWithBuf()
	cmd.Use = "rc"
	cmd.Flags().Bool(cmdFlagNameForce, true, "")

	require.NoError(t, cmdCtx.StorageSchemaEncryptionRotateRunE(cmd, nil))

	assert.Equal(t, "Completed the 'rc' key rotation successfully and cleanly truncated the 'recovery_codes' table.\n", buf.String())

	store, err := storage.NewProvider(cmdCtx.config, nil)
	require.NoError(t, err)

	defer store.Close()

	codes, err := store.LoadRecoveryCodesByUsername(context.Background(), "john")
	require.NoError(t, err)
	assert.Empty(t, codes)
}

func TestStorageUserRecoveryCodesRunEErrors(t *testing.T) {
	type runE func(ctx *CmdCtx) func(cmd *cobra.Command, args []string) error

	status := func(ctx *CmdCtx) func(cmd *cobra.Command, args []string) error {
		return ctx.StorageUserRecoveryCodesStatusRunE
	}

	list := func(ctx *CmdCtx) func(cmd *cobra.Command, args []string) error {
		return ctx.StorageUserRecoveryCodesListRunE
	}

	generate := func(ctx *CmdCtx) func(cmd *cobra.Command, args []string) error {
		return ctx.StorageUserRecoveryCodesGenerateRunE
	}

	del := func(ctx *CmdCtx) func(cmd *cobra.Command, args []string) error {
		return ctx.StorageUserRecoveryCodesDeleteRunE
	}

	testCases := []struct {
		name  string
		run   runE
		setup func(store *mocks.MockStorage, rnd *mocks.MockRandom)
		err   string
	}{
		{
			"ShouldErrStatusCheckSchema",
			status,
			func(store *mocks.MockStorage, rnd *mocks.MockRandom) {
				store.EXPECT().SchemaVersion(gomock.Any()).Return(0, errors.New("version failed"))
			},
			"version failed",
		},
		{
			"ShouldErrStatusOutdatedSchema",
			status,
			func(store *mocks.MockStorage, rnd *mocks.MockRandom) {
				store.EXPECT().SchemaVersion(gomock.Any()).Return(1, nil)
				store.EXPECT().SchemaLatestVersion().Return(2, nil)
			},
			"command requires the use of a up to date schema version: storage schema outdated: version 1 is outdated please migrate to version 2 in order to use this command or use an older binary",
		},
		{
			"ShouldErrStatusLoad",
			status,
			func(store *mocks.MockStorage, rnd *mocks.MockRandom) {
				expectCheckSchemaSuccess(store)
				store.EXPECT().LoadRecoveryCodesByUsername(gomock.Any(), "john").Return(nil, errors.New("load failed"))
			},
			"failed to load recovery codes for user 'john': load failed",
		},
		{
			"ShouldErrListCheckSchema",
			list,
			func(store *mocks.MockStorage, rnd *mocks.MockRandom) {
				store.EXPECT().SchemaVersion(gomock.Any()).Return(0, errors.New("version failed"))
			},
			"version failed",
		},
		{
			"ShouldErrListLoad",
			list,
			func(store *mocks.MockStorage, rnd *mocks.MockRandom) {
				expectCheckSchemaSuccess(store)
				store.EXPECT().LoadRecoveryCodesByUsername(gomock.Any(), "john").Return(nil, errors.New("load failed"))
			},
			"failed to load recovery codes for user 'john': load failed",
		},
		{
			"ShouldErrGenerateCheckSchema",
			generate,
			func(store *mocks.MockStorage, rnd *mocks.MockRandom) {
				store.EXPECT().SchemaVersion(gomock.Any()).Return(0, errors.New("version failed"))
			},
			"version failed",
		},
		{
			"ShouldErrGenerateLoadHMACKey",
			generate,
			func(store *mocks.MockStorage, rnd *mocks.MockRandom) {
				expectCheckSchemaSuccess(store)
				store.EXPECT().LoadRecoveryCodeHMACKey(gomock.Any()).Return(errors.New("hmac failed"))
			},
			"hmac failed",
		},
		{
			"ShouldErrGenerateRandom",
			generate,
			func(store *mocks.MockStorage, rnd *mocks.MockRandom) {
				expectCheckSchemaSuccess(store)
				store.EXPECT().LoadRecoveryCodeHMACKey(gomock.Any()).Return(nil)
				rnd.EXPECT().BytesCustomErr(gomock.Any(), gomock.Any()).Return(nil, errors.New("entropy failed"))
			},
			"failed to generate recovery code 0 for user 'john': failed to generate recovery code random bytes: entropy failed",
		},
		{
			"ShouldErrGenerateReplace",
			generate,
			func(store *mocks.MockStorage, rnd *mocks.MockRandom) {
				expectCheckSchemaSuccess(store)
				store.EXPECT().LoadRecoveryCodeHMACKey(gomock.Any()).Return(nil)
				rnd.EXPECT().BytesCustomErr(gomock.Any(), gomock.Any()).Return([]byte(strings.Repeat("A", model.RecoveryCodeLength)), nil).Times(model.RecoveryCodeBatchSize)
				store.EXPECT().ReplaceRecoveryCodesByUsername(gomock.Any(), "john", gomock.Len(model.RecoveryCodeBatchSize), model.NullIP{}).Return(errors.New("replace failed"))
			},
			"failed to replace recovery codes for user 'john': replace failed",
		},
		{
			"ShouldErrDeleteCheckSchema",
			del,
			func(store *mocks.MockStorage, rnd *mocks.MockRandom) {
				store.EXPECT().SchemaVersion(gomock.Any()).Return(0, errors.New("version failed"))
			},
			"version failed",
		},
		{
			"ShouldErrDeleteRevoke",
			del,
			func(store *mocks.MockStorage, rnd *mocks.MockRandom) {
				expectCheckSchemaSuccess(store)
				store.EXPECT().RevokeRecoveryCodesByUsername(gomock.Any(), "john", model.NullIP{}).Return(errors.New("revoke failed"))
			},
			"failed to revoke recovery codes for user 'john': revoke failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			store := mocks.NewMockStorage(ctrl)
			rnd := mocks.NewMockRandom(ctrl)

			tc.setup(store, rnd)
			store.EXPECT().Close().Return(nil)

			cmdCtx := NewCmdCtx()
			cmdCtx.providers.StorageProvider = store
			cmdCtx.providers.Random = rnd

			cmd, buf := newTestCmdWithBuf()

			assert.EqualError(t, tc.run(cmdCtx)(cmd, []string{"john"}), tc.err)
			assert.Empty(t, buf.String())
		})
	}
}

func TestStorageUserRecoveryCodesRunEShouldPanicOnCloseError(t *testing.T) {
	testCases := []struct {
		name string
		run  func(ctx *CmdCtx, cmd *cobra.Command, args []string) error
	}{
		{"Status", func(ctx *CmdCtx, cmd *cobra.Command, args []string) error {
			return ctx.StorageUserRecoveryCodesStatusRunE(cmd, args)
		}},
		{"List", func(ctx *CmdCtx, cmd *cobra.Command, args []string) error {
			return ctx.StorageUserRecoveryCodesListRunE(cmd, args)
		}},
		{"Generate", func(ctx *CmdCtx, cmd *cobra.Command, args []string) error {
			return ctx.StorageUserRecoveryCodesGenerateRunE(cmd, args)
		}},
		{"Delete", func(ctx *CmdCtx, cmd *cobra.Command, args []string) error {
			return ctx.StorageUserRecoveryCodesDeleteRunE(cmd, args)
		}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			store := mocks.NewMockStorage(ctrl)

			store.EXPECT().SchemaVersion(gomock.Any()).Return(0, errors.New("version failed"))
			store.EXPECT().Close().Return(errors.New("close failed"))

			cmdCtx := NewCmdCtx()
			cmdCtx.providers.StorageProvider = store

			cmd, _ := newTestCmdWithBuf()

			assert.PanicsWithError(t, "close failed", func() {
				_ = tc.run(cmdCtx, cmd, []string{"john"})
			})
		})
	}
}

func expectCheckSchemaSuccess(store *mocks.MockStorage) {
	store.EXPECT().SchemaVersion(gomock.Any()).Return(1, nil)
	store.EXPECT().SchemaLatestVersion().Return(1, nil)
	store.EXPECT().SchemaEncryptionCheckKey(gomock.Any(), false).Return(storage.EncryptionValidationResult{}, nil)
}

func seedRecoveryCodes(t *testing.T, store storage.Provider, username string, createdAt time.Time, plaintexts ...string) {
	t.Helper()

	batch := make([]*model.RecoveryCode, 0, len(plaintexts))

	for _, plaintext := range plaintexts {
		batch = append(batch, &model.RecoveryCode{Username: username, Plaintext: plaintext, CreatedAt: createdAt})
	}

	require.NoError(t, store.ReplaceRecoveryCodesByUsername(context.Background(), username, batch, model.NullIP{}))
}

func seedRecoveryCodesMixed(t *testing.T, store storage.Provider, username string, generatedAt time.Time) {
	t.Helper()

	ctx := context.Background()

	seedRecoveryCodes(t, store, username, generatedAt.Add(-24*time.Hour), "AAAAA-11111", "BBBBB-22222")

	codes, err := store.LoadRecoveryCodesByUsername(ctx, username)
	require.NoError(t, err)
	require.Len(t, codes, 2)

	require.NoError(t, store.ConsumeRecoveryCode(ctx, codes[0].ID, model.NullIP{}))

	seedRecoveryCodes(t, store, username, generatedAt, "CCCCC-33333", "DDDDD-44444")
}
