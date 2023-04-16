package validator

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"fmt"
	"strings"

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

func strJoinOr(items []string) string {
	return strJoinComma("or", items)
}

func strJoinAnd(items []string) string {
	return strJoinComma("and", items)
}

func strJoinComma(word string, items []string) string {
	if word == "" {
		return buildJoinedString(",", "", "'", items)
	}

	return buildJoinedString(",", word, "'", items)
}

func buildJoinedString(sep, sepFinal, quote string, items []string) string {
	n := len(items)

	if n == 0 {
		return ""
	}

	b := &strings.Builder{}

	for i := 0; i < n; i++ {
		if quote != "" {
			b.WriteString(quote)
		}

		b.WriteString(items[i])

		if quote != "" {
			b.WriteString(quote)
		}

		if i == (n - 1) {
			continue
		}

		if sep != "" {
			if sepFinal == "" || n != 2 {
				b.WriteString(sep)
			}

			b.WriteString(" ")
		}

		if sepFinal != "" && i == (n-2) {
			b.WriteString(strings.Trim(sepFinal, " "))
			b.WriteString(" ")
		}
	}

	return b.String()
}

func validateList(values, valid []string, chkDuplicate bool) (invalid, duplicates []string) { //nolint:unparam
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
		return nil, fmt.Errorf("private key is nil")
	case *rsa.PrivateKey:
		return &JWKProperties{oidc.KeyUseSignature, oidc.SigningAlgRSAUsingSHA256, key.Size(), nil}, nil
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
