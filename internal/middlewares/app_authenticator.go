// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"context"
	"strings"
	"time"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/storage"
)

const (
	appAuthenticatorGarbageCollectionFrequency = 5 * time.Minute

	pathAppAuthenticatorRegister        = "/api/register/app"
	pathAppAuthenticatorRegisterConfirm = "/api/register/app/confirm/"
	pathAppAuthenticatorChallenge       = "/api/secondfactor/app/challenge/"
	pathAppAuthenticatorConfirmation    = "/api/secondfactor/app/confirmation/"
	pathAppAuthenticatorDevice          = "/api/secondfactor/app/device"

	valueRedacted = "{redacted}"
)

// NewAppAuthenticatorGarbageCollector returns a GarbageCollectorProvider which prunes the expired app authenticator
// request nonces, registrations and challenges from the storage provider.
func NewAppAuthenticatorGarbageCollector(provider storage.Provider, clock clock.Provider) *AppAuthenticatorGarbageCollector {
	return &AppAuthenticatorGarbageCollector{provider: provider, clock: clock}
}

// AppAuthenticatorGarbageCollector is a GarbageCollectorProvider for the app authenticator storage.
type AppAuthenticatorGarbageCollector struct {
	provider storage.Provider
	clock    clock.Provider
}

// GarbageCollection prunes the expired app authenticator storage.
func (c *AppAuthenticatorGarbageCollector) GarbageCollection(ctx context.Context) (err error) {
	return c.provider.PruneAppAuthenticator(ctx, c.clock.Now())
}

// GarbageCollectionFrequency returns the frequency of the app authenticator storage garbage collection.
func (c *AppAuthenticatorGarbageCollector) GarbageCollectionFrequency(ctx context.Context) (frequency time.Duration) {
	return appAuthenticatorGarbageCollectionFrequency
}

// IsAppAuthenticatorDevicePath returns true if the path, with any base path removed, belongs to the app
// authenticator device endpoints.
func IsAppAuthenticatorDevicePath(path string) bool {
	_, ok := appAuthenticatorDevicePathRedacted(path)

	return ok
}

// RedactAppAuthenticatorDevicePath returns the path with the secret portion of any app authenticator device endpoint
// redacted. Unlike IsAppAuthenticatorDevicePath the path may include a base path.
func RedactAppAuthenticatorDevicePath(path string) string {
	for i := 0; i < len(path); i++ {
		if path[i] != '/' || !strings.HasPrefix(path[i:], "/api/") {
			continue
		}

		if redacted, ok := appAuthenticatorDevicePathRedacted(path[i:]); ok {
			return path[:i] + redacted
		}
	}

	return path
}

func appAuthenticatorDevicePathRedacted(path string) (redacted string, ok bool) {
	switch path {
	case pathAppAuthenticatorRegister, pathAppAuthenticatorDevice:
		return path, true
	}

	for _, prefix := range []string{pathAppAuthenticatorRegisterConfirm, pathAppAuthenticatorRegister + "/", pathAppAuthenticatorChallenge, pathAppAuthenticatorConfirmation, pathAppAuthenticatorDevice + "/"} {
		if strings.HasPrefix(path, prefix) {
			return prefix + valueRedacted, true
		}
	}

	return path, false
}
