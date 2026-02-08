package validator

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

const unexistingFilePath = "/tmp/unexisting_file"

func TestShouldSetDefaultServerValues(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.Configuration{}

	ValidateServer(config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Len(t, validator.Warnings(), 0)

	assert.Equal(t, schema.DefaultServerConfiguration.Address, config.Server.Address)
	assert.Equal(t, schema.DefaultServerConfiguration.Buffers.Read, config.Server.Buffers.Read)
	assert.Equal(t, schema.DefaultServerConfiguration.Buffers.Write, config.Server.Buffers.Write)
	assert.Equal(t, schema.DefaultServerConfiguration.Timeouts.Read, config.Server.Timeouts.Read)
	assert.Equal(t, schema.DefaultServerConfiguration.Timeouts.Write, config.Server.Timeouts.Write)
	assert.Equal(t, schema.DefaultServerConfiguration.Timeouts.Idle, config.Server.Timeouts.Idle)
	assert.Equal(t, schema.DefaultServerConfiguration.TLS.Key, config.Server.TLS.Key)
	assert.Equal(t, schema.DefaultServerConfiguration.TLS.Certificate, config.Server.TLS.Certificate)
	assert.Equal(t, schema.DefaultServerConfiguration.Endpoints.EnableExpvars, config.Server.Endpoints.EnableExpvars)
	assert.Equal(t, schema.DefaultServerConfiguration.Endpoints.EnablePprof, config.Server.Endpoints.EnablePprof)
	assert.Equal(t, schema.DefaultServerConfiguration.Endpoints.Authz, config.Server.Endpoints.Authz)
}

func TestShouldSetDefaultConfigRateLimits(t *testing.T) {
	testCases := []struct {
		name     string
		config   *schema.Configuration
		expected schema.ServerEndpointRateLimits
	}{
		{
			name: "ShouldSetDefaults",
			config: &schema.Configuration{
				IdentityValidation: schema.IdentityValidation{
					ResetPassword: schema.IdentityValidationResetPassword{
						JWTSecret: "dfgdfgdzfgdgdfgdfg",
					},
				},
			},
			expected: schema.ServerEndpointRateLimits{
				ResetPasswordStart: schema.ServerEndpointRateLimit{
					Buckets: []schema.ServerEndpointRateLimitBucket{
						{Period: 10 * time.Minute, Requests: 5},
						{Period: 15 * time.Minute, Requests: 10},
						{Period: 30 * time.Minute, Requests: 15},
					},
				},
				ResetPasswordFinish: schema.ServerEndpointRateLimit{
					Buckets: []schema.ServerEndpointRateLimitBucket{
						{Period: 1 * time.Minute, Requests: 10},
						{Period: 2 * time.Minute, Requests: 15},
					},
				},
				SecondFactorTOTP: schema.ServerEndpointRateLimit{
					Buckets: []schema.ServerEndpointRateLimitBucket{
						{Period: 1 * time.Minute, Requests: 30},
						{Period: 2 * time.Minute, Requests: 40},
						{Period: 10 * time.Minute, Requests: 50},
					},
				},
				SecondFactorDuo: schema.ServerEndpointRateLimit{
					Buckets: []schema.ServerEndpointRateLimitBucket{
						{Period: 1 * time.Minute, Requests: 10},
						{Period: 2 * time.Minute, Requests: 15},
					},
				},
				SessionElevationStart: schema.ServerEndpointRateLimit{
					Buckets: []schema.ServerEndpointRateLimitBucket{
						{Period: schema.DefaultIdentityValidation.ElevatedSession.CodeLifespan * 1, Requests: 3},
						{Period: schema.DefaultIdentityValidation.ElevatedSession.CodeLifespan * 2, Requests: 5},
						{Period: schema.DefaultIdentityValidation.ElevatedSession.CodeLifespan * 12, Requests: 15},
					},
				},
				SessionElevationFinish: schema.ServerEndpointRateLimit{
					Buckets: []schema.ServerEndpointRateLimitBucket{
						{Period: schema.DefaultIdentityValidation.ElevatedSession.ElevationLifespan * 1, Requests: 3},
						{Period: schema.DefaultIdentityValidation.ElevatedSession.ElevationLifespan * 2, Requests: 5},
						{Period: schema.DefaultIdentityValidation.ElevatedSession.ElevationLifespan * 6, Requests: 15},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator := schema.NewStructValidator()

			ValidateIdentityValidation(tc.config, validator)
			ValidateServer(tc.config, validator)

			assert.Len(t, validator.Errors(), 0)
			assert.Len(t, validator.Warnings(), 0)

			assert.Equal(t, tc.expected.ResetPasswordStart, tc.config.Server.Endpoints.RateLimits.ResetPasswordStart)
			assert.Equal(t, tc.expected.ResetPasswordFinish, tc.config.Server.Endpoints.RateLimits.ResetPasswordFinish)
			assert.Equal(t, tc.expected.SecondFactorTOTP, tc.config.Server.Endpoints.RateLimits.SecondFactorTOTP)
			assert.Equal(t, tc.expected.SecondFactorDuo, tc.config.Server.Endpoints.RateLimits.SecondFactorDuo)
			assert.Equal(t, tc.expected.SessionElevationStart, tc.config.Server.Endpoints.RateLimits.SessionElevationStart)
			assert.Equal(t, tc.expected.SessionElevationFinish, tc.config.Server.Endpoints.RateLimits.SessionElevationFinish)
		})
	}
}

func TestValidateRateLimitErrors(t *testing.T) {
	have := &schema.Configuration{
		Server: schema.Server{
			Endpoints: schema.ServerEndpoints{
				RateLimits: schema.ServerEndpointRateLimits{
					ResetPasswordStart: schema.ServerEndpointRateLimit{
						Buckets: []schema.ServerEndpointRateLimitBucket{
							{Period: time.Second, Requests: 5},
						},
					},
					ResetPasswordFinish: schema.ServerEndpointRateLimit{
						Buckets: []schema.ServerEndpointRateLimitBucket{
							{Period: time.Second, Requests: 5},
						},
					},
					SecondFactorTOTP: schema.ServerEndpointRateLimit{
						Buckets: []schema.ServerEndpointRateLimitBucket{
							{Period: time.Second, Requests: 5},
						},
					},
					SecondFactorDuo: schema.ServerEndpointRateLimit{
						Buckets: []schema.ServerEndpointRateLimitBucket{
							{Period: time.Second, Requests: 5},
						},
					},
					SessionElevationStart: schema.ServerEndpointRateLimit{
						Buckets: []schema.ServerEndpointRateLimitBucket{
							{Period: time.Second, Requests: 5},
						},
					},
					SessionElevationFinish: schema.ServerEndpointRateLimit{
						Buckets: []schema.ServerEndpointRateLimitBucket{
							{Period: time.Duration(0), Requests: 0},
						},
					},
				},
			},
		},
	}

	validator := schema.NewStructValidator()

	ValidateServer(have, validator)

	require.Len(t, validator.Errors(), 7)
	assert.EqualError(t, validator.Errors()[0], "server: endpoints: rate_limits: reset_password_start: bucket 1: option 'period' has a value of '1s' but it must be greater than 10 seconds")
	assert.EqualError(t, validator.Errors()[1], "server: endpoints: rate_limits: reset_password_finish: bucket 1: option 'period' has a value of '1s' but it must be greater than 10 seconds")
	assert.EqualError(t, validator.Errors()[2], "server: endpoints: rate_limits: second_factor_totp: bucket 1: option 'period' has a value of '1s' but it must be greater than 10 seconds")
	assert.EqualError(t, validator.Errors()[3], "server: endpoints: rate_limits: second_factor_duo: bucket 1: option 'period' has a value of '1s' but it must be greater than 10 seconds")
	assert.EqualError(t, validator.Errors()[4], "server: endpoints: rate_limits: session_elevation_start: bucket 1: option 'period' has a value of '1s' but it must be greater than 10 seconds")
	assert.EqualError(t, validator.Errors()[5], "server: endpoints: rate_limits: session_elevation_finish: bucket 1: option 'period' must have a value")
	assert.EqualError(t, validator.Errors()[6], "server: endpoints: rate_limits: session_elevation_finish: bucket 1: option 'requests' has a value of '0' but it must be greater than 1")
}

func TestValidateSeverAddress(t *testing.T) {
	config := &schema.Configuration{
		Server: schema.Server{
			Address: &schema.AddressTCP{Address: MustParseAddress("tcp://:9091/path/")},
		},
	}

	validator := schema.NewStructValidator()
	ValidateServer(config, validator)

	require.Len(t, validator.Errors(), 1)
	assert.Len(t, validator.Warnings(), 0)

	assert.EqualError(t, validator.Errors()[0], "server: option 'address' must not have a path with a forward slash but it's configured as '/path/'")
}

func TestValidateServerShouldCorrectlyIdentifyValidAddressSchemes(t *testing.T) {
	testCases := []struct {
		have     string
		expected string
	}{
		{schema.AddressSchemeTCP, ""},
		{schema.AddressSchemeTCP4, ""},
		{schema.AddressSchemeTCP6, ""},
		{schema.AddressSchemeUDP, "server: option 'address' with value 'udp://:9091' is invalid: scheme must be one of 'tcp', 'tcp4', 'tcp6', 'unix', or 'fd' but is configured as 'udp'"},
		{schema.AddressSchemeUDP4, "server: option 'address' with value 'udp4://:9091' is invalid: scheme must be one of 'tcp', 'tcp4', 'tcp6', 'unix', or 'fd' but is configured as 'udp4'"},
		{schema.AddressSchemeUDP6, "server: option 'address' with value 'udp6://:9091' is invalid: scheme must be one of 'tcp', 'tcp4', 'tcp6', 'unix', or 'fd' but is configured as 'udp6'"},
		{schema.AddressSchemeUnix, ""},
		{"http", "server: option 'address' with value 'http://:9091' is invalid: scheme must be one of 'tcp', 'tcp4', 'tcp6', 'unix', or 'fd' but is configured as 'http'"},
	}

	validator := schema.NewStructValidator()

	for _, tc := range testCases {
		t.Run(tc.have, func(t *testing.T) {
			have := &schema.Configuration{
				Server: schema.Server{
					Buffers: schema.ServerBuffers{
						Read:  -1,
						Write: -1,
					},
					Timeouts: schema.ServerTimeouts{
						Read:  time.Second * -1,
						Write: time.Second * -1,
						Idle:  time.Second * -1,
					},
				},
			}

			validator.Clear()

			switch tc.have {
			case schema.AddressSchemeUnix:
				have.Server.Address = &schema.AddressTCP{Address: schema.NewAddressUnix("/path/to/authelia.sock")}
			default:
				have.Server.Address = &schema.AddressTCP{Address: schema.NewAddressFromNetworkValues(tc.have, "", 9091)}
			}

			ValidateServer(have, validator)

			assert.Len(t, validator.Warnings(), 0)

			if tc.expected == "" {
				assert.Len(t, validator.Errors(), 0)
			} else {
				require.Len(t, validator.Errors(), 1)
				assert.EqualError(t, validator.Errors()[0], tc.expected)
			}
		})
	}
}

func TestShouldDefaultOnNegativeValues(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.Configuration{
		Server: schema.Server{
			Buffers: schema.ServerBuffers{
				Read:  -1,
				Write: -1,
			},
			Timeouts: schema.ServerTimeouts{
				Read:  time.Second * -1,
				Write: time.Second * -1,
				Idle:  time.Second * -1,
			},
		},
	}

	ValidateServer(config, validator)

	require.Len(t, validator.Errors(), 0)

	assert.Equal(t, schema.DefaultServerConfiguration.Buffers.Read, config.Server.Buffers.Read)
	assert.Equal(t, schema.DefaultServerConfiguration.Buffers.Write, config.Server.Buffers.Write)

	assert.Equal(t, schema.DefaultServerConfiguration.Timeouts.Read, config.Server.Timeouts.Read)
	assert.Equal(t, schema.DefaultServerConfiguration.Timeouts.Write, config.Server.Timeouts.Write)
	assert.Equal(t, schema.DefaultServerConfiguration.Timeouts.Idle, config.Server.Timeouts.Idle)
}

func TestShouldRaiseOnNonAlphanumericCharsInPath(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.Configuration{
		Server: schema.Server{
			Address: &schema.AddressTCP{
				Address: MustParseAddress("tcp://:9091/app-le"),
			},
		},
	}

	ValidateServer(config, validator)

	require.Len(t, validator.Errors(), 1)

	assert.EqualError(t, validator.Errors()[0], "server: option 'address' must have a path with only alphanumeric characters but it's configured as '/app-le'")
}

func TestShouldRaiseOnForwardSlashInPath(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.Configuration{
		Server: schema.Server{
			Address: &schema.AddressTCP{
				Address: MustParseAddress("tcp://:9091/app/le"),
			},
		},
	}

	ValidateServer(config, validator)

	require.Len(t, validator.Errors(), 1)

	assert.Error(t, validator.Errors()[0], "server path must not contain any forward slashes")
}

func TestShouldValidateAndUpdateAddress(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()
	config.Server.Address = nil

	ValidateServer(&config, validator)

	require.Len(t, validator.Errors(), 0)
	assert.Equal(t, "tcp://:9091/", config.Server.Address.String())
}

func TestShouldDisableHealthcheckForUnixSocketAndFileDescriptor(t *testing.T) {
	testCases := []struct {
		name     string
		address  *schema.AddressTCP
		expected bool
	}{
		{
			name:     "ShouldDisableHealthcheckForUnixSocket",
			address:  &schema.AddressTCP{Address: schema.NewAddressUnix("/path/to/authelia.sock")},
			expected: true,
		},
		{
			name:     "ShouldDisableHealthcheckForFileDescriptor",
			address:  &schema.AddressTCP{Address: MustParseAddress("fd://3")},
			expected: true,
		},
		{
			name:     "ShouldNotDisableHealthcheckForTCP",
			address:  &schema.AddressTCP{Address: schema.NewAddressFromNetworkValues("tcp", "127.0.0.1", 9091)},
			expected: false,
		},
		{
			name:     "ShouldNotDisableHealthcheckForTCP4",
			address:  &schema.AddressTCP{Address: schema.NewAddressFromNetworkValues("tcp4", "127.0.0.1", 9091)},
			expected: false,
		},
		{
			name:     "ShouldNotDisableHealthcheckForTCP6",
			address:  &schema.AddressTCP{Address: schema.NewAddressFromNetworkValues("tcp6", "::1", 9091)},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator := schema.NewStructValidator()
			config := &schema.Configuration{
				Server: schema.Server{
					Address:            tc.address,
					DisableHealthcheck: false,
				},
			}

			ValidateServerAddress(config, validator)

			assert.Len(t, validator.Errors(), 0)
			assert.Equal(t, tc.expected, config.Server.DisableHealthcheck)
		})
	}
}

func TestShouldRaiseErrorWhenTLSCertWithoutKeyIsProvided(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()

	file, err := os.CreateTemp("", "cert")
	require.NoError(t, err)

	defer os.Remove(file.Name())

	config.Server.TLS.Certificate = file.Name()

	ValidateServer(&config, validator)
	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "server: tls: option 'certificate' must also be accompanied by option 'key'")
}

func TestShouldRaiseErrorWhenTLSCertDoesNotExist(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()

	file, err := os.CreateTemp("", "key")
	require.NoError(t, err)

	defer os.Remove(file.Name())

	config.Server.TLS.Certificate = unexistingFilePath
	config.Server.TLS.Key = file.Name()

	ValidateServer(&config, validator)
	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "server: tls: option 'certificate' with path '/tmp/unexisting_file' refers to a file that doesn't exist")
}

func TestShouldRaiseErrorWhenTLSKeyWithoutCertIsProvided(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()

	file, err := os.CreateTemp("", "key")
	require.NoError(t, err)

	defer os.Remove(file.Name())

	config.Server.TLS.Key = file.Name()

	ValidateServer(&config, validator)
	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "server: tls: option 'key' must also be accompanied by option 'certificate'")
}

func TestShouldRaiseErrorWhenTLSKeyDoesNotExist(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()

	file, err := os.CreateTemp("", "key")
	require.NoError(t, err)

	defer os.Remove(file.Name())

	config.Server.TLS.Key = unexistingFilePath
	config.Server.TLS.Certificate = file.Name()

	ValidateServer(&config, validator)
	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "server: tls: option 'key' with path '/tmp/unexisting_file' refers to a file that doesn't exist")
}

func TestShouldNotRaiseErrorWhenBothTLSCertificateAndKeyAreProvided(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()

	certFile, err := os.CreateTemp("", "cert")
	require.NoError(t, err)

	defer os.Remove(certFile.Name())

	keyFile, err := os.CreateTemp("", "key")
	require.NoError(t, err)

	defer os.Remove(keyFile.Name())

	config.Server.TLS.Certificate = certFile.Name()
	config.Server.TLS.Key = keyFile.Name()

	ValidateServer(&config, validator)
	require.Len(t, validator.Errors(), 0)
}

func TestShouldRaiseErrorWhenTLSClientCertificateDoesNotExist(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()

	certFile, err := os.CreateTemp("", "cert")
	require.NoError(t, err)

	defer os.Remove(certFile.Name())

	keyFile, err := os.CreateTemp("", "key")
	require.NoError(t, err)

	defer os.Remove(keyFile.Name())

	config.Server.TLS.Certificate = certFile.Name()
	config.Server.TLS.Key = keyFile.Name()
	config.Server.TLS.ClientCertificates = []string{"/tmp/unexisting"}

	ValidateServer(&config, validator)
	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "server: tls: option 'client_certificates' with path '/tmp/unexisting' refers to a file that doesn't exist")
}

func TestShouldRaiseErrorWhenTLSClientAuthIsDefinedButNotServerCertificate(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()

	certFile, err := os.CreateTemp("", "cert")
	require.NoError(t, err)

	defer os.Remove(certFile.Name())

	config.Server.TLS.ClientCertificates = []string{certFile.Name()}

	ValidateServer(&config, validator)
	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "server: tls: client authentication cannot be configured if no server certificate and key are provided")
}

func TestShouldNotUpdateConfig(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()

	ValidateServer(&config, validator)

	require.Len(t, validator.Errors(), 0)
	assert.Equal(t, "tcp://127.0.0.1:9090/", config.Server.Address.String())
}

func TestServerEndpointsDevelShouldWarn(t *testing.T) {
	config := &schema.Configuration{
		Server: schema.Server{
			Endpoints: schema.ServerEndpoints{
				EnablePprof:   true,
				EnableExpvars: true,
			},
		},
	}

	validator := schema.NewStructValidator()

	ValidateServer(config, validator)

	require.Len(t, validator.Warnings(), 2)
	assert.Len(t, validator.Errors(), 0)

	assert.EqualError(t, validator.Warnings()[0], "server: endpoints: option 'enable_expvars' should not be enabled in production")
	assert.EqualError(t, validator.Warnings()[1], "server: endpoints: option 'enable_pprof' should not be enabled in production")
}

func TestServerAuthzEndpointErrors(t *testing.T) {
	testCases := []struct {
		name string
		have map[string]schema.ServerEndpointsAuthz
		errs []string
	}{
		{"ShouldAllowDefaultEndpoints", schema.DefaultServerConfiguration.Endpoints.Authz, nil},
		{"ShouldAllowSetDefaultEndpoints", nil, nil},
		{
			"ShouldErrorOnInvalidEndpointImplementations",
			map[string]schema.ServerEndpointsAuthz{
				"example": {Implementation: "zero"},
			},
			[]string{
				"server: endpoints: authz: example: option 'implementation' must be one of 'AuthRequest', 'ForwardAuth', 'ExtAuthz', or 'Legacy' but it's configured as 'zero'",
			},
		},
		{
			"ShouldErrorOnInvalidEndpointImplementationLegacy",
			map[string]schema.ServerEndpointsAuthz{
				"legacy": {Implementation: "zero"},
			},
			[]string{
				"server: endpoints: authz: legacy: option 'implementation' must be one of 'AuthRequest', 'ForwardAuth', 'ExtAuthz', or 'Legacy' but it's configured as 'zero'",
			},
		},
		{
			"ShouldErrorOnInvalidEndpointLegacyImplementation",
			map[string]schema.ServerEndpointsAuthz{
				"legacy": {Implementation: "ExtAuthz"},
			},
			[]string{"server: endpoints: authz: legacy: option 'implementation' is invalid: the endpoint with the name 'legacy' must use the 'Legacy' implementation"},
		},
		{
			"ShouldErrorOnInvalidAuthnStrategies",
			map[string]schema.ServerEndpointsAuthz{
				"example": {Implementation: "ExtAuthz", AuthnStrategies: []schema.ServerEndpointsAuthzAuthnStrategy{{Name: "bad-name"}}},
			},
			[]string{
				"server: endpoints: authz: example: authn_strategies: option 'name' must be one of 'CookieSession', 'HeaderAuthorization', 'HeaderProxyAuthorization', 'HeaderAuthRequestProxyAuthorization', or 'HeaderLegacy' but it's configured as 'bad-name'",
			},
		},
		{
			"ShouldErrorOnDuplicateName",
			map[string]schema.ServerEndpointsAuthz{
				"example": {Implementation: "ExtAuthz", AuthnStrategies: []schema.ServerEndpointsAuthzAuthnStrategy{{Name: "CookieSession"}, {Name: "CookieSession"}}},
			},
			[]string{"server: endpoints: authz: example: authn_strategies: duplicate strategy name detected with name 'CookieSession'"},
		},
		{
			"ShouldErrorOnSchemesForInvalidStrategy",
			map[string]schema.ServerEndpointsAuthz{
				"example": {Implementation: "ForwardAuth", AuthnStrategies: []schema.ServerEndpointsAuthzAuthnStrategy{{Name: "CookieSession", Schemes: []string{"basic"}}}},
			},
			[]string{"server: endpoints: authz: example: authn_strategies: strategy #1 (CookieSession): option 'schemes' is not valid for the strategy"},
		},
		{
			"ShouldNotErrorOnSchemeCase",
			map[string]schema.ServerEndpointsAuthz{
				"example": {Implementation: "ForwardAuth", AuthnStrategies: []schema.ServerEndpointsAuthzAuthnStrategy{{Name: "HeaderAuthorization", Schemes: []string{"basIc"}}}},
			},
			nil,
		},
		{
			"ShouldErrorOnInvalidStrategySchemesAndUnnamedStrategy",
			map[string]schema.ServerEndpointsAuthz{
				"example": {Implementation: "ForwardAuth", AuthnStrategies: []schema.ServerEndpointsAuthzAuthnStrategy{{Name: "HeaderAuthorization", Schemes: []string{"basic", "bearer", "abc"}}}},
			},
			[]string{
				"server: endpoints: authz: example: authn_strategies: strategy #1 (HeaderAuthorization): option 'schemes' must only include the values 'basic' or 'bearer' but has 'abc'",
			},
		},
		{
			"ShouldErrorOnUnnamedStrategy",
			map[string]schema.ServerEndpointsAuthz{
				"example": {Implementation: "ForwardAuth", AuthnStrategies: []schema.ServerEndpointsAuthzAuthnStrategy{{Name: "", Schemes: []string{"basic", "bearer", "abc"}}}},
			},
			[]string{
				"server: endpoints: authz: example: authn_strategies: strategy #1: option 'name' must be configured",
			},
		},
		{
			"ShouldErrorOnInvalidSchemeOption",
			map[string]schema.ServerEndpointsAuthz{
				"example": {Implementation: "ForwardAuth", AuthnStrategies: []schema.ServerEndpointsAuthzAuthnStrategy{{Name: "HeaderAuthorization", SchemeBasicCacheLifespan: time.Minute, Schemes: []string{"bearer"}}}},
			},
			[]string{
				"server: endpoints: authz: example: authn_strategies: strategy #1: option 'scheme_basic_cache_lifespan' can't be configured unless the 'basic' scheme is configured but only the 'bearer' schemes are configured",
			},
		},
		{
			"ShouldNotErrorOnValidSchemeCase",
			map[string]schema.ServerEndpointsAuthz{
				"example": {Implementation: "ForwardAuth", AuthnStrategies: []schema.ServerEndpointsAuthzAuthnStrategy{{Name: "HeaderAuthorization", SchemeBasicCacheLifespan: time.Minute, Schemes: []string{"BASIC"}}}},
			},
			[]string{},
		},
		{
			"ShouldErrorOnInvalidChars",
			map[string]schema.ServerEndpointsAuthz{
				"/abc":  {Implementation: "ForwardAuth"},
				"/abc/": {Implementation: "ForwardAuth"},
				"abc/":  {Implementation: "ForwardAuth"},
				"1abc":  {Implementation: "ForwardAuth"},
				"1abc1": {Implementation: "ForwardAuth"},
				"abc1":  {Implementation: "ForwardAuth"},
				"-abc":  {Implementation: "ForwardAuth"},
				"-abc-": {Implementation: "ForwardAuth"},
				"abc-":  {Implementation: "ForwardAuth"},
			},
			[]string{
				"server: endpoints: authz: -abc: contains invalid characters",
				"server: endpoints: authz: -abc-: contains invalid characters",
				"server: endpoints: authz: /abc: contains invalid characters",
				"server: endpoints: authz: /abc/: contains invalid characters",
				"server: endpoints: authz: 1abc: contains invalid characters",
				"server: endpoints: authz: 1abc1: contains invalid characters",
				"server: endpoints: authz: abc-: contains invalid characters",
				"server: endpoints: authz: abc/: contains invalid characters",
				"server: endpoints: authz: abc1: contains invalid characters",
			},
		},
		{
			"ShouldErrorOnEndpointsWithDuplicatePrefix",
			map[string]schema.ServerEndpointsAuthz{
				"apple":         {Implementation: "ForwardAuth"},
				"apple/abc":     {Implementation: "ForwardAuth"},
				"pear/abc":      {Implementation: "ExtAuthz"},
				"pear":          {Implementation: "ExtAuthz"},
				"another":       {Implementation: "ExtAuthz"},
				"another/test":  {Implementation: "ForwardAuth"},
				"anotherb/test": {Implementation: "ForwardAuth"},
				"anothe":        {Implementation: "ExtAuthz"},
				"anotherc/test": {Implementation: "ForwardAuth"},
				"anotherc":      {Implementation: "ExtAuthz"},
				"anotherd/test": {Implementation: "ForwardAuth"},
				"anotherd":      {Implementation: "Legacy"},
				"anothere/test": {Implementation: "ExtAuthz"},
				"anothere":      {Implementation: "ExtAuthz"},
			},
			[]string{
				"server: endpoints: authz: another/test: endpoint starts with the same prefix as the 'another' endpoint with the 'ExtAuthz' implementation which accepts prefixes as part of its implementation",
				"server: endpoints: authz: anotherc/test: endpoint starts with the same prefix as the 'anotherc' endpoint with the 'ExtAuthz' implementation which accepts prefixes as part of its implementation",
				"server: endpoints: authz: anotherd/test: endpoint starts with the same prefix as the 'anotherd' endpoint with the 'Legacy' implementation which accepts prefixes as part of its implementation",
				"server: endpoints: authz: anothere/test: endpoint starts with the same prefix as the 'anothere' endpoint with the 'ExtAuthz' implementation which accepts prefixes as part of its implementation",
				"server: endpoints: authz: pear/abc: endpoint starts with the same prefix as the 'pear' endpoint with the 'ExtAuthz' implementation which accepts prefixes as part of its implementation",
			},
		},
	}

	validator := schema.NewStructValidator()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator.Clear()

			config := newDefaultConfig()

			config.Server.Endpoints.Authz = tc.have

			ValidateServerEndpoints(&config, validator)

			if tc.errs == nil {
				assert.Len(t, validator.Warnings(), 0)
				assert.Len(t, validator.Errors(), 0)
			} else {
				require.Len(t, validator.Errors(), len(tc.errs))

				for i, expected := range tc.errs {
					assert.EqualError(t, validator.Errors()[i], expected)
				}
			}
		})
	}
}

func TestServerAuthzEndpointDefaults(t *testing.T) {
	testCases := []struct {
		name     string
		have     map[string]schema.ServerEndpointsAuthz
		expected map[string]schema.ServerEndpointsAuthz
	}{
		{
			"ShouldSetDefaultSchemes",
			map[string]schema.ServerEndpointsAuthz{
				"example": {
					Implementation: "ForwardAuth",
					AuthnStrategies: []schema.ServerEndpointsAuthzAuthnStrategy{
						{
							Name:    "HeaderAuthorization",
							Schemes: []string{},
						},
					}},
			},
			map[string]schema.ServerEndpointsAuthz{
				"example": {
					Implementation: "ForwardAuth",
					AuthnStrategies: []schema.ServerEndpointsAuthzAuthnStrategy{
						{
							Name:    "HeaderAuthorization",
							Schemes: []string{"basic"},
						},
					}},
			},
		},
		{
			"ShouldSetDefaultStrategies",
			map[string]schema.ServerEndpointsAuthz{
				"example": {
					Implementation: "ForwardAuth",
				},
			},
			map[string]schema.ServerEndpointsAuthz{
				"example": {
					Implementation: "ForwardAuth",
					AuthnStrategies: []schema.ServerEndpointsAuthzAuthnStrategy{
						{
							Name:    "HeaderAuthorization",
							Schemes: []string{"basic"},
						},
						{
							Name: "CookieSession",
						},
					}},
			},
		},
	}

	validator := schema.NewStructValidator()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator.Clear()

			config := newDefaultConfig()

			config.Server.Endpoints.Authz = tc.have

			ValidateServerEndpoints(&config, validator)

			assert.Len(t, validator.Warnings(), 0)
			assert.Len(t, validator.Errors(), 0)

			assert.Equal(t, tc.expected, config.Server.Endpoints.Authz)
		})
	}
}

func TestServerAuthzEndpointLegacyAsImplementationLegacyWhenBlank(t *testing.T) {
	have := map[string]schema.ServerEndpointsAuthz{
		"legacy": {},
	}

	config := newDefaultConfig()

	config.Server.Endpoints.Authz = have

	validator := schema.NewStructValidator()

	ValidateServerEndpoints(&config, validator)

	assert.Len(t, validator.Warnings(), 0)
	assert.Len(t, validator.Errors(), 0)

	assert.Equal(t, schema.AuthzImplementationLegacy, config.Server.Endpoints.Authz[legacy].Implementation)
}

func TestValidateTLSPathStatInvalidArgument(t *testing.T) {
	validator := schema.NewStructValidator()

	validateServerTLSFileExists("key", string([]byte{0x0, 0x1}), validator)

	require.Len(t, validator.Errors(), 1)

	assert.EqualError(t, validator.Errors()[0], "server: tls: option 'key' with path '\x00\x01' could not be verified due to a file system error: stat \x00\x01: invalid argument")
}

func TestValidateTLSPathIsDir(t *testing.T) {
	dir := t.TempDir()

	validator := schema.NewStructValidator()

	validateServerTLSFileExists("key", dir, validator)

	require.Len(t, validator.Errors(), 1)

	assert.EqualError(t, validator.Errors()[0], fmt.Sprintf("server: tls: option 'key' with path '%s' refers to a directory but it should refer to a file", dir))
}

func TestValidateServerAssets(t *testing.T) {
	testCases := []struct {
		name   string
		have   string
		setup  func(t *testing.T, have string) string
		errors []any
	}{
		{
			name: "ShouldValidateEmbedded",
			have: "../../server",
		},
		{
			name: "ShouldValidateEmptyPath",
			have: "",
		},
		{
			name: "ShouldValidateNoPath",
			have: "../../nopath",
			errors: []any{
				"server: asset_path: error occurred reading the '../../nopath' directory: the directory does not exist",
			},
		},
		{
			name: "ShouldValidateValueExcludePlaceholder",
			have: "../test_resources/i18n/example1",
			errors: []any{
				"server: asset_path: error occurred decoding the '../test_resources/i18n/example1/locales/en/portal.json' file: translation key 'Powered by {{authelia}}' has a value which is missing a required placeholder",
			},
		},
		{
			name: "ShouldValidateValueExcludePlaceholder",
			have: "../test_resources/i18n/example2",
			errors: []any{
				"server: asset_path: error occurred decoding the '../test_resources/i18n/example2/locales/en/portal.json' file: translation key 'Powered by {{authelia}}' has a value which is not the required type",
			},
		},
		{
			name: "ShouldErrorReadDirectory",
			setup: func(t *testing.T, have string) (out string) {
				out = t.TempDir()

				require.NoError(t, os.Mkdir(filepath.Join(out, "locales"), 0000))

				return out
			},
			errors: []any{
				regexp.MustCompile(`server: asset_path: error occurred reading the '[\w#/\\]+/locales' directory: open [\w#/\\]+locales: permission denied`),
			},
		},
		{
			name: "ShouldErrorReadLocaleDirectory",
			setup: func(t *testing.T, have string) (out string) {
				out = t.TempDir()

				require.NoError(t, os.Mkdir(filepath.Join(out, "locales"), 0777))
				require.NoError(t, os.Mkdir(filepath.Join(out, "locales", "en"), 0000))

				return out
			},
			errors: []any{
				regexp.MustCompile(`server: asset_path: error occurred reading the '[\w#/\\]+/locales/en' directory: open [\w#/\\]+locales/en: permission denied`),
			},
		},
		{
			name: "ShouldErrorReadNamespaceFile",
			setup: func(t *testing.T, have string) (out string) {
				out = t.TempDir()

				require.NoError(t, os.Mkdir(filepath.Join(out, "locales"), 0777))
				require.NoError(t, os.Mkdir(filepath.Join(out, "locales", "en"), 0777))

				f, err := os.Create(filepath.Join(out, "locales", "en", "portal.json"))
				require.NoError(t, err)

				require.NoError(t, f.Chmod(0000))

				require.NoError(t, f.Close())

				return out
			},
			errors: []any{
				regexp.MustCompile(`server: asset_path: error occurred reading the '[\w#/\\]+/locales/en/portal.json' file: open [\w#/\\]+locales/en/portal.json: permission denied`),
			},
		},
		{
			name: "ShouldErrorDecodeJSON",
			setup: func(t *testing.T, have string) (out string) {
				out = t.TempDir()

				require.NoError(t, os.Mkdir(filepath.Join(out, "locales"), 0777))
				require.NoError(t, os.Mkdir(filepath.Join(out, "locales", "en"), 0777))
				require.NoError(t, os.Mkdir(filepath.Join(out, "locales", "en", "notafile"), 0777))

				f, err := os.Create(filepath.Join(out, "locales", "notadir"))
				require.NoError(t, err)
				require.NoError(t, f.Close())

				f, err = os.Create(filepath.Join(out, "locales", "en", "x.notjson"))
				require.NoError(t, err)
				require.NoError(t, f.Close())

				f, err = os.Create(filepath.Join(out, "locales", "en", "portal.json"))
				require.NoError(t, err)

				_, err = f.Write([]byte("not json"))
				require.NoError(t, err)

				require.NoError(t, f.Close())

				return out
			},
			errors: []any{
				regexp.MustCompile(`server: asset_path: error occurred decoding the '[\w#/\\]+/locales/en/portal.json' file: invalid character 'o' in literal null \(expecting 'u'\)`),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			have := tc.have

			if tc.setup != nil {
				have = tc.setup(t, have)
			}

			config := &schema.Configuration{Server: schema.Server{AssetPath: have}}

			validator := schema.NewStructValidator()

			validateServerAssets(config, validator)

			warnings := validator.Warnings()
			errors := validator.Errors()

			assert.Len(t, warnings, 0)
			require.Len(t, errors, len(tc.errors))

			for i, expected := range tc.errors {
				switch e := expected.(type) {
				case *regexp.Regexp:
					assert.Regexp(t, e, errors[i])
				case string:
					assert.EqualError(t, errors[i], e)
				default:
					t.Fatal("Expected regex or string for error type")
				}
			}
		})
	}
}
