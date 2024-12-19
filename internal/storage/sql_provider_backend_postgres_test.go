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
	standardAddress, err := schema.NewAddress("tcp://postgres")
	require.NoError(t, err)

	standardAddressWithPort, err := schema.NewAddress("tcp://postgres:5432")
	require.NoError(t, err)

	testCases := []struct {
		name string
		have *schema.Configuration
	}{
		{
			"ShouldHandleBasic",
			&schema.Configuration{
				Storage: schema.Storage{
					PostgreSQL: &schema.StoragePostgreSQL{
						StorageSQL: schema.StorageSQL{
							Address:  &schema.AddressTCP{Address: *standardAddress},
							Database: "authelia",
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
							Address:  &schema.AddressTCP{Address: *standardAddress},
							Database: "authelia",
							TLS: &schema.TLS{
								MinimumVersion: schema.TLSVersion{Value: tls.VersionTLS12},
								MaximumVersion: schema.TLSVersion{Value: tls.VersionTLS13},
								SkipVerify:     false,
								ServerName:     "postgres",
							},
						},
					},
				},
			},
		},
		{
			"ShouldHandleTLSLegacyDisable",
			&schema.Configuration{
				Storage: schema.Storage{
					PostgreSQL: &schema.StoragePostgreSQL{
						StorageSQL: schema.StorageSQL{
							Address:  &schema.AddressTCP{Address: *standardAddress},
							Database: "authelia",
						},
						SSL: &schema.StoragePostgreSQLSSL{
							Mode: "disable",
						},
					},
				},
			},
		},
		{
			"ShouldHandleTLSLegacyRequireNoRoot",
			&schema.Configuration{
				Storage: schema.Storage{
					PostgreSQL: &schema.StoragePostgreSQL{
						StorageSQL: schema.StorageSQL{
							Address:  &schema.AddressTCP{Address: *standardAddress},
							Database: "authelia",
						},
						SSL: &schema.StoragePostgreSQLSSL{
							Mode: "require",
						},
					},
				},
			},
		},
		{
			"ShouldHandleTLSLegacyRequire",
			&schema.Configuration{
				Storage: schema.Storage{
					PostgreSQL: &schema.StoragePostgreSQL{
						StorageSQL: schema.StorageSQL{
							Address:  &schema.AddressTCP{Address: *standardAddress},
							Database: "authelia",
						},
						SSL: &schema.StoragePostgreSQLSSL{
							Mode:            "require",
							RootCertificate: "../configuration/test_resources/crypto/ca.rsa.2048.crt",
						},
					},
				},
			},
		},
		{
			"ShouldHandleTLSLegacyRequireUsingKeyForCert",
			&schema.Configuration{
				Storage: schema.Storage{
					PostgreSQL: &schema.StoragePostgreSQL{
						StorageSQL: schema.StorageSQL{
							Address:  &schema.AddressTCP{Address: *standardAddress},
							Database: "authelia",
						},
						SSL: &schema.StoragePostgreSQLSSL{
							Mode:            "require",
							RootCertificate: "../configuration/test_resources/crypto/ca.rsa.2048.pem",
						},
					},
				},
			},
		},
		{
			"ShouldHandleTLSLegacyRequireUsingNonPem",
			&schema.Configuration{
				Storage: schema.Storage{
					PostgreSQL: &schema.StoragePostgreSQL{
						StorageSQL: schema.StorageSQL{
							Address:  &schema.AddressTCP{Address: *standardAddress},
							Database: "authelia",
						},
						SSL: &schema.StoragePostgreSQLSSL{
							Mode:            "require",
							RootCertificate: "../configuration/test_resources/crypto/gen.sh",
						},
					},
				},
			},
		},
		{
			"ShouldHandleTLSLegacyRequireBadPath",
			&schema.Configuration{
				Storage: schema.Storage{
					PostgreSQL: &schema.StoragePostgreSQL{
						StorageSQL: schema.StorageSQL{
							Address:  &schema.AddressTCP{Address: *standardAddress},
							Database: "authelia",
						},
						SSL: &schema.StoragePostgreSQLSSL{
							Mode:            "require",
							RootCertificate: "../configuration/test_resources/crypto/ca.rsa.2048.bad",
						},
					},
				},
			},
		},
		{
			"ShouldHandleTLSLegacyCertKey",
			&schema.Configuration{
				Storage: schema.Storage{
					PostgreSQL: &schema.StoragePostgreSQL{
						StorageSQL: schema.StorageSQL{
							Address:  &schema.AddressTCP{Address: *standardAddress},
							Database: "authelia",
						},
						SSL: &schema.StoragePostgreSQLSSL{
							Mode:            "require",
							RootCertificate: "../configuration/test_resources/crypto/ca.rsa.4096.crt",
							Certificate:     "../configuration/test_resources/crypto/rsa.2048.crt",
							Key:             "../configuration/test_resources/crypto/rsa.2048.pem",
						},
					},
				},
			},
		},
		{
			"ShouldHandleTLSLegacyFullCertKey",
			&schema.Configuration{
				Storage: schema.Storage{
					PostgreSQL: &schema.StoragePostgreSQL{
						StorageSQL: schema.StorageSQL{
							Address:  &schema.AddressTCP{Address: *standardAddress},
							Database: "authelia",
						},
						SSL: &schema.StoragePostgreSQLSSL{
							Mode:            "verify-full",
							RootCertificate: "../configuration/test_resources/crypto/ca.rsa.4096.crt",
							Certificate:     "../configuration/test_resources/crypto/rsa.2048.crt",
							Key:             "../configuration/test_resources/crypto/rsa.2048.pem",
						},
					},
				},
			},
		},
		{
			"ShouldHandleTLSLegacyCertKeyBadKey",
			&schema.Configuration{
				Storage: schema.Storage{
					PostgreSQL: &schema.StoragePostgreSQL{
						StorageSQL: schema.StorageSQL{
							Address:  &schema.AddressTCP{Address: *standardAddress},
							Database: "authelia",
						},
						SSL: &schema.StoragePostgreSQLSSL{
							Mode:            "require",
							RootCertificate: "../configuration/test_resources/crypto/ca.rsa.4096.crt",
							Certificate:     "../configuration/test_resources/crypto/rsa.2048.crt",
							Key:             "../configuration/test_resources/crypto/rsa.2048.bad",
						},
					},
				},
			},
		},
		{
			"ShouldHandleTLSLegacyCertKeyBadCert",
			&schema.Configuration{
				Storage: schema.Storage{
					PostgreSQL: &schema.StoragePostgreSQL{
						StorageSQL: schema.StorageSQL{
							Address:  &schema.AddressTCP{Address: *standardAddress},
							Database: "authelia",
						},
						SSL: &schema.StoragePostgreSQLSSL{
							Mode:            "require",
							RootCertificate: "../configuration/test_resources/crypto/ca.rsa.4096.crt",
							Certificate:     "../configuration/test_resources/crypto/rsa.2048.bad",
							Key:             "../configuration/test_resources/crypto/rsa.2048.pem",
						},
					},
				},
			},
		},
		{
			"ShouldHandleTLSLegacyCertMismatched",
			&schema.Configuration{
				Storage: schema.Storage{
					PostgreSQL: &schema.StoragePostgreSQL{
						StorageSQL: schema.StorageSQL{
							Address:  &schema.AddressTCP{Address: *standardAddress},
							Database: "authelia",
						},
						SSL: &schema.StoragePostgreSQLSSL{
							Mode:            "require",
							RootCertificate: "../configuration/test_resources/crypto/ca.rsa.4096.crt",
							Certificate:     "../configuration/test_resources/crypto/rsa.2048.crt",
							Key:             "../configuration/test_resources/crypto/rsa.4096.pem",
						},
					},
				},
			},
		},
		{
			"ShouldHandleFallbacks",
			&schema.Configuration{
				Storage: schema.Storage{
					PostgreSQL: &schema.StoragePostgreSQL{
						StorageSQL: schema.StorageSQL{
							Address:  &schema.AddressTCP{Address: *standardAddress},
							Database: "authelia",
						},
						Servers: []schema.StoragePostgreSQLServer{
							{
								Address: &schema.AddressTCP{Address: *standardAddress},
							},
							{
								Address: &schema.AddressTCP{Address: *standardAddressWithPort},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		pool := x509.NewCertPool()

		t.Run(tc.name, func(t *testing.T) {
			assert.NotNil(t, NewPostgreSQLProvider(tc.have, pool))
		})
	}
}
