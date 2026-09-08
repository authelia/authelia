//go:build go1.27

package oidc

import (
	"crypto/mldsa"
)

// SigningAlgsMLDSA are the ML-DSA JWS algorithms this build supports, in increasing order of parameter set. ML-DSA
// requires crypto/mldsa which landed in Go 1.27, so a build produced by an earlier toolchain supports none of them
// and this is empty. Every list of algorithms this Authorization Server advertises or accepts is extended with it
// rather than naming the algorithms directly, so what is offered can never exceed what the binary can perform.
var SigningAlgsMLDSA = []string{SigningAlgMLDSA44, SigningAlgMLDSA65, SigningAlgMLDSA87}

// SigningAlgFromMLDSAKey returns the JWS 'alg' value for the given ML-DSA public or private key. An ML-DSA JSON Web
// Key carries no curve, so the parameter set of the key is the only thing which identifies the algorithm it belongs
// to. The ok result is false when the key is not an ML-DSA key at all.
func SigningAlgFromMLDSAKey(key any) (alg string, ok bool) {
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
		return SigningAlgMLDSA44, true
	case mldsa.MLDSA65():
		return SigningAlgMLDSA65, true
	case mldsa.MLDSA87():
		return SigningAlgMLDSA87, true
	default:
		return "", false
	}
}
