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

	fjwt "authelia.com/provider/oauth2/token/jwt"
	"authelia.com/provider/oauth2/x/errorsx"
	"github.com/go-jose/go-jose/v4"
	"github.com/golang-jwt/jwt/v5"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// NewKeyManager news up a KeyManager.
func NewKeyManager(config *schema.IdentityProvidersOpenIDConnect) (manager *KeyManager) {
	manager = &KeyManager{
		alg2kid: config.Discovery.DefaultKeyIDs,
		kids:    map[string]*JWK{},
		algs:    map[string]*JWK{},
	}

	for _, sjwk := range config.JSONWebKeys {
		jwk := NewJWK(sjwk)

		manager.kids[sjwk.KeyID] = jwk
		manager.algs[jwk.alg.Alg()] = jwk
	}

	return manager
}

// The KeyManager type handles JWKs and signing operations.
type KeyManager struct {
	alg2kid map[string]string
	kids    map[string]*JWK
	algs    map[string]*JWK
}

// GetDefaultKeyID returns the default key id.
func (m *KeyManager) GetDefaultKeyID(ctx context.Context) string {
	return m.alg2kid[SigningAlgRSAUsingSHA256]
}

// GetKeyID returns the JWK Key ID given an kid/alg or the default if it doesn't exist.
func (m *KeyManager) GetKeyID(ctx context.Context, kid, alg string) string {
	if kid != "" {
		if jwk, ok := m.kids[kid]; ok {
			return jwk.KeyID()
		}
	}

	if jwk, ok := m.algs[alg]; ok {
		return jwk.KeyID()
	}

	return m.alg2kid[SigningAlgRSAUsingSHA256]
}

// GetKeyIDFromAlgStrict returns the key id given an alg or an error if it doesn't exist.
func (m *KeyManager) GetKeyIDFromAlgStrict(ctx context.Context, alg string) (kid string, err error) {
	if jwks, ok := m.algs[alg]; ok {
		return jwks.kid, nil
	}

	return "", fmt.Errorf("alg not found")
}

// GetKeyIDFromAlg returns the key id given an alg or the default if it doesn't exist.
func (m *KeyManager) GetKeyIDFromAlg(ctx context.Context, alg string) string {
	if jwks, ok := m.algs[alg]; ok {
		return jwks.kid
	}

	return m.alg2kid[SigningAlgRSAUsingSHA256]
}

// Get returns the JWK given an kid/alg or nil if it doesn't exist.
func (m *KeyManager) Get(ctx context.Context, kid, alg string) *JWK {
	if kid != "" {
		return m.kids[kid]
	}

	if jwk, ok := m.algs[alg]; ok {
		return jwk
	}

	return nil
}

// GetByAlg returns the JWK given an alg or nil if it doesn't exist.
func (m *KeyManager) GetByAlg(ctx context.Context, alg string) *JWK {
	if jwk, ok := m.algs[alg]; ok {
		return jwk
	}

	return nil
}

// GetByKID returns the JWK given an key id or nil if it doesn't exist. If given a blank string it returns the default.
func (m *KeyManager) GetByKID(ctx context.Context, kid string) *JWK {
	if kid == "" {
		return m.kids[m.alg2kid[SigningAlgRSAUsingSHA256]]
	}

	if jwk, ok := m.kids[kid]; ok {
		return jwk
	}

	return nil
}

// GetByHeader returns the JWK a JWT header with the appropriate kid value or returns an error.
func (m *KeyManager) GetByHeader(ctx context.Context, header fjwt.Mapper) (jwk *JWK, err error) {
	var (
		kid, alg string
		ok       bool
	)

	if header == nil {
		return nil, fmt.Errorf("jwt header was nil")
	}

	kid, _ = header.Get(JWTHeaderKeyIdentifier).(string)
	alg, _ = header.Get(JWTHeaderKeyAlgorithm).(string)

	if len(kid) != 0 {
		if jwk, ok = m.kids[kid]; ok {
			return jwk, nil
		}

		return nil, fmt.Errorf("jwt header '%s' with value '%s' does not match a managed jwk", JWTHeaderKeyIdentifier, kid)
	}

	if len(alg) != 0 {
		if jwk, ok = m.algs[alg]; ok {
			return jwk, nil
		}

		return nil, fmt.Errorf("jwt header '%s' with value '%s' does not match a managed jwk", JWTHeaderKeyAlgorithm, alg)
	}

	return nil, fmt.Errorf("jwt header did not match a known jwk")
}

// GetByTokenString does an invalidated decode of a token to get the  header, then calls GetByHeader.
func (m *KeyManager) GetByTokenString(ctx context.Context, tokenString string) (jwk *JWK, err error) {
	var (
		token *jwt.Token
	)

	if token, _, err = jwt.NewParser().ParseUnverified(tokenString, jwt.MapClaims{}); err != nil {
		return nil, err
	}

	return m.GetByHeader(ctx, &fjwt.Headers{Extra: token.Header})
}

// Set returns the *jose.JSONWebKeySet.
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

// Generate implements the fosite jwt.Signer interface and automatically maps the underlying keys based on the JWK Header kid.
func (m *KeyManager) Generate(ctx context.Context, claims fjwt.MapClaims, header fjwt.Mapper) (tokenString string, sig string, err error) {
	var jwk *JWK

	if jwk, err = m.GetByHeader(ctx, header); err != nil {
		return "", "", fmt.Errorf("error getting jwk from header: %w", err)
	}

	extra := header.ToMap()

	extra[JWTHeaderKeyIdentifier] = jwk.KeyID()
	extra[JWTHeaderKeyAlgorithm] = jwk.Algorithm()

	return jwk.Strategy().Generate(ctx, claims, &fjwt.Headers{Extra: extra})
}

// Validate implements the fosite jwt.Signer interface and automatically maps the underlying keys based on the JWK Header kid.
func (m *KeyManager) Validate(ctx context.Context, tokenString string) (sig string, err error) {
	var jwk *JWK

	if jwk, err = m.GetByTokenString(ctx, tokenString); err != nil {
		return "", fmt.Errorf("error getting jwk from token string: %w", err)
	}

	return jwk.Strategy().Validate(ctx, tokenString)
}

// Hash implements the fosite jwt.Signer interface.
func (m *KeyManager) Hash(ctx context.Context, in []byte) (sum []byte, err error) {
	return m.GetByKID(ctx, "").Strategy().Hash(ctx, in)
}

// Decode implements the fosite jwt.Signer interface and automatically maps the underlying keys based on the JWK Header kid.
func (m *KeyManager) Decode(ctx context.Context, tokenString string) (token *fjwt.Token, err error) {
	var jwk *JWK

	if jwk, err = m.GetByTokenString(ctx, tokenString); err != nil {
		return nil, fmt.Errorf("error getting jwk from token string: %w", err)
	}

	return jwk.Strategy().Decode(ctx, tokenString)
}

// GetSignature implements the fosite jwt.Signer interface.
func (m *KeyManager) GetSignature(ctx context.Context, tokenString string) (sig string, err error) {
	return getTokenSignature(tokenString)
}

// GetSigningMethodLength implements the fosite jwt.Signer interface.
func (m *KeyManager) GetSigningMethodLength(ctx context.Context) (size int) {
	return m.GetByKID(ctx, "").Strategy().GetSigningMethodLength(ctx)
}

// NewPublicJSONWebKeySetFromSchemaJWK creates a *jose.JSONWebKeySet from a slice of schema.JWK.
func NewPublicJSONWebKeySetFromSchemaJWK(sjwks []schema.JWK) (jwks *jose.JSONWebKeySet) {
	n := len(sjwks)

	if n == 0 {
		return nil
	}

	keys := make([]jose.JSONWebKey, n)

	for i := 0; i < n; i++ {
		jwk := jose.JSONWebKey{
			KeyID:        sjwks[i].KeyID,
			Algorithm:    sjwks[i].Algorithm,
			Use:          sjwks[i].Use,
			Certificates: sjwks[i].CertificateChain.Certificates(),
		}

		switch key := sjwks[i].Key.(type) {
		case *rsa.PublicKey:
			jwk.Key = key
		case rsa.PublicKey:
			jwk.Key = &key
		case *rsa.PrivateKey:
			jwk.Key = key.PublicKey
		case *ecdsa.PublicKey:
			jwk.Key = key
		case ecdsa.PublicKey:
			jwk.Key = &key
		case *ecdsa.PrivateKey:
			jwk.Key = key.PublicKey
		}

		keys[i] = jwk
	}

	return &jose.JSONWebKeySet{
		Keys: keys,
	}
}

// NewJWK creates a *JWK f rom a schema.JWK.
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

// JWK is a representation layer over the *jose.JSONWebKey for convenience.
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

// GetSigningMethod returns the jwt.SigningMethod for this *JWK.
func (j *JWK) GetSigningMethod() jwt.SigningMethod {
	return j.alg
}

// GetPrivateKey returns the Private Key for this *JWK.
func (j *JWK) GetPrivateKey(ctx context.Context) (any, error) {
	return j.PrivateJWK(), nil
}

// KeyID returns the Key ID for this *JWK.
func (j *JWK) KeyID() string {
	return j.kid
}

// Algorithm returns the Algorithm for this *JWK.
func (j *JWK) Algorithm() string {
	return j.alg.Alg()
}

// DirectJWK directly returns the *JWK as a jose.JSONWebKey with the private key if appropriate.
func (j *JWK) DirectJWK() (jwk jose.JSONWebKey) {
	return jose.JSONWebKey{
		Key:                         j.key,
		KeyID:                       j.kid,
		Algorithm:                   j.alg.Alg(),
		Use:                         j.use,
		Certificates:                j.chain.Certificates(),
		CertificateThumbprintSHA1:   j.thumbprintsha1,
		CertificateThumbprintSHA256: j.thumbprint,
	}
}

// PrivateJWK directly returns the *JWK as a *jose.JSONWebKey with the private key if appropriate.
func (j *JWK) PrivateJWK() (jwk *jose.JSONWebKey) {
	value := j.DirectJWK()

	return &value
}

// JWK directly returns the *JWK as a jose.JSONWebKey specifically without the private key.
func (j *JWK) JWK() (jwk jose.JSONWebKey) {
	if jwk = j.DirectJWK(); jwk.IsPublic() {
		return jwk
	}

	return jwk.Public()
}

// Strategy returns the fosite jwt.Signer.
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

// GetPublicKey returns the PublicKey for this Signer.
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
		return generateToken(jwt.MapClaims(claims), header, j.alg, t.Key)
	case jose.JSONWebKey:
		return generateToken(jwt.MapClaims(claims), header, j.alg, t.Key)
	case *rsa.PrivateKey, *ecdsa.PrivateKey:
		return generateToken(jwt.MapClaims(claims), header, j.alg, t)
	case jose.OpaqueSigner:
		switch tt := t.Public().Key.(type) {
		case *rsa.PrivateKey, *ecdsa.PrivateKey:
			return generateToken(jwt.MapClaims(claims), header, j.alg, t)
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

func generateToken(claims jwt.MapClaims, header fjwt.Mapper, signingMethod jwt.SigningMethod, key any) (rawToken string, sig string, err error) {
	if header == nil || claims == nil {
		return "", "", errors.New("either claims or header is nil")
	}

	token := jwt.NewWithClaims(signingMethod, claims)

	assign(token.Header, header.ToMap())

	if typ := header.Get(JWTHeaderKeyType); typ != nil {
		token.Header[JWTHeaderKeyType] = typ
	}

	if rawToken, err = token.SignedString(key); err != nil {
		return "", "", err
	}

	if sig, err = getTokenSignature(rawToken); err != nil {
		return "", "", err
	}

	return rawToken, sig, nil
}

func decodeToken(tokenString string, key any) (token *fjwt.Token, err error) {
	return fjwt.ParseWithClaims(tokenString, fjwt.MapClaims{}, keyFromValue(key))
}

func keyFromValue(key any) func(*fjwt.Token) (any, error) {
	return func(_ *fjwt.Token) (any, error) {
		return key, nil
	}
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
