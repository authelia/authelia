package schema

import (
	"crypto"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/authelia/jsonschema"
	"github.com/go-crypt/crypt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v4"
)

func TestPasswordDigest_MarshalYAML(t *testing.T) {
	type Example struct {
		Value    bool            `yaml:"value"`
		Password *PasswordDigest `yaml:"password,omitempty"`
	}

	password, err := crypt.Decode("$pbkdf2-sha256$310000$C./EitMdCemqoluAK4Kapw$TTb4uTnL09mJsfbVnypCzJjGvICiiqO56i8VlU5zx6Q")
	require.NoError(t, err)

	testCases := []struct {
		name     string
		have     Example
		expected string
	}{
		{
			"ShouldMarshalValue",
			Example{
				Password: &PasswordDigest{
					Digest: password,
				},
			},
			"value: false\npassword: $pbkdf2-sha256$310000$C./EitMdCemqoluAK4Kapw$TTb4uTnL09mJsfbVnypCzJjGvICiiqO56i8VlU5zx6Q\n",
		},
		{
			"ShouldOmitValue",
			Example{
				Password: nil,
			},
			"value: false\n",
		},
		{
			"ShouldOmitValueNil",
			Example{
				Password: &PasswordDigest{},
			},
			"value: false\npassword: null\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := yaml.Marshal(tc.have)
			require.NoError(t, err)

			assert.Equal(t, tc.expected, string(data))
		})
	}
}

func TestPasswordDigest_UnmarshalYAML(t *testing.T) {
	type Example struct {
		Password *PasswordDigest `yaml:"password,omitempty"`
	}

	password, err := crypt.Decode("$pbkdf2-sha256$310000$C./EitMdCemqoluAK4Kapw$TTb4uTnL09mJsfbVnypCzJjGvICiiqO56i8VlU5zx6Q")
	require.NoError(t, err)

	testCases := []struct {
		name     string
		have     string
		expected Example
		err      string
	}{
		{
			"ShouldUnmarshalValue",
			"password: $pbkdf2-sha256$310000$C./EitMdCemqoluAK4Kapw$TTb4uTnL09mJsfbVnypCzJjGvICiiqO56i8VlU5zx6Q\n",
			Example{
				Password: &PasswordDigest{
					Digest: password,
				},
			},
			"",
		},
		{
			"ShouldErrUnmarshalValue",
			"password: $p-sha256$310000$C./EitMdCemqoluAK4Kapw$TTb4uTnL09mJsfbVnypCzJjGvICiiqO56i8VlU5zx6Q\n",
			Example{},
			"yaml: construct errors:\n  line 1: provided encoded hash has an invalid identifier: the identifier 'p-sha256' is unknown to the decoder",
		},
		{
			"ShouldErrUnmarshalValueType",
			"password: 1\n",
			Example{},
			"yaml: construct errors:\n  line 1: provided encoded hash has an invalid format: the digest doesn't begin with the delimiter '$' and is not one of the other understood formats",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := Example{}

			err = yaml.Unmarshal([]byte(tc.have), &actual)

			if tc.err == "" {
				require.NoError(t, yaml.Unmarshal([]byte(tc.have), &actual))
				assert.Equal(t, tc.expected, actual)
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestNewTLSVersion(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		expected *TLSVersion
		err      string
	}{
		{
			"ShouldParseTLS1.3",
			"TLS1.3",
			&TLSVersion{Value: tls.VersionTLS13},
			"",
		},
		{
			"ShouldParseTLS_1.3",
			"TLS 1.3",
			&TLSVersion{Value: tls.VersionTLS13},
			"",
		},
		{
			"ShouldParse1.3",
			"1.3",
			&TLSVersion{Value: tls.VersionTLS13},
			"",
		},
		{
			"ShouldParseTLS1.2",
			"TLS1.2",
			&TLSVersion{Value: tls.VersionTLS12},
			"",
		},
		{
			"ShouldParseTLS_1.2",
			"TLS 1.2",
			&TLSVersion{Value: tls.VersionTLS12},
			"",
		},
		{
			"ShouldParse1.2",
			"1.2",
			&TLSVersion{Value: tls.VersionTLS12},
			"",
		},
		{
			"ShouldParseTLS1.1",
			"TLS1.1",
			&TLSVersion{Value: tls.VersionTLS11},
			"",
		},
		{
			"ShouldParseTLS_1.1",
			"TLS 1.1",
			&TLSVersion{Value: tls.VersionTLS11},
			"",
		},
		{
			"ShouldParse1.1",
			"1.1",
			&TLSVersion{Value: tls.VersionTLS11},
			"",
		},
		{
			"ShouldParseTLS1.0",
			"TLS1.0",
			&TLSVersion{Value: tls.VersionTLS10},
			"",
		},
		{
			"ShouldParseTLS_1.0",
			"TLS 1.0",
			&TLSVersion{Value: tls.VersionTLS10},
			"",
		},
		{
			"ShouldParse1.0",
			"1.0",
			&TLSVersion{Value: tls.VersionTLS10},
			"",
		},
		{
			"ShouldParseSSL3.0",
			"SSL3.0",
			&TLSVersion{Value: tls.VersionSSL30}, //nolint:staticcheck
			"",
		},
		{
			"ShouldParseSSLv3",
			"SSLv3",
			&TLSVersion{Value: tls.VersionSSL30}, //nolint:staticcheck
			"",
		},
		{
			"ShouldNotParse3.0",
			"3.0",
			nil,
			"supplied tls version isn't supported",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, theError := NewTLSVersion(tc.have)

			if tc.err == "" {
				assert.NoError(t, theError)
				assert.Equal(t, tc.expected, actual)
			} else {
				assert.EqualError(t, theError, tc.err)
				assert.Nil(t, actual)
			}
		})
	}
}

func TestTLSVersion_Functions(t *testing.T) {
	type expected struct {
		min, max uint16
		str      string
	}

	testCases := []struct {
		name     string
		have     *TLSVersion
		expected expected
	}{
		{
			"ShouldReturnCorrectValuesNotConfigured",
			&TLSVersion{},
			expected{
				tls.VersionTLS12,
				tls.VersionTLS13,
				"",
			},
		},
		{
			"ShouldReturnCorrectValueTLS1.3",
			&TLSVersion{Value: tls.VersionTLS13},
			expected{
				tls.VersionTLS13,
				tls.VersionTLS13,
				"TLS 1.3",
			},
		},
		{
			"ShouldReturnCorrectValueTLS1.2",
			&TLSVersion{Value: tls.VersionTLS12},
			expected{
				tls.VersionTLS12,
				tls.VersionTLS12,
				"TLS 1.2",
			},
		},
		{
			"ShouldReturnCorrectValueTLS1.1",
			&TLSVersion{Value: tls.VersionTLS11},
			expected{
				tls.VersionTLS11,
				tls.VersionTLS11,
				"TLS 1.1",
			},
		},
		{
			"ShouldReturnCorrectValueTLS1.0",
			&TLSVersion{Value: tls.VersionTLS10},
			expected{
				tls.VersionTLS10,
				tls.VersionTLS10,
				"TLS 1.0",
			},
		},
		{
			"ShouldReturnCorrectValueSSL3.0",
			&TLSVersion{Value: tls.VersionSSL30}, //nolint:staticcheck
			expected{
				tls.VersionSSL30, //nolint:staticcheck
				tls.VersionSSL30, //nolint:staticcheck
				"SSLv3",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected.min, tc.have.MinVersion())
			assert.Equal(t, tc.expected.max, tc.have.MaxVersion())
			assert.Equal(t, tc.expected.str, tc.have.String())
		})
	}
}

func TestNewX509CertificateChain(t *testing.T) {
	testCases := []struct {
		name             string
		have             string
		thumbprintSHA256 string
		err              string
	}{
		{"ShouldParseCertificate", x509CertificateRSA2048,
			"68a0522fba5df4ec95206ea7f0851f59255617f7abccf42cb1ccc224273ffcfe", ""},
		{"ShouldParseCertificateChain", x509CertificateRSA2048 + "\n" + x509CACertificateRSA2048,
			"68a0522fba5df4ec95206ea7f0851f59255617f7abccf42cb1ccc224273ffcfe", ""},
		{"ShouldNotParseInvalidCertificate", x509CertificateRSAInvalid, "",
			"the PEM data chain contains an invalid certificate: x509: malformed certificate"},
		{"ShouldNotParseInvalidCertificateBlock", x509CertificateRSAInvalidBlock, "", "invalid PEM block"},
		{"ShouldNotParsePrivateKey", x509PrivateKeyRSA2048, "",
			"the PEM data chain contains a PRIVATE KEY but only certificates are expected"},
		{"ShouldNotParseEmptyPEMBlock", x509CertificateEmpty, "", "invalid PEM block"},
		{"ShouldNotParseEmptyData", "", "", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := NewX509CertificateChain(tc.have)

			switch len(tc.err) {
			case 0:
				switch len(tc.have) {
				case 0:
					assert.Nil(t, actual)
				default:
					assert.NotNil(t, actual)
					assert.Equal(t, tc.thumbprintSHA256, fmt.Sprintf("%x", actual.Thumbprint(crypto.SHA256)))
					assert.True(t, actual.HasCertificates())
					assert.NotNil(t, actual.Leaf())
					assert.NotNil(t, actual.CertificatesRaw())
				}

				assert.NoError(t, err)
			default:
				assert.Nil(t, actual)
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestTX509CertificateChain_EncodePEM(t *testing.T) {
	testCases := []struct {
		name string
		have string
	}{
		{
			"ShouldEncodeSingle",
			x509CertificateRSA2048,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			have, err := NewX509CertificateChain(tc.have)

			require.NoError(t, err)
			require.NotNil(t, have)

			actual, err := have.EncodePEM()
			assert.NoError(t, err)

			assert.Equal(t, tc.have, string(actual))
		})
	}
}

func TestX509CertificateChain_Empty(t *testing.T) {
	have := &X509CertificateChain{}

	assert.Nil(t, have.Leaf())
	assert.Nil(t, have.CertificatesRaw())

	encoded, err := have.EncodePEM()

	assert.Nil(t, encoded)
	assert.Nil(t, err)
}

func TestNewX509CertificateChainFromCerts(t *testing.T) {
	have := NewX509CertificateChainFromCerts(nil)
	assert.NotNil(t, have)
}

func TestX509CertificateChain(t *testing.T) {
	chain := &X509CertificateChain{}

	assert.Nil(t, chain.Thumbprint(crypto.SHA256))
	assert.False(t, chain.HasCertificates())
	assert.Len(t, chain.Certificates(), 0)

	assert.False(t, chain.Equal(nil))
	assert.False(t, chain.Equal(&x509.Certificate{}))

	assert.False(t, chain.EqualKey(nil))
	assert.False(t, chain.EqualKey(&rsa.PrivateKey{}))

	cert := MustParseCertificate(x509CertificateRSA4096)
	cacert := MustParseCertificate(x509CACertificateRSA4096)

	chain = MustParseX509CertificateChain(x509CertificateRSA4096 + "\n" + x509CACertificateRSA4096)
	key := MustParsePKCS8RSAPrivateKey(x509PrivateKeyRSA4096)

	thumbprint := chain.Thumbprint(crypto.SHA256)
	assert.NotNil(t, thumbprint)
	assert.Equal(t, "2d1e64e9dd3d3ebd352e1e4a86f4584745f4e2d83e4011efadd43bc078a1e78b", fmt.Sprintf("%x", thumbprint))

	assert.True(t, chain.Equal(cert))
	assert.False(t, chain.Equal(cacert))
	assert.True(t, chain.EqualKey(key))

	assert.NoError(t, chain.Validate())

	chain = MustParseX509CertificateChain(x509CertificateRSA1024 + "\n" + x509CertificateRSA1024)
	assert.EqualError(t, chain.Validate(), "certificate #1 in chain is not signed properly by certificate #2 in chain: x509: invalid signature: parent certificate cannot sign this kind of certificate")

	chain = MustParseX509CertificateChain(x509CertificateRSAExpired + "\n" + x509CACertificateRSAExpired)

	err := chain.Validate()
	require.NotNil(t, err)
	assert.Regexp(t, regexp.MustCompile(`^certificate #1 in chain is invalid after 31536000 but the time is \d+$`), err.Error())

	chain = MustParseX509CertificateChain(x509CertificateRSANotBefore + "\n" + x509CACertificateRSAotBefore)

	err = chain.Validate()
	require.NotNil(t, err)
	assert.Regexp(t, regexp.MustCompile(`^certificate #1 in chain is invalid before 13569465600 but the time is \d+$`), err.Error())
}

func TestPasswordDigest_IsPlainText(t *testing.T) {
	digest, err := DecodePasswordDigest("$plaintext$exam")
	assert.NoError(t, err)
	assert.True(t, digest.IsPlainText())

	value, err := digest.GetPlainTextValue()
	assert.NoError(t, err)
	assert.Equal(t, string(value), "exam")

	digest = &PasswordDigest{}

	assert.False(t, digest.IsPlainText())

	value, err = digest.GetPlainTextValue()
	assert.Nil(t, value)
	assert.EqualError(t, err, "error: nil value")

	digest, err = DecodePasswordDigest("$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng")
	assert.NoError(t, err)

	assert.False(t, digest.IsPlainText())

	value, err = digest.GetPlainTextValue()
	assert.Nil(t, value)
	assert.EqualError(t, err, "error: digest isn't plaintext")
}

func TestPasswordDigest_PlainText(t *testing.T) {
	digest, err := DecodePasswordDigest("$plaintext$exam")
	assert.NoError(t, err)

	v, ok := digest.PlainText()

	assert.NotNil(t, v)
	assert.True(t, ok)

	digest = &PasswordDigest{}

	assert.False(t, digest.IsPlainText())

	digest, err = DecodePasswordDigest("$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng")
	assert.NoError(t, err)

	v, ok = digest.PlainText()

	assert.Nil(t, v)
	assert.False(t, ok)
}

func TestJSONSchema(t *testing.T) {
	testCases := []customSchemaImpl{
		&Address{},
		&AddressSMTP{},
		&AddressTCP{},
		&AddressLDAP{},
		&AddressUDP{},
		&PasswordDigest{},
		&TLSVersion{},
		&X509CertificateChain{},
		&AccessControlRuleDomains{},
		&AccessControlRuleMethods{},
		&AccessControlRuleRegex{},
		&AccessControlRuleSubjects{},
		&IdentityProvidersOpenIDConnectClientURIs{},
	}

	for _, tc := range testCases {
		t.Run(reflect.TypeOf(tc).String(), func(t *testing.T) {
			assert.NotNil(t, tc.JSONSchema())
		})
	}
}

func MustParseX509CertificateChain(data string) *X509CertificateChain {
	chain, err := NewX509CertificateChain(data)
	if err != nil {
		panic(err)
	}

	if chain == nil {
		panic("nil chain")
	}

	return chain
}

func MustParseCertificate(data string) *x509.Certificate {
	block, x := pem.Decode([]byte(data))
	if block == nil {
		panic("not pem")
	}

	if len(x) != 0 {
		panic("extra data")
	}

	if block.Type != blockCERTIFICATE {
		panic(fmt.Sprintf("not certificate block: %s", block.Type))
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		panic(err)
	}

	return cert
}

func MustParsePKCS8RSAPrivateKey(data string) *rsa.PrivateKey {
	return MustParsePKCS8PrivateKey(data).(*rsa.PrivateKey)
}

func MustParsePKCS8PrivateKey(data string) CryptographicPrivateKey {
	block, _ := pem.Decode([]byte(data))
	if block == nil || block.Bytes == nil || len(block.Bytes) == 0 {
		panic("not pem encoded")
	}

	if block.Type != "PRIVATE KEY" {
		panic("not PKCS8 private key")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	if pkey, ok := key.(CryptographicPrivateKey); ok {
		return pkey
	}

	panic("key does not implement the required members")
}

var (
	// Valid from 1970 to 1971 (years).
	x509CertificateRSAExpired = `-----BEGIN CERTIFICATE-----
MIIC5jCCAc6gAwIBAgIRAPKEPEnRO1hurtNAdEuDJA8wDQYJKoZIhvcNAQELBQAw
EzERMA8GA1UEChMIQXV0aGVsaWEwHhcNNzAwMTAxMDAwMDAwWhcNNzEwMTAxMDAw
MDAwWjATMREwDwYDVQQKEwhBdXRoZWxpYTCCASIwDQYJKoZIhvcNAQEBBQADggEP
ADCCAQoCggEBAN+qPAlnoqHMeBeXUF7qnaZXvHS4p8m2N9+hU8us6GYl3mYdFRDy
PGBYWewtIE0RsexBa7UYOV6IXdfheipsmRZZzUxjPbP/VfNuafxdZMVgQzWZZAtt
JJHRhLBATSfkutoPe3eUXxFonEhvl5ErU4327M9cZlLPRsIiVoTWOmigTT0jctx+
u/3IyEVtV982SpttYnpCZ9lCvaSgjpvf1Mim+dbGF0KPKitAbuFnNpWsbRzIYfiy
rGMvxuftkywJ/e6Lx34HJjq/4+K1qII9clIiwAxa1RTnLbBuSLzVHxmj3L5hQhap
jf7HMhLReW2XLJNw4xUShSKpvapBRGbly18CAwEAAaM1MDMwDgYDVR0PAQH/BAQD
AgWgMBMGA1UdJQQMMAoGCCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwDQYJKoZIhvcN
AQELBQADggEBAE5cRfJFTGOozfa9vOedNjif6MLqCtnVv9SRy7k/pdz0KSpYfpRo
dPJ0kBIV2JwIVAXfge6SNn5DK/fyDzERq/kQaCzgBZ6//hCf6Y1BApGajLvOmhPx
KijvMtqQ8IZnqwrURN04pQCt936grsas3Y6haxNyfdJDnQ5FgTe6kDOJU+rXyLp1
N7H6hMXAd9+T++F50PM8AwtRZM6jSUVEhyrniKQSdkOQnXO5ng9if/7GntNzn56o
7cV3sBeenxEmvaXsR30C+A+Ankxr8HBlVOCYcJpbtsCmOB2PVRq9Q5KJAylHRWE1
JedOdWjWvrVaP2IqRopS9mV3Ckf1E19YWFg=
-----END CERTIFICATE-----`

	/*
			// Private Key for x509CertificateRSAExpired.
			x509PrivateKeyRSAExpired = `-----BEGIN RSA PRIVATE KEY-----
		MIIEowIBAAKCAQEA36o8CWeiocx4F5dQXuqdple8dLinybY336FTy6zoZiXeZh0V
		EPI8YFhZ7C0gTRGx7EFrtRg5Xohd1+F6KmyZFlnNTGM9s/9V825p/F1kxWBDNZlk
		C20kkdGEsEBNJ+S62g97d5RfEWicSG+XkStTjfbsz1xmUs9GwiJWhNY6aKBNPSNy
		3H67/cjIRW1X3zZKm21iekJn2UK9pKCOm9/UyKb51sYXQo8qK0Bu4Wc2laxtHMhh
		+LKsYy/G5+2TLAn97ovHfgcmOr/j4rWogj1yUiLADFrVFOctsG5IvNUfGaPcvmFC
		FqmN/scyEtF5bZcsk3DjFRKFIqm9qkFEZuXLXwIDAQABAoIBADBgYrHqD3wNfKAl
		o0WUW1riOSnJ0sjHN9iPzU8NbArEABF4Etlie3qfQXva2tSwkho2oDRANBBlUF7k
		LwdEC+yQqd3uzSbEgHOxmwzxql0ikAbk0YXDKpi7h4aTsdyCFYQauyrHFbTvOnZU
		ZKUKiPz4vomvQ5Z/rJ9KzAnZSDLeqbJfBXPPitlE8DAiYypGKDUmX0unMJh/x0Pw
		mIP/DTd+nMl+QpoSR0nS8r8Pr+4oBJ8K6k9Oni2DKdIW8IvoQJBBa9cm8Y0fHkSl
		hB7fncY5bE0lOZ8jBlSNuGfZHjVihwBA+rYAcWpyzdBx3SHRSe5AH4RKBPaERgSt
		SBV35PECgYEA7ayQs2SMggOiEK4Wf9AzywieaHiHa2ObJRg2dHlIgVkUDL3zB/b6
		57jPMXAtMyGQDW6pZF6Oq3mgYP2A9alB6QKjpX1OGFmqZJxtRMAm0KPs2C2inWzg
		dz9OW18jDlKKsHR00JktqsNgOZC8ldE2cyqgwBNXT/P9GyUMC9RmYhkCgYEA8Okk
		9u8IoIHJEWbtmmh0d1CEmPz4zQosTgUl2GLbNaCE/zDvA82YUQmi1yaF1FHjXMoa
		tD0Plkixoj/ezASeSE+duVpgXflYL4IHbqQq9JQg39vyaSuU1g3wP+hnmnCT62vb
		z4v7ugDLLkSlvNeEQLR3GvZvInZnfdwI5/mjeDcCgYA1UGlhJGP0YjY/gZ2gbCbC
		G5vVGXxfFYfeyVClzfL6uO2rcgyLM9bSlf08PMqW1qeGq9Upo6BjTLQyLYt5D8+u
		Ih5tZ+9VvP9g9En6ixPp52ugjpQUtjCf7z53dp7ZfqCHtofhpwq8bHkwUIxNGxIY
		wW4vx+blE3kqVqQeHzYcOQKBgBgBAv/fvVpQ1Dn5qX8THVeuHCgqPJghhVyYwraW
		0wS648WRmJ8mYyDf9uu9GOSY7DCYqqR+2Qi+YYSrHIXzh9nopOyNBsEWUSUarabm
		kKkiAUyM29CC2Sei5+dWPsxynyp76sD5T7Gu1o/boy/3wWO5F40GNPiYF6PAwtpq
		U1FtAoGBANfr5OcnCIdtHLCEVRCaJdzTkQj5X1g9dF0D/gWkBIF0hibcs9yV2i1Z
		JtxBrOvctkRsY7/C8dCms1gkfwDyTpKuMk9iDd3wfDGP3LdD1+V10pCm5ShHIGNm
		/pRFpN45nR5iCX9mnvr8YJLUsrBkh7N4c4ao8xsXzOLBOk8WvtXL
		-----END RSA PRIVATE KEY-----`
	*/

	// Valid from 1970 to 1971 (years).
	x509CACertificateRSAExpired = `-----BEGIN CERTIFICATE-----
MIIDAzCCAeugAwIBAgIQB07G1WhPAiAHaM6FogkZ7TANBgkqhkiG9w0BAQsFADAT
MREwDwYDVQQKEwhBdXRoZWxpYTAeFw03MDAxMDEwMDAwMDBaFw03MTAxMDEwMDAw
MDBaMBMxETAPBgNVBAoTCEF1dGhlbGlhMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A
MIIBCgKCAQEArly1jZXqLaAQJwTaoT0QaePV7SejQZj418Id7X7Gx1bgjmIS0mL8
058c8Av1ThRaUmVAMkbBc4MED9KnKKckU+qMBV0+2xovrn7TvAKHG3yLUtasn/Uz
2UzBXMgbCIMVCloWkPQ6BuSaNsZnSBmTj/IC16bAZwm5lIS9ZjzQ+QnZg8x5ftNR
Jx8Gar8xb4tQbhb6uZU5zxfuV1+4qk04lf6E48IVrf57NBXpJhdzxuvdBj0l46k3
zf44gybrXpr9O0n0Eb/H8lkIoIDM+vanoRvdg868QQw8C/r06M4E8gJJzZk7Ad2c
oCq66peom6eLUSo2DVfVmmQ2KjLUyW0L6QIDAQABo1MwUTAOBgNVHQ8BAf8EBAMC
AqQwDwYDVR0lBAgwBgYEVR0lADAPBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBSQ
qzXTP6jV517cAknynvP0vr4ChzANBgkqhkiG9w0BAQsFAAOCAQEAGvX14cJM+uxh
rSBYa3qHEfSvlSwbQsCxZv5VXLJj6SLYUXqMLxCuPH16ZKD8MygmnF5/8Mp3ZIdF
Aesehtaoth1H06q6iXWw+wgT9fZKGIHL4RftvjFWncD+Lbk8hP/DkLqsGt+Mj23u
JAByhiG8QmbXu/X7kfXSvXjhQ7f7C+bNKxb03r7mT5gI8mCUp5MyLp3DPk3dKTwa
uby/wjlFMHi92HjfQ6mCn5ijc/ltiMh1wtXf53IEESYvrWV5ABjV8xnumI4j4idB
7yHjCn5id379go8e8a+W8jODNzUORzSio8tDhL4c13tiD4PzlMJ1tUr+CIoCzqB5
m99SvwJ74A==
-----END CERTIFICATE-----`

	/*
			// Private Key for x509CACertificateRSAExpired.
			x509CAPrivateKeyRSAExpired = `-----BEGIN RSA PRIVATE KEY-----
		MIIEowIBAAKCAQEArly1jZXqLaAQJwTaoT0QaePV7SejQZj418Id7X7Gx1bgjmIS
		0mL8058c8Av1ThRaUmVAMkbBc4MED9KnKKckU+qMBV0+2xovrn7TvAKHG3yLUtas
		n/Uz2UzBXMgbCIMVCloWkPQ6BuSaNsZnSBmTj/IC16bAZwm5lIS9ZjzQ+QnZg8x5
		ftNRJx8Gar8xb4tQbhb6uZU5zxfuV1+4qk04lf6E48IVrf57NBXpJhdzxuvdBj0l
		46k3zf44gybrXpr9O0n0Eb/H8lkIoIDM+vanoRvdg868QQw8C/r06M4E8gJJzZk7
		Ad2coCq66peom6eLUSo2DVfVmmQ2KjLUyW0L6QIDAQABAoIBAAiOZBpekO9MO36u
		rkvbQ0Lu+0B4AXrmls9/pxhQcFC34q0aAvJwCRgZZsIg1BjQxt3kOhI9hqC0fS6J
		l8pW6WF00QoyWTNHRa+6aYmAVkDzC6M1BaOT1MeFDLgQ2cLBK/cmFJVoZrCP50Fo
		2wieuK8HoTwT4r0rrP+sw96QfXC7BjC1VSL9GXYemKz0RXEUvXXmzGGc9YE8vCt/
		PXOb4TV30TIQrivkywSTJi8A1jUjYI2rPgo6JCl6GZGmc7hVX4jJ9lbBhUH76ozO
		KS1Yzo/veWL4rVspc2exT5cuX7JIuFCjVi0Nlv1MKv39jpfTfKQh0ug6pHlxUzqX
		Rl6Ln0ECgYEAwX0HtsmKMoSIrU4k90TKyhzjNoL8EaMKrSoZ4nK599igd8g+y6eD
		jc1qO60lOHObyLPFor0rQT7K2KCD7GKYSX5+sh9leQEASl7cJkSAG17WBlrf9Pas
		nUXjTRx3moEILAWmuov4UrYpkuEFk65d98xP3uPtDylFj57Bc+a8DOcCgYEA5rHK
		qdjE8c0/1kgmItDJjmKxrJ5K5hY4w7SkpZeg4Rwr2WAjv3je/nx34D8S7m6R2uzp
		NQYAAHXdzHt4iegupyW/3UXJboEscSTuC6/v3llawAozh02nDsMrdC1LreQ1IiFy
		mKDmPZWxiAZXxEJ0hi0YMCcQnBY673eAleostq8CgYEAtia/mVvYh0Bv7z9O253e
		jzFs0ce0B+KGzYiB/8XjvyknwDw6qbzUwy0romyZSrDDasma+F7AFtdHXXKXX3Ve
		SmoUWhnmjGjd3iW5eSkptRqtwCPTDKkgzZqapuBy1Hg+ujrDwIC+0Rb+wnCmsGYJ
		vpuQYZQPeyNugguBsVv5kucCgYBToTRE6k5LEgsIVVNt356RvXmHiELCsl+VotDl
		Ltilgp7qyI1tBhZgzyJt6q+kO/UoFiZckHZDtHbZgBEsfT0cXvT09C2Xn8BKrAaX
		ugoM4vuhDpGrhR0AnwQLs7fxq/8PBm0So5GT1cZr91Ct1yGC2qogGqlMzEpFMV8t
		+ZyIBQKBgFyI2cZ3/uHQMkWUguml9w134bIGpqGdp8jf9YTvWs3Ax8/qxAVmFmtm
		fot+QiamambblrhdT6pau11Mp06FzytorQH0qKd8mPAqvqtvcSoDZzqXjrkUnZIx
		uLUcfb41clRhlfGDUj5MWimfi7d/Vakh86zAa8pg5WBCtXr0bZ/j
		-----END RSA PRIVATE KEY-----`
	*/

	// Valid from 2400 to 2401 (years).
	x509CertificateRSANotBefore = `-----BEGIN CERTIFICATE-----
MIIC6TCCAdGgAwIBAgIQYQWHYM90zNnwv22xhOPkszANBgkqhkiG9w0BAQsFADAT
MREwDwYDVQQKEwhBdXRoZWxpYTAiGA8yNDAwMDEwMTAwMDAwMFoYDzI0MDAxMjMx
MDAwMDAwWjATMREwDwYDVQQKEwhBdXRoZWxpYTCCASIwDQYJKoZIhvcNAQEBBQAD
ggEPADCCAQoCggEBAPFYYindnRfs8aJfXbaX5IRVj10uKlT4i0BwJ5IYaC4O/3UQ
km7do8lL2Ea2N2L5tQJhk2d+yoWGPeaUyuYP692jPA+4BW6RPuroSPB9WEU+x1ir
it/AzJtavg0Lu2fGZkXxAZJj2MlrXT7csaGwRAvvPEHS6EJW4UtERYIqfpKGB39I
sUhKNvY3edF9sosUAJmiZ8Q4K/uYoyCxyiE1QKLaiIjcZJxtzXkzwVBy1ZlmG+r5
VNNguQQFsS8f7uRlOmo0o3hDG9dByUn7PgFEExbgBtdmNoIPk/pfMFM8NIHK+wOC
q/SO2e/MX0IhJZXfq2VTZFgrisPovg8GpHSHRCkCAwEAAaM1MDMwDgYDVR0PAQH/
BAQDAgWgMBMGA1UdJQQMMAoGCCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwDQYJKoZI
hvcNAQELBQADggEBADBTRbfg/UQeJpdMogm9tleXJBcHqgOgiBxkKYxGSlRg4vlr
tM8USAr24whLvb8KDhT6PaSY8wyPuCxqwqiKR84eselxOAcgDLV9n36OcWRm+oFl
Th1S1JUtjbrctU7i6pp4BwUmBwkVALrbrj+cGVG7uBfbP8L/onmjg9KkY5ttnbyb
Qi/Pv9zYLEo394hL3oeaphkP0iE6cHOvII/qFhnpLREINGp3g8V2I46Id/xYDi22
WRabgMuFGpel7Q26yh+YXyHoIkKdiOXMNNXTsuvp5EDBGybTIXWK3xrqTsREMVDr
EmiaOgL8c7+PpSWuUggJLb/JXDYnPtvekH3gPao=
-----END CERTIFICATE-----`

	/*
			// Private Key for x509CertificateRSANotBefore.
			x509PrivateKeyRSANotBefore = `-----BEGIN RSA PRIVATE KEY-----
		MIIEowIBAAKCAQEA8VhiKd2dF+zxol9dtpfkhFWPXS4qVPiLQHAnkhhoLg7/dRCS
		bt2jyUvYRrY3Yvm1AmGTZ37KhYY95pTK5g/r3aM8D7gFbpE+6uhI8H1YRT7HWKuK
		38DMm1q+DQu7Z8ZmRfEBkmPYyWtdPtyxobBEC+88QdLoQlbhS0RFgip+koYHf0ix
		SEo29jd50X2yixQAmaJnxDgr+5ijILHKITVAotqIiNxknG3NeTPBUHLVmWYb6vlU
		02C5BAWxLx/u5GU6ajSjeEMb10HJSfs+AUQTFuAG12Y2gg+T+l8wUzw0gcr7A4Kr
		9I7Z78xfQiElld+rZVNkWCuKw+i+DwakdIdEKQIDAQABAoIBACGcZHdeJK2bUv+A
		9oUiXDHN1JxufHi+8G218NzYx1F6xzrfZvVHqrKy/FjEsav4CKxfOG8Wak/0JRTC
		rgsiNn/0Zr3tq9v9IF0IonfTjQJ/vrVrlniY2iXcmlEozB2ktMOSz9w6SYurhx3l
		EFvrN17OH38vRydOACxCQsfg8SWofY6SV0gcvlCcuM4lKBiuOBWGcf+xwIs3B+Bs
		Frd282jRWtlcYd+zDE+vLxugNizLGpRKCMEdcKPRw9fkBKDI/f56WegNTUZYYFrV
		LEmYIbOwMawvbi0mOdLsp27CfmeUjkEbwzgdNwjFrWIFAk0wT3QvDrKxDYDLM2Z3
		+PtBMwECgYEA9ICYgzPMbN3CsQ+eWSQXXNk35V7PlMl1DC4UIHhi53BMT1ZhvkHf
		D+eqXQ3BSqOUR7b417VBGkK8UtQuQXh9FwwVU0RhVkjpX0nTBhe8gGF3f3094rX4
		Ckhm8XYQEWCUA9HNhCW+KSNVWqgw9Qi0awEY7HaiR628br39/EckvokCgYEA/LHI
		HA9ixEBeTjds52rK6n9bPHeI87qxF62lLQYXvosyJij9/ruUfXwfJjG3EvCfcW7N
		fr2EvgzPbCozC1V6gI9k5CXhOsf+wD6M8A7g7YHUa/dPq2B8bfqaMD0vW7OoZiLQ
		NpfMtBvZxd1wukPGypLGWabPLo8u6bSfxTqz8KECgYAudnmFBUTls0aaKyOmQOuH
		o2ex2NCNr7Lke6UrfnUdEgQOV5X/d7kR5q5DPKfsrSUyc5zaMQGMIf5zpwqbOnBa
		/trWlfoBUZ23k+ncEIqrwtnYik5GVNor6hJV9F+dTcMS7r2lTR7T5nkD305eYicW
		5oB7/xdbk7JpQQWQ+VwMMQKBgQC3hbKs1mvH1mvnaI+aftACgR5VCweW4/bsGHwG
		+A7Unyl713eowrk0ban9xkuM4N8btfpe2uuGT61xhDBwQdNnfT0sCWrLkyasnoEj
		c9reA9Wv1/yvnbKg+Ul0UWuMsS1TiGMp0xOjlzqRXqMZVFITG4gc4m5EBU9wAnOq
		/VhkIQKBgBUwpoL+K1OswQKV/SHH7n5b/By02mLuMkQR0NlpWRJY6eTDk3FAUkHn
		+T996U0a+7OY+mATVqrfLBcsa6i+HGpb4jZL+kkdtfmtHUnN8YOAwopWp3uoPRmF
		lpakAQq8NPVcrX6PLDkHeOlKhE4ercYFqcCRKcnMB3nD+re6x5l6
		-----END RSA PRIVATE KEY-----`
	*/

	// Valid from 2400 to 2401 (years).
	x509CACertificateRSAotBefore = `-----BEGIN CERTIFICATE-----
MIIDBzCCAe+gAwIBAgIQeR2/TbyH9gEzyjuTijMGVzANBgkqhkiG9w0BAQsFADAT
MREwDwYDVQQKEwhBdXRoZWxpYTAiGA8yNDAwMDEwMTAwMDAwMFoYDzI0MDAxMjMx
MDAwMDAwWjATMREwDwYDVQQKEwhBdXRoZWxpYTCCASIwDQYJKoZIhvcNAQEBBQAD
ggEPADCCAQoCggEBAJ7vpRSDXVvwOLGmjbZdoG25OdsWgmhVAWpFCUifotqorz+z
+VgwPnvVDp1cTp07y+mDJK++GNBOG1pS5G4or6Y1HAlT3nGpE0FYhrtDBQmhvqV0
6mPM/Dq5JIuGiju4LX0KBlaFJugJevw3ySnoPUu0BQ9mTZUgggNwetqsAX7TioFj
TkVIMtgranigOvWjJQyLlmiK1TOgMqgWYNR20SE4CkmIp9SjOdeW7kNVMOojRx9a
VgElA2TN57/Je+/tLbvTDDCP9o59SIXxn5N6JQ2/XdDZPNBHrxnmchQVDXWCkP0A
gkV7V1dl8ur85iEdN+F31Kvd0nzCCaC3YUMxgmcCAwEAAaNTMFEwDgYDVR0PAQH/
BAQDAgKkMA8GA1UdJQQIMAYGBFUdJQAwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4E
FgQUB4/oEXJWhMwKBpMesSOnQdD+G0UwDQYJKoZIhvcNAQELBQADggEBAIlzOnWB
xIhMm3zpLfpJGBi62d9J7Rlf5NitWoztyHdJpQ9y99s67QonR7UY7n11PMdtsLja
hWy1/iZ1o5F2zGKPzSS8pscdIuo4B+TocLiHkEx7ttyQ0MepoDt1RlTOjqilqbfD
A4GyGidns1VZuH8wP8NpZNlWajsXgvkYT433RzPgKe7qoI3DFQwc72SBuZSHHyjE
9SVgdN0KmfFXMum4BurwftelF1etGR+4II3cDG80CH2ZvYdqCURPoa+ny/qqMtzq
W2CnwP59TrotQgKCFJS5EdL3MXaZSvK9z2LERdxDvp4OSoJYoxSMJawfkVwZ15rk
apA21VwIrpFg54A=
-----END CERTIFICATE-----`

	/*
			// Private Key for x509CACertificateRSAotBefore.
			x509CAPrivateKeyRSANotBefore = `-----BEGIN RSA PRIVATE KEY-----
		MIIEpAIBAAKCAQEAnu+lFINdW/A4saaNtl2gbbk52xaCaFUBakUJSJ+i2qivP7P5
		WDA+e9UOnVxOnTvL6YMkr74Y0E4bWlLkbiivpjUcCVPecakTQViGu0MFCaG+pXTq
		Y8z8Orkki4aKO7gtfQoGVoUm6Al6/DfJKeg9S7QFD2ZNlSCCA3B62qwBftOKgWNO
		RUgy2CtqeKA69aMlDIuWaIrVM6AyqBZg1HbRITgKSYin1KM515buQ1Uw6iNHH1pW
		ASUDZM3nv8l77+0tu9MMMI/2jn1IhfGfk3olDb9d0Nk80EevGeZyFBUNdYKQ/QCC
		RXtXV2Xy6vzmIR034XfUq93SfMIJoLdhQzGCZwIDAQABAoIBAEneATBOeYZwWDkg
		un5Gd3hnfN85T/SjhVvZqB3rq6nKemC2Ca4WBgRRmlBChXsIPpZR0CwpwqiVlJrf
		KbGVEUXDKzuekiTrOrrFJSFFXcMDPHLzqrglnhjA0Z5TMk3dJK8XiKiPi+yN823j
		k4f5mvtjOHLWzjn/+M0WatLU3IEPnqpnE+pEKrkZQa7Mg+xHprvt67Q4aCgP5lfy
		A04eoUo7+TMRsK718vb02E81ZQSLgSbQMd0W8Dkt7vRkYRNL0OKBlPQcP9qZlw5s
		swy4ne9jgmJKY3mmdnURTjvdJb20dUsSSzseZ8Tj6UYUDntXrB62YhZvC0ZRhGY/
		Nnf10IECgYEAzSi1l1G6ZblV2g2jPqqD4EsdUvitnN5t592dw52+SyizNj7j+pLt
		OPi+bt5HW4orHQHlPi6wt1BQIZ7UVHmljKQq7ByxOX0SRrRg428JPlFMjCp3bG1t
		zmRQwADGkfqn3JQcmY9VDjtn5oCx18bNpDR1gNiK6zImK/jmWBOiaKcCgYEAxlKN
		vYeG70ZrtVBz6jmX6yOtN8/hBA0mifZVHmkKX6earU3Ds2Uj2Hg+AgMqoFD2aRSp
		wodEYzV6hSpvdLzqBi6SklnfF4NBqJ51TEFBWaVMUZjTwConOcc+vvvSDbCkmnoF
		yTcqVm2p91HD7859ACcO+m8nsFJGldbJFl8RkEECgYEAkSLajEkyH2Kk3JTHRr7k
		eplJDniEgbRNdjmusUN36r3JQnftWkf08FfwiIhRXO37IBNGNN5c/+IePhqZxYUl
		W8CL6OtHaQ8VDdXvsRXNKTvkdkhYoeksRFVtVtd1orH7bK2PKgdfOalHEKc8qRSo
		SCEge101sbuRi4wSkH6bZ4MCgYAerDXv0j40U5fk+wRyfWXZoDLyJtyOW9pSDB8u
		DODl2m45z4UtAb+Bg1dTyFmXYe46Yk+/HlydW3APmHiUfYNUYW+Z4vx2Dn7hLWDG
		4nDRBJfBJvnZBqv6a65wq1HZfDB5E9ZBQJ7zrxJShfrf4/fBRkkywm5I/vCbzBRd
		uWZmAQKBgQCJE5rx3rWZz+srunmmkg5LXBMveD+1HRvlwlj/gELG2b2ytNUmnVey
		2naApHnvW7lZdrADpbzLKGEDB/EsaIPJjqQw45OoIZwPdM4bm0/w5c1ZLnStXGCz
		Th/7Sva6x6FW7tHY6ldqybcMj8w3kA4ByQEOg2BtPnWTm1NX/qcr8Q==
		-----END RSA PRIVATE KEY-----`
	*/

	x509CertificateRSAInvalid = `-----BEGIN CERTIFICATE-----
mIIC5jCCAc6gAwIBAgIRAK4Sj7FiN6PXo/urpfO4E7owDQYJKoZIhvcNAQELBQAw
EzERMA8GA1UEChMIQXV0aGVsaWEwHhcNNzAwMTAxMDAwMDAwWhcNNzEwMTAxMDAw
MDAwWjATMREwDwYDVQQKEwhBdXRoZWxpYTCCASIwDQYJKoZIhvcNAQEBBQADggEP
ADCCAQoCggEBAPKv3pSyP4ozGEiVLJ14dIwFCEGEgq7WUMI0SZZqQA2ID0L59U/Q
/Usyy7uC9gfMUzODTpANtkOjFQcQAsxlR1FOjVBrX5QgjSvXwbQn3DtwMA7XWSl6
LuYx2rBYSlMSN5UZQm/RxMtXfLK2b51WgEEYDFi+nECSqKzR4R54eOPkBEWRfvuY
91AMjlhpivg8e4JWkq4LVQUKbmiFYwIdK8XQiN4blY9WwXwJFYs5sQ/UYMwBFi0H
kWOh7GEjfxgoUOPauIueZSMSlQp7zqAH39n0ZSYb6cS0Npj57QoWZSY3ak87ebcR
Nf4rCvZLby7LoN7qYCKxmCaDD3x2+NYpWH8CAwEAAaM1MDMwDgYDVR0PAQH/BAQD
AgWgMBMGA1UdJQQMMAoGCCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwDQYJKoZIhvcN
AQELBQADggEBAHSITqIQSNzonFl3DzxHPEzr2hp6peo45buAAtu8FZHoA+U7Icfh
/ZXjPg7Xz+hgFwM/DTNGXkMWacQA/PaNWvZspgRJf2AXvNbMSs2UQODr7Tbv+Fb4
lyblmMUNYFMCFVAMU0eIxXAFq2qcwv8UMcQFT0Z/35s6PVOakYnAGGQjTfp5Ljuq
wsdc/xWmM0cHWube6sdRRUD7SY20KU/kWzl8iFO0VbSSrDf1AlEhnLEkp1SPaxXg
OdBnl98MeoramNiJ7NT6Jnyb3zZ578fjaWfThiBpagItI8GZmG4s4Ovh2JbheN8i
ZsjNr9jqHTjhyLVbDRlmJzcqoj4JhbKs6/I=
-----END CERTIFICATE-----`

	x509CertificateRSAInvalidBlock = `-----BEGIN CERTIFICATE-----
MIIC5jCCAc6gAwIBAgIRAK4Sj7FiN6PXo/urPfO4E7owDQYJKoZIhvcNAQELBQAw
EzERMA8GA1UEChMIQXV0aGVsaWEwHhcNNzAwMTAxMDAwMDAwWhcNNzEwMTAxMDAw
MDAwWjATMREwDwYDVQQKEwhBdXRoZWxpYTCCASIwDQYJKoZIhvcNAQEBBQADggEP
ADCCAQoCggEBAPKv3pSyP4ozGEiVLJ14dIWFCEGEgq7WUMI0SZZqQA2ID0L59U/Q
/Usyy7uC9gfMUzODTpANtkOjFQcQAsxlR1FOjVBrX5QgjSvXwbQn3DtwMA7XWSl6
LuYx2rBYSlMSN5UZQm/RxMtXf^K2b51WgEEYDFi+nECSqKzR4R54eOPkBEWRfvuY
91AMjlhpivg8e4JWkq4LVQUKbmiFYwIdK8XQiN4blY9WwXwJFYs5sQ/UYMwBFi0H
kWOh7GEjfxgoUOPauIueZSMSlQp7zqAH39N0ZSYb6cS0Npj57QoWZSY3ak87ebcR
Nf4rCvZLby7LoN7qYCKxmCaDD3x2+NYpWH8CAwEAAaM1MDMwDgYDVR0PAQH/BAQD
AgWgMBMGA1UdJQQMMAoGCCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwDQYJKoZIhvcN
AQELBQADggEBAHSITqIQSNzonFl3DzxHPEzr2hp6peo45buAAtu8FZHoA+U7Icfh
/ZXjPg7Xz+hgFwM/DTNGXkMWacQA/PaNWvZspgRJf2AXvNbMSs2UQODr7Tbv+Fb4
lyblmMUNYFMCFVAMU0eIxXAFq2qcwv8UMcQFT0Z/35s6PVOakYnAGGQjTfp5Ljuq
wsdc/xWmM0cHWube6sdRRUD7SY20KU/kWzl8iFO0VbSSrDf1AlEhnLEkp1SPaxXg
OdBnl98MeoramNiJ7NT6Jnyb3zZ578fjaWfThiBpagItI8GZmG4s4Ovh2JbheN8i
ZsjNr9jqHTjhyLVbDRlmJzcqoj4JhbKs6/I=
-----END CERTIFICATE-----`

	x509CertificateEmpty = `-----BEGIN CERTIFICATE-----
-----END CERTIFICATE-----`
)

type customSchemaImpl interface {
	JSONSchema() *jsonschema.Schema
}

const (
	pathCrypto = "../test_resources/crypto/%s.%s"
)

func MustLoadCryptoSet(alg string, legacy bool, extra ...string) (certCA, keyCA, cert, key string) {
	extraAlt := make([]string, len(extra))

	copy(extraAlt, extra)

	if legacy {
		extraAlt = append(extraAlt, "legacy")
	}

	return MustLoadCryptoRaw(true, alg, "crt", extra...), MustLoadCryptoRaw(true, alg, "pem", extra...), MustLoadCryptoRaw(false, alg, "crt", extraAlt...), MustLoadCryptoRaw(false, alg, "pem", extraAlt...)
}

func MustLoadCryptoRaw(ca bool, alg, ext string, extra ...string) string {
	var fparts []string

	if ca {
		fparts = append(fparts, "ca")
	}

	fparts = append(fparts, strings.ToLower(alg))

	if len(extra) != 0 {
		fparts = append(fparts, extra...)
	}

	var (
		data []byte
		err  error
	)
	if data, err = os.ReadFile(fmt.Sprintf(pathCrypto, strings.Join(fparts, "."), ext)); err != nil {
		panic(err)
	}

	return string(data)
}

var (
	x509CertificateRSA1024, x509CertificateRSA2048, x509CertificateRSA4096 string
	x509CACertificateRSA2048, x509CACertificateRSA4096                     string
	x509PrivateKeyRSA2048, x509PrivateKeyRSA4096                           string
)

func init() {
	_, _, x509CertificateRSA1024, _ = MustLoadCryptoSet("RSA", false, "1024")
	x509CACertificateRSA2048, _, x509CertificateRSA2048, x509PrivateKeyRSA2048 = MustLoadCryptoSet("RSA", false, "2048")
	x509CACertificateRSA4096, _, x509CertificateRSA4096, x509PrivateKeyRSA4096 = MustLoadCryptoSet("RSA", false, "4096")
}
