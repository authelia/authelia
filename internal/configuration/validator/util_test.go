package validator

import (
	"crypto/elliptic"
	"crypto/rsa"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestIsCookieDomainValid(t *testing.T) {
	testCases := []struct {
		domain   string
		expected bool
	}{
		{"example.com", false},
		{".example.com", false},
		{"*.example.com", false},
		{"authelia.com", false},
		{"duckdns.org", true},
		{".duckdns.org", true},
		{"example.duckdns.org", false},
		{"192.168.2.1", false},
		{"localhost", true},
		{"com", true},
		{"randomnada", true},
	}

	for _, tc := range testCases {
		name := "ShouldFail"

		if tc.expected {
			name = "ShouldPass"
		}

		t.Run(tc.domain, func(t *testing.T) {
			t.Run(name, func(t *testing.T) {
				assert.Equal(t, tc.expected, isCookieDomainAPublicSuffix(tc.domain))
			})
		})
	}
}

func TestBuildStringFuncsMissingTests(t *testing.T) {
	assert.Equal(t, "", buildJoinedString(".", ":", "'", nil))
	assert.Equal(t, "'abc', '123'", strJoinComma("", []string{"abc", "123"}))
}

func TestSchemaJWKGetPropertiesMissingTests(t *testing.T) {
	props, err := schemaJWKGetProperties(schema.JWK{Key: keyECDSAP224})

	assert.NoError(t, err)
	assert.Equal(t, oidc.KeyUseSignature, props.Use)
	assert.Equal(t, "", props.Algorithm)
	assert.Equal(t, elliptic.P224(), props.Curve)
	assert.Equal(t, -1, props.Bits)

	props, err = schemaJWKGetProperties(schema.JWK{Key: keyECDSAP224.Public()})

	assert.NoError(t, err)
	assert.Equal(t, oidc.KeyUseSignature, props.Use)
	assert.Equal(t, "", props.Algorithm)
	assert.Equal(t, elliptic.P224(), props.Curve)
	assert.Equal(t, -1, props.Bits)

	rsa := &rsa.PrivateKey{}

	*rsa = *keyRSA2048
	rsa.PublicKey.N = nil

	props, err = schemaJWKGetProperties(schema.JWK{Key: rsa})

	assert.NoError(t, err)
	assert.Equal(t, oidc.KeyUseSignature, props.Use)
	assert.Equal(t, oidc.SigningAlgRSAUsingSHA256, props.Algorithm)
	assert.Equal(t, nil, props.Curve)
	assert.Equal(t, 0, props.Bits)
}

func TestGetResponseObjectAlgFromKID(t *testing.T) {
	c := &schema.IdentityProvidersOpenIDConnect{
		JSONWebKeys: []schema.JWK{
			{KeyID: "abc", Algorithm: "EX256"},
			{KeyID: "123", Algorithm: "EX512"},
		},
	}

	assert.Equal(t, "EX256", getResponseObjectAlgFromKID(c, "abc", "not"))
	assert.Equal(t, "EX512", getResponseObjectAlgFromKID(c, "123", "not"))
	assert.Equal(t, "not", getResponseObjectAlgFromKID(c, "111111", "not"))
}
