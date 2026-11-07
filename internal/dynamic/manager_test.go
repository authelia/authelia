// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package dynamic

import (
	"net/url"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/configuration/validator"
)

const aclOneFactor = `
access_control:
  default_policy: deny
  rules:
    - domain: app.example.com
      policy: one_factor
`

const aclDeny = `
access_control:
  default_policy: deny
  rules:
    - domain: app.example.com
      policy: deny
`

type testEnv struct {
	t          *testing.T
	path       string
	manager    *Manager
	authorizer *authorization.Authorizer
	clients    map[string]struct{}
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()

	env := &testEnv{t: t, path: filepath.Join(t.TempDir(), "acl.yml"), clients: map[string]struct{}{}}

	env.write(aclOneFactor)

	log := logrus.NewEntry(logrus.New())

	env.manager = NewManager([]string{env.path}, nil, func() (validator.DynamicReferences, error) {
		return validator.DynamicReferences{OAuth2ClientIDs: env.clients}, nil
	}, log)

	config, _, err := env.manager.Load()
	require.NoError(t, err)

	env.authorizer = authorization.NewAuthorizer(&schema.Configuration{AccessControl: config.AccessControl})
	env.manager.consumers = []Consumer{NewAuthorizerConsumer(env.authorizer)}
	env.manager.SetCurrent(config)

	return env
}

func (e *testEnv) write(content string) {
	e.t.Helper()

	require.NoError(e.t, os.WriteFile(e.path, []byte(content), 0o600))
}

func (e *testEnv) level(authorizer *authorization.Authorizer) authorization.Level {
	e.t.Helper()

	object := authorization.NewObjectRaw(&url.URL{Scheme: "https", Host: "app.example.com", Path: "/"}, []byte("GET"))

	_, level := authorizer.GetRequiredLevel(authorization.Subject{}, object)

	return level
}

func TestManagerReloadAppliesValidChange(t *testing.T) {
	env := newTestEnv(t)

	assert.Equal(t, authorization.OneFactor, env.level(env.authorizer))

	inflight := env.authorizer.Snapshot()

	env.write(aclDeny)

	reloaded, err := env.manager.Reload()
	require.NoError(t, err)
	assert.True(t, reloaded)

	assert.Equal(t, authorization.Denied, env.level(env.authorizer), "new requests use the new rules")
	assert.Equal(t, authorization.OneFactor, env.level(inflight), "in-flight requests keep the old rules")
	assert.Equal(t, uint64(2), env.manager.Generation())
}

func TestManagerReloadUnchangedIsNoop(t *testing.T) {
	env := newTestEnv(t)

	reloaded, err := env.manager.Reload()
	require.NoError(t, err)
	assert.False(t, reloaded)
	assert.Equal(t, uint64(1), env.manager.Generation())
	assert.Equal(t, uint64(1), env.authorizer.Generation())
}

func TestManagerReloadInvalidKeepsPrevious(t *testing.T) {
	testCases := []struct {
		name    string
		content string
		err     string
	}{
		{"InvalidYAML", "access_control: [", ""},
		{"InvalidPolicy", "access_control:\n  default_policy: deny\n  rules:\n    - domain: app.example.com\n      policy: nope\n", "policy"},
		{"NoDomain", "access_control:\n  default_policy: deny\n  rules:\n    - policy: deny\n", "domain"},
		{"UnknownKey", "access_control:\n  default_policy: deny\n  rulez: []\n", "rulez"},
		{"UnsupportedSection", aclDeny + "\nserver:\n  address: tcp://:9091/\n", "server.address"},
		{"UnknownOIDCClient", "access_control:\n  default_policy: deny\n  rules:\n    - domain: app.example.com\n      policy: one_factor\n      subject: 'oauth2:client:missing'\n", "missing"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			env := newTestEnv(t)

			env.write(tc.content)

			reloaded, err := env.manager.Reload()
			assert.False(t, reloaded)
			require.Error(t, err)

			if tc.err != "" {
				assert.Contains(t, err.Error(), tc.err)
			}

			assert.Contains(t, err.Error(), "previous configuration")
			assert.Equal(t, authorization.OneFactor, env.level(env.authorizer))
			assert.Equal(t, uint64(1), env.manager.Generation())

			// Fixing the file recovers without a restart.
			env.write(aclDeny)

			reloaded, err = env.manager.Reload()
			require.NoError(t, err)
			assert.True(t, reloaded)
			assert.Equal(t, authorization.Denied, env.level(env.authorizer))
		})
	}
}

func TestManagerReloadMissingFileKeepsPrevious(t *testing.T) {
	env := newTestEnv(t)

	require.NoError(t, os.Remove(env.path))

	reloaded, err := env.manager.Reload()
	assert.False(t, reloaded)
	assert.Error(t, err)
	assert.Equal(t, authorization.OneFactor, env.level(env.authorizer))
}

func TestManagerReloadRefreshesOIDCClientReferences(t *testing.T) {
	env := newTestEnv(t)

	env.write("access_control:\n  default_policy: deny\n  rules:\n    - domain: app.example.com\n      policy: one_factor\n      subject: 'oauth2:client:new'\n")

	_, err := env.manager.Reload()
	require.Error(t, err)

	env.clients["new"] = struct{}{}

	reloaded, err := env.manager.Reload()
	require.NoError(t, err)
	assert.True(t, reloaded)
}

func TestManagerReloadNamedNetworks(t *testing.T) {
	env := newTestEnv(t)

	env.write(`
definitions:
  network:
    office:
      - 10.0.0.0/8
access_control:
  default_policy: deny
  rules:
    - domain: app.example.com
      policy: bypass
      networks:
        - office
`)

	reloaded, err := env.manager.Reload()
	require.NoError(t, err)
	assert.True(t, reloaded)

	object := authorization.NewObjectRaw(&url.URL{Scheme: "https", Host: "app.example.com", Path: "/"}, []byte("GET"))

	_, inside := env.authorizer.GetRequiredLevel(authorization.Subject{IP: []byte{10, 1, 2, 3}}, object)
	_, outside := env.authorizer.GetRequiredLevel(authorization.Subject{IP: []byte{192, 168, 0, 1}}, object)

	assert.Equal(t, authorization.Bypass, inside)
	assert.Equal(t, authorization.Denied, outside)
}

func TestThrottleLeadingAndTrailing(t *testing.T) {
	var calls atomic.Int32

	throttle := NewThrottle(150*time.Millisecond, func() (bool, error) {
		calls.Add(1)

		return true, nil
	}, logrus.NewEntry(logrus.New()))

	reloaded, err := throttle.Reload()
	require.NoError(t, err)
	assert.True(t, reloaded, "first trigger runs immediately")
	assert.Equal(t, int32(1), calls.Load())

	for i := 0; i < 5; i++ {
		reloaded, err = throttle.Reload()
		require.NoError(t, err)
		assert.False(t, reloaded, "triggers inside the interval are deferred")
	}

	assert.Equal(t, int32(1), calls.Load())

	assert.Eventually(t, func() bool { return calls.Load() == 2 }, 2*time.Second, 10*time.Millisecond, "a single trailing run")

	time.Sleep(300 * time.Millisecond)

	assert.Equal(t, int32(2), calls.Load(), "the burst collapses into exactly one trailing run")

	throttle.Stop()

	reloaded, err = throttle.Reload()
	require.NoError(t, err)
	assert.False(t, reloaded)
}
