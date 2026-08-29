//go:build !go1.27

package oidc

// This file provides the ML-DSA seam for toolchains without crypto/mldsa, which was added in Go 1.27. Every hook
// reports that it did not recognize the key and no algorithm is offered, so a build produced by an earlier toolchain
// neither advertises ML-DSA nor accepts it in configuration.
//
// Each identifier here must have a counterpart in mldsa.go.

// SigningAlgsMLDSA are the ML-DSA JWS algorithms this build supports, which is none of them.
var SigningAlgsMLDSA []string

// SigningAlgFromMLDSAKey returns the JWS 'alg' value for the given ML-DSA public or private key.
func SigningAlgFromMLDSAKey(key any) (alg string, ok bool) {
	return "", false
}
