package validator

import (
	"fmt"
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

	assert.Equal(t, "", config.Server.Host) //nolint:staticcheck
	assert.Equal(t, 0, config.Server.Port)  //nolint:staticcheck
	assert.Equal(t, schema.DefaultServerConfiguration.Address, config.Server.Address)
	assert.Equal(t, schema.DefaultServerConfiguration.Buffers.Read, config.Server.Buffers.Read)
	assert.Equal(t, schema.DefaultServerConfiguration.Buffers.Write, config.Server.Buffers.Write)
	assert.Equal(t, schema.DefaultServerConfiguration.TLS.Key, config.Server.TLS.Key)
	assert.Equal(t, schema.DefaultServerConfiguration.TLS.Certificate, config.Server.TLS.Certificate)
	assert.Equal(t, schema.DefaultServerConfiguration.Path, config.Server.Path)
	assert.Equal(t, schema.DefaultServerConfiguration.Endpoints.EnableExpvars, config.Server.Endpoints.EnableExpvars)
	assert.Equal(t, schema.DefaultServerConfiguration.Endpoints.EnablePprof, config.Server.Endpoints.EnablePprof)
	assert.Equal(t, schema.DefaultServerConfiguration.Endpoints.Authz, config.Server.Endpoints.Authz)
}

func TestShouldSetDefaultServerValuesWithLegacyAddress(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.Configuration{
		Server: schema.ServerConfiguration{
			Host: "abc",
			Port: 123,
		},
	}

	ValidateServer(config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Len(t, validator.Warnings(), 0)

	assert.Equal(t, "abc", config.Server.Host) //nolint:staticcheck
	assert.Equal(t, 123, config.Server.Port)   //nolint:staticcheck
	assert.Equal(t, &schema.AddressTCP{Address: MustParseAddress("tcp://abc:123")}, config.Server.Address)

	config = &schema.Configuration{
		Server: schema.ServerConfiguration{
			Host: "abc",
		},
	}

	ValidateServer(config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Len(t, validator.Warnings(), 0)

	assert.Equal(t, "abc", config.Server.Host) //nolint:staticcheck
	assert.Equal(t, 0, config.Server.Port)     //nolint:staticcheck
	assert.Equal(t, &schema.AddressTCP{Address: MustParseAddress("tcp://abc:9091")}, config.Server.Address)

	config = &schema.Configuration{
		Server: schema.ServerConfiguration{
			Port: 123,
		},
	}

	ValidateServer(config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Len(t, validator.Warnings(), 0)

	assert.Equal(t, "", config.Server.Host)  //nolint:staticcheck
	assert.Equal(t, 123, config.Server.Port) //nolint:staticcheck
	assert.Equal(t, &schema.AddressTCP{Address: MustParseAddress("tcp://:123")}, config.Server.Address)
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

func TestValidateServerShouldCorrectlyIdentifyValidAddressSchemes(t *testing.T) {
	testCases := []struct {
		have     string
		expected string
	}{
		{schema.AddressSchemeTCP, ""},
		{schema.AddressSchemeTCP4, ""},
		{schema.AddressSchemeTCP6, ""},
		{schema.AddressSchemeUDP, "server: option 'address' with value 'udp://:9091' is invalid: scheme must be one of 'tcp', 'tcp4', 'tcp6', or 'unix' but is configured as 'udp'"},
		{schema.AddressSchemeUDP4, "server: option 'address' with value 'udp4://:9091' is invalid: scheme must be one of 'tcp', 'tcp4', 'tcp6', or 'unix' but is configured as 'udp4'"},
		{schema.AddressSchemeUDP6, "server: option 'address' with value 'udp6://:9091' is invalid: scheme must be one of 'tcp', 'tcp4', 'tcp6', or 'unix' but is configured as 'udp6'"},
		{schema.AddressSchemeUnix, ""},
		{"http", "server: option 'address' with value 'http://:9091' is invalid: scheme must be one of 'tcp', 'tcp4', 'tcp6', or 'unix' but is configured as 'http'"},
	}

	have := &schema.Configuration{
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

	validator := schema.NewStructValidator()

	for _, tc := range testCases {
		t.Run(tc.have, func(t *testing.T) {
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

func TestShouldValidateAndUpdateAddress(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()
	config.Server.Address = nil

	ValidateServer(&config, validator)

	require.Len(t, validator.Errors(), 0)
	assert.Equal(t, "tcp://:9091", config.Server.Address.String())
}

func TestShouldRaiseErrorOnLegacyAndModernValues(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()
	config.Server.Host = local25 //nolint:staticcheck
	config.Server.Port = 9999    //nolint:staticcheck

	ValidateServer(&config, validator)

	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "server: option 'host' and 'port' can't be configured at the same time as 'address'")
}

func TestShouldValidateAndUpdateAddressWithOldValues(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()
	config.Server.Address = nil
	config.Server.Host = local25 //nolint:staticcheck
	config.Server.Port = 9999    //nolint:staticcheck

	ValidateServer(&config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Len(t, validator.Warnings(), 0)

	require.NotNil(t, config.Server.Address)
	assert.Equal(t, "tcp://127.0.0.25:9999", config.Server.Address.String())
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
	assert.Equal(t, "tcp://127.0.0.1:9090", config.Server.Address.String())
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
			[]string{
				"server: endpoints: authz: example: option 'implementation' must be one of 'AuthRequest', 'ForwardAuth', 'ExtAuthz', or 'Legacy' but it's configured as 'zero'",
			},
		},
		{
			"ShouldErrorOnInvalidEndpointImplementationLegacy",
			map[string]schema.ServerAuthzEndpoint{
				"legacy": {Implementation: "zero"},
			},
			[]string{
				"server: endpoints: authz: legacy: option 'implementation' must be one of 'AuthRequest', 'ForwardAuth', 'ExtAuthz', or 'Legacy' but it's configured as 'zero'",
			},
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
			[]string{
				"server: endpoints: authz: example: authn_strategies: option 'name' must be one of 'CookieSession', 'HeaderAuthorization', 'HeaderProxyAuthorization', 'HeaderAuthRequestProxyAuthorization', or 'HeaderLegacy' but it's configured as 'bad-name'",
			},
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

func TestValidateTLSPathStatInvalidArgument(t *testing.T) {
	val := schema.NewStructValidator()

	validateServerTLSFileExists("key", string([]byte{0x0, 0x1}), val)

	require.Len(t, val.Errors(), 1)

	assert.EqualError(t, val.Errors()[0], "server: tls: option 'key' with path '\x00\x01' could not be verified due to a file system error: stat \x00\x01: invalid argument")
}

func TestValidateTLSPathIsDir(t *testing.T) {
	dir := t.TempDir()

	val := schema.NewStructValidator()

	validateServerTLSFileExists("key", dir, val)

	require.Len(t, val.Errors(), 1)

	assert.EqualError(t, val.Errors()[0], fmt.Sprintf("server: tls: option 'key' with path '%s' refers to a directory but it should refer to a file", dir))
}
