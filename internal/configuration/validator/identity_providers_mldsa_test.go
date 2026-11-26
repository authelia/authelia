// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

//go:build go1.27

package validator

import (
	"crypto/mldsa"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestValidateOIDCClientJWKSMLDSA(t *testing.T) {
	key, err := mldsa.GenerateKey(mldsa.MLDSA65())
	require.NoError(t, err)

	testCases := []struct {
		name     string
		key      any
		expected func(t *testing.T, config *schema.IdentityProvidersOpenIDConnect)
		errs     []string
	}{
		{
			"ShouldAllowPublicKey",
			key.PublicKey(),
			func(t *testing.T, config *schema.IdentityProvidersOpenIDConnect) {
				assert.Equal(t, oidc.KeyUseSignature, config.Clients[0].JSONWebKeys[0].Use)
				assert.Equal(t, oidc.SigningAlgMLDSA65, config.Clients[0].JSONWebKeys[0].Algorithm)
				assert.Equal(t, []string{oidc.SigningAlgMLDSA65}, config.Clients[0].Discovery.RequestObjectSigningAlgs)
			},
			nil,
		},
		{
			"ShouldFailOnPrivateKey",
			key,
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': jwks: key #1 with key id 'test': option 'key' must be a RSA, ECDSA, Ed25519, or ML-DSA public key but it's type is *mldsa.PrivateKey",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &schema.IdentityProvidersOpenIDConnect{
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:          "test",
						JSONWebKeys: []schema.JWK{{KeyID: "test", Key: tc.key}},
					},
				},
			}

			validator := schema.NewStructValidator()

			validateOIDCClientPublicKeys(0, config, validator)

			if tc.expected != nil {
				tc.expected(t, config)
			}

			assert.Len(t, validator.Warnings(), 0)

			require.Len(t, validator.Errors(), len(tc.errs))

			for i, err := range tc.errs {
				assert.EqualError(t, validator.Errors()[i], err)
			}
		})
	}
}
