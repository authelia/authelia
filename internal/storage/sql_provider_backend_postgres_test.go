package storage

import (
	"crypto/tls"
	"crypto/x509"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestNewPostgreSQLProvider(t *testing.T) {
	address, err := schema.NewAddress("tcp://localhost:5432")
	require.NoError(t, err)

	testCases := []struct {
		name string
		have *schema.Configuration
	}{
		{
			"ShouldHandleSimple",
			&schema.Configuration{
				Storage: schema.Storage{
					PostgreSQL: &schema.StoragePostgreSQL{
						StorageSQL: schema.StorageSQL{
							Address: &schema.AddressTCP{Address: *address},
						},
					},
				},
			},
		},
		{
			"ShouldHandleTLS",
			&schema.Configuration{
				Storage: schema.Storage{
					PostgreSQL: &schema.StoragePostgreSQL{
						StorageSQL: schema.StorageSQL{
							Address: &schema.AddressTCP{Address: *address},
						},
						TLS: &schema.TLS{
							MinimumVersion: schema.TLSVersion{Value: tls.VersionTLS12},
							MaximumVersion: schema.TLSVersion{Value: tls.VersionTLS13},
						},
					},
				},
			},
		},
		{
			"ShouldHandleLegacyTLSVerifyFull",
			&schema.Configuration{
				Storage: schema.Storage{
					PostgreSQL: &schema.StoragePostgreSQL{
						StorageSQL: schema.StorageSQL{
							Address: &schema.AddressTCP{Address: *address},
						},
						SSL: &schema.StoragePostgreSQLSSL{
							Mode: "verify-full",
						},
					},
				},
			},
		},
		{
			"ShouldHandleLegacyTLSVerifyCA",
			&schema.Configuration{
				Storage: schema.Storage{
					PostgreSQL: &schema.StoragePostgreSQL{
						StorageSQL: schema.StorageSQL{
							Address: &schema.AddressTCP{Address: *address},
						},
						SSL: &schema.StoragePostgreSQLSSL{
							Mode: "verify-ca",
						},
					},
				},
			},
		},
		{
			"ShouldHandleLegacyTLSRequire",
			&schema.Configuration{
				Storage: schema.Storage{
					PostgreSQL: &schema.StoragePostgreSQL{
						StorageSQL: schema.StorageSQL{
							Address: &schema.AddressTCP{Address: *address},
						},
						SSL: &schema.StoragePostgreSQLSSL{
							Mode: "require",
						},
					},
				},
			},
		},
		{
			"ShouldHandleLegacyTLSDisabled",
			&schema.Configuration{
				Storage: schema.Storage{
					PostgreSQL: &schema.StoragePostgreSQL{
						StorageSQL: schema.StorageSQL{
							Address: &schema.AddressTCP{Address: *address},
						},
						SSL: &schema.StoragePostgreSQLSSL{
							Mode: "disable",
						},
					},
				},
			},
		},
		{
			"ShouldHandleLegacyTLSVerifyCARootCA",
			&schema.Configuration{
				Storage: schema.Storage{
					PostgreSQL: &schema.StoragePostgreSQL{
						StorageSQL: schema.StorageSQL{
							Address: &schema.AddressTCP{Address: *address},
						},
						SSL: &schema.StoragePostgreSQLSSL{
							Mode:            "verify-ca",
							RootCertificate: "../configuration/test_resources/crypto/ca.rsa.2048.crt",
						},
					},
				},
			},
		},
		{
			"ShouldHandleLegacyTLSVerifyCAAllCertificates",
			&schema.Configuration{
				Storage: schema.Storage{
					PostgreSQL: &schema.StoragePostgreSQL{
						StorageSQL: schema.StorageSQL{
							Address: &schema.AddressTCP{Address: *address},
						},
						SSL: &schema.StoragePostgreSQLSSL{
							Mode:            "verify-ca",
							RootCertificate: "../configuration/test_resources/crypto/ca.rsa.2048.crt",
							Certificate:     "../configuration/test_resources/crypto/rsa.2048.crt",
							Key:             "../configuration/test_resources/crypto/rsa.2048.pem",
						},
					},
				},
			},
		},
		{
			"ShouldHandleLegacyTLSVerifyCAAllCertificatesFailReadFileCA",
			&schema.Configuration{
				Storage: schema.Storage{
					PostgreSQL: &schema.StoragePostgreSQL{
						StorageSQL: schema.StorageSQL{
							Address: &schema.AddressTCP{Address: *address},
						},
						SSL: &schema.StoragePostgreSQLSSL{
							Mode:            "verify-ca",
							RootCertificate: "../configuration/test_resources/crypto/ca.rsa.2048.cert",
							Certificate:     "../configuration/test_resources/crypto/rsa.2048.crt",
							Key:             "../configuration/test_resources/crypto/rsa.2048.pem",
						},
					},
				},
			},
		},
		{
			"ShouldHandleLegacyTLSVerifyCAAllCertificatesFailReadFileKey",
			&schema.Configuration{
				Storage: schema.Storage{
					PostgreSQL: &schema.StoragePostgreSQL{
						StorageSQL: schema.StorageSQL{
							Address: &schema.AddressTCP{Address: *address},
						},
						SSL: &schema.StoragePostgreSQLSSL{
							Mode:            "verify-ca",
							RootCertificate: "../configuration/test_resources/crypto/ca.rsa.2048.crt",
							Certificate:     "../configuration/test_resources/crypto/rsa.2048.crt",
							Key:             "../configuration/test_resources/crypto/rsa.2048.key",
						},
					},
				},
			},
		},
		{
			"ShouldHandleLegacyTLSVerifyCAAllCertificatesFailReadFileCertificate",
			&schema.Configuration{
				Storage: schema.Storage{
					PostgreSQL: &schema.StoragePostgreSQL{
						StorageSQL: schema.StorageSQL{
							Address: &schema.AddressTCP{Address: *address},
						},
						SSL: &schema.StoragePostgreSQLSSL{
							Mode:            "verify-ca",
							RootCertificate: "../configuration/test_resources/crypto/ca.rsa.2048.crt",
							Certificate:     "../configuration/test_resources/crypto/rsa.2048.cert",
							Key:             "../configuration/test_resources/crypto/rsa.2048.pem",
						},
					},
				},
			},
		},
		{
			"ShouldHandleLegacyTLSVerifyCAAllCertificatesFailPair",
			&schema.Configuration{
				Storage: schema.Storage{
					PostgreSQL: &schema.StoragePostgreSQL{
						StorageSQL: schema.StorageSQL{
							Address: &schema.AddressTCP{Address: *address},
						},
						SSL: &schema.StoragePostgreSQLSSL{
							Mode:            "verify-ca",
							RootCertificate: "../configuration/test_resources/crypto/ca.rsa.2048.crt",
							Certificate:     "../configuration/test_resources/crypto/rsa.2048.crt",
							Key:             "../configuration/test_resources/crypto/rsa.4096.pem",
						},
					},
				},
			},
		},
		{
			"ShouldHandleLegacyTLSVerifyCAAllCertificatesFailReadCACertificateFromPrivateKey",
			&schema.Configuration{
				Storage: schema.Storage{
					PostgreSQL: &schema.StoragePostgreSQL{
						StorageSQL: schema.StorageSQL{
							Address: &schema.AddressTCP{Address: *address},
						},
						SSL: &schema.StoragePostgreSQLSSL{
							Mode:            "verify-ca",
							RootCertificate: "../configuration/test_resources/crypto/ca.rsa.2048.pem",
							Certificate:     "../configuration/test_resources/crypto/rsa.2048.crt",
							Key:             "../configuration/test_resources/crypto/rsa.2048.pem",
						},
					},
				},
			},
		},
	}

	t.Parallel()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := NewPostgreSQLProvider(tc.have, x509.NewCertPool())

			assert.NotNil(t, provider)
		})
	}
}

func TestDSNConfigPostgreSQLHostPort(t *testing.T) {
	testCases := []struct {
		name      string
		have      string
		hexpected string
		pexpected uint16
	}{
		{
			"ShouldParseDirectoryDefaultPort",
			"unix:///tmp",
			"/tmp",
			5432,
		},
		{
			"ShouldParseURLPort",
			"unix://:255/tmp",
			"/tmp",
			255,
		},
		{
			"ShouldParseAbsolutePort",
			"unix:///tmp/.s.PGSQL.25432",
			"/tmp",
			25432,
		},
		{
			"ShouldParseAbsolutePortWithURLPort",
			"unix://:2455/tmp/.s.PGSQL.25432",
			"/tmp",
			25432,
		},
		{
			"ShouldParseAbsolutePortInvalidWithURLPort",
			"unix://:2455/tmp/.s.PGSQL.233335432",
			"/tmp/.s.PGSQL.233335432",
			2455,
		},
		{
			"ShouldParseAbsolutePortInvalid",
			"unix:///tmp/.s.PGSQL.233335432",
			"/tmp/.s.PGSQL.233335432",
			5432,
		},
	}

	t.Parallel()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			address, err := schema.NewAddress(tc.have)
			require.NotNil(t, address)
			require.NoError(t, err)

			host, port := dsnPostgreSQLHostPort(&schema.AddressTCP{Address: *address})
			assert.Equal(t, tc.hexpected, host)
			assert.Equal(t, tc.pexpected, port)
		})
	}
}
