//go:build !go1.27

package oidc

// SigningAlgsMLDSA are the ML-DSA JWS algorithms this build supports, which is none of them.
var SigningAlgsMLDSA []string

// SigningAlgFromMLDSAKey returns the JWS 'alg' value for the given ML-DSA public or private key.
func SigningAlgFromMLDSAKey(key any) (alg string, ok bool) {
	return "", false
}
