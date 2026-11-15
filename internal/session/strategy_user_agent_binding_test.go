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

func TestDefaultStrategyUserAgentBinding(t *testing.T) {
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
				ctx.cookies[testNameUserAgentBinding] = tc.cookie
			}

			binding, err := strategy.UserAgentBinding(ctx)
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

func TestDefaultStrategyUserAgentBindingShouldBeStableForTheUserAgent(t *testing.T) {
	strategy := newTestStrategy(t, nil)

	ctx := newTestContext()

	first, err := strategy.UserAgentBinding(ctx)
	require.NoError(t, err)

	strategy.IssueUserAgentBinding(ctx, first)

	next := newTestContext()
	next.cookies[testNameUserAgentBinding] = ctx.jar[testNameUserAgentBinding].Value

	second, err := strategy.UserAgentBinding(next)
	require.NoError(t, err)

	assert.Equal(t, first, second)
	assert.True(t, strategy.VerifyUserAgentBinding(next, UserAgentBindingDigest(first)))
}

func TestDefaultStrategyIssueUserAgentBinding(t *testing.T) {
	strategy := newTestStrategy(t, nil)

	ctx := newTestContext()

	binding, err := strategy.UserAgentBinding(ctx)
	require.NoError(t, err)

	strategy.IssueUserAgentBinding(ctx, binding)

	cookie, ok := ctx.jar[testNameUserAgentBinding]
	require.True(t, ok)

	assert.Equal(t, testNameUserAgentBinding, cookie.Name)
	assert.Equal(t, binding, cookie.Value)
	assert.Equal(t, "/", cookie.Path)
	assert.Empty(t, cookie.Domain)
	assert.Equal(t, 3600, cookie.MaxAge)
	assert.True(t, cookie.Expires.IsZero())
	assert.True(t, cookie.Secure)
	assert.True(t, cookie.HttpOnly)
	assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
	assert.True(t, strategy.VerifyUserAgentBinding(ctx, UserAgentBindingDigest(binding)))
}

func TestDefaultStrategyUserAgentBindingShouldBeUnique(t *testing.T) {
	strategy := newTestStrategy(t, nil)

	first, err := strategy.UserAgentBinding(newTestContext())
	require.NoError(t, err)

	second, err := strategy.UserAgentBinding(newTestContext())
	require.NoError(t, err)

	assert.NotEqual(t, first, second)
}

func TestDefaultStrategyUserAgentBindingShouldHandleGenerateError(t *testing.T) {
	strategy := newTestStrategyWithCodec(t, &failingCodec{Codec: newTestCodec(t), errSessionID: errors.New("bad random")}, newTestRepository())

	ctx := newTestContext()

	binding, err := strategy.UserAgentBinding(ctx)

	assert.EqualError(t, err, "error occurred generating user agent binding: bad random")
	assert.Empty(t, binding)
	assert.Equal(t, 0, ctx.set)
}

func TestDefaultStrategyVerifyUserAgentBinding(t *testing.T) {
	binding := newTestCookie("binding")

	testCases := []struct {
		name     string
		cookie   string
		binding  string
		expected bool
	}{
		{"ShouldVerifyMatchingBinding", binding, UserAgentBindingDigest(binding), true},
		{"ShouldNotVerifyBindingAgainstItself", binding, binding, false},
		{"ShouldNotVerifyDifferentBinding", binding, UserAgentBindingDigest(newTestCookie("other")), false},
		{"ShouldNotVerifyWithoutCookie", "", UserAgentBindingDigest(binding), false},
		{"ShouldNotVerifyEmptyBinding", binding, "", false},
		{"ShouldNotVerifyMalformedCookie", "not-a-cookie-value", UserAgentBindingDigest("not-a-cookie-value"), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			strategy := newTestStrategy(t, nil)

			ctx := newTestContext()

			if tc.cookie != "" {
				ctx.cookies[testNameUserAgentBinding] = tc.cookie
			}

			assert.Equal(t, tc.expected, strategy.VerifyUserAgentBinding(ctx, tc.binding))
		})
	}
}

func TestUserAgentBindingCookieName(t *testing.T) {
	assert.Equal(t, "__Host-Http-authelia_session_ua", UserAgentBindingCookieName("authelia_session"))
}

func TestUserAgentBindingDigest(t *testing.T) {
	digest := UserAgentBindingDigest("binding")

	assert.Equal(t, "80f70afeef3caa57646fd20afb95be9c3f2c03d38e091906de31838813dcc22e", digest)
	assert.NotEqual(t, UserAgentBindingDigest("other"), digest)
}
