package oidc

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	fjwt "github.com/ory/fosite/token/jwt"
	"github.com/ory/x/errorsx"

	"gopkg.in/square/go-jose.v2"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// NewKeyManager news up a KeyManager.
func NewKeyManager(config *schema.OpenIDConnectConfiguration) (manager *KeyManager) {
	manager = &KeyManager{
		kids: map[string]*JWK{},
		algs: map[string]*JWK{},
	}

	for _, sjwk := range config.IssuerJWKS {
		jwk := NewJWK(sjwk)

		manager.kids[sjwk.KeyID] = jwk
		manager.algs[jwk.alg.Alg()] = jwk

		if jwk.kid == config.Discovery.DefaultKeyID {
			manager.kid = jwk.kid
		}
	}

	return manager
}

// The KeyManager type handles JWKs and signing operations.
type KeyManager struct {
	kid  string
	kids map[string]*JWK
	algs map[string]*JWK
}

func (m *KeyManager) GetKIDFromAlgStrict(ctx context.Context, alg string) (kid string, err error) {
	if jwks, ok := m.algs[alg]; ok {
		return jwks.kid, nil
	}

	return "", fmt.Errorf("alg not found")
}

func (m *KeyManager) GetKIDFromAlg(ctx context.Context, alg string) string {
	if jwks, ok := m.algs[alg]; ok {
		return jwks.kid
	}

	return m.kid
}

func (m *KeyManager) GetByAlg(ctx context.Context, alg string) *JWK {
	if jwk, ok := m.algs[alg]; ok {
		return jwk
	}

	return nil
}

func (m *KeyManager) GetByKID(ctx context.Context, kid string) *JWK {
	if kid == "" {
		return m.kids[m.kid]
	}

	if jwk, ok := m.kids[kid]; ok {
		return jwk
	}

	return nil
}

func (m *KeyManager) GetByHeader(ctx context.Context, header fjwt.Mapper) (jwk *JWK, err error) {
	var (
		kid string
		ok  bool
	)

	if header == nil {
		return nil, fmt.Errorf("jwt header was nil")
	}

	if kid, ok = header.Get(JWTHeaderKeyIdentifier).(string); !ok {
		return nil, fmt.Errorf("jwt header did not have a kid")
	}

	if jwk, ok = m.kids[kid]; !ok {
		return nil, fmt.Errorf("jwt header '%s' with value '%s' does not match a managed jwk", JWTHeaderKeyIdentifier, kid)
	}

	return jwk, nil
}

func (m *KeyManager) GetByTokenString(ctx context.Context, tokenString string) (jwk *JWK, err error) {
	var (
		token *jwt.Token
	)

	if token, _, err = jwt.NewParser().ParseUnverified(tokenString, jwt.MapClaims{}); err != nil {
		return nil, err
	}

	return m.GetByHeader(ctx, &fjwt.Headers{Extra: token.Header})
}

func (m *KeyManager) Set(ctx context.Context) *jose.JSONWebKeySet {
	keys := make([]jose.JSONWebKey, 0, len(m.kids))

	for _, jwk := range m.kids {
		keys = append(keys, jwk.JWK())
	}

	sort.Sort(SortedJSONWebKey(keys))

	return &jose.JSONWebKeySet{
		Keys: keys,
	}
}

func (m *KeyManager) Generate(ctx context.Context, claims fjwt.MapClaims, header fjwt.Mapper) (tokenString string, sig string, err error) {
	var jwk *JWK

	if jwk, err = m.GetByHeader(ctx, header); err != nil {
		return "", "", fmt.Errorf("error getting jwk from header: %w", err)
	}

	return jwk.Strategy().Generate(ctx, claims, header)
}

func (m *KeyManager) Validate(ctx context.Context, tokenString string) (sig string, err error) {
	var jwk *JWK

	if jwk, err = m.GetByTokenString(ctx, tokenString); err != nil {
		return "", fmt.Errorf("error getting jwk from token string: %w", err)
	}

	return jwk.Strategy().Validate(ctx, tokenString)
}

func (m *KeyManager) Hash(ctx context.Context, in []byte) (sum []byte, err error) {
	return m.GetByKID(ctx, "").Strategy().Hash(ctx, in)
}

func (m *KeyManager) Decode(ctx context.Context, tokenString string) (token *fjwt.Token, err error) {
	var jwk *JWK

	if jwk, err = m.GetByTokenString(ctx, tokenString); err != nil {
		return nil, fmt.Errorf("error getting jwk from token string: %w", err)
	}

	return jwk.Strategy().Decode(ctx, tokenString)
}

func (m *KeyManager) GetSignature(ctx context.Context, tokenString string) (sig string, err error) {
	return getTokenSignature(tokenString)
}

func (m *KeyManager) GetSigningMethodLength(ctx context.Context) (size int) {
	return m.GetByKID(ctx, "").Strategy().GetSigningMethodLength(ctx)
}

func NewJWK(s schema.JWK) (jwk *JWK) {
	jwk = &JWK{
		kid: s.KeyID,
		use: s.Use,
		alg: jwt.GetSigningMethod(s.Algorithm),
		key: s.Key.(schema.CryptographicPrivateKey),

		chain:          s.CertificateChain,
		thumbprint:     s.CertificateChain.Thumbprint(crypto.SHA256),
		thumbprintsha1: s.CertificateChain.Thumbprint(crypto.SHA1),
	}

	switch jwk.alg {
	case jwt.SigningMethodRS256, jwt.SigningMethodPS256, jwt.SigningMethodES256:
		jwk.hash = crypto.SHA256
	case jwt.SigningMethodRS384, jwt.SigningMethodPS384, jwt.SigningMethodES384:
		jwk.hash = crypto.SHA384
	case jwt.SigningMethodRS512, jwt.SigningMethodPS512, jwt.SigningMethodES512:
		jwk.hash = crypto.SHA512
	default:
		jwk.hash = crypto.SHA256
	}

	return jwk
}

type JWK struct {
	kid  string
	use  string
	alg  jwt.SigningMethod
	hash crypto.Hash

	key            schema.CryptographicPrivateKey
	chain          schema.X509CertificateChain
	thumbprintsha1 []byte
	thumbprint     []byte
}

func (j *JWK) GetPrivateKey(ctx context.Context) (any, error) {
	return j.PrivateJWK(), nil
}

func (j *JWK) KeyID() string {
	return j.kid
}

func (j *JWK) PrivateJWK() (jwk *jose.JSONWebKey) {
	return &jose.JSONWebKey{
		Key:                         j.key,
		KeyID:                       j.kid,
		Algorithm:                   j.alg.Alg(),
		Use:                         j.use,
		Certificates:                j.chain.Certificates(),
		CertificateThumbprintSHA1:   j.thumbprintsha1,
		CertificateThumbprintSHA256: j.thumbprint,
	}
}

func (j *JWK) JWK() (jwk jose.JSONWebKey) {
	return j.PrivateJWK().Public()
}

func (j *JWK) Strategy() (strategy fjwt.Signer) {
	return &Signer{
		hash:          j.hash,
		alg:           j.alg,
		GetPrivateKey: j.GetPrivateKey,
	}
}

// Signer is responsible for generating and validating JWT challenges.
type Signer struct {
	hash crypto.Hash
	alg  jwt.SigningMethod

	GetPrivateKey fjwt.GetPrivateKeyFunc
}

func (j *Signer) GetPublicKey(ctx context.Context) (key crypto.PublicKey, err error) {
	var k any

	if k, err = j.GetPrivateKey(ctx); err != nil {
		return nil, err
	}

	switch t := k.(type) {
	case *jose.JSONWebKey:
		return t.Public().Key, nil
	case jose.OpaqueSigner:
		return t.Public().Key, nil
	case schema.CryptographicPrivateKey:
		return t.Public(), nil
	default:
		return nil, errors.New("invalid private key type")
	}
}

// Generate generates a new authorize code or returns an error. set secret.
func (j *Signer) Generate(ctx context.Context, claims fjwt.MapClaims, header fjwt.Mapper) (tokenString string, sig string, err error) {
	var key any

	if key, err = j.GetPrivateKey(ctx); err != nil {
		return "", "", err
	}

	switch t := key.(type) {
	case *jose.JSONWebKey:
		return generateToken(claims, header, j.alg, t.Key)
	case jose.JSONWebKey:
		return generateToken(claims, header, j.alg, t.Key)
	case *rsa.PrivateKey, *ecdsa.PrivateKey:
		return generateToken(claims, header, j.alg, t)
	case jose.OpaqueSigner:
		switch tt := t.Public().Key.(type) {
		case *rsa.PrivateKey, *ecdsa.PrivateKey:
			return generateToken(claims, header, j.alg, t)
		default:
			return "", "", fmt.Errorf("unsupported private / public key pairs: %T, %T", t, tt)
		}
	default:
		return "", "", fmt.Errorf("unsupported private key type: %T", t)
	}
}

// Validate validates a token and returns its signature or an error if the token is not valid.
func (j *Signer) Validate(ctx context.Context, tokenString string) (sig string, err error) {
	var (
		key crypto.PublicKey
	)

	if key, err = j.GetPublicKey(ctx); err != nil {
		return "", err
	}

	return validateToken(tokenString, key)
}

// Decode will decode a JWT token.
func (j *Signer) Decode(ctx context.Context, tokenString string) (token *fjwt.Token, err error) {
	var (
		key crypto.PublicKey
	)

	if key, err = j.GetPublicKey(ctx); err != nil {
		return nil, err
	}

	return decodeToken(tokenString, key)
}

// GetSignature will return the signature of a token.
func (j *Signer) GetSignature(ctx context.Context, tokenString string) (sig string, err error) {
	return getTokenSignature(tokenString)
}

// Hash will return a given hash based on the byte input or an error upon fail.
func (j *Signer) Hash(ctx context.Context, in []byte) (sum []byte, err error) {
	hash := j.hash.New()

	if _, err = hash.Write(in); err != nil {
		return []byte{}, errorsx.WithStack(err)
	}

	return hash.Sum([]byte{}), nil
}

// GetSigningMethodLength will return the length of the signing method.
func (j *Signer) GetSigningMethodLength(ctx context.Context) (size int) {
	return j.hash.Size()
}

func generateToken(claims fjwt.MapClaims, header fjwt.Mapper, signingMethod jwt.SigningMethod, key any) (rawToken string, sig string, err error) {
	if header == nil || claims == nil {
		return "", "", errors.New("either claims or header is nil")
	}

	token := jwt.NewWithClaims(signingMethod, claims)

	token.Header = assign(token.Header, header.ToMap())

	if rawToken, err = token.SignedString(key); err != nil {
		return "", "", err
	}

	if sig, err = getTokenSignature(rawToken); err != nil {
		return "", "", err
	}

	return rawToken, sig, nil
}

func decodeToken(tokenString string, key any) (token *fjwt.Token, err error) {
	return fjwt.ParseWithClaims(tokenString, fjwt.MapClaims{}, func(*fjwt.Token) (any, error) {
		return key, nil
	})
}

func validateToken(tokenString string, key any) (sig string, err error) {
	if _, err = decodeToken(tokenString, key); err != nil {
		return "", err
	}

	return getTokenSignature(tokenString)
}

func getTokenSignature(tokenString string) (sig string, err error) {
	parts := strings.Split(tokenString, ".")

	if len(parts) != 3 {
		return "", errors.New("header, body and signature must all be set")
	}

	return parts[2], nil
}

func assign(a, b map[string]any) map[string]any {
	for k, w := range b {
		if _, ok := a[k]; ok {
			continue
		}

		a[k] = w
	}

	return a
}
