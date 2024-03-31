package validator

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"fmt"
	"strings"

	"github.com/go-jose/go-jose/v4"
	"golang.org/x/net/publicsuffix"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/utils"
)

func isCookieDomainAPublicSuffix(domain string) (valid bool) {
	var suffix string

	suffix, _ = publicsuffix.PublicSuffix(domain)

	return len(strings.TrimLeft(domain, ".")) == len(suffix)
}

func validateListNotAllowed(values, filter []string) (invalid []string) {
	for _, value := range values {
		if utils.IsStringInSlice(value, filter) {
			invalid = append(invalid, value)
		}
	}

	return invalid
}

func validateList(values, valid []string, chkDuplicate bool) (invalid, duplicates []string) {
	chkValid := len(valid) != 0

	for i, value := range values {
		if chkValid {
			if !utils.IsStringInSlice(value, valid) {
				invalid = append(invalid, value)

				// Skip checking duplicates for invalid values.
				continue
			}
		}

		if chkDuplicate {
			for j, valueAlt := range values {
				if i == j {
					continue
				}

				if value != valueAlt {
					continue
				}

				if utils.IsStringInSlice(value, duplicates) {
					continue
				}

				duplicates = append(duplicates, value)
			}
		}
	}

	return
}

type JWKProperties struct {
	Use       string
	Algorithm string
	Bits      int
	Curve     elliptic.Curve
}

func schemaJWKGetProperties(jwk schema.JWK) (properties *JWKProperties, err error) {
	switch key := jwk.Key.(type) {
	case nil:
		return nil, nil
	case ed25519.PrivateKey, ed25519.PublicKey:
		return &JWKProperties{}, nil
	case *rsa.PrivateKey:
		if key.PublicKey.N == nil {
			return &JWKProperties{oidc.KeyUseSignature, oidc.SigningAlgRSAUsingSHA256, 0, nil}, nil
		}

		return &JWKProperties{oidc.KeyUseSignature, oidc.SigningAlgRSAUsingSHA256, key.Size(), nil}, nil
	case *rsa.PublicKey:
		if key.N == nil {
			return &JWKProperties{oidc.KeyUseSignature, oidc.SigningAlgRSAUsingSHA256, 0, nil}, nil
		}

		return &JWKProperties{oidc.KeyUseSignature, oidc.SigningAlgRSAUsingSHA256, key.Size(), nil}, nil
	case *ecdsa.PublicKey:
		switch key.Curve {
		case elliptic.P256():
			return &JWKProperties{oidc.KeyUseSignature, oidc.SigningAlgECDSAUsingP256AndSHA256, -1, key.Curve}, nil
		case elliptic.P384():
			return &JWKProperties{oidc.KeyUseSignature, oidc.SigningAlgECDSAUsingP384AndSHA384, -1, key.Curve}, nil
		case elliptic.P521():
			return &JWKProperties{oidc.KeyUseSignature, oidc.SigningAlgECDSAUsingP521AndSHA512, -1, key.Curve}, nil
		default:
			return &JWKProperties{oidc.KeyUseSignature, "", -1, key.Curve}, nil
		}
	case *ecdsa.PrivateKey:
		switch key.Curve {
		case elliptic.P256():
			return &JWKProperties{oidc.KeyUseSignature, oidc.SigningAlgECDSAUsingP256AndSHA256, -1, key.Curve}, nil
		case elliptic.P384():
			return &JWKProperties{oidc.KeyUseSignature, oidc.SigningAlgECDSAUsingP384AndSHA384, -1, key.Curve}, nil
		case elliptic.P521():
			return &JWKProperties{oidc.KeyUseSignature, oidc.SigningAlgECDSAUsingP521AndSHA512, -1, key.Curve}, nil
		default:
			return &JWKProperties{oidc.KeyUseSignature, "", -1, key.Curve}, nil
		}
	default:
		return nil, fmt.Errorf("the key type '%T' is unknown or not valid for the configuration", key)
	}
}

func jwkCalculateKID(key schema.CryptographicKey, props *JWKProperties, alg string) (kid string, err error) {
	j := jose.JSONWebKey{}

	switch k := key.(type) {
	case schema.CryptographicPrivateKey:
		j.Key = k.Public()
	case *rsa.PublicKey, *ecdsa.PublicKey, ed25519.PublicKey:
		j.Key = k
	default:
		return "", nil
	}

	if alg == "" {
		alg = props.Algorithm
	}

	var thumbprint []byte

	if thumbprint, err = j.Thumbprint(crypto.SHA256); err != nil {
		return "", err
	}

	if alg == "" {
		return fmt.Sprintf("%x", thumbprint)[:6], nil
	}

	return fmt.Sprintf("%s-%s", fmt.Sprintf("%x", thumbprint)[:6], strings.ToLower(alg)), nil
}

func getResponseObjectAlgFromKID(config *schema.IdentityProvidersOpenIDConnect, kid, alg string) string {
	for _, jwk := range config.JSONWebKeys {
		if kid == jwk.KeyID {
			return jwk.Algorithm
		}
	}

	return alg
}
