package oidc

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

const (
	pathCrypto    = "../configuration/test_resources/crypto/%s.%s"
	myclient      = "myclient"
	myclientdesc  = "My Client"
	onefactor     = "one_factor"
	twofactor     = "two_factor"
	examplecom    = "https://example.com"
	examplecomsid = "example.com"
	badsecret     = "$plaintext$a_bad_secret"
	badhmac       = "asbdhaaskmdlkamdklasmdlkams"
)

func MustDecodeSecret(value string) *schema.PasswordDigest {
	if secret, err := schema.DecodePasswordDigest(value); err != nil {
		panic(err)
	} else {
		return secret
	}
}

func MustParseRequestURI(input string) *url.URL {
	if requestURI, err := url.ParseRequestURI(input); err != nil {
		panic(err)
	} else {
		return requestURI
	}
}

func MustLoadCrypto(alg, mod, ext string, extra ...string) any {
	fparts := []string{alg, mod}
	if len(extra) != 0 {
		fparts = append(fparts, extra...)
	}

	var (
		data    []byte
		decoded any
		err     error
	)

	if data, err = os.ReadFile(fmt.Sprintf(pathCrypto, strings.Join(fparts, "_"), ext)); err != nil {
		panic(err)
	}

	if decoded, err = utils.ParseX509FromPEMRecursive(data); err != nil {
		panic(err)
	}

	return decoded
}

func MustLoadCertificateChain(alg, op string) schema.X509CertificateChain {
	decoded := MustLoadCrypto(alg, op, "crt")

	switch cert := decoded.(type) {
	case *x509.Certificate:
		return schema.NewX509CertificateChainFromCerts([]*x509.Certificate{cert})
	case []*x509.Certificate:
		return schema.NewX509CertificateChainFromCerts(cert)
	default:
		panic(fmt.Errorf("the key was not a *x509.Certificate or []*x509.Certificate, it's a %T", cert))
	}
}

func MustLoadCertificate(alg, op string) *x509.Certificate {
	decoded := MustLoadCrypto(alg, op, "crt")

	cert, ok := decoded.(*x509.Certificate)
	if !ok {
		panic(fmt.Errorf("the key was not a *x509.Certificate, it's a %T", cert))
	}

	return cert
}

func MustLoadEd15519PrivateKey(curve string, extra ...string) ed25519.PrivateKey {
	decoded := MustLoadCrypto("ED25519", curve, "pem", extra...)

	key, ok := decoded.(ed25519.PrivateKey)
	if !ok {
		panic(fmt.Errorf("the key was not a ed25519.PrivateKey, it's a %T", key))
	}

	return key
}

func MustLoadECDSAPrivateKey(curve string, extra ...string) *ecdsa.PrivateKey {
	decoded := MustLoadCrypto("ECDSA", curve, "pem", extra...)

	key, ok := decoded.(*ecdsa.PrivateKey)
	if !ok {
		panic(fmt.Errorf("the key was not a *ecdsa.PrivateKey, it's a %T", key))
	}

	return key
}

func MustLoadRSAPublicKey(bits string, extra ...string) *rsa.PublicKey {
	decoded := MustLoadCrypto("RSA", bits, "pem", extra...)

	key, ok := decoded.(*rsa.PublicKey)
	if !ok {
		panic(fmt.Errorf("the key was not a *rsa.PublicKey, it's a %T", key))
	}

	return key
}

func MustLoadRSAPrivateKey(bits string, extra ...string) *rsa.PrivateKey {
	decoded := MustLoadCrypto("RSA", bits, "pem", extra...)

	key, ok := decoded.(*rsa.PrivateKey)
	if !ok {
		panic(fmt.Errorf("the key was not a *rsa.PrivateKey, it's a %T", key))
	}

	return key
}

var (
	tOpenIDConnectPBKDF2ClientSecret, tOpenIDConnectPlainTextClientSecret *schema.PasswordDigest

	// Standard RSA key / certificate pairs.
	keyRSA1024, keyRSA2048, keyRSA4096    *rsa.PrivateKey
	certRSA1024, certRSA2048, certRSA4096 schema.X509CertificateChain

	// Standard ECDSA key / certificate pairs.
	keyECDSAP224, keyECDSAP256, keyECDSAP384, keyECDSAP521     *ecdsa.PrivateKey
	certECDSAP224, certECDSAP256, certECDSAP384, certECDSAP521 schema.X509CertificateChain
)

func init() {
	tOpenIDConnectPBKDF2ClientSecret = MustDecodeSecret("$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng")
	tOpenIDConnectPlainTextClientSecret = MustDecodeSecret("$plaintext$example")

	keyRSA1024 = MustLoadRSAPrivateKey("1024")
	keyRSA2048 = MustLoadRSAPrivateKey("2048")
	keyRSA4096 = MustLoadRSAPrivateKey("4096")
	keyECDSAP224 = MustLoadECDSAPrivateKey("P224")
	keyECDSAP256 = MustLoadECDSAPrivateKey("P256")
	keyECDSAP384 = MustLoadECDSAPrivateKey("P384")
	keyECDSAP521 = MustLoadECDSAPrivateKey("P521")

	certRSA1024 = MustLoadCertificateChain("RSA", "1024")
	certRSA2048 = MustLoadCertificateChain("RSA", "2048")
	certRSA4096 = MustLoadCertificateChain("RSA", "4096")
	certECDSAP224 = MustLoadCertificateChain("ECDSA", "P224")
	certECDSAP256 = MustLoadCertificateChain("ECDSA", "P256")
	certECDSAP384 = MustLoadCertificateChain("ECDSA", "P384")
	certECDSAP521 = MustLoadCertificateChain("ECDSA", "P521")
}
