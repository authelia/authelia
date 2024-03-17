package oidc_test

import (
	"bytes"
	"crypto/ecdsa"
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
	pathCrypto     = "../configuration/test_resources/crypto/%s.%s"
	myclient       = "myclient"
	myclientdesc   = "My Client"
	onefactor      = "one_factor"
	twofactor      = "two_factor"
	examplecom     = "https://example.com"
	examplecomsid  = "example.com"
	badhmac        = "asbdhaaskmdlkamdklasmdlkams"
	badTokenString = "badTokenString"
)

const (
	abc = "abc"
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

var (
	tOpenIDConnectPBKDF2ClientSecret, tOpenIDConnectPlainTextClientSecret *schema.PasswordDigest
)

func MustLoadRSACryptoSet(legacy bool, extra ...string) (chain schema.X509CertificateChain, key *rsa.PrivateKey) {
	c, cc, k := MustLoadCryptoSet("RSA", legacy, extra...)

	chain = MustParseCertificateChain(c, cc)

	var (
		decoded any
		err     error
		ok      bool
	)

	if decoded, err = utils.ParseX509FromPEMRecursive(k); err != nil {
		panic(err)
	}

	if key, ok = decoded.(*rsa.PrivateKey); ok {
		if chain.EqualKey(key) {
			return chain, key
		}

		panic("key not valid for chain")
	}

	panic("invalid key")
}

func MustLoadECDSACryptoSet(legacy bool, extra ...string) (chain schema.X509CertificateChain, key *ecdsa.PrivateKey) {
	c, cc, k := MustLoadCryptoSet("ECDSA", legacy, extra...)

	chain = MustParseCertificateChain(c, cc)

	var (
		decoded any
		err     error
		ok      bool
	)

	if decoded, err = utils.ParseX509FromPEMRecursive(k); err != nil {
		panic(err)
	}

	if key, ok = decoded.(*ecdsa.PrivateKey); ok {
		if chain.EqualKey(key) {
			return chain, key
		}

		panic("key not valid for chain")
	}

	panic("invalid key")
}

func MustLoadCryptoSet(alg string, legacy bool, extra ...string) (cert, certCA, key []byte) {
	extraAlt := make([]string, len(extra))

	copy(extraAlt, extra)

	if legacy {
		extraAlt = append(extraAlt, "legacy")
	}

	return MustLoadCryptoRaw(false, alg, "crt", extraAlt...), MustLoadCryptoRaw(true, alg, "crt", extra...), MustLoadCryptoRaw(false, alg, "pem", extraAlt...)
}

func MustLoadCryptoRaw(ca bool, alg, ext string, extra ...string) []byte {
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

	return data
}

func MustParseCertificateChain(blocks ...[]byte) schema.X509CertificateChain {
	buf := &bytes.Buffer{}

	for _, block := range blocks {
		buf.Write(block)
	}

	var (
		decoded any
		err     error
	)

	if decoded, err = utils.ParseX509FromPEMRecursive(buf.Bytes()); err != nil {
		panic(err)
	}

	switch cert := decoded.(type) {
	case *x509.Certificate:
		return schema.NewX509CertificateChainFromCerts([]*x509.Certificate{cert})
	case []*x509.Certificate:
		return schema.NewX509CertificateChainFromCerts(cert)
	default:
		panic(fmt.Errorf("the key was not a *x509.Certificate or []*x509.Certificate, it's a %T", cert))
	}
}

var (
	x509CertificateChainRSA2048, x509CertificateChainRSA4096 schema.X509CertificateChain
	x509PrivateKeyRSA2048, x509PrivateKeyRSA4096             *rsa.PrivateKey

	x509CertificateChainECDSAP256, x509CertificateChainECDSAP384, x509CertificateChainECDSAP521 schema.X509CertificateChain
	x509PrivateKeyECDSAP256, x509PrivateKeyECDSAP384, x509PrivateKeyECDSAP521                   *ecdsa.PrivateKey
)

func init() {
	tOpenIDConnectPBKDF2ClientSecret = MustDecodeSecret("$pbkdf2-sha512$100000$cfNEo93VkIUIvaXHqetFoQ$O6qFLAlwCMz6.hv9XqUEPnMtrFxODw70T7bmnfTzfNPi3iXbgUEmGiyA6msybOfmj7m3QJS6lLy4DglgJifkKw")
	tOpenIDConnectPlainTextClientSecret = MustDecodeSecret("$plaintext$client-secret")

	x509CertificateChainRSA2048, x509PrivateKeyRSA2048 = MustLoadRSACryptoSet(false, "2048")
	x509CertificateChainRSA4096, x509PrivateKeyRSA4096 = MustLoadRSACryptoSet(false, "4096")

	x509CertificateChainECDSAP256, x509PrivateKeyECDSAP256 = MustLoadECDSACryptoSet(false, "P256")
	x509CertificateChainECDSAP384, x509PrivateKeyECDSAP384 = MustLoadECDSACryptoSet(false, "P384")
	x509CertificateChainECDSAP521, x509PrivateKeyECDSAP521 = MustLoadECDSACryptoSet(false, "P521")
}
