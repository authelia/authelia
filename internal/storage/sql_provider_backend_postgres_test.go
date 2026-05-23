package storage

import (
	"crypto/tls"
	"crypto/x509"
	"testing"

	"github.com/jackc/pgx/v5"
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
							TLS: &schema.TLS{
								MinimumVersion: schema.TLSVersion{Value: tls.VersionTLS12},
								MaximumVersion: schema.TLSVersion{Value: tls.VersionTLS13},
							},
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

func TestDSNPostgreSQLFallbacks(t *testing.T) {
	mkAddress := func(t *testing.T, raw string) *schema.AddressTCP {
		t.Helper()

		address, err := schema.NewAddress(raw)
		require.NoError(t, err)

		return &schema.AddressTCP{Address: *address}
	}

	testCases := []struct {
		name    string
		servers []schema.StoragePostgreSQLServer
		assert  func(t *testing.T, dsnConfig *pgx.ConnConfig)
	}{
		{
			name:    "ShouldHandleEmptyServers",
			servers: []schema.StoragePostgreSQLServer{},
			assert: func(t *testing.T, dsnConfig *pgx.ConnConfig) {
				assert.NotNil(t, dsnConfig.Fallbacks)
				assert.Empty(t, dsnConfig.Fallbacks)
			},
		},
		{
			name: "ShouldHandleSingleTCPServer",
			servers: []schema.StoragePostgreSQLServer{
				{Address: mkAddress(t, "tcp://db1.example.com:5432")},
			},
			assert: func(t *testing.T, dsnConfig *pgx.ConnConfig) {
				require.Len(t, dsnConfig.Fallbacks, 1)
				assert.Equal(t, "db1.example.com", dsnConfig.Fallbacks[0].Host)
				assert.Equal(t, uint16(5432), dsnConfig.Fallbacks[0].Port)
				assert.Nil(t, dsnConfig.Fallbacks[0].TLSConfig)
			},
		},
		{
			name: "ShouldHandleMultipleTCPServers",
			servers: []schema.StoragePostgreSQLServer{
				{Address: mkAddress(t, "tcp://db1.example.com:5432")},
				{Address: mkAddress(t, "tcp://db2.example.com:6543")},
				{Address: mkAddress(t, "tcp://db3.example.com:7654")},
			},
			assert: func(t *testing.T, dsnConfig *pgx.ConnConfig) {
				require.Len(t, dsnConfig.Fallbacks, 3)
				assert.Equal(t, "db1.example.com", dsnConfig.Fallbacks[0].Host)
				assert.Equal(t, uint16(5432), dsnConfig.Fallbacks[0].Port)
				assert.Equal(t, "db2.example.com", dsnConfig.Fallbacks[1].Host)
				assert.Equal(t, uint16(6543), dsnConfig.Fallbacks[1].Port)
				assert.Equal(t, "db3.example.com", dsnConfig.Fallbacks[2].Host)
				assert.Equal(t, uint16(7654), dsnConfig.Fallbacks[2].Port)
			},
		},
		{
			name: "ShouldHandleTCPServerWithTLS",
			servers: []schema.StoragePostgreSQLServer{
				{
					Address: mkAddress(t, "tcp://db1.example.com:5432"),
					TLS: &schema.TLS{
						MinimumVersion: schema.TLSVersion{Value: tls.VersionTLS12},
						MaximumVersion: schema.TLSVersion{Value: tls.VersionTLS13},
					},
				},
			},
			assert: func(t *testing.T, dsnConfig *pgx.ConnConfig) {
				require.Len(t, dsnConfig.Fallbacks, 1)
				require.NotNil(t, dsnConfig.Fallbacks[0].TLSConfig)
				assert.Equal(t, uint16(tls.VersionTLS12), dsnConfig.Fallbacks[0].TLSConfig.MinVersion)
				assert.Equal(t, uint16(tls.VersionTLS13), dsnConfig.Fallbacks[0].TLSConfig.MaxVersion)
			},
		},
		{
			name: "ShouldDefaultPortToFiveFourThreeTwoWhenZeroAndTCP",
			servers: []schema.StoragePostgreSQLServer{
				{Address: mkAddress(t, "tcp://db.example.com")},
			},
			assert: func(t *testing.T, dsnConfig *pgx.ConnConfig) {
				require.Len(t, dsnConfig.Fallbacks, 1)
				assert.Equal(t, "db.example.com", dsnConfig.Fallbacks[0].Host)
				assert.Equal(t, uint16(5432), dsnConfig.Fallbacks[0].Port)
			},
		},
		{
			name: "ShouldHandleUnixSocketServer",
			servers: []schema.StoragePostgreSQLServer{
				{Address: mkAddress(t, "unix:///var/run/postgresql")},
			},
			assert: func(t *testing.T, dsnConfig *pgx.ConnConfig) {
				require.Len(t, dsnConfig.Fallbacks, 1)
				assert.Equal(t, "/var/run/postgresql", dsnConfig.Fallbacks[0].Host)
				assert.Equal(t, uint16(5432), dsnConfig.Fallbacks[0].Port)
			},
		},
		{
			name: "ShouldHandleUnixSocketServerWithAbsolutePort",
			servers: []schema.StoragePostgreSQLServer{
				{Address: mkAddress(t, "unix:///tmp/.s.PGSQL.25432")},
			},
			assert: func(t *testing.T, dsnConfig *pgx.ConnConfig) {
				require.Len(t, dsnConfig.Fallbacks, 1)
				assert.Equal(t, "/tmp", dsnConfig.Fallbacks[0].Host)
				assert.Equal(t, uint16(25432), dsnConfig.Fallbacks[0].Port)
			},
		},
		{
			name: "ShouldHandleMixedTCPAndUnixSocketServers",
			servers: []schema.StoragePostgreSQLServer{
				{Address: mkAddress(t, "tcp://db.example.com:6543")},
				{Address: mkAddress(t, "unix:///var/run/postgresql")},
			},
			assert: func(t *testing.T, dsnConfig *pgx.ConnConfig) {
				require.Len(t, dsnConfig.Fallbacks, 2)
				assert.Equal(t, "db.example.com", dsnConfig.Fallbacks[0].Host)
				assert.Equal(t, uint16(6543), dsnConfig.Fallbacks[0].Port)
				assert.Equal(t, "/var/run/postgresql", dsnConfig.Fallbacks[1].Host)
				assert.Equal(t, uint16(5432), dsnConfig.Fallbacks[1].Port)
			},
		},
		{
			name: "ShouldHandleServerWithoutTLS",
			servers: []schema.StoragePostgreSQLServer{
				{Address: mkAddress(t, "tcp://db.example.com:5432")},
			},
			assert: func(t *testing.T, dsnConfig *pgx.ConnConfig) {
				require.Len(t, dsnConfig.Fallbacks, 1)
				assert.Nil(t, dsnConfig.Fallbacks[0].TLSConfig)
			},
		},
		{
			name: "ShouldOverwriteExistingFallbacks",
			servers: []schema.StoragePostgreSQLServer{
				{Address: mkAddress(t, "tcp://db.example.com:5432")},
			},
			assert: func(t *testing.T, dsnConfig *pgx.ConnConfig) {
				require.Len(t, dsnConfig.Fallbacks, 1)
				assert.Equal(t, "db.example.com", dsnConfig.Fallbacks[0].Host)
			},
		},
	}

	t.Parallel()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dsnConfig, err := pgx.ParseConfig("")
			require.NoError(t, err)

			config := &schema.StoragePostgreSQL{
				Servers: tc.servers,
			}

			dsnPostgreSQLFallbacks(config, x509.NewCertPool(), dsnConfig)

			tc.assert(t, dsnConfig)
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
