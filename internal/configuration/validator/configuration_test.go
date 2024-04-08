package validator

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func newDefaultConfig() schema.Configuration {
	config := schema.Configuration{}
	config.Server.Address = &schema.AddressTCP{Address: schema.NewAddressFromNetworkValues("tcp", loopback, 9090)}
	config.Log.Level = "info"
	config.Log.Format = "text"
	config.IdentityValidation.ResetPassword = schema.IdentityValidationResetPassword{
		JWTSecret: testJWTSecret,
	}

	config.AuthenticationBackend.File = &schema.AuthenticationBackendFile{
		Path: "/a/path",
	}
	config.AccessControl = schema.AccessControl{
		DefaultPolicy: "two_factor",
	}
	config.Session = schema.Session{
		Secret: "secret",
		Cookies: []schema.SessionCookie{
			{
				SessionCookieCommon: schema.SessionCookieCommon{
					Name: "authelia_session",
				},
				Domain:      exampleDotCom,
				AutheliaURL: &url.URL{Scheme: schemeHTTPS, Host: authdot + exampleDotCom},
			},
		},
	}
	config.Storage.EncryptionKey = testEncryptionKey
	config.Storage.Local = &schema.StorageLocal{
		Path: "abc",
	}
	config.Notifier = schema.Notifier{
		FileSystem: &schema.NotifierFileSystem{
			Filename: "/tmp/file",
		},
	}

	return config
}

func TestShouldEnsureNotifierConfigIsProvided(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()

	ValidateConfiguration(&config, validator)
	require.Len(t, validator.Errors(), 0)

	config = newDefaultConfig()

	config.Notifier.SMTP = nil
	config.Notifier.FileSystem = nil

	ValidateConfiguration(&config, validator)
	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "notifier: you must ensure either the 'smtp' or 'filesystem' notifier is configured")
}

func TestShouldAddDefaultAccessControl(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()

	config.AccessControl.DefaultPolicy = ""
	config.AccessControl.Rules = []schema.AccessControlRule{
		{
			Policy: "bypass",
			Domains: []string{
				"public.example.com",
			},
		},
	}

	ValidateConfiguration(&config, validator)
	require.Len(t, validator.Errors(), 0)
	assert.NotNil(t, config.AccessControl)
	assert.Equal(t, "deny", config.AccessControl.DefaultPolicy)
}

func TestShouldRaiseErrorWithBadDefaultRedirectionURL(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()
	config.Session.Cookies[0].DefaultRedirectionURL = &url.URL{Host: "localhost"}

	ValidateConfiguration(&config, validator)
	require.Len(t, validator.Errors(), 2)
	require.Len(t, validator.Warnings(), 1)

	assert.EqualError(t, validator.Errors()[0], "session: domain config #1 (domain 'example.com'): option 'default_redirection_url' is not absolute with a value of '//localhost'")
	assert.EqualError(t, validator.Errors()[1], "session: domain config #1 (domain 'example.com'): option 'default_redirection_url' does not share a cookie scope with domain 'example.com' with a value of '//localhost'")
	assert.EqualError(t, validator.Warnings()[0], "access_control: no rules have been specified so the 'default_policy' of 'two_factor' is going to be applied to all requests")
}

func TestShouldRaiseErrorWithLegacyDefaultRedirectionURL(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()
	config.DefaultRedirectionURL = &url.URL{Host: "localhost"} //nolint:staticcheck

	ValidateConfiguration(&config, validator)
	require.Len(t, validator.Errors(), 1)
	require.Len(t, validator.Warnings(), 1)

	assert.EqualError(t, validator.Errors()[0], "session: option 'cookies' must be configured with the per cookie option 'default_redirection_url' but the global one is configured which is not supported")
	assert.EqualError(t, validator.Warnings()[0], "access_control: no rules have been specified so the 'default_policy' of 'two_factor' is going to be applied to all requests")
}

func TestShouldAllowLegacyDefaultRedirectionURL(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()

	config.Session.Cookies = nil

	config.DefaultRedirectionURL = &url.URL{Scheme: "https", Host: "www.example.com"} //nolint:staticcheck
	config.Session.Domain = "example.com"                                             //nolint:staticcheck

	ValidateConfiguration(&config, validator)
	require.Len(t, validator.Errors(), 0)
	require.Len(t, validator.Warnings(), 2)

	assert.EqualError(t, validator.Warnings()[0], "access_control: no rules have been specified so the 'default_policy' of 'two_factor' is going to be applied to all requests")
	assert.EqualError(t, validator.Warnings()[1], "session: option 'domain' is deprecated in v4.38.0 and has been replaced by a multi-domain configuration: this has automatically been mapped for you but you will need to adjust your configuration to remove this message and receive the latest messages")

	assert.Equal(t, "example.com", config.Session.Cookies[0].Domain)
	assert.Equal(t, &url.URL{Scheme: schemeHTTPS, Host: "www.example.com"}, config.Session.Cookies[0].DefaultRedirectionURL)
}

func TestShouldNotOverrideCertificatesDirectoryAndShouldPassWhenBlank(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()

	ValidateConfiguration(&config, validator)

	assert.Len(t, validator.Errors(), 0)
	require.Len(t, validator.Warnings(), 1)

	require.Equal(t, "", config.CertificatesDirectory)

	assert.EqualError(t, validator.Warnings()[0], "access_control: no rules have been specified so the 'default_policy' of 'two_factor' is going to be applied to all requests")
}

func TestShouldRaiseErrorOnInvalidCertificatesDirectory(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()
	config.CertificatesDirectory = "not-a-real-file.go"

	ValidateConfiguration(&config, validator)

	require.Len(t, validator.Errors(), 1)
	require.Len(t, validator.Warnings(), 1)

	if runtime.GOOS == "windows" {
		assert.EqualError(t, validator.Errors()[0], "the location 'certificates_directory' could not be inspected: CreateFile not-a-real-file.go: The system cannot find the file specified.")
	} else {
		assert.EqualError(t, validator.Errors()[0], "the location 'certificates_directory' could not be inspected: stat not-a-real-file.go: no such file or directory")
	}

	assert.EqualError(t, validator.Warnings()[0], "access_control: no rules have been specified so the 'default_policy' of 'two_factor' is going to be applied to all requests")

	config = newDefaultConfig()

	validator = schema.NewStructValidator()
	config.CertificatesDirectory = "const.go"

	ValidateConfiguration(&config, validator)

	require.Len(t, validator.Errors(), 1)
	require.Len(t, validator.Warnings(), 1)

	assert.EqualError(t, validator.Errors()[0], "the location 'certificates_directory' refers to 'const.go' is not a directory")
	assert.EqualError(t, validator.Warnings()[0], "access_control: no rules have been specified so the 'default_policy' of 'two_factor' is going to be applied to all requests")
}

func TestShouldNotRaiseErrorOnValidCertificatesDirectory(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()
	config.CertificatesDirectory = "../../suites/common/pki"

	ValidateConfiguration(&config, validator)

	assert.Len(t, validator.Errors(), 0)
	require.Len(t, validator.Warnings(), 1)

	assert.EqualError(t, validator.Warnings()[0], "access_control: no rules have been specified so the 'default_policy' of 'two_factor' is going to be applied to all requests")
}

func TestValidateDefault2FAMethod(t *testing.T) {
	testCases := []struct {
		desc         string
		have         *schema.Configuration
		expectedErrs []string
	}{
		{
			desc: "ShouldAllowConfiguredMethodTOTP",
			have: &schema.Configuration{
				Default2FAMethod: "totp",
				DuoAPI: schema.DuoAPI{
					SecretKey:      "a key",
					IntegrationKey: "another key",
					Hostname:       "none",
				},
			},
		},
		{
			desc: "ShouldAllowConfiguredMethodWebAuthn",
			have: &schema.Configuration{
				Default2FAMethod: "webauthn",
				DuoAPI: schema.DuoAPI{
					SecretKey:      "a key",
					IntegrationKey: "another key",
					Hostname:       "none",
				},
			},
		},
		{
			desc: "ShouldAllowConfiguredMethodMobilePush",
			have: &schema.Configuration{
				Default2FAMethod: "mobile_push",
				DuoAPI: schema.DuoAPI{
					SecretKey:      "a key",
					IntegrationKey: "another key",
					Hostname:       "none",
				},
			},
		},
		{
			desc: "ShouldNotAllowDisabledMethodTOTP",
			have: &schema.Configuration{
				Default2FAMethod: "totp",
				DuoAPI: schema.DuoAPI{
					SecretKey:      "a key",
					IntegrationKey: "another key",
					Hostname:       "none",
				},
				TOTP: schema.TOTP{Disable: true},
			},
			expectedErrs: []string{
				"option 'default_2fa_method' must be one of the enabled options 'webauthn' or 'mobile_push' but it's configured as 'totp'",
			},
		},
		{
			desc: "ShouldNotAllowDisabledMethodWebAuthn",
			have: &schema.Configuration{
				Default2FAMethod: "webauthn",
				DuoAPI: schema.DuoAPI{
					SecretKey:      "a key",
					IntegrationKey: "another key",
					Hostname:       "none",
				},
				WebAuthn: schema.WebAuthn{Disable: true},
			},
			expectedErrs: []string{
				"option 'default_2fa_method' must be one of the enabled options 'totp' or 'mobile_push' but it's configured as 'webauthn'",
			},
		},
		{
			desc: "ShouldNotAllowDisabledMethodMobilePush",
			have: &schema.Configuration{
				Default2FAMethod: "mobile_push",
				DuoAPI:           schema.DuoAPI{Disable: true},
			},
			expectedErrs: []string{
				"option 'default_2fa_method' must be one of the enabled options 'totp' or 'webauthn' but it's configured as 'mobile_push'",
			},
		},
		{
			desc: "ShouldNotAllowInvalidMethodDuo",
			have: &schema.Configuration{
				Default2FAMethod: "duo",
			},
			expectedErrs: []string{
				"option 'default_2fa_method' must be one of 'totp', 'webauthn', or 'mobile_push' but it's configured as 'duo'",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			validator := schema.NewStructValidator()

			validateDefault2FAMethod(tc.have, validator)

			assert.Len(t, validator.Warnings(), 0)

			errs := validator.Errors()

			require.Len(t, errs, len(tc.expectedErrs))

			for i, expected := range tc.expectedErrs {
				t.Run(fmt.Sprintf("Err%d", i+1), func(t *testing.T) {
					assert.EqualError(t, errs[i], expected)
				})
			}
		})
	}
}

func TestNewValidateCtx(t *testing.T) {
	ctx := NewValidateCtx()
	require.NotNil(t, ctx)
	assert.NotNil(t, ctx.Context)
}

func TestValidateCtx_GetHTTPClient_DefaultTLSConfig(t *testing.T) {
	ctx := NewValidateCtx()
	client := ctx.GetHTTPClient()

	require.NotNil(t, client)
	assert.IsType(t, &http.Client{}, client)
	assert.NotNil(t, client.Transport)
	transport, ok := client.Transport.(*http.Transport)
	require.True(t, ok)
	assert.Nil(t, transport.TLSClientConfig)
}

func TestValidateCtx_GetHTTPClient_CustomTLSConfig(t *testing.T) {
	customTLSConfig := &tls.Config{
		InsecureSkipVerify: true, //nolint:gosec
		MinVersion:         tls.VersionTLS12,
		MaxVersion:         tls.VersionTLS13,
	}

	ctx := NewValidateCtx()
	WithTLSConfig(customTLSConfig)(ctx)
	client := ctx.GetHTTPClient()

	require.NotNil(t, client)
	transport, ok := client.Transport.(*http.Transport)
	require.True(t, ok)
	assert.Equal(t, true, transport.TLSClientConfig.InsecureSkipVerify)
}

func TestValidateCtx_CacheSectorIdentifierURIs(t *testing.T) {
	ctx := NewValidateCtx()
	ctx.cacheSectorIdentifierURIs = make(map[string][]string)

	uri := "https://example.com"
	ctx.cacheSectorIdentifierURIs[uri] = []string{"redirect_uri1", "redirect_uri2"}

	cachedURIs, exists := ctx.cacheSectorIdentifierURIs[uri]
	require.True(t, exists)
	assert.Equal(t, []string{"redirect_uri1", "redirect_uri2"}, cachedURIs)
}
