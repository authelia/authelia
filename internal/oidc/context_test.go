// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestNewDetachedContext(t *testing.T) {
	ctx := &TestContext{
		Context:       t.Context(),
		MockIssuerURL: MustParseRequestURI("https://auth.example.com"),
		Clock:         clock.New(),
	}

	detached, cancel, err := oidc.NewDetachedContext(ctx, time.Minute)
	require.NoError(t, err)

	defer cancel()

	issuer, err := detached.IssuerURL()
	require.NoError(t, err)

	assert.Equal(t, "https://auth.example.com", issuer.String())

	ctx.MockIssuerURL = MustParseRequestURI("https://elsewhere.example.com")

	issuer, err = detached.IssuerURL()
	require.NoError(t, err)
	assert.Equal(t, "https://auth.example.com", issuer.String())
	assert.Same(t, detached, detached.Value(model.CtxKeyAutheliaCtx))

	config := &oidc.Config{}

	assert.Equal(t, "https://auth.example.com", config.GetIDTokenIssuer(detached))

	deadline, ok := detached.Deadline()

	assert.True(t, ok)
	assert.False(t, deadline.IsZero())
	assert.NoError(t, detached.Err())

	cancel()

	assert.ErrorIs(t, detached.Err(), context.Canceled)

	assert.Equal(t, ctx.GetClock(), detached.GetClock())
}

func TestNewDetachedContextShouldErrorWithoutAnIssuer(t *testing.T) {
	ctx := &TestContext{Context: t.Context(), IssuerURLFunc: func() (*url.URL, error) { return nil, errors.New("no issuer") }}

	detached, cancel, err := oidc.NewDetachedContext(ctx, time.Minute)

	assert.EqualError(t, err, "no issuer")
	assert.Nil(t, detached)
	assert.Nil(t, cancel)
}
