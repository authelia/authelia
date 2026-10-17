// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package middlewares_test

import (
	"crypto/sha256"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/cache"
	"github.com/authelia/authelia/v4/internal/expression"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/storage"
)

func TestProvidersFinalize(t *testing.T) {
	names := []string{session.HMACKeyNameSession, session.HMACKeyNameSessionAuthz, session.HMACKeyNameSessionCSRF}

	hmacKeys := func(failing string) func(mock *mocks.MockAutheliaCtx) {
		return func(mock *mocks.MockAutheliaCtx) {
			for i, name := range names {
				if name == failing {
					mock.StorageMock.EXPECT().LoadHMACKey(mock.Ctx, name, sha256.BlockSize).Return(nil, errors.New("failed to load the key"))

					return
				}

				key := make([]byte, sha256.BlockSize)
				key[0] = byte(i + 1)

				mock.StorageMock.EXPECT().LoadHMACKey(mock.Ctx, name, sha256.BlockSize).Return(key, nil)
			}
		}
	}

	hmacKey := hmacKeys("")

	testCases := []struct {
		name     string
		storage  string
		setup    func(mock *mocks.MockAutheliaCtx)
		expected any
		err      string
	}{
		{"ShouldUseTheStorageByDefault", "", hmacKey, storage.SessionRepository{}, ""},
		{"ShouldUseTheCache", "cache", hmacKey, cache.SessionRepository{}, ""},
		{"ShouldUseTheStorage", "internal", hmacKey, storage.SessionRepository{}, ""},
		{"ShouldErrorOnUnknownStorage", "unknown", nil, nil, "unknown session storage 'unknown'"},
		{"ShouldErrorWhenTheSessionHMACKeyCannotBeLoaded", "internal", hmacKeys(session.HMACKeyNameSession), nil, "failed to load the key"},
		{"ShouldErrorWhenTheSessionAuthzHMACKeyCannotBeLoaded", "internal", hmacKeys(session.HMACKeyNameSessionAuthz), nil, "failed to load the key"},
		{"ShouldErrorWhenTheSessionCSRFHMACKeyCannotBeLoaded", "internal", hmacKeys(session.HMACKeyNameSessionCSRF), nil, "failed to load the key"},
		{
			"ShouldErrorWhenTheHMACKeysAreNotDistinct",
			"internal",
			func(mock *mocks.MockAutheliaCtx) {
				for _, name := range names {
					mock.StorageMock.EXPECT().LoadHMACKey(mock.Ctx, name, sha256.BlockSize).Return(make([]byte, sha256.BlockSize), nil)
				}
			},
			nil,
			"error occurred creating the session codec: every hmac key must be distinct",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			mock.Ctx.Configuration.Session.Storage = tc.storage

			if tc.setup != nil {
				tc.setup(mock)
			}

			providers := middlewares.Providers{
				StorageProvider: mock.StorageMock,
				Cache:           cache.NewMemory(),
				Clock:           mock.Ctx.Providers.Clock,
				Random:          mock.Ctx.Providers.Random,
			}

			err := providers.Finalize(mock.Ctx)

			if tc.err != "" {
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, providers.Session)
				assert.Nil(t, providers.SessionRepository)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, providers.Session)

			assert.IsType(t, tc.expected, providers.SessionRepository)

			strategy, err := providers.Session.GetStrategy("example.com")

			require.NoError(t, err)
			assert.NotNil(t, strategy)
		})
	}
}

func TestProvidersStartupChecksShouldCheckTheCache(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	defer mock.Close()

	mock.Ctx.Configuration.Server.DisableHealthcheck = true
	mock.Ctx.Configuration.Notifier.DisableStartupCheck = true
	mock.Ctx.Configuration.NTP.DisableStartupCheck = true

	mock.Ctx.Providers.UserAttributeResolver = expression.NewUserAttributes(&mock.Ctx.Configuration)

	mock.StorageMock.EXPECT().StartupCheck().Return(nil).Times(2)
	mock.UserProviderMock.EXPECT().StartupCheck().Return(nil).Times(2)

	mock.Ctx.Providers.Cache = cache.NewMemory()

	assert.NoError(t, mock.Ctx.Providers.StartupChecks(mock.Ctx, true))

	mock.Ctx.Providers.Cache = nil

	assert.ErrorContains(t, mock.Ctx.Providers.StartupChecks(mock.Ctx, true), middlewares.ProviderNameCache)
}
