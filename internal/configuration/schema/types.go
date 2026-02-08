package schema

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/authelia/jsonschema"
	"github.com/go-crypt/crypt"
	"github.com/go-crypt/crypt/algorithm"
	"github.com/go-crypt/crypt/algorithm/plaintext"
	"github.com/valyala/fasthttp"
	"go.yaml.in/yaml/v4"
)

var cdecoder algorithm.DecoderRegister

// DecodePasswordDigest returns a new PasswordDigest if it can be decoded.
func DecodePasswordDigest(encodedDigest string) (digest *PasswordDigest, err error) {
	var d algorithm.Digest

	if d, err = DecodeAlgorithmDigest(encodedDigest); err != nil {
		return nil, err
	}

	return NewPasswordDigest(d), nil
}

// DecodeAlgorithmDigest returns a new algorithm.Digest if it can be decoded.
func DecodeAlgorithmDigest(encodedDigest string) (digest algorithm.Digest, err error) {
	if cdecoder == nil {
		if cdecoder, err = crypt.NewDefaultDecoder(); err != nil {
			return nil, fmt.Errorf("failed to initialize decoder: %w", err)
		}

		if err = plaintext.RegisterDecoderPlainText(cdecoder); err != nil {
			return nil, fmt.Errorf("failed to initialize decoder: could not register the plaintext decoder: %w", err)
		}
	}

	return cdecoder.Decode(encodedDigest)
}

// NewPasswordDigest returns a new *PasswordDigest from an algorithm.Digest.
func NewPasswordDigest(digest algorithm.Digest) *PasswordDigest {
	return &PasswordDigest{Digest: digest}
}

// PasswordDigest is a configuration type for the crypt.Digest.
type PasswordDigest struct {
	algorithm.Digest
}

// JSONSchema returns the JSON Schema information for the PasswordDigest type.
func (PasswordDigest) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:    jsonschema.TypeString,
		Pattern: `^\$((argon2(id|i|d)\$v=19\$m=\d+,t=\d+,p=\d+|scrypt\$ln=\d+,r=\d+,p=\d+)\$[a-zA-Z0-9\/+]+\$[a-zA-Z0-9\/+]+|pbkdf2(-sha(224|256|384|512))?\$\d+\$[a-zA-Z0-9\/.]+\$[a-zA-Z0-9\/.]+|bcrypt-sha256\$v=2,t=2b,r=\d+\$[a-zA-Z0-9\/.]+\$[a-zA-Z0-9\/.]+|2(a|b|y)?\$\d+\$[a-zA-Z0-9.\/]+|(5|6)\$rounds=\d+\$[a-zA-Z0-9.\/]+\$[a-zA-Z0-9.\/]+|plaintext\$.+|base64\$[a-zA-Z0-9.=\/]+)$`,
	}
}

// PlainText returns a *plaintext.Digest and boolean. If the PasswordDigest is not a plaintext.Digest then it returns
// nil, false, otherwise it returns the value and true.
func (d *PasswordDigest) PlainText() (digest *plaintext.Digest, ok bool) {
	switch raw := d.Digest.(type) {
	case *plaintext.Digest:
		return raw, true
	default:
		return nil, false
	}
}

// IsPlainText returns true if the underlying algorithm.Digest is a *plaintext.Digest.
func (d *PasswordDigest) IsPlainText() (is bool) {
	if d == nil || d.Digest == nil {
		return false
	}

	switch d.Digest.(type) {
	case *plaintext.Digest:
		return true
	default:
		return false
	}
}

// Valid returns true if this digest has a value.
func (d *PasswordDigest) Valid() (valid bool) {
	return d != nil && d.Digest != nil
}

// GetPlainTextValue returns a *plaintext.Digest's byte value from Key() and an error. If the PasswordDigest is not a
// plaintext.Digest then it returns nil and an error, otherwise it returns the value and nil.
func (d *PasswordDigest) GetPlainTextValue() (value []byte, err error) {
	if d == nil || d.Digest == nil {
		return nil, errors.New("error: nil value")
	}

	switch digest := d.Digest.(type) {
	case *plaintext.Digest:
		return digest.Key(), nil
	default:
		return nil, errors.New("error: digest isn't plaintext")
	}
}

func (d *PasswordDigest) UnmarshalYAML(value *yaml.Node) (err error) {
	digestRaw := ""

	if err = value.Decode(&digestRaw); err != nil {
		return err
	}

	if d.Digest, err = DecodeAlgorithmDigest(digestRaw); err != nil {
		return err
	}

	return nil
}

func (d *PasswordDigest) MarshalYAML() (value any, err error) {
	if !d.Valid() {
		return nil, nil
	}

	return d.String(), nil
}

// NewX509CertificateChain creates a new *X509CertificateChain from a given string, parsing each PEM block one by one.
func NewX509CertificateChain(in string) (chain *X509CertificateChain, err error) {
	if in == "" {
		return nil, nil
	}

	chain = &X509CertificateChain{
		certs: []*x509.Certificate{},
	}

	data := []byte(in)

	var (
		block *pem.Block
		cert  *x509.Certificate
	)

	for {
		block, data = pem.Decode(data)

		if block == nil || len(block.Bytes) == 0 {
			return nil, fmt.Errorf("invalid PEM block")
		}

		if block.Type != blockCERTIFICATE {
			return nil, fmt.Errorf("the PEM data chain contains a %s but only certificates are expected", block.Type)
		}

		if cert, err = x509.ParseCertificate(block.Bytes); err != nil {
			return nil, fmt.Errorf("the PEM data chain contains an invalid certificate: %w", err)
		}

		chain.certs = append(chain.certs, cert)

		if len(data) == 0 {
			break
		}
	}

	return chain, nil
}

// NewX509CertificateChainFromCerts returns a chain from a given list of certificates without validation.
func NewX509CertificateChainFromCerts(in []*x509.Certificate) (chain X509CertificateChain) {
	return X509CertificateChain{certs: in}
}

// NewTLSVersion returns a new TLSVersion given a string.
func NewTLSVersion(input string) (version *TLSVersion, err error) {
	switch strings.ReplaceAll(strings.ToUpper(input), " ", "") {
	case TLSVersion13, Version13, tls.VersionName(tls.VersionTLS13):
		return &TLSVersion{tls.VersionTLS13}, nil
	case TLSVersion12, Version12, tls.VersionName(tls.VersionTLS12):
		return &TLSVersion{tls.VersionTLS12}, nil
	case TLSVersion11, Version11, tls.VersionName(tls.VersionTLS11):
		return &TLSVersion{tls.VersionTLS11}, nil
	case TLSVersion10, Version10, tls.VersionName(tls.VersionTLS10):
		return &TLSVersion{tls.VersionTLS10}, nil
	case SSLVersion30, strings.ToUpper(tls.VersionName(tls.VersionSSL30)): //nolint:staticcheck
		return &TLSVersion{tls.VersionSSL30}, nil //nolint:staticcheck
	}

	return nil, ErrTLSVersionNotSupported
}

// TLSVersion is a struct which handles tls.Config versions.
type TLSVersion struct {
	Value uint16
}

// JSONSchema returns the JSON Schema information for the TLSVersion type.
func (TLSVersion) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type: jsonschema.TypeString,
		Enum: []any{
			"TLS 1.0",
			"TLS1.0",
			"TLS 1.1",
			"TLS1.1",
			"TLS 1.2",
			"TLS1.2",
			"TLS 1.3",
			"TLS1.3",
		},
	}
}

// MaxVersion returns the value of this as a MaxVersion value.
func (v *TLSVersion) MaxVersion() uint16 {
	if v == nil || v.Value == 0 {
		return tls.VersionTLS13
	}

	return v.Value
}

// MinVersion returns the value of this as a MinVersion value.
func (v *TLSVersion) MinVersion() uint16 {
	if v == nil || v.Value == 0 {
		return tls.VersionTLS12
	}

	return v.Value
}

// String provides the Stringer.
func (v *TLSVersion) String() string {
	if name := tls.VersionName(v.Value); !strings.HasPrefix(name, "0x") {
		return name
	}

	return ""
}

func (v TLSVersion) MarshalYAML() (any, error) {
	return v.String(), nil
}

// CryptographicPrivateKey represents the actual crypto.PrivateKey interface.
type CryptographicPrivateKey interface {
	Public() crypto.PublicKey
	Equal(x crypto.PrivateKey) bool
}

// CryptographicKey represents an artificial cryptographic public or private key.
type CryptographicKey any

// X509CertificateChain is a helper struct that holds a list of *x509.Certificate's.
type X509CertificateChain struct {
	certs []*x509.Certificate
}

// JSONSchema returns the JSON Schema information for the X509CertificateChain type.
func (X509CertificateChain) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:    jsonschema.TypeString,
		Pattern: `^(-{5}BEGIN CERTIFICATE-{5}\n([a-zA-Z0-9\/+]{1,64}\n)+([a-zA-Z0-9\/+]{1,64}[=]{0,2})\n-{5}END CERTIFICATE-{5}\n?)+$`,
	}
}

// Thumbprint returns the Thumbprint for the first certificate.
func (c *X509CertificateChain) Thumbprint(hash crypto.Hash) []byte {
	if len(c.certs) == 0 {
		return nil
	}

	h := hash.New()

	h.Write(c.certs[0].Raw)

	return h.Sum(nil)
}

// HasCertificates returns true if the chain has any certificates.
func (c *X509CertificateChain) HasCertificates() (has bool) {
	return len(c.certs) != 0
}

// Equal checks if the provided *x509.Certificate is equal to the first *x509.Certificate in the chain.
func (c *X509CertificateChain) Equal(other *x509.Certificate) (equal bool) {
	if len(c.certs) == 0 {
		return false
	}

	return c.certs[0].Equal(other)
}

// EqualKey checks if the provided key (public or private) has a public key equal to the first public key in this chain.
//
//nolint:gocyclo // This is an adequately clear function even with the complexity.
func (c *X509CertificateChain) EqualKey(other any) (equal bool) {
	if len(c.certs) == 0 || other == nil {
		return false
	}

	switch key := other.(type) {
	case *rsa.PublicKey:
		return key.Equal(c.certs[0].PublicKey)
	case rsa.PublicKey:
		return key.Equal(c.certs[0].PublicKey)
	case *rsa.PrivateKey:
		return key.PublicKey.Equal(c.certs[0].PublicKey)
	case rsa.PrivateKey:
		return key.PublicKey.Equal(c.certs[0].PublicKey)
	case *ecdsa.PublicKey:
		return key.Equal(c.certs[0].PublicKey)
	case ecdsa.PublicKey:
		return key.Equal(c.certs[0].PublicKey)
	case *ecdsa.PrivateKey:
		return key.PublicKey.Equal(c.certs[0].PublicKey)
	case ecdsa.PrivateKey:
		return key.PublicKey.Equal(c.certs[0].PublicKey)
	case *ed25519.PublicKey:
		return key.Equal(c.certs[0].PublicKey)
	case ed25519.PublicKey:
		return key.Equal(c.certs[0].PublicKey)
	case *ed25519.PrivateKey:
		switch pub := key.Public().(type) {
		case *ed25519.PublicKey:
			return pub.Equal(c.certs[0].PublicKey)
		case ed25519.PublicKey:
			return pub.Equal(c.certs[0].PublicKey)
		default:
			return false
		}
	case ed25519.PrivateKey:
		switch pub := key.Public().(type) {
		case *ed25519.PublicKey:
			return pub.Equal(c.certs[0].PublicKey)
		case ed25519.PublicKey:
			return pub.Equal(c.certs[0].PublicKey)
		default:
			return false
		}
	default:
		return false
	}
}

// Certificates for this X509CertificateChain.
func (c *X509CertificateChain) Certificates() (certificates []*x509.Certificate) {
	return c.certs
}

// CertificatesRaw for this X509CertificateChain.
func (c *X509CertificateChain) CertificatesRaw() (certificates [][]byte) {
	if !c.HasCertificates() {
		return nil
	}

	for _, cert := range c.certs {
		certificates = append(certificates, cert.Raw)
	}

	return certificates
}

// Leaf returns the first certificate if available for use with tls.Certificate.
func (c *X509CertificateChain) Leaf() (leaf *x509.Certificate) {
	if !c.HasCertificates() {
		return nil
	}

	return c.certs[0]
}

// EncodePEM encodes the entire chain as PEM bytes.
func (c *X509CertificateChain) EncodePEM() (encoded []byte, err error) {
	if !c.HasCertificates() {
		return nil, nil
	}

	buf := &bytes.Buffer{}

	for _, cert := range c.certs {
		block := pem.Block{
			Type:  blockCERTIFICATE,
			Bytes: cert.Raw,
		}

		if err = pem.Encode(buf, &block); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

// Validate the X509CertificateChain ensuring the certificates were provided in the correct order
// (with nth being signed by the nth+1), and that all of the certificates are valid based on the current time.
func (c *X509CertificateChain) Validate() (err error) {
	n := len(c.certs)
	now := time.Now()

	for i, cert := range c.certs {
		if !cert.NotBefore.IsZero() && cert.NotBefore.After(now) {
			return fmt.Errorf("certificate #%d in chain is invalid before %d but the time is %d", i+1, cert.NotBefore.Unix(), now.Unix())
		}

		if cert.NotAfter.Before(now) {
			return fmt.Errorf("certificate #%d in chain is invalid after %d but the time is %d", i+1, cert.NotAfter.Unix(), now.Unix())
		}

		if i+1 >= n {
			break
		}

		if err = cert.CheckSignatureFrom(c.certs[i+1]); err != nil {
			return fmt.Errorf("certificate #%d in chain is not signed properly by certificate #%d in chain: %w", i+1, i+2, err)
		}
	}

	return nil
}

// NewRefreshIntervalDuration returns a RefreshIntervalDuration given a time.Duration.
func NewRefreshIntervalDuration(value time.Duration) RefreshIntervalDuration {
	return RefreshIntervalDuration{value: value, valid: true}
}

// NewRefreshIntervalDurationAlways returns a RefreshIntervalDuration with an always value.
func NewRefreshIntervalDurationAlways() RefreshIntervalDuration {
	return RefreshIntervalDuration{valid: true, always: true}
}

// NewRefreshIntervalDurationNever returns a RefreshIntervalDuration with a never value.
func NewRefreshIntervalDurationNever() RefreshIntervalDuration {
	return RefreshIntervalDuration{valid: true, never: true}
}

// RefreshIntervalDuration is a special time.Duration for the refresh interval.
type RefreshIntervalDuration struct {
	value  time.Duration
	valid  bool
	always bool
	never  bool
}

// Valid returns true if the value was correctly newed up.
func (d RefreshIntervalDuration) Valid() bool {
	return d.valid
}

// Update returns true if the session could require updates.
func (d RefreshIntervalDuration) Update() bool {
	return !d.never && !d.always
}

// Always returns true if the interval is always.
func (d RefreshIntervalDuration) Always() bool {
	return d.always
}

// Never returns true if the interval is never.
func (d RefreshIntervalDuration) Never() bool {
	return d.never
}

// Value returns the time.Duration.
func (d RefreshIntervalDuration) Value() time.Duration {
	return d.value
}

// JSONSchema provides the json-schema formatting.
func (RefreshIntervalDuration) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Default: "5 minutes",
		OneOf: []*jsonschema.Schema{
			{
				Type: jsonschema.TypeString,
				Enum: []any{"always", "never"},
			},
			{
				Type:    jsonschema.TypeString,
				Pattern: `^\d+\s*(y|M|w|d|h|m|s|ms|((year|month|week|day|hour|minute|second|millisecond)s?))(\s*(\s+and\s+)?\d+\s*(y|M|w|d|h|m|s|ms|((year|month|week|day|hour|minute|second|millisecond)s?)))*$`,
			},
			{
				Type:        jsonschema.TypeInteger,
				Description: "The duration in seconds",
			},
		},
	}
}

type IdentityProvidersOpenIDConnectClientURIs []string

func (IdentityProvidersOpenIDConnectClientURIs) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		OneOf: []*jsonschema.Schema{
			&jsonschemaURI,
			{
				Type:        jsonschema.TypeArray,
				Items:       &jsonschemaURI,
				UniqueItems: true,
			},
		},
	}
}

type AccessControlRuleDomains []string

func (AccessControlRuleDomains) JSONSchema() *jsonschema.Schema {
	return &jsonschemaWeakStringUniqueSlice
}

type AccessControlRuleMethods []string

func (AccessControlRuleMethods) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		OneOf: []*jsonschema.Schema{
			&jsonschemaACLMethod,
			{
				Type:        jsonschema.TypeArray,
				Items:       &jsonschemaACLMethod,
				UniqueItems: true,
			},
		},
	}
}

// AccessControlRuleRegex represents the ACL AccessControlRuleSubjects type.
type AccessControlRuleRegex []regexp.Regexp

func (AccessControlRuleRegex) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		OneOf: []*jsonschema.Schema{
			{
				Type:   jsonschema.TypeString,
				Format: jsonschema.FormatStringRegex,
			},
			{
				Type: jsonschema.TypeArray,
				Items: &jsonschema.Schema{
					Type:   jsonschema.TypeString,
					Format: jsonschema.FormatStringRegex,
				},
				UniqueItems: true,
			},
		},
	}
}

// AccessControlRuleSubjects represents the ACL AccessControlRuleSubjects type.
type AccessControlRuleSubjects [][]string

func (AccessControlRuleSubjects) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		OneOf: []*jsonschema.Schema{
			&jsonschemaACLSubject,
			{
				Type:  jsonschema.TypeArray,
				Items: &jsonschemaACLSubject,
			},
			{
				Type: jsonschema.TypeArray,
				Items: &jsonschema.Schema{
					Type:  jsonschema.TypeArray,
					Items: &jsonschemaACLSubject,
				},
				UniqueItems: true,
			},
		},
	}
}

type CSPTemplate string

var jsonschemaURI = jsonschema.Schema{
	Type:   jsonschema.TypeString,
	Format: jsonschema.FormatStringURI,
}

var jsonschemaWeakStringUniqueSlice = jsonschema.Schema{
	OneOf: []*jsonschema.Schema{
		{
			Type: jsonschema.TypeString,
		},
		{
			Type: jsonschema.TypeArray,
			Items: &jsonschema.Schema{
				Type: jsonschema.TypeString,
			},
			UniqueItems: true,
		},
	},
}

var jsonschemaACLSubject = jsonschema.Schema{
	Type:    jsonschema.TypeString,
	Pattern: "^(user|group|oauth2:client):.+$",
}

var jsonschemaACLMethod = jsonschema.Schema{
	Type: jsonschema.TypeString,
	Enum: []any{
		fasthttp.MethodGet,
		fasthttp.MethodHead,
		fasthttp.MethodPost,
		fasthttp.MethodPut,
		fasthttp.MethodPatch,
		fasthttp.MethodDelete,
		fasthttp.MethodTrace,
		fasthttp.MethodConnect,
		fasthttp.MethodOptions,
		"COPY",
		"LOCK",
		"MKCOL",
		"MOVE",
		"PROPFIND",
		"PROPPATCH",
		"UNLOCK",
	},
}
