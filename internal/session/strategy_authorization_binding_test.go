// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"encoding/base64"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultStrategyAuthorizationBinding(t *testing.T) {
	existing := newTestCookie("binding")

	testCases := []struct {
		name   string
		cookie string
		reused bool
	}{
		{"ShouldGenerateWhenAbsent", "", false},
		{"ShouldReuseWhenPresent", existing, true},
		{"ShouldGenerateWhenMalformed", "not-a-cookie-value", false},
		{"ShouldGenerateWhenTheWrongLength", base64.RawURLEncoding.EncodeToString([]byte("short")), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			strategy := newTestStrategy(t, nil)

			ctx := newTestContext()

			if tc.cookie != "" {
				ctx.cookies[testNameAuthorizationBinding] = tc.cookie
			}

			binding, err := strategy.AuthorizationBinding(ctx)
			require.NoError(t, err)

			assert.Equal(t, 0, ctx.set)

			id, err := base64.RawURLEncoding.DecodeString(binding)
			require.NoError(t, err)
			assert.Len(t, id, sessionIDLength)

			if tc.reused {
				assert.Equal(t, tc.cookie, binding)
			} else {
				assert.NotEqual(t, tc.cookie, binding)
			}
		})
	}
}

func TestDefaultStrategyAuthorizationBindingShouldBeStableForTheUserAgent(t *testing.T) {
	strategy := newTestStrategy(t, nil)

	ctx := newTestContext()

	first, err := strategy.AuthorizationBinding(ctx)
	require.NoError(t, err)

	strategy.IssueAuthorizationBinding(ctx, first)

	next := newTestContext()
	next.cookies[testNameAuthorizationBinding] = ctx.jar[testNameAuthorizationBinding].Value

	second, err := strategy.AuthorizationBinding(next)
	require.NoError(t, err)

	assert.Equal(t, first, second)
	assert.True(t, strategy.VerifyAuthorizationBinding(next, first))
}

func TestDefaultStrategyIssueAuthorizationBinding(t *testing.T) {
	strategy := newTestStrategy(t, nil)

	ctx := newTestContext()

	binding, err := strategy.AuthorizationBinding(ctx)
	require.NoError(t, err)

	strategy.IssueAuthorizationBinding(ctx, binding)

	cookie, ok := ctx.jar[testNameAuthorizationBinding]
	require.True(t, ok)

	assert.Equal(t, testNameAuthorizationBinding, cookie.Name)
	assert.Equal(t, binding, cookie.Value)
	assert.Equal(t, "/", cookie.Path)
	assert.Empty(t, cookie.Domain)
	assert.Equal(t, 3600, cookie.MaxAge)
	assert.True(t, cookie.Expires.IsZero())
	assert.True(t, cookie.Secure)
	assert.True(t, cookie.HttpOnly)
	assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
	assert.True(t, strategy.VerifyAuthorizationBinding(ctx, binding))
}

func TestDefaultStrategyAuthorizationBindingShouldBeUnique(t *testing.T) {
	strategy := newTestStrategy(t, nil)

	first, err := strategy.AuthorizationBinding(newTestContext())
	require.NoError(t, err)

	second, err := strategy.AuthorizationBinding(newTestContext())
	require.NoError(t, err)

	assert.NotEqual(t, first, second)
}

func TestDefaultStrategyAuthorizationBindingShouldHandleGenerateError(t *testing.T) {
	strategy := newTestStrategyWithCodec(t, &failingCodec{Codec: newTestCodec(t), errSessionID: errors.New("bad random")}, newTestRepository())

	ctx := newTestContext()

	binding, err := strategy.AuthorizationBinding(ctx)

	assert.EqualError(t, err, "error occurred generating authorization binding: bad random")
	assert.Empty(t, binding)
	assert.Equal(t, 0, ctx.set)
}

func TestDefaultStrategyVerifyAuthorizationBinding(t *testing.T) {
	binding := newTestCookie("binding")

	testCases := []struct {
		name     string
		cookie   string
		binding  string
		expected bool
	}{
		{"ShouldVerifyMatchingBinding", binding, binding, true},
		{"ShouldNotVerifyDifferentBinding", binding, newTestCookie("other"), false},
		{"ShouldNotVerifyWithoutCookie", "", binding, false},
		{"ShouldNotVerifyEmptyBinding", binding, "", false},
		{"ShouldNotVerifyMalformedCookie", "not-a-cookie-value", "not-a-cookie-value", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			strategy := newTestStrategy(t, nil)

			ctx := newTestContext()

			if tc.cookie != "" {
				ctx.cookies[testNameAuthorizationBinding] = tc.cookie
			}

			assert.Equal(t, tc.expected, strategy.VerifyAuthorizationBinding(ctx, tc.binding))
		})
	}
}

func TestAuthorizationBindingCookieName(t *testing.T) {
	assert.Equal(t, "__Host-Http-authelia_session_ua", AuthorizationBindingCookieName("authelia_session"))
}
