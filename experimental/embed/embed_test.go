// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package embed

import (
	"context"
	"crypto/sha256"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/cache"
	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/expression"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/random"
)

func TestShouldPanicNilCtx(t *testing.T) {
	assert.Panics(t, func() {
		_ = ProvidersStartupCheck(nil, false)
	})

	ctx := &ctxEmbed{}

	assert.Panics(t, func() {
		_ = ProvidersStartupCheck(ctx, false)
	})
}

func TestProvidersStartupCheckShouldRetainFinalizedProviders(t *testing.T) {
	ctx, storage := newStartupCheckContext(t)

	storage.EXPECT().LoadHMACKey(gomock.Any(), "session", sha256.BlockSize).Return(make([]byte, sha256.BlockSize), nil)

	require.Nil(t, ctx.GetProviders().Session)

	require.NoError(t, ProvidersStartupCheck(ctx, false))

	providers := ctx.GetProviders()

	assert.NotNil(t, providers.Session)
	assert.NotNil(t, providers.SessionRepository)
}

func TestProvidersStartupCheckShouldErrorWhenContextCannotRetainProviders(t *testing.T) {
	ctx, _ := newStartupCheckContext(t)

	assert.EqualError(t, ProvidersStartupCheck(struct{ Context }{ctx}, false), "the context does not support retaining the finalized providers as it does not implement SetProviders")
}

func newStartupCheckContext(t *testing.T) (ctx *ctxEmbed, storage *mocks.MockStorage) {
	t.Helper()

	ctrl := gomock.NewController(t)

	t.Cleanup(ctrl.Finish)

	storage = mocks.NewMockStorage(ctrl)
	storage.EXPECT().StartupCheck().Return(nil)

	users := mocks.NewMockUserProvider(ctrl)
	users.EXPECT().StartupCheck().Return(nil)

	config := &Configuration{}

	config.Server.DisableHealthcheck = true
	config.Notifier.DisableStartupCheck = true
	config.NTP.DisableStartupCheck = true

	return &ctxEmbed{
		Configuration: config,
		Logger:        logrus.NewEntry(logrus.StandardLogger()),
		Providers: Providers{
			StorageProvider:       storage,
			Cache:                 cache.NewMemory(),
			UserProvider:          users,
			UserAttributeResolver: expression.NewUserAttributes(config.ToInternal()),
			Clock:                 clock.New(),
			Random:                random.New(),
		},
		Context: context.Background(),
	}, storage
}
