package storage

import (
	"crypto/tls"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestNewMySQLProvider(t *testing.T) {
	standardAddress, err := schema.NewAddress("tcp://mysql")
	require.NoError(t, err)

	testCases := []struct {
		name string
		have *schema.Configuration
	}{
		{
			"ShouldHandleBasic",
			&schema.Configuration{
				Storage: schema.Storage{
					MySQL: &schema.StorageMySQL{
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
					MySQL: &schema.StorageMySQL{
						StorageSQL: schema.StorageSQL{
							Address:  &schema.AddressTCP{Address: *standardAddress},
							Database: "authelia",
							TLS: &schema.TLS{
								MinimumVersion: schema.TLSVersion{Value: tls.VersionTLS12},
								MaximumVersion: schema.TLSVersion{Value: tls.VersionTLS13},
								SkipVerify:     false,
								ServerName:     "mysql",
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.NotNil(t, NewMySQLProvider(tc.have, nil))
		})
	}
}
