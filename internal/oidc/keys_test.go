package oidc

import (
	"crypto"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyManager_AddActiveKeyData(t *testing.T) {
	manager := NewKeyManager()
	assert.Nil(t, manager.strategy)
	assert.Nil(t, manager.Strategy())

	key, wk, err := manager.AddActivePrivateKeyData(exampleIssuerPrivateKey)
	require.NoError(t, err)
	require.NotNil(t, key)
	require.NotNil(t, wk)
	require.NotNil(t, manager.strategy)
	require.NotNil(t, manager.Strategy())

	thumbprint, err := wk.Thumbprint(crypto.SHA1)
	assert.NoError(t, err)

	kid := strings.ToLower(fmt.Sprintf("%x", thumbprint)[0:6])
	assert.Equal(t, manager.activeKeyID, kid)
	assert.Equal(t, kid, wk.KeyID)
	assert.Len(t, manager.keys, 1)
	assert.Len(t, manager.keySet.Keys, 1)
	assert.Contains(t, manager.keys, kid)

	keys := manager.keySet.Key(kid)
	assert.Equal(t, keys[0].KeyID, kid)

	privKey, err := manager.GetActivePrivateKey()
	assert.NoError(t, err)
	assert.NotNil(t, privKey)

	pubKey, err := manager.GetActiveKey()
	assert.NoError(t, err)
	assert.NotNil(t, pubKey)

	webKey, err := manager.GetActiveWebKey()
	assert.NoError(t, err)
	assert.NotNil(t, webKey)

	keySet := manager.GetKeySet()
	assert.NotNil(t, keySet)
	assert.Equal(t, kid, manager.GetActiveKeyID())
}
