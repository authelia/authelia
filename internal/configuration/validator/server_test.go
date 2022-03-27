package validator

import (
	"io/ioutil"
	"os"
	"testing"

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
	assert.Equal(t, schema.DefaultServerConfiguration.ReadBufferSize, config.Server.ReadBufferSize)
	assert.Equal(t, schema.DefaultServerConfiguration.WriteBufferSize, config.Server.WriteBufferSize)
	assert.Equal(t, schema.DefaultServerConfiguration.TLS.Key, config.Server.TLS.Key)
	assert.Equal(t, schema.DefaultServerConfiguration.TLS.Certificate, config.Server.TLS.Certificate)
	assert.Equal(t, schema.DefaultServerConfiguration.Path, config.Server.Path)
	assert.Equal(t, schema.DefaultServerConfiguration.EnableExpvars, config.Server.EnableExpvars)
	assert.Equal(t, schema.DefaultServerConfiguration.EnablePprof, config.Server.EnablePprof)
}

func TestShouldSetDefaultConfig(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.Configuration{}

	ValidateServer(config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Len(t, validator.Warnings(), 0)

	assert.Equal(t, schema.DefaultServerConfiguration.ReadBufferSize, config.Server.ReadBufferSize)
	assert.Equal(t, schema.DefaultServerConfiguration.WriteBufferSize, config.Server.WriteBufferSize)
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

func TestShouldRaiseOnNegativeValues(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.Configuration{
		Server: schema.ServerConfiguration{
			ReadBufferSize:  -1,
			WriteBufferSize: -1,
		},
	}

	ValidateServer(config, validator)

	require.Len(t, validator.Errors(), 2)

	assert.EqualError(t, validator.Errors()[0], "server: option 'read_buffer_size' must be above 0 but it is configured as '-1'")
	assert.EqualError(t, validator.Errors()[1], "server: option 'write_buffer_size' must be above 0 but it is configured as '-1'")
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

	file, err := ioutil.TempFile("", "cert")
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

	file, err := ioutil.TempFile("", "key")
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

	file, err := ioutil.TempFile("", "key")
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

	file, err := ioutil.TempFile("", "key")
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

	certFile, err := ioutil.TempFile("", "cert")
	require.NoError(t, err)

	defer os.Remove(certFile.Name())

	keyFile, err := ioutil.TempFile("", "key")
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

	certFile, err := ioutil.TempFile("", "cert")
	require.NoError(t, err)

	defer os.Remove(certFile.Name())

	keyFile, err := ioutil.TempFile("", "key")
	require.NoError(t, err)

	defer os.Remove(keyFile.Name())

	config.Server.TLS.Certificate = certFile.Name()
	config.Server.TLS.Key = keyFile.Name()
	config.Server.TLS.ClientCertificates = []string{"/tmp/unexisting"}

	ValidateServer(&config, validator)
	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "server: tls: client_auth: certificates: file path /tmp/unexisting does not exist")
}

func TestShouldRaiseErrorWhenTLSClientAuthIsDefinedButNotServerCertificate(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()

	certFile, err := ioutil.TempFile("", "cert")
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
