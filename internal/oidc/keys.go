package oidc

import (
	"context"
	"crypto"
	"crypto/rsa"
	"errors"
	"fmt"
	"strings"

	"github.com/ory/fosite/token/jwt"
	jose "gopkg.in/square/go-jose.v2"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// NewKeyManagerWithConfiguration when provided a schema.OpenIDConnectConfiguration creates a new KeyManager and adds an
// initial key to the manager.
func NewKeyManagerWithConfiguration(config *schema.OpenIDConnectConfiguration) (manager *KeyManager, err error) {
	manager = NewKeyManager()

	if _, err = manager.AddActiveJWK(config.IssuerCertificateChain, config.IssuerPrivateKey); err != nil {
		return nil, err
	}

	return manager, nil
}

// NewKeyManager creates a new empty KeyManager.
func NewKeyManager() (manager *KeyManager) {
	return &KeyManager{
		jwks: &jose.JSONWebKeySet{},
	}
}

// Strategy returns the fosite jwt.JWTStrategy.
func (m *KeyManager) Strategy() (strategy jwt.Signer) {
	if m.jwk == nil {
		return nil
	}

	return m.jwk.Strategy()
}

// GetKeySet returns the joseJSONWebKeySet containing the rsa.PublicKey types.
func (m *KeyManager) GetKeySet() (jwks *jose.JSONWebKeySet) {
	return m.jwks
}

// GetActiveJWK obtains the currently active jose.JSONWebKey.
func (m *KeyManager) GetActiveJWK() (jwk *jose.JSONWebKey, err error) {
	if m.jwks == nil || m.jwk == nil {
		return nil, errors.New("could not obtain the active JWK from an improperly configured key manager")
	}

	jwks := m.jwks.Key(m.jwk.id)

	if len(jwks) == 1 {
		return &jwks[0], nil
	}

	if len(jwks) == 0 {
		return nil, errors.New("could not find a key with the active key id")
	}

	return nil, errors.New("multiple keys with the same key id")
}

// GetActiveKeyID returns the key id of the currently active key.
func (m *KeyManager) GetActiveKeyID() (keyID string) {
	if m.jwk == nil {
		return ""
	}

	return m.jwk.id
}

// GetActivePrivateKey returns the rsa.PrivateKey of the currently active key.
func (m *KeyManager) GetActivePrivateKey() (key *rsa.PrivateKey, err error) {
	if m.jwk == nil {
		return nil, errors.New("failed to retrieve active private key")
	}

	return m.jwk.key, nil
}

// AddActiveJWK is used to add a cert and key pair.
func (m *KeyManager) AddActiveJWK(chain schema.X509CertificateChain, key *rsa.PrivateKey) (jwk *JWK, err error) {
	// TODO: Add a mutex when implementing key rotation to be utilized here and in methods which retrieve the JWK or JWKS.
	if m.jwk, err = NewJWK(chain, key); err != nil {
		return nil, err
	}

	m.jwks.Keys = append(m.jwks.Keys, *m.jwk.JSONWebKey())

	return m.jwk, nil
}

// JWTStrategy is a decorator struct for the fosite jwt.JWTStrategy.
type JWTStrategy struct {
	jwt.Signer

	id string
}

// KeyID returns the key id.
func (s *JWTStrategy) KeyID() (id string) {
	return s.id
}

// GetPublicKeyID is a decorator func for the underlying fosite RS256JWTStrategy.
func (s *JWTStrategy) GetPublicKeyID(_ context.Context) (string, error) {
	return s.id, nil
}

// NewJWK creates a new JWK.
func NewJWK(chain schema.X509CertificateChain, key *rsa.PrivateKey) (j *JWK, err error) {
	if key == nil {
		return nil, fmt.Errorf("JWK is not properly initialized: missing key")
	}

	j = &JWK{
		key:   key,
		chain: chain,
	}

	jwk := &jose.JSONWebKey{
		Algorithm: SigningAlgorithmRSAWithSHA256,
		Use:       "sig",
		Key:       &key.PublicKey,
	}

	var thumbprint []byte

	if thumbprint, err = jwk.Thumbprint(crypto.SHA1); err != nil {
		return nil, fmt.Errorf("failed to calculate SHA1 thumbprint for certificate: %w", err)
	}

	j.id = strings.ToLower(fmt.Sprintf("%x", thumbprint))

	if len(j.id) >= 7 {
		j.id = j.id[:6]
	}

	if len(j.id) >= 7 {
		j.id = j.id[:6]
	}

	return j, nil
}

// JWK is a utility wrapper for JSON Web Key's.
type JWK struct {
	id    string
	key   *rsa.PrivateKey
	chain schema.X509CertificateChain
}

// Strategy returns the relevant jwt.JWTStrategy for this JWT.
func (j *JWK) Strategy() (strategy jwt.Signer) {
	return &JWTStrategy{id: j.id, Signer: &jwt.DefaultSigner{GetPrivateKey: j.GetPrivateKey}}
}

func (j *JWK) GetPrivateKey(ctx context.Context) (key any, err error) {
	return j.key, nil
}

// JSONWebKey returns the relevant *jose.JSONWebKey for this JWT.
func (j *JWK) JSONWebKey() (jwk *jose.JSONWebKey) {
	jwk = &jose.JSONWebKey{
		Key:          &j.key.PublicKey,
		KeyID:        j.id,
		Algorithm:    "RS256",
		Use:          "sig",
		Certificates: j.chain.Certificates(),
	}

	if len(jwk.Certificates) != 0 {
		jwk.CertificateThumbprintSHA1, jwk.CertificateThumbprintSHA256 = j.chain.Thumbprint(crypto.SHA1), j.chain.Thumbprint(crypto.SHA256)
	}

	return jwk
}
