// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package validator

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"fmt"
	"strings"

	"github.com/weppos/publicsuffix-go/publicsuffix"

	"authelia.com/provider/jose"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/expression"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/utils"
)

func isUserAttributeDefinitionNameValid(attribute string, config *schema.Configuration) bool {
	if expression.IsReservedAttribute(attribute) {
		return false
	}

	if config.AuthenticationBackend.LDAP != nil {
		for attrname, attr := range config.AuthenticationBackend.LDAP.Attributes.Extra {
			if attr.Name != "" {
				if attr.Name == attribute {
					return false
				}
			} else if attrname == attribute {
				return false
			}
		}
	}

	if config.AuthenticationBackend.File != nil {
		for attrname := range config.AuthenticationBackend.File.ExtraAttributes {
			if attrname == attribute {
				return false
			}
		}
	}

	return true
}

func boolApply(current, new bool) bool {
	if current || new {
		return true
	}

	return false
}

func isCookieDomainAPublicSuffix(domain string) (valid bool) {
	domain = strings.TrimLeft(domain, ".")

	_, err := publicsuffix.Domain(domain)
	if err != nil {
		return err.Error() == fmt.Sprintf(errFmtCookieDomainInPSL, domain)
	}

	return false
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

// JWKProperties represents the properties of a JSON Web Key.
type JWKProperties struct {
	Use       string
	Algorithm string
	Bits      int
	Curve     elliptic.Curve
}

//nolint:gocyclo
func schemaJWKGetProperties(jwk schema.JWK) (properties *JWKProperties, err error) {
	if jwk.Use == oidc.KeyUseEncryption {
		return schemaJWKGetPropertiesEnc(jwk)
	}

	switch key := jwk.Key.(type) {
	case nil:
		return nil, nil
	case []byte:
		return nil, fmt.Errorf("symmetric keys are not permitted for signing")
	case ed25519.PrivateKey, ed25519.PublicKey:
		return &JWKProperties{oidc.KeyUseSignature, oidc.SigningAlgEd25519, -1, nil}, nil
	case *rsa.PrivateKey:
		if key.N == nil {
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
		if alg, ok := oidc.SigningAlgFromMLDSAKey(key); ok {
			return &JWKProperties{oidc.KeyUseSignature, alg, -1, nil}, nil
		}

		return nil, fmt.Errorf("the key type '%T' is unknown or not valid for the configuration", key)
	}
}

func schemaJWKGetPropertiesEnc(jwk schema.JWK) (properties *JWKProperties, err error) {
	switch key := jwk.Key.(type) {
	case nil:
		return nil, nil
	case []byte:
		switch n := len(key); n {
		case 256:
			return &JWKProperties{oidc.KeyUseEncryption, oidc.EncryptionAlgA256GCMKW, n, nil}, nil
		case 192:
			return &JWKProperties{oidc.KeyUseEncryption, oidc.EncryptionAlgA192GCMKW, n, nil}, nil
		case 128:
			return &JWKProperties{oidc.KeyUseEncryption, oidc.EncryptionAlgA128GCMKW, n, nil}, nil
		default:
			if n < 32 {
				return nil, fmt.Errorf("invalid symmetric key length of %d but the minimum is 32", n)
			}

			return &JWKProperties{oidc.KeyUseEncryption, oidc.EncryptionAlgDirect, n, nil}, nil
		}
	case ed25519.PrivateKey, ed25519.PublicKey:
		return &JWKProperties{}, nil
	case *rsa.PrivateKey:
		if key.N == nil {
			return &JWKProperties{oidc.KeyUseEncryption, oidc.EncryptionAlgRSAOAEP256, 0, nil}, nil
		}

		return &JWKProperties{oidc.KeyUseEncryption, oidc.EncryptionAlgRSAOAEP256, key.Size(), nil}, nil
	case *rsa.PublicKey:
		if key.N == nil {
			return &JWKProperties{oidc.KeyUseEncryption, oidc.EncryptionAlgRSAOAEP256, 0, nil}, nil
		}

		return &JWKProperties{oidc.KeyUseEncryption, oidc.EncryptionAlgRSAOAEP256, key.Size(), nil}, nil
	case *ecdsa.PublicKey:
		return &JWKProperties{oidc.KeyUseEncryption, oidc.EncryptionAlgECDHESA256KW, -1, key.Curve}, nil
	case *ecdsa.PrivateKey:
		return &JWKProperties{oidc.KeyUseEncryption, oidc.EncryptionAlgECDHESA256KW, -1, key.Curve}, nil
	default:
		return nil, fmt.Errorf("the key type '%T' is unknown or not valid for the configuration", key)
	}
}

func jwkValidateJOSE(jwk schema.JWK) (err error) {
	key := jose.JSONWebKey{
		Key:       jwk.Key,
		KeyID:     jwk.KeyID,
		Algorithm: jwk.Algorithm,
		Use:       jwk.Use,
	}

	if !key.Valid() {
		return fmt.Errorf("the key parameters are invalid or inconsistent")
	}

	public := key.Public()

	if _, err = public.MarshalJSON(); err != nil {
		return fmt.Errorf("the public key can't be published: %w", err)
	}

	if jwk.Use != oidc.KeyUseSignature || !utils.IsStringInSlice(jwk.Algorithm, validOIDCIssuerJWKSigningAlgs) {
		return nil
	}

	if err = jwkSigningAlgSupportsKey(jwk.Algorithm, jwk.Key); err != nil || key.IsPublic() {
		return err
	}

	var (
		signer    jose.Signer
		signature *jose.JSONWebSignature
	)

	if signer, err = jose.NewSigner(jose.SigningKey{Algorithm: jose.SignatureAlgorithm(jwk.Algorithm), Key: &key}, nil); err != nil {
		return fmt.Errorf("the key can't sign using the '%s' algorithm: %w", jwk.Algorithm, err)
	}

	if signature, err = signer.Sign([]byte("authelia")); err != nil {
		return fmt.Errorf("the key can't sign using the '%s' algorithm: %w", jwk.Algorithm, err)
	}

	if _, err = signature.Verify(&public); err != nil {
		return fmt.Errorf("the key produced a signature using the '%s' algorithm which can't be verified using its public key: %w", jwk.Algorithm, err)
	}

	return nil
}

func jwkSigningAlgSupportsKey(alg string, key schema.CryptographicKey) (err error) {
	switch alg {
	case oidc.SigningAlgRSAUsingSHA256, oidc.SigningAlgRSAUsingSHA384, oidc.SigningAlgRSAUsingSHA512,
		oidc.SigningAlgRSAPSSUsingSHA256, oidc.SigningAlgRSAPSSUsingSHA384, oidc.SigningAlgRSAPSSUsingSHA512:
		switch key.(type) {
		case *rsa.PublicKey, *rsa.PrivateKey:
			return nil
		}

		return fmt.Errorf("the '%s' algorithm requires an RSA key but the key is a %s", alg, jwkKeyDescription(key))
	case oidc.SigningAlgECDSAUsingP256AndSHA256, oidc.SigningAlgECDSAUsingP384AndSHA384, oidc.SigningAlgECDSAUsingP521AndSHA512:
		curve := map[string]elliptic.Curve{
			oidc.SigningAlgECDSAUsingP256AndSHA256: elliptic.P256(),
			oidc.SigningAlgECDSAUsingP384AndSHA384: elliptic.P384(),
			oidc.SigningAlgECDSAUsingP521AndSHA512: elliptic.P521(),
		}[alg]

		var actual elliptic.Curve

		switch k := key.(type) {
		case *ecdsa.PublicKey:
			actual = k.Curve
		case *ecdsa.PrivateKey:
			actual = k.Curve
		}

		if actual == curve {
			return nil
		}

		return fmt.Errorf("the '%s' algorithm requires an ECDSA %s key but the key is a %s", alg, curve.Params().Name, jwkKeyDescription(key))
	case oidc.SigningAlgEd25519, oidc.SigningAlgEdDSA:
		switch key.(type) {
		case ed25519.PublicKey, ed25519.PrivateKey:
			return nil
		}

		return fmt.Errorf("the '%s' algorithm requires an Ed25519 key but the key is a %s", alg, jwkKeyDescription(key))
	default:
		if actual, ok := oidc.SigningAlgFromMLDSAKey(key); ok && actual == alg {
			return nil
		}

		return fmt.Errorf("the '%s' algorithm requires a %s key but the key is a %s", alg, alg, jwkKeyDescription(key))
	}
}

func jwkKeyDescription(key schema.CryptographicKey) string {
	switch k := key.(type) {
	case *rsa.PublicKey, *rsa.PrivateKey:
		return "RSA key"
	case *ecdsa.PublicKey:
		return fmt.Sprintf("ECDSA %s key", k.Curve.Params().Name)
	case *ecdsa.PrivateKey:
		return fmt.Sprintf("ECDSA %s key", k.Curve.Params().Name)
	case ed25519.PublicKey, ed25519.PrivateKey:
		return "Ed25519 key"
	default:
		if alg, ok := oidc.SigningAlgFromMLDSAKey(key); ok {
			return fmt.Sprintf("%s key", alg)
		}

		return fmt.Sprintf("%T", key)
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
		if utils.IsMLDSAPublicKey(k) {
			j.Key = k

			break
		}

		return "", nil
	}

	if alg == "" {
		alg = props.Algorithm
	}

	alg = strings.ReplaceAll(alg, "+", ".")

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
		if kid != jwk.KeyID {
			continue
		}

		if paired, ok := oidc.SigningAlgEdwardsPair(alg); ok && paired == jwk.Algorithm {
			return alg
		}

		return jwk.Algorithm
	}

	return alg
}

func appendOIDCSigningAlg(algs []string, alg string) (out []string) {
	out = algs

	if !utils.IsStringInSlice(alg, out) {
		out = append(out, alg)
	}

	if paired, ok := oidc.SigningAlgEdwardsPair(alg); ok && !utils.IsStringInSlice(paired, out) {
		out = append(out, paired)
	}

	return out
}
