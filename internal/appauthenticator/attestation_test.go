// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package appauthenticator

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func newStubVerifier(mode string, stubErr error) *Verifier {
	v := NewVerifier(schema.AppAuthenticatorAttestation{Mode: mode})

	v.formats[AttestationFormatAndroidKey] = func(_ context.Context, _ AttestationInput) (AttestationResult, error) {
		if stubErr != nil {
			return AttestationResult{}, stubErr
		}

		return AttestationResult{Format: AttestationFormatAndroidKey, Identity: "com.example.app", Detail: "StrongBox", Attested: true}, nil
	}

	return v
}

func TestShouldApplyAttestationPolicy(t *testing.T) {
	testCases := []struct {
		name     string
		mode     string
		format   string
		err      string
		attested bool
	}{
		{"ShouldRejectNoneWhenRequired", schema.AppAuthenticatorAttestationModeRequired, AttestationFormatNone, "attestation format 'none' is not accepted in mode 'required'", false},
		{"ShouldVerifyFormatWhenRequired", schema.AppAuthenticatorAttestationModeRequired, AttestationFormatAndroidKey, "", true},
		{"ShouldRejectUnknownWhenRequired", schema.AppAuthenticatorAttestationModeRequired, "foo", "attestation format 'foo' is not supported", false},
		{"ShouldAcceptNoneWhenPreferred", schema.AppAuthenticatorAttestationModePreferred, AttestationFormatNone, "", false},
		{"ShouldVerifyFormatWhenPreferred", schema.AppAuthenticatorAttestationModePreferred, AttestationFormatAndroidKey, "", true},
		{"ShouldRejectUnknownWhenPreferred", schema.AppAuthenticatorAttestationModePreferred, "foo", "attestation format 'foo' is not supported", false},
		{"ShouldAcceptNoneWhenNone", schema.AppAuthenticatorAttestationModeNone, AttestationFormatNone, "", false},
		{"ShouldRejectFormatWhenNone", schema.AppAuthenticatorAttestationModeNone, AttestationFormatAndroidKey, "attestation format 'android-key' is not accepted in mode 'none'", false},
		{"ShouldRejectUnknownWhenNone", schema.AppAuthenticatorAttestationModeNone, "foo", "attestation format 'foo' is not supported", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v := newStubVerifier(tc.mode, nil)

			result, err := v.Verify(context.Background(), AttestationInput{Format: tc.format})

			if tc.err != "" {
				assert.EqualError(t, err, tc.err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.format, result.Format)
			assert.Equal(t, tc.attested, result.Attested)

			if tc.attested {
				assert.Equal(t, "com.example.app", result.Identity)
			} else {
				assert.Empty(t, result.Identity)
			}
		})
	}
}

func TestShouldRejectFailedVerificationWhenPreferred(t *testing.T) {
	for _, mode := range []string{schema.AppAuthenticatorAttestationModePreferred, schema.AppAuthenticatorAttestationModeRequired} {
		t.Run(mode, func(t *testing.T) {
			v := newStubVerifier(mode, errors.New("bad chain"))

			_, err := v.Verify(context.Background(), AttestationInput{Format: AttestationFormatAndroidKey})

			assert.EqualError(t, err, "error verifying attestation format 'android-key': bad chain")
		})
	}
}

func TestShouldTreatEmptyModeAsRequired(t *testing.T) {
	v := newStubVerifier("", nil)

	_, err := v.Verify(context.Background(), AttestationInput{Format: AttestationFormatNone})

	assert.EqualError(t, err, "attestation format 'none' is not accepted in mode 'required'")
}
