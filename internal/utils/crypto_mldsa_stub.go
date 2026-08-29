//go:build !go1.27

package utils

import (
	"crypto/x509"
	"fmt"
)

// This file provides the ML-DSA seam for toolchains without crypto/mldsa, which was added in Go 1.27. No parameter
// set is offered and every hook reports that it did not recognize the key, so a build produced by an earlier
// toolchain neither generates ML-DSA keys nor claims to understand one.
//
// Each identifier here must have a counterpart in crypto_mldsa.go.

// MLDSAParameterSets returns the names of the ML-DSA parameter sets this build supports, which is none of them.
func MLDSAParameterSets() []string {
	return nil
}

// GenerateMLDSAKey returns a new ML-DSA private key for the named parameter set.
func GenerateMLDSAKey(parameters string) (key any, err error) {
	return nil, fmt.Errorf("generating ML-DSA private keys requires this binary to be built with Go 1.27 or later")
}

// MLDSASignatureAlgorithmFromString returns the [x509.SignatureAlgorithm] for the named ML-DSA parameter set.
func MLDSASignatureAlgorithmFromString(parameters string) (alg x509.SignatureAlgorithm) {
	return x509.UnknownSignatureAlgorithm
}

// MLDSAParameterSetFromKey returns the name of the parameter set of the given ML-DSA public or private key.
func MLDSAParameterSetFromKey(key any) (parameters string, ok bool) {
	return "", false
}

// IsMLDSAPrivateKey returns true if the given key is an ML-DSA private key.
func IsMLDSAPrivateKey(key any) (ok bool) {
	return false
}

// IsMLDSAPublicKey returns true if the given key is an ML-DSA public key.
func IsMLDSAPublicKey(key any) (ok bool) {
	return false
}

// PublicKeyAlgorithmMLDSA returns the [x509.PublicKeyAlgorithm] for ML-DSA.
func PublicKeyAlgorithmMLDSA() (alg x509.PublicKeyAlgorithm, ok bool) {
	return x509.UnknownPublicKeyAlgorithm, false
}
