package oidc

import (
	"strings"

	"github.com/ory/fosite"
	"gopkg.in/square/go-jose.v2"
)

// IsPushedAuthorizedRequest returns true if the requester has a PushedAuthorizationRequest redirect_uri value.
func IsPushedAuthorizedRequest(r fosite.Requester, prefix string) bool {
	return strings.HasPrefix(r.GetRequestForm().Get(FormParameterRequestURI), prefix)
}

// SortedSigningAlgs is a sorting type which allows the use of sort.Sort to order a list of OAuth 2.0 Signing Algs.
// Sorting occurs in the order of from within the RFC's.
type SortedSigningAlgs []string

func (algs SortedSigningAlgs) Len() int {
	return len(algs)
}

func (algs SortedSigningAlgs) Less(i, j int) bool {
	return isSigningAlgLess(algs[i], algs[j])
}

func (algs SortedSigningAlgs) Swap(i, j int) {
	algs[i], algs[j] = algs[j], algs[i]
}

type SortedJSONWebKey []jose.JSONWebKey

func (jwks SortedJSONWebKey) Len() int {
	return len(jwks)
}

func (jwks SortedJSONWebKey) Less(i, j int) bool {
	if jwks[i].Algorithm == jwks[j].Algorithm {
		return jwks[i].KeyID < jwks[j].KeyID
	}

	return isSigningAlgLess(jwks[i].Algorithm, jwks[j].Algorithm)
}

func (jwks SortedJSONWebKey) Swap(i, j int) {
	jwks[i], jwks[j] = jwks[j], jwks[i]
}

//nolint:gocyclo // Low importance func.
func isSigningAlgLess(i, j string) bool {
	switch {
	case i == j:
		return false
	case i == SigningAlgNone:
		return false
	case j == SigningAlgNone:
		return true
	default:
		var (
			ip, jp string
			it, jt bool
		)

		if len(i) > 2 {
			it = true
			ip = i[:2]
		}

		if len(j) > 2 {
			jt = true
			jp = j[:2]
		}

		switch {
		case it && jt && ip == jp:
			return i < j
		case ip == SigningAlgPrefixHMAC:
			return true
		case jp == SigningAlgPrefixHMAC:
			return false
		case ip == SigningAlgPrefixRSAPSS:
			return false
		case jp == SigningAlgPrefixRSAPSS:
			return true
		case ip == SigningAlgPrefixRSA:
			return true
		case jp == SigningAlgPrefixRSA:
			return false
		case ip == SigningAlgPrefixECDSA:
			return true
		case jp == SigningAlgPrefixECDSA:
			return false
		default:
			return false
		}
	}
}
