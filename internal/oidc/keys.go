package oidc

import (
	"context"
	"crypto"
	"crypto/rsa"
	"errors"
	"fmt"
	"strings"

	"github.com/ory/fosite/token/jwt"
	"gopkg.in/square/go-jose.v2"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// NewKeyManagerWithConfiguration when provided a schema.OpenIDConnectConfiguration creates a new KeyManager and adds an
// initial key to the manager.
func NewKeyManagerWithConfiguration(configuration *schema.OpenIDConnectConfiguration) (manager *KeyManager, err error) {
	manager = NewKeyManager()

	_, _, err = manager.AddActivePrivateKeyData(configuration.IssuerPrivateKey)
	if err != nil {
		return nil, err
	}

	return manager, nil
}

// NewKeyManager creates a new empty KeyManager.
func NewKeyManager() (manager *KeyManager) {
	manager = new(KeyManager)
	manager.keys = map[string]*rsa.PrivateKey{}
	manager.keySet = new(jose.JSONWebKeySet)

	return manager
}

// Strategy returns the RS256JWTStrategy.
func (m KeyManager) Strategy() (strategy *RS256JWTStrategy) {
	return m.strategy
}

// GetKeySet returns the joseJSONWebKeySet containing the rsa.PublicKey types.
func (m KeyManager) GetKeySet() (keySet *jose.JSONWebKeySet) {
	return m.keySet
}

// GetActiveWebKey obtains the currently active jose.JSONWebKey.
func (m KeyManager) GetActiveWebKey() (webKey *jose.JSONWebKey, err error) {
	webKeys := m.keySet.Key(m.activeKeyID)
	if len(webKeys) == 1 {
		return &webKeys[0], nil
	}

	if len(webKeys) == 0 {
		return nil, errors.New("could not find a key with the active key id")
	}

	return &webKeys[0], errors.New("multiple keys with the same key id")
}

// GetActiveKeyID returns the key id of the currently active key.
func (m KeyManager) GetActiveKeyID() (keyID string) {
	return m.activeKeyID
}

// GetActiveKey returns the rsa.PublicKey of the currently active key.
func (m KeyManager) GetActiveKey() (key *rsa.PublicKey, err error) {
	if key, ok := m.keys[m.activeKeyID]; ok {
		return &key.PublicKey, nil
	}

	return nil, errors.New("failed to retrieve active public key")
}

// GetActivePrivateKey returns the rsa.PrivateKey of the currently active key.
func (m KeyManager) GetActivePrivateKey() (key *rsa.PrivateKey, err error) {
	if key, ok := m.keys[m.activeKeyID]; ok {
		return key, nil
	}

	return nil, errors.New("failed to retrieve active private key")
}

// AddActivePrivateKeyData adds a rsa.PublicKey given the key in the PEM string format, then sets it to the active key.
func (m *KeyManager) AddActivePrivateKeyData(data string) (key *rsa.PrivateKey, webKey *jose.JSONWebKey, err error) {
	ikey, err := utils.ParseX509FromPEM([]byte(data))
	if err != nil {
		return nil, nil, err
	}

	var ok bool

	if key, ok = ikey.(*rsa.PrivateKey); !ok {
		return nil, nil, errors.New("key must be an RSA private key")
	}

	webKey, err = m.AddActivePrivateKey(key)

	return key, webKey, err
}

// AddActivePrivateKey adds a rsa.PublicKey, then sets it to the active key.
func (m *KeyManager) AddActivePrivateKey(key *rsa.PrivateKey) (webKey *jose.JSONWebKey, err error) {
	wk := jose.JSONWebKey{
		Key:       &key.PublicKey,
		Algorithm: "RS256",
		Use:       "sig",
	}

	keyID, err := wk.Thumbprint(crypto.SHA1)
	if err != nil {
		return nil, err
	}

	strKeyID := strings.ToLower(fmt.Sprintf("%x", keyID))
	if len(strKeyID) >= 7 {
		// Shorten the key if it's greater than 7 to a length of exactly 7.
		strKeyID = strKeyID[0:6]
	}

	if _, ok := m.keys[strKeyID]; ok {
		return nil, fmt.Errorf("key id %s already exists", strKeyID)
	}

	// TODO: Add Mutex here when implementing key rotation.
	wk.KeyID = strKeyID
	m.keySet.Keys = append(m.keySet.Keys, wk)
	m.keys[strKeyID] = key
	m.activeKeyID = strKeyID

	m.strategy, err = NewRS256JWTStrategy(wk.KeyID, key)
	if err != nil {
		return &wk, err
	}

	return &wk, nil
}

// NewRS256JWTStrategy returns a new RS256JWTStrategy.
func NewRS256JWTStrategy(id string, key *rsa.PrivateKey) (strategy *RS256JWTStrategy, err error) {
	strategy = new(RS256JWTStrategy)
	strategy.JWTStrategy = new(jwt.RS256JWTStrategy)

	strategy.SetKey(id, key)

	return strategy, nil
}

// RS256JWTStrategy is a decorator struct for the fosite RS256JWTStrategy.
type RS256JWTStrategy struct {
	JWTStrategy *jwt.RS256JWTStrategy

	keyID string
}

// KeyID returns the key id.
func (s RS256JWTStrategy) KeyID() (id string) {
	return s.keyID
}

// SetKey sets the provided key id and key as the active key (this is what triggers fosite to use it).
func (s *RS256JWTStrategy) SetKey(id string, key *rsa.PrivateKey) {
	s.keyID = id
	s.JWTStrategy.PrivateKey = key
}

// Hash is a decorator func for the underlying fosite RS256JWTStrategy.
func (s *RS256JWTStrategy) Hash(ctx context.Context, in []byte) ([]byte, error) {
	return s.JWTStrategy.Hash(ctx, in)
}

// GetSigningMethodLength is a decorator func for the underlying fosite RS256JWTStrategy.
func (s *RS256JWTStrategy) GetSigningMethodLength() int {
	return s.JWTStrategy.GetSigningMethodLength()
}

// GetSignature is a decorator func for the underlying fosite RS256JWTStrategy.
func (s *RS256JWTStrategy) GetSignature(ctx context.Context, token string) (string, error) {
	return s.JWTStrategy.GetSignature(ctx, token)
}

// Generate is a decorator func for the underlying fosite RS256JWTStrategy.
func (s *RS256JWTStrategy) Generate(ctx context.Context, claims jwt.MapClaims, header jwt.Mapper) (string, string, error) {
	return s.JWTStrategy.Generate(ctx, claims, header)
}

// Validate is a decorator func for the underlying fosite RS256JWTStrategy.
func (s *RS256JWTStrategy) Validate(ctx context.Context, token string) (string, error) {
	return s.JWTStrategy.Validate(ctx, token)
}

// Decode is a decorator func for the underlying fosite RS256JWTStrategy.
func (s *RS256JWTStrategy) Decode(ctx context.Context, token string) (*jwt.Token, error) {
	return s.JWTStrategy.Decode(ctx, token)
}

// GetPublicKeyID is a decorator func for the underlying fosite RS256JWTStrategy.
func (s *RS256JWTStrategy) GetPublicKeyID(_ context.Context) (string, error) {
	return s.keyID, nil
}
