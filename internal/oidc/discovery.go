// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc

// SigningAlgValuesSupported returns the JWS algorithms this Authorization Server can produce or verify a signature
// with, followed by any extra values a particular metadata member also permits such as SigningAlgNone. The ML-DSA
// algorithms are appended from SigningAlgsMLDSA rather than named here, so a build whose toolchain has no
// crypto/mldsa advertises only what it can actually perform.
func SigningAlgValuesSupported(extra ...string) (algs []string) {
	algs = []string{
		SigningAlgHMACUsingSHA256,
		SigningAlgHMACUsingSHA384,
		SigningAlgHMACUsingSHA512,
		SigningAlgRSAUsingSHA256,
		SigningAlgRSAUsingSHA384,
		SigningAlgRSAUsingSHA512,
		SigningAlgECDSAUsingP256AndSHA256,
		SigningAlgECDSAUsingP384AndSHA384,
		SigningAlgECDSAUsingP521AndSHA512,
		SigningAlgRSAPSSUsingSHA256,
		SigningAlgRSAPSSUsingSHA384,
		SigningAlgRSAPSSUsingSHA512,
		SigningAlgEd25519,
		SigningAlgEdDSA,
	}

	algs = append(algs, SigningAlgsMLDSA...)

	return append(algs, extra...)
}

// DPoPSigningAlgValuesSupported returns the default JWS algorithms an RFC9449 proof may be signed with, used when the
// allowed algorithms are not configured. The symmetric algorithms SigningAlgValuesSupported lists are deliberately
// absent: a proof is signed by a key the client holds and proves possession of, which a shared secret cannot
// demonstrate.
func DPoPSigningAlgValuesSupported() (algs []string) {
	return []string{
		SigningAlgRSAUsingSHA256,
		SigningAlgRSAUsingSHA384,
		SigningAlgRSAUsingSHA512,
		SigningAlgECDSAUsingP256AndSHA256,
		SigningAlgECDSAUsingP384AndSHA384,
		SigningAlgECDSAUsingP521AndSHA512,
		SigningAlgRSAPSSUsingSHA256,
		SigningAlgRSAPSSUsingSHA384,
		SigningAlgRSAPSSUsingSHA512,
		SigningAlgEd25519,
		SigningAlgEdDSA,
	}
}

// EncryptionAlgValuesSupported returns the JWE key management algorithms this Authorization Server supports.
func EncryptionAlgValuesSupported() (algs []string) {
	return []string{
		EncryptionAlgRSA15,
		EncryptionAlgRSAOAEP,
		EncryptionAlgRSAOAEP256,
		EncryptionAlgA128KW,
		EncryptionAlgA192KW,
		EncryptionAlgA256KW,
		EncryptionAlgDirect,
		EncryptionAlgECDHES,
		EncryptionAlgECDHESA128KW,
		EncryptionAlgECDHESA192KW,
		EncryptionAlgECDHESA256KW,
		EncryptionAlgA128GCMKW,
		EncryptionAlgA192GCMKW,
		EncryptionAlgA256GCMKW,
		EncryptionAlgPBES2HS256A128KW,
		EncryptionAlgPBES2HS284A192KW,
		EncryptionAlgPBES2HS512A256KW,
	}
}

// EncryptionEncValuesSupported returns the JWE content encryption algorithms this Authorization Server supports.
func EncryptionEncValuesSupported() (encs []string) {
	return []string{
		EncryptionEncA128CBCHS256,
		EncryptionEncA192CBCHS384,
		EncryptionEncA256CBCHS512,
		EncryptionEncA128GCM,
		EncryptionEncA192GCM,
		EncryptionEncA256GCM,
	}
}
