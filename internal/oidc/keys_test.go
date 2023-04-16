package oidc

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/square/go-jose.v2"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

func TestJWTEncoding(t *testing.T) {
	certificate, err := schema.NewX509CertificateChain(ecP256Certificate)

	assert.NoError(t, err)

	require.NotNil(t, certificate)

	key, err := utils.ParseX509FromPEM([]byte(ecP256Key))

	assert.NoError(t, err)

	require.NotNil(t, key)

	have := schema.JWK{
		KeyID:            "example",
		Use:              KeyUseSignature,
		Algorithm:        SigningAlgECDSAUsingP256AndSHA256,
		Key:              key,
		CertificateChain: *certificate,
	}

	out := jose.JSONWebKey{}

	jwk := NewJWK(have)
	data, err := json.Marshal(jwk.JWK())

	assert.NoError(t, err)
	require.NotNil(t, data)
	assert.NoError(t, json.Unmarshal(data, &out))

	assert.True(t, out.IsPublic())
	assert.Equal(t, "example", out.KeyID)
	assert.Equal(t, KeyUseSignature, out.Use)
	assert.Equal(t, SigningAlgECDSAUsingP256AndSHA256, out.Algorithm)
	assert.NotNil(t, out.Key)
	assert.NotNil(t, out.Certificates)
	assert.NotNil(t, out.CertificateThumbprintSHA1)
	assert.NotNil(t, out.CertificateThumbprintSHA256)
	assert.True(t, out.Valid())

	data, err = json.Marshal(jwk.PrivateJWK())

	assert.NoError(t, err)
	require.NotNil(t, data)
	assert.NoError(t, json.Unmarshal(data, &out))

	assert.False(t, out.IsPublic())
	assert.Equal(t, "example", out.KeyID)
	assert.Equal(t, KeyUseSignature, out.Use)
	assert.Equal(t, SigningAlgECDSAUsingP256AndSHA256, out.Algorithm)
	assert.NotNil(t, out.Key)
	assert.NotNil(t, out.Certificates)
	assert.NotNil(t, out.CertificateThumbprintSHA1)
	assert.NotNil(t, out.CertificateThumbprintSHA256)
	assert.True(t, out.Valid())
}

const (
	ecP256Key = `
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIMG0XC0KHg2kEcmHy3+l2h09n1HMmMNO5WOqDtIq9gDzoAoGCCqGSM49
AwEHoUQDQgAE5psG+0pKr9MPAUPh2g1CRCSsV1Ku4HZYpL+NF7pePQQVd5Q4MDnZ
Il+gUSRO4cjae7mJJ4CTK7tyC3Y2cmJ3Jg==
-----END EC PRIVATE KEY-----`

	ecP256Certificate = `
-----BEGIN CERTIFICATE-----
MIIBWDCB/6ADAgECAhBTisKItcCwgi81yzJVPRwgMAoGCCqGSM49BAMCMBMxETAP
BgNVBAoTCEF1dGhlbGlhMB4XDTIzMDQxNzA3MjE1MloXDTI0MDQxNjA3MjE1Mlow
EzERMA8GA1UEChMIQXV0aGVsaWEwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAATm
mwb7Skqv0w8BQ+HaDUJEJKxXUq7gdlikv40Xul49BBV3lDgwOdkiX6BRJE7hyNp7
uYkngJMru3ILdjZyYncmozUwMzAOBgNVHQ8BAf8EBAMCBaAwEwYDVR0lBAwwCgYI
KwYBBQUHAwEwDAYDVR0TAQH/BAIwADAKBggqhkjOPQQDAgNIADBFAiAgLV+In0Q7
s8CgoeuYUnD18Assm7RqrHMOcYw2Kga5AAIhAJbiCgirlNUYVPgr0julBTvHdK00
ygJhFOMN13bY8jwi
-----END CERTIFICATE-----`
)
