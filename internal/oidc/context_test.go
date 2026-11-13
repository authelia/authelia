// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestNewDetachedContextFromProviders(t *testing.T) {
	ctx := &TestContext{
		Context:       t.Context(),
		MockIssuerURL: MustParseRequestURI("https://elsewhere.example.com"),
		Clock:         clock.New(),
	}

	providers := oidc.NewDetachedContextProviders(ctx)

	assert.Equal(t, ctx.GetClock(), providers.Clock)
	assert.NotNil(t, providers.Random)
	assert.Equal(t, ctx.GetConfiguration(), providers.Config)

	parent, cancelParent := context.WithCancel(t.Context())

	detached, cancel := oidc.NewDetachedContextFromProviders(parent, time.Minute, MustParseRequestURI("https://auth.example.com"), providers)

	defer cancel()

	issuer, err := detached.IssuerURL()

	assert.NoError(t, err)
	assert.Equal(t, "https://auth.example.com", issuer.String())
	assert.Equal(t, "https://auth.example.com", (&oidc.Config{}).GetIssuer(detached))
	assert.Same(t, detached, detached.Value(model.CtxKeyAutheliaCtx))
	assert.Equal(t, providers.Clock, detached.GetClock())
	assert.Equal(t, providers.Random, detached.GetRandom())
	assert.Equal(t, providers.Config, detached.GetConfiguration())
	assert.Nil(t, detached.GetProviderStorage())
	assert.Nil(t, detached.GetUserProvider())
	assert.Equal(t, providers.Resolver, detached.GetProviderUserAttributeResolver())

	deadline, ok := detached.Deadline()

	assert.True(t, ok)
	assert.False(t, deadline.IsZero())
	assert.NoError(t, detached.Err())

	cancelParent()

	assert.ErrorIs(t, detached.Err(), context.Canceled)
}
