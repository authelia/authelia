//go:build go1.27

package utils

import (
	"crypto/mldsa"
	"crypto/x509"
	"fmt"
	"strings"
)

// MLDSAParameterSets returns the names of the ML-DSA parameter sets this build supports, in increasing order of
// security category. ML-DSA requires crypto/mldsa which landed in Go 1.27, so a build produced by an earlier
// toolchain supports none of them and this is empty.
func MLDSAParameterSets() []string {
	return []string{KeyMLDSAParameters44, KeyMLDSAParameters65, KeyMLDSAParameters87}
}

// MLDSAParametersFromString returns the crypto/mldsa parameter set for the given name. The bare security level is
// accepted alongside the full name, so '44' and 'ML-DSA-44' both select the same parameter set. The ok result is
// false when the name matches no parameter set, and is always false on a build without ML-DSA support.
func MLDSAParametersFromString(parameters string) (params mldsa.Parameters, ok bool) {
	switch strings.TrimPrefix(strings.ToUpper(parameters), "ML-DSA-") {
	case "44":
		return mldsa.MLDSA44(), true
	case "65":
		return mldsa.MLDSA65(), true
	case "87":
		return mldsa.MLDSA87(), true
	default:
		return mldsa.Parameters{}, false
	}
}

// GenerateMLDSAKey returns a new ML-DSA private key for the named parameter set. Unlike the other key types this
// does not take a randomness source: crypto/mldsa seeds a key from crypto/rand itself and exposes no way to supply
// one.
func GenerateMLDSAKey(parameters string) (key any, err error) {
	params, ok := MLDSAParametersFromString(parameters)
	if !ok {
		return nil, fmt.Errorf("invalid parameters '%s' were specified: parameters must be %s", parameters, StringJoinOr(MLDSAParameterSets()))
	}

	if key, err = mldsa.GenerateKey(params); err != nil {
		return nil, err
	}

	return key, nil
}

// MLDSASignatureAlgorithmFromString returns the [x509.SignatureAlgorithm] for the named ML-DSA parameter set. ML-DSA
// offers no choice of hash or padding, so the parameter set alone determines it.
func MLDSASignatureAlgorithmFromString(parameters string) (alg x509.SignatureAlgorithm) {
	switch strings.TrimPrefix(strings.ToUpper(parameters), "ML-DSA-") {
	case "44":
		return x509.MLDSA44
	case "65":
		return x509.MLDSA65
	case "87":
		return x509.MLDSA87
	default:
		return x509.UnknownSignatureAlgorithm
	}
}

// MLDSAParameterSetFromKey returns the name of the parameter set of the given ML-DSA public or private key. The ok
// result is false when the key is not an ML-DSA key at all.
func MLDSAParameterSetFromKey(key any) (parameters string, ok bool) {
	var params mldsa.Parameters

	switch k := key.(type) {
	case *mldsa.PrivateKey:
		params = k.PublicKey().Parameters()
	case *mldsa.PublicKey:
		params = k.Parameters()
	default:
		return "", false
	}

	switch params {
	case mldsa.MLDSA44():
		return KeyMLDSAParameters44, true
	case mldsa.MLDSA65():
		return KeyMLDSAParameters65, true
	case mldsa.MLDSA87():
		return KeyMLDSAParameters87, true
	default:
		return "", false
	}
}

// IsMLDSAPrivateKey returns true if the given key is an ML-DSA private key.
func IsMLDSAPrivateKey(key any) (ok bool) {
	_, ok = key.(*mldsa.PrivateKey)

	return ok
}

// IsMLDSAPublicKey returns true if the given key is an ML-DSA public key.
func IsMLDSAPublicKey(key any) (ok bool) {
	_, ok = key.(*mldsa.PublicKey)

	return ok
}

// PublicKeyAlgorithmMLDSA returns the [x509.PublicKeyAlgorithm] for ML-DSA. The ok result is false on a build
// without ML-DSA support, where the constant does not exist.
func PublicKeyAlgorithmMLDSA() (alg x509.PublicKeyAlgorithm, ok bool) {
	return x509.MLDSA, true
}
