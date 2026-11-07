// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"net/url"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/configuration/validator"
	"github.com/authelia/authelia/v4/internal/dynamic"
)

const (
	testDynamicOneFactor = "access_control:\n  default_policy: deny\n  rules:\n    - domain: app.example.com\n      policy: one_factor\n"
	testDynamicDeny      = "access_control:\n  default_policy: deny\n  rules:\n    - domain: app.example.com\n      policy: deny\n"
)

func newDynamicTestCtx(t *testing.T, watch bool) (ctx *mockServiceCtx, path string, authorizer *authorization.Authorizer) {
	t.Helper()

	path = filepath.Join(t.TempDir(), "acl.yml")

	require.NoError(t, os.WriteFile(path, []byte(testDynamicOneFactor), 0o600))

	ctx = newMockServiceCtx()

	manager := dynamic.NewManager([]string{path}, nil, func() (validator.DynamicReferences, error) {
		return validator.DynamicReferences{}, nil
	}, ctx.logger)

	config, _, err := manager.Load()
	require.NoError(t, err)

	authorizer = authorization.NewAuthorizer(&schema.Configuration{AccessControl: config.AccessControl})

	manager.AddConsumer(dynamic.NewAuthorizerConsumer(authorizer))
	manager.SetCurrent(config)

	ctx.config.DynamicConfiguration = schema.DynamicConfigurationSource{Paths: []string{path}, Watch: watch}
	ctx.providers.Authorizer = authorizer
	ctx.providers.DynamicConfiguration = manager

	return ctx, path, authorizer
}

func dynamicTestLevel(authorizer *authorization.Authorizer) authorization.Level {
	object := authorization.NewObjectRaw(&url.URL{Scheme: "https", Host: "app.example.com", Path: "/"}, []byte("GET"))

	_, level := authorizer.GetRequiredLevel(authorization.Subject{}, object)

	return level
}

func TestProvisionApplicationReloadSignalDynamicConfigurationReloadsWithoutRestart(t *testing.T) {
	ctx, path, authorizer := newDynamicTestCtx(t, false)

	service, err := ProvisionApplicationReloadSignal(ctx)
	require.NoError(t, err)
	require.NotNil(t, service)

	errCh := make(chan error, 1)

	go func() {
		errCh <- service.Run()
	}()

	time.Sleep(100 * time.Millisecond)

	require.Equal(t, authorization.OneFactor, dynamicTestLevel(authorizer))

	// An invalid file must be rejected, must not stop the service, and must leave the previous rules in place.
	require.NoError(t, os.WriteFile(path, []byte("access_control: ["), 0o600))
	require.NoError(t, syscall.Kill(os.Getpid(), syscall.SIGHUP))

	time.Sleep(300 * time.Millisecond)

	assert.Equal(t, authorization.OneFactor, dynamicTestLevel(authorizer))

	select {
	case err = <-errCh:
		t.Fatalf("the service must not return on a dynamic reload, got: %v", err)
	default:
	}

	// A valid file is applied without the service returning ErrApplicationReload.
	require.NoError(t, os.WriteFile(path, []byte(testDynamicDeny), 0o600))
	require.NoError(t, syscall.Kill(os.Getpid(), syscall.SIGHUP))

	assert.Eventually(t, func() bool { return dynamicTestLevel(authorizer) == authorization.Denied }, 5*time.Second, 20*time.Millisecond)

	select {
	case err = <-errCh:
		t.Fatalf("the service must not return on a dynamic reload, got: %v", err)
	default:
	}

	service.Shutdown()

	select {
	case err = <-errCh:
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("service did not shut down")
	}
}

func TestProvisionDynamicConfigFileWatcherDisabled(t *testing.T) {
	ctx, _, _ := newDynamicTestCtx(t, false)

	service, err := ProvisionDynamicConfigFileWatcher(ctx)
	assert.NoError(t, err)
	assert.Nil(t, service)

	service, err = ProvisionDynamicConfigFileWatcher(newMockServiceCtx())
	assert.NoError(t, err)
	assert.Nil(t, service)
}

func TestProvisionDynamicConfigFileWatcherReloadsOnChange(t *testing.T) {
	ctx, path, authorizer := newDynamicTestCtx(t, true)

	service, err := ProvisionDynamicConfigFileWatcher(ctx)
	require.NoError(t, err)
	require.NotNil(t, service)

	assert.Equal(t, "dynamic-configuration", service.ServiceName())

	errCh := make(chan error, 1)

	go func() {
		errCh <- service.Run()
	}()

	time.Sleep(100 * time.Millisecond)

	require.Equal(t, authorization.OneFactor, dynamicTestLevel(authorizer))

	require.NoError(t, os.WriteFile(path, []byte(testDynamicDeny), 0o600))

	assert.Eventually(t, func() bool { return dynamicTestLevel(authorizer) == authorization.Denied }, 5*time.Second, 20*time.Millisecond)

	// An atomic replace, as performed by many editors, is also detected.
	replacement := path + ".tmp"

	require.NoError(t, os.WriteFile(replacement, []byte(testDynamicOneFactor), 0o600))
	require.NoError(t, os.Rename(replacement, path))

	assert.Eventually(t, func() bool { return dynamicTestLevel(authorizer) == authorization.OneFactor }, 5*time.Second, 20*time.Millisecond)

	service.Shutdown()

	select {
	case err = <-errCh:
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("service did not shut down")
	}
}
