package validator

import (
	"os"
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

	assert.Equal(t, schema.DefaultServerConfiguration.Host, config.Server.Host)
	assert.Equal(t, schema.DefaultServerConfiguration.Port, config.Server.Port)
	assert.Equal(t, schema.DefaultServerConfiguration.Buffers.Read, config.Server.Buffers.Read)
	assert.Equal(t, schema.DefaultServerConfiguration.Buffers.Write, config.Server.Buffers.Write)
	assert.Equal(t, schema.DefaultServerConfiguration.TLS.Key, config.Server.TLS.Key)
	assert.Equal(t, schema.DefaultServerConfiguration.TLS.Certificate, config.Server.TLS.Certificate)
	assert.Equal(t, schema.DefaultServerConfiguration.Path, config.Server.Path)
	assert.Equal(t, schema.DefaultServerConfiguration.Endpoints.EnableExpvars, config.Server.Endpoints.EnableExpvars)
	assert.Equal(t, schema.DefaultServerConfiguration.Endpoints.EnablePprof, config.Server.Endpoints.EnablePprof)
	assert.Equal(t, schema.DefaultServerConfiguration.Endpoints.Authz, config.Server.Endpoints.Authz)
}

func TestShouldSetDefaultConfig(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.Configuration{}

	ValidateServer(config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Len(t, validator.Warnings(), 0)

	assert.Equal(t, schema.DefaultServerConfiguration.Buffers.Read, config.Server.Buffers.Read)
	assert.Equal(t, schema.DefaultServerConfiguration.Buffers.Write, config.Server.Buffers.Write)
}

func TestShouldParsePathCorrectly(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.Configuration{
		Server: schema.ServerConfiguration{
			Path: "apple",
		},
	}

	ValidateServer(config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Len(t, validator.Warnings(), 0)

	assert.Equal(t, "/apple", config.Server.Path)
}

func TestShouldDefaultOnNegativeValues(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.Configuration{
		Server: schema.ServerConfiguration{
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
		Server: schema.ServerConfiguration{
			Path: "app le",
		},
	}

	ValidateServer(config, validator)

	require.Len(t, validator.Errors(), 1)

	assert.Error(t, validator.Errors()[0], "server path must only be alpha numeric characters")
}

func TestShouldRaiseOnForwardSlashInPath(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.Configuration{
		Server: schema.ServerConfiguration{
			Path: "app/le",
		},
	}

	ValidateServer(config, validator)

	assert.Len(t, validator.Errors(), 1)

	assert.Error(t, validator.Errors()[0], "server path must not contain any forward slashes")
}

func TestShouldValidateAndUpdateHost(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()
	config.Server.Host = ""

	ValidateServer(&config, validator)

	require.Len(t, validator.Errors(), 0)
	assert.Equal(t, "0.0.0.0", config.Server.Host)
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
	assert.EqualError(t, validator.Errors()[0], "server: tls: file path /tmp/unexisting_file provided in 'certificate' does not exist")
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
	assert.EqualError(t, validator.Errors()[0], "server: tls: file path /tmp/unexisting_file provided in 'key' does not exist")
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
	assert.EqualError(t, validator.Errors()[0], "server: tls: client_certificates: certificates: file path /tmp/unexisting does not exist")
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
	assert.Equal(t, 9090, config.Server.Port)
	assert.Equal(t, loopback, config.Server.Host)
}

func TestShouldValidateAndUpdatePort(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()
	config.Server.Port = 0

	ValidateServer(&config, validator)

	require.Len(t, validator.Errors(), 0)
	assert.Equal(t, 9091, config.Server.Port)
}

func TestServerEndpointsDevelShouldWarn(t *testing.T) {
	config := &schema.Configuration{
		Server: schema.ServerConfiguration{
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
		have map[string]schema.ServerAuthzEndpoint
		errs []string
	}{
		{"ShouldAllowDefaultEndpoints", schema.DefaultServerConfiguration.Endpoints.Authz, nil},
		{"ShouldAllowSetDefaultEndpoints", nil, nil},
		{
			"ShouldErrorOnInvalidEndpointImplementations",
			map[string]schema.ServerAuthzEndpoint{
				"example": {Implementation: "zero"},
			},
			[]string{"server: endpoints: authz: example: option 'implementation' must be one of 'AuthRequest', 'ForwardAuth', 'ExtAuthz', 'Legacy' but is configured as 'zero'"},
		},
		{
			"ShouldErrorOnInvalidEndpointImplementationLegacy",
			map[string]schema.ServerAuthzEndpoint{
				"legacy": {Implementation: "zero"},
			},
			[]string{"server: endpoints: authz: legacy: option 'implementation' must be one of 'AuthRequest', 'ForwardAuth', 'ExtAuthz', 'Legacy' but is configured as 'zero'"},
		},
		{
			"ShouldErrorOnInvalidEndpointLegacyImplementation",
			map[string]schema.ServerAuthzEndpoint{
				"legacy": {Implementation: "ExtAuthz"},
			},
			[]string{"server: endpoints: authz: legacy: option 'implementation' is invalid: the endpoint with the name 'legacy' must use the 'Legacy' implementation"},
		},
		{
			"ShouldErrorOnInvalidAuthnStrategies",
			map[string]schema.ServerAuthzEndpoint{
				"example": {Implementation: "ExtAuthz", AuthnStrategies: []schema.ServerAuthzEndpointAuthnStrategy{{Name: "bad-name"}}},
			},
			[]string{"server: endpoints: authz: example: authn_strategies: option 'name' must be one of 'CookieSession', 'HeaderAuthorization', 'HeaderProxyAuthorization', 'HeaderAuthRequestProxyAuthorization', 'HeaderLegacy' but is configured as 'bad-name'"},
		},
		{
			"ShouldErrorOnDuplicateName",
			map[string]schema.ServerAuthzEndpoint{
				"example": {Implementation: "ExtAuthz", AuthnStrategies: []schema.ServerAuthzEndpointAuthnStrategy{{Name: "CookieSession"}, {Name: "CookieSession"}}},
			},
			[]string{"server: endpoints: authz: example: authn_strategies: duplicate strategy name detected with name 'CookieSession'"},
		},
		{
			"ShouldErrorOnInvalidChars",
			map[string]schema.ServerAuthzEndpoint{
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
			map[string]schema.ServerAuthzEndpoint{
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

func TestServerAuthzEndpointLegacyAsImplementationLegacyWhenBlank(t *testing.T) {
	have := map[string]schema.ServerAuthzEndpoint{
		"legacy": {},
	}

	config := newDefaultConfig()

	config.Server.Endpoints.Authz = have

	validator := schema.NewStructValidator()

	ValidateServerEndpoints(&config, validator)

	assert.Len(t, validator.Warnings(), 0)
	assert.Len(t, validator.Errors(), 0)

	assert.Equal(t, authzImplementationLegacy, config.Server.Endpoints.Authz[legacy].Implementation)
}
