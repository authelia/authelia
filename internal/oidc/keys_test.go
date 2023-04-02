// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"crypto"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestKeyManager_AddActiveJWK(t *testing.T) {
	manager := NewKeyManager()
	assert.Nil(t, manager.jwk)
	assert.Nil(t, manager.Strategy())

	j, err := manager.AddActiveJWK(schema.X509CertificateChain{}, mustParseRSAPrivateKey(exampleIssuerPrivateKey))
	require.NoError(t, err)
	require.NotNil(t, j)
	require.NotNil(t, manager.jwk)
	require.NotNil(t, manager.Strategy())

	thumbprint, err := j.JSONWebKey().Thumbprint(crypto.SHA1)
	assert.NoError(t, err)

	kid := strings.ToLower(fmt.Sprintf("%x", thumbprint)[:6])
	assert.Equal(t, manager.jwk.id, kid)
	assert.Equal(t, kid, j.JSONWebKey().KeyID)
	assert.Len(t, manager.jwks.Keys, 1)

	keys := manager.jwks.Key(kid)
	assert.Equal(t, keys[0].KeyID, kid)

	privKey, err := manager.GetActivePrivateKey()
	assert.NoError(t, err)
	assert.NotNil(t, privKey)

	webKey, err := manager.GetActiveJWK()
	assert.NoError(t, err)
	assert.NotNil(t, webKey)

	keySet := manager.GetKeySet()
	assert.NotNil(t, keySet)
	assert.Equal(t, kid, manager.GetActiveKeyID())
}
