// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/metrics"
)

func TestSetContentType(t *testing.T) {
	testCases := []struct {
		name     string
		fn       func(*fasthttp.RequestCtx)
		expected string
	}{
		{
			name:     "ShouldSetContentTypeApplicationJSON",
			fn:       SetContentTypeApplicationJSON,
			expected: "application/json; charset=utf-8",
		},
		{
			name:     "ShouldSetContentTypeTextPlain",
			fn:       SetContentTypeTextPlain,
			expected: "text/plain; charset=utf-8",
		},
	}

	for i := range testCases {
		t.Run(testCases[i].name, func(t *testing.T) {
			tc := testCases[i]

			var ctx fasthttp.RequestCtx

			tc.fn(&ctx)

			require.Equal(t, tc.expected, string(ctx.Response.Header.ContentType()))
		})
	}
}

func TestNewAuthenticationProvider(t *testing.T) {
	testCases := []struct {
		name   string
		config schema.Configuration
	}{
		{
			name:   "ShouldReturnNilProviderWhenNoBackendConfigured",
			config: schema.Configuration{},
		},
	}

	for i := range testCases {
		t.Run(testCases[i].name, func(t *testing.T) {
			tc := testCases[i]
			provider := NewAuthenticationProvider(&tc.config, nil)
			require.Nil(t, provider)
		})
	}
}

func TestNewAuthenticationProviderFile(t *testing.T) {
	dir := t.TempDir()

	dbPath := filepath.Join(dir, "users.yml")

	require.NoError(t, os.WriteFile(dbPath, []byte("users: {}\n"), 0600))

	config := schema.Configuration{
		AuthenticationBackend: schema.AuthenticationBackend{
			File: &schema.AuthenticationBackendFile{
				Path:     dbPath,
				Password: schema.DefaultCIPasswordConfig,
			},
		},
	}

	provider := NewAuthenticationProvider(&config, nil)

	assert.NotNil(t, provider)
}

func TestNewAuthenticationProviderLDAP(t *testing.T) {
	config := schema.Configuration{
		AuthenticationBackend: schema.AuthenticationBackend{
			LDAP: &schema.AuthenticationBackendLDAP{
				Address: &schema.AddressLDAP{},
			},
		},
	}

	provider := NewAuthenticationProvider(&config, nil)

	assert.NotNil(t, provider)
}

func TestNewProvidersBasic(t *testing.T) {
	providers := NewProvidersBasic()

	assert.NotNil(t, providers.Clock)
	assert.NotNil(t, providers.Random)
	assert.Nil(t, providers.StorageProvider)
	assert.Nil(t, providers.Authorizer)
	assert.Nil(t, providers.UserProvider)
	assert.Nil(t, providers.Session)
	assert.Nil(t, providers.Notifier)
	assert.Nil(t, providers.OpenIDConnect)
	assert.Nil(t, providers.Metrics)
}

func TestNewProviders(t *testing.T) {
	testCases := []struct {
		name           string
		config         func(t *testing.T) *schema.Configuration
		expectErrs     int
		expectWarns    int
		expectNotifier bool
		expectMetrics  bool
	}{
		{
			"ShouldCreateWithMinimalConfig",
			func(t *testing.T) *schema.Configuration {
				return &schema.Configuration{}
			},
			0,
			0,
			false,
			false,
		},
		{
			"ShouldCreateWithFileNotifier",
			func(t *testing.T) *schema.Configuration {
				return &schema.Configuration{
					Notifier: schema.Notifier{
						FileSystem: &schema.NotifierFileSystem{
							Filename: filepath.Join(t.TempDir(), "notification.txt"),
						},
					},
				}
			},
			0,
			0,
			true,
			false,
		},
		{
			"ShouldCreateWithSMTPNotifier",
			func(t *testing.T) *schema.Configuration {
				return &schema.Configuration{
					Notifier: schema.Notifier{
						SMTP: &schema.NotifierSMTP{
							Address: &schema.AddressSMTP{},
						},
					},
				}
			},
			0,
			0,
			true,
			false,
		},
		{
			"ShouldCreateWithMetrics",
			func(t *testing.T) *schema.Configuration {
				return &schema.Configuration{
					Telemetry: schema.Telemetry{
						Metrics: schema.TelemetryMetrics{
							Enabled: true,
						},
					},
				}
			},
			0,
			0,
			false,
			true,
		},
		{
			"ShouldCreateWithFileAuthBackend",
			func(t *testing.T) *schema.Configuration {
				dir := t.TempDir()
				dbPath := filepath.Join(dir, "users.yml")

				require.NoError(t, os.WriteFile(dbPath, []byte("users: {}\n"), 0600))

				return &schema.Configuration{
					AuthenticationBackend: schema.AuthenticationBackend{
						File: &schema.AuthenticationBackendFile{
							Path:     dbPath,
							Password: schema.DefaultCIPasswordConfig,
						},
					},
				}
			},
			0,
			0,
			false,
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := tc.config(t)

			providers, warns, errs := NewProviders(config, nil)

			assert.Len(t, errs, tc.expectErrs)
			assert.Len(t, warns, tc.expectWarns)

			assert.NotNil(t, providers.Clock)
			assert.NotNil(t, providers.Random)
			assert.NotNil(t, providers.Authorizer)
			assert.NotNil(t, providers.Regulator)

			if tc.expectNotifier {
				assert.NotNil(t, providers.Notifier)
			} else {
				assert.Nil(t, providers.Notifier)
			}

			if tc.expectMetrics {
				assert.NotNil(t, providers.Metrics)
			} else {
				assert.Nil(t, providers.Metrics)
			}
		})
	}
}

func TestNewProvidersRecoveryCodeGauges(t *testing.T) {
	testCases := []struct {
		name       string
		disable    bool
		registered bool
	}{
		{"ShouldRegisterGaugesWhenRecoveryCodesEnabled", false, true},
		{"ShouldNotRegisterGaugesWhenRecoveryCodesDisabled", true, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &schema.Configuration{
				Storage: schema.Storage{
					EncryptionKey: "a_not_so_secure_encryption_key_for_testing",
					Local: &schema.StorageLocal{
						Path: filepath.Join(t.TempDir(), "db.sqlite3"),
					},
				},
				Telemetry: schema.Telemetry{
					Metrics: schema.TelemetryMetrics{
						Enabled: true,
					},
				},
				RecoveryCodes: schema.RecoveryCodes{
					Disable: tc.disable,
				},
			}

			providers, _, errs := NewProviders(config, nil)

			require.Len(t, errs, 0)
			require.NotNil(t, providers.StorageProvider)
			require.NotNil(t, providers.Metrics)

			t.Cleanup(func() {
				assert.NoError(t, providers.StorageProvider.Close())
			})

			require.NoError(t, providers.StorageProvider.StartupCheck())

			prom, ok := providers.Metrics.(*metrics.Prometheus)
			require.True(t, ok)

			families, err := prom.GetGatherer().Gather()
			require.NoError(t, err)

			names := make([]string, 0, len(families))

			for _, family := range families {
				names = append(names, family.GetName())
			}

			for _, name := range []string{"authelia_recovery_codes_users_with_codes", "authelia_recovery_codes_users_low", "authelia_recovery_codes_users_depleted"} {
				if tc.registered {
					assert.Contains(t, names, name)
				} else {
					assert.NotContains(t, names, name)
				}
			}
		})
	}
}

func TestFormatRequestMethod(t *testing.T) {
	testCases := []struct {
		name     string
		method   string
		expected string
	}{
		{"ShouldReturnRFC7231Method", fasthttp.MethodGet, fasthttp.MethodGet},
		{"ShouldReturnRFC7231MethodPatch", fasthttp.MethodPatch, fasthttp.MethodPatch},
		{"ShouldReturnRFC4918Method", schema.MethodPropFind, schema.MethodPropFind},
		{"ShouldReturnIANAMethod", schema.MethodQuery, schema.MethodQuery},
		{"ShouldReturnIANAMethodWithHyphen", schema.MethodVersionControl, schema.MethodVersionControl},
		{"ShouldUppercaseLowercaseMethod", "get", fasthttp.MethodGet},
		{"ShouldUppercaseMixedCaseMethod", "PropFind", schema.MethodPropFind},
		{"ShouldReturnEmptyForUnknownMethod", "FOO", ""},
		{"ShouldReturnEmptyForEmptyMethod", "", ""},
		{"ShouldReturnEmptyForMethodWithWhitespace", " GET", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, formatRequestMethod(tc.method))
		})
	}
}
