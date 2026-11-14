// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package middlewares_test

import (
	"errors"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/mocks"
)

func TestAppAuthenticatorGarbageCollectorShouldPrune(t *testing.T) {
	ctrl := gomock.NewController(t)

	store := mocks.NewMockStorage(ctrl)
	now := time.Date(2026, 11, 12, 1, 2, 3, 0, time.UTC)

	collector := middlewares.NewAppAuthenticatorGarbageCollector(store, clock.NewFixed(now))

	assert.Equal(t, 5*time.Minute, collector.GarbageCollectionFrequency(t.Context()))

	store.EXPECT().PruneAppAuthenticator(gomock.Any(), now).Return(nil)
	assert.NoError(t, collector.GarbageCollection(t.Context()))

	store.EXPECT().PruneAppAuthenticator(gomock.Any(), now).Return(errors.New("bad"))
	assert.EqualError(t, collector.GarbageCollection(t.Context()), "bad")
}

func TestNewProvidersShouldConfigureAppAuthenticator(t *testing.T) {
	newConfig := func(t *testing.T, disable bool) *schema.Configuration {
		return &schema.Configuration{
			Storage: schema.Storage{
				EncryptionKey: "authelia-test-key-not-a-secret-authelia-test-key-not-a-secret",
				Local:         &schema.StorageLocal{Path: filepath.Join(t.TempDir(), "db.sqlite3")},
			},
			AppAuthenticator: schema.AppAuthenticator{Disable: disable},
		}
	}

	t.Run("ShouldConfigureWhenEnabled", func(t *testing.T) {
		providers, _, errs := middlewares.NewProviders(newConfig(t, false), nil)
		require.Empty(t, errs)

		assert.NotNil(t, providers.AppAuthenticator)
		assert.Equal(t, 1, providers.GarbageCollector.Len())
	})

	t.Run("ShouldNotConfigureWhenDisabled", func(t *testing.T) {
		providers, _, errs := middlewares.NewProviders(newConfig(t, true), nil)
		require.Empty(t, errs)

		assert.Nil(t, providers.AppAuthenticator)
		assert.Equal(t, 0, providers.GarbageCollector.Len())
	})
}

func TestNewRequestLoggerShouldRedactAppAuthenticatorDevicePaths(t *testing.T) {
	testCases := []struct {
		name, base, path, expected, expectedRaw string
	}{
		{"ShouldRedactBegin", "", "/api/register/app/secret", "/api/register/app/{redacted}", ""},
		{"ShouldRedactConfirm", "", "/api/register/app/confirm/secret", "/api/register/app/confirm/{redacted}", ""},
		{"ShouldRedactCheck", "", "/api/secondfactor/app/challenge/secret", "/api/secondfactor/app/challenge/{redacted}", ""},
		{"ShouldRedactConfirmation", "", "/api/secondfactor/app/confirmation/secret/extra", "/api/secondfactor/app/confirmation/{redacted}", ""},
		{"ShouldKeepDevice", "", "/api/secondfactor/app/device", "/api/secondfactor/app/device", ""},
		{"ShouldRedactRawWithBasePath", "/authelia", "/api/secondfactor/app/challenge/secret", "/api/secondfactor/app/challenge/{redacted}", "/authelia/api/secondfactor/app/challenge/{redacted}"},
		{"ShouldNotRedactOtherPaths", "/authelia", "/api/secondfactor/totp", "/api/secondfactor/totp", "/authelia/api/secondfactor/totp?x=secret"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				ctx fasthttp.RequestCtx
				req fasthttp.Request
			)

			req.SetRequestURI(tc.path)
			ctx.Init(&req, &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}, nil)

			if tc.base != "" {
				ctx.SetUserValue(middlewares.UserValueKeyBaseURL, tc.base)
				ctx.SetUserValue(middlewares.UserValueKeyRawURI, tc.base+tc.path+"?x=secret")
			}

			entry := middlewares.NewRequestLogger(&ctx)

			assert.Equal(t, tc.expected, entry.Data["path"])

			if tc.expectedRaw != "" {
				assert.Equal(t, tc.expectedRaw, entry.Data["path_raw"])
			} else {
				assert.NotContains(t, entry.Data, "path_raw")
			}
		})
	}
}

func TestRedactAppAuthenticatorDevicePath(t *testing.T) {
	testCases := []struct {
		name     string
		path     string
		expected string
	}{
		{"ShouldRedactRegistration", "/api/register/app/abc", "/api/register/app/{redacted}"},
		{"ShouldRedactRegistrationConfirm", "/api/register/app/confirm/abc", "/api/register/app/confirm/{redacted}"},
		{"ShouldRedactChallenge", "/api/secondfactor/app/challenge/abc", "/api/secondfactor/app/challenge/{redacted}"},
		{"ShouldRedactConfirmation", "/api/secondfactor/app/confirmation/abc", "/api/secondfactor/app/confirmation/{redacted}"},
		{"ShouldKeepDevice", "/api/secondfactor/app/device", "/api/secondfactor/app/device"},
		{"ShouldRedactUnderBasePath", "/auth/api/secondfactor/app/challenge/abc", "/auth/api/secondfactor/app/challenge/{redacted}"},
		{"ShouldRedactUnderNestedBasePath", "/a/api/b/api/register/app/abc", "/a/api/b/api/register/app/{redacted}"},
		{"ShouldNotRedactOtherPaths", "/api/secondfactor/app/status", "/api/secondfactor/app/status"},
		{"ShouldNotRedactRoot", "/", "/"},
		{"ShouldNotRedactEmpty", "", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, middlewares.RedactAppAuthenticatorDevicePath(tc.path))
		})
	}
}
