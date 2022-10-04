package oidc

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"

	"github.com/ory/fosite/token/jwt"
	"gopkg.in/square/go-jose.v2"
	jjwt "gopkg.in/square/go-jose.v2/jwt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func NewKeyStrategy(config *schema.OpenIDConnectConfiguration) *KeyStrategy {
	k := &KeyStrategy{
		alg:  map[string]string{},
		keys: map[string]*JSONWebKey{},
	}

	k.AddJSONWebKeyRSA(config.IssuerPrivateKey, config.IssuerCertificateChain)

	for _, pair := range config.IssuerECDSA {
		k.AddJSONWebKeyECDSA(pair.PrivateKey, pair.CertificateChain)
	}

	return k
}

type PrivateKey interface {
	Public() crypto.PublicKey
	Equal(x crypto.PrivateKey) bool
}

func NewJSONWebKey(key PrivateKey, chain schema.X509CertificateChain, use string, alg jose.SignatureAlgorithm) *JSONWebKey {
	jwk := &JSONWebKey{
		alg:            alg,
		use:            use,
		key:            key,
		chain:          chain,
		thumbprint:     chain.Thumbprint(crypto.SHA256),
		thumbprintSHA1: chain.Thumbprint(crypto.SHA1),
	}

	jwk.id = MustGenerateKeyID(jwk.Public())

	return jwk
}

type JSONWebKey struct {
	id, use                    string
	alg                        jose.SignatureAlgorithm
	key                        PrivateKey
	chain                      schema.X509CertificateChain
	thumbprint, thumbprintSHA1 []byte
}

func (k *JSONWebKey) Key() PrivateKey {
	return k.key
}

func (k *JSONWebKey) Public() *jose.JSONWebKey {
	return &jose.JSONWebKey{
		Key:                         k.key.Public(),
		KeyID:                       k.id,
		Algorithm:                   string(k.alg),
		Use:                         k.use,
		Certificates:                k.chain.Certificates(),
		CertificateThumbprintSHA1:   k.thumbprintSHA1,
		CertificateThumbprintSHA256: k.thumbprint,
	}
}

type KeyStrategy struct {
	alg  map[string]string
	keys map[string]*JSONWebKey
}

func (j *KeyStrategy) SigningAlgValues() []string {
	values := make([]string, len(j.alg))

	var i int

	for alg := range j.alg {
		values[i] = alg

		i++
	}

	return values
}

func (j *KeyStrategy) AddJSONWebKeyECDSA(key *ecdsa.PrivateKey, chain schema.X509CertificateChain) {
	if key == nil {
		return
	}

	switch key.Curve {
	case elliptic.P256():
		j.AddJSONWebKeyWithAlgorithm(key, chain, jose.ES256)
	case elliptic.P384():
		j.AddJSONWebKeyWithAlgorithm(key, chain, jose.ES384)
	case elliptic.P521():
		j.AddJSONWebKeyWithAlgorithm(key, chain, jose.ES512)
	}
}

func (j *KeyStrategy) AddJSONWebKeyRSA(key *rsa.PrivateKey, chain schema.X509CertificateChain) {
	if key == nil {
		return
	}

	for _, alg := range []jose.SignatureAlgorithm{jose.RS256, jose.RS384, jose.RS512, jose.PS256, jose.PS384, jose.PS512} {
		j.AddJSONWebKeyWithAlgorithm(key, chain, alg)
	}
}

func (j *KeyStrategy) AddJSONWebKeyWithAlgorithm(key PrivateKey, chain schema.X509CertificateChain, alg jose.SignatureAlgorithm) {
	jwk := NewJSONWebKey(key, chain, JSONWebKeyUseSignature, alg)

	j.keys[jwk.id] = jwk
	j.alg[string(jwk.alg)] = jwk.id
}

func (j *KeyStrategy) JSONWebKeySet() (set jose.JSONWebKeySet) {
	set.Keys = make([]jose.JSONWebKey, len(j.keys))

	var i int

	for _, key := range j.keys {
		set.Keys[i] = *key.Public()

		i++
	}

	return set
}

func (j *KeyStrategy) GetKIDFromJWA(jwa string) (kid string, err error) {
	var ok bool

	if kid, ok = j.alg[jwa]; ok {
		if _, ok = j.keys[kid]; !ok {
			return "", fmt.Errorf("couldn't find key for JWA '%s' and kid '%s'", jwa, kid)
		}

		return kid, nil
	}

	return "", fmt.Errorf("couldn't find kid for JWA '%s'", jwa)
}

func (j *KeyStrategy) GetJWKFromJWA(jwa string) (key *jose.JSONWebKey, err error) {
	if kid, ok := j.alg[jwa]; ok {
		k, ok := j.keys[kid]
		if !ok {
			return nil, fmt.Errorf("couldn't find key for JWA '%s' and kid '%s'", jwa, kid)
		}

		return k.Public(), nil
	}

	return nil, fmt.Errorf("couldn't find kid for JWA '%s'", jwa)
}

func (j *KeyStrategy) GetJWKFromRawToken(rawToken string) (key *jose.JSONWebKey, err error) {
	token, err := jjwt.ParseSigned(rawToken)
	if err != nil {
		return nil, err
	}

	if len(token.Headers) < 1 || len(token.Headers[0].KeyID) == 0 {
		return nil, fmt.Errorf("could not determine key")
	}

	kid := token.Headers[0].KeyID
	k, ok := j.keys[kid]

	if !ok {
		return nil, fmt.Errorf("could not find key with id '%s'", kid)
	}

	if string(k.alg) != token.Headers[0].Algorithm {
		return nil, fmt.Errorf("key with alg '%s' can't decode a token with alg '%s'", k.alg, token.Headers[0].Algorithm)
	}

	return k.Public(), nil
}

func (j *KeyStrategy) GetJWKFromHeader(header jwt.Mapper) (key *JSONWebKey, err error) {
	var (
		kid string
		ok  bool
	)

	if kid, ok = header.Get(JWTHeaderKeyIdentifier).(string); !ok {
		return nil, fmt.Errorf("could not retrieve kid from headers")
	}

	if key, ok = j.keys[kid]; !ok {
		return nil, fmt.Errorf("could not find key with id '%s'", kid)
	}

	return key, nil
}

func (j *KeyStrategy) Generate(ctx context.Context, claims jwt.MapClaims, header jwt.Mapper) (rawToken string, sig string, err error) {
	if header == nil || claims == nil {
		err = errors.New("Either claims or header is nil.")
		return
	}

	var jwk *JSONWebKey

	if jwk, err = j.GetJWKFromHeader(header); err != nil {
		return "", "", fmt.Errorf("error getting jwk from header: %w", err)
	}

	token := jwt.NewWithClaims(jwk.alg, claims)

	token.Header = assign(token.Header, header.ToMap())

	if rawToken, err = token.SignedString(jwk.key); err != nil {
		return "", "", fmt.Errorf("error using signed string: %w", err)
	}

	sig, err = getTokenSignature(rawToken)
	return
}

// Validate validates a token and returns its signature or an error if the token is not valid.
func (j *KeyStrategy) Validate(ctx context.Context, rawToken string) (sig string, err error) {
	key, err := j.GetJWKFromRawToken(rawToken)
	if err != nil {
		return
	}

	return validateToken(rawToken, key.Key)
}

func (j *KeyStrategy) Decode(ctx context.Context, rawToken string) (token *jwt.Token, err error) {
	key, err := j.GetJWKFromRawToken(rawToken)
	if err != nil {
		return
	}

	return decodeToken(rawToken, key.Key)
}

func (j *KeyStrategy) Hash(ctx context.Context, in []byte) (sum []byte, err error) {
	return hashSHA256(in)
}

func (j *KeyStrategy) GetSignature(ctx context.Context, token string) (sig string, err error) {
	return getTokenSignature(token)
}

func (j *KeyStrategy) GetSigningMethodLength() (length int) {
	return crypto.SHA256.Size()
}

func decodeToken(token string, verificationKey interface{}) (*jwt.Token, error) {
	keyFunc := func(*jwt.Token) (interface{}, error) { return verificationKey, nil }
	return jwt.ParseWithClaims(token, jwt.MapClaims{}, keyFunc)
}

func validateToken(tokenStr string, verificationKey interface{}) (string, error) {
	_, err := decodeToken(tokenStr, verificationKey)
	if err != nil {
		return "", err
	}
	return getTokenSignature(tokenStr)
}

func getTokenSignature(token string) (string, error) {
	split := strings.Split(token, ".")
	if len(split) != 3 {
		return "", errors.New("Header, body and signature must all be set")
	}
	return split[2], nil
}

func hashSHA256(in []byte) ([]byte, error) {
	hash := sha256.New()
	_, err := hash.Write(in)
	if err != nil {
		return []byte{}, err
	}
	return hash.Sum([]byte{}), nil
}

func assign(a, b map[string]interface{}) map[string]interface{} {
	for k, w := range b {
		if _, ok := a[k]; ok {
			continue
		}
		a[k] = w
	}
	return a
}

func MustGenerateKeyID(jwk *jose.JSONWebKey) string {
	var (
		thumbprint []byte
		err        error
	)

	if thumbprint, err = jwk.Thumbprint(crypto.SHA256); err != nil {
		panic(err)
	}

	h := hmac.New(crypto.SHA1.New, thumbprint)

	h.Write([]byte(jwk.Use))
	h.Write([]byte(jwk.Algorithm))
	h.Write(thumbprint)

	if thumbprint, err = jwk.Thumbprint(crypto.SHA1); err != nil {
		panic(err)
	}

	h.Write(thumbprint)

	return fmt.Sprintf("%x", h.Sum(nil))[:6]
}

/*
// JWTStrategy is a decorator struct for the fosite jwt.JWTStrategy.
type JWTStrategy struct {
	jwt.JWTStrategy

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
func (j *JWK) Strategy() (strategy jwt.JWTStrategy) {
	return &JWTStrategy{id: j.id, JWTStrategy: &jwt.RS256JWTStrategy{PrivateKey: j.key}}
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
}.


*/
