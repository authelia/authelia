// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package appauthenticator

import (
	"context"
	"crypto/x509"
	"fmt"
	"net/http"
	"time"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

const (
	// AttestationFormatNone is the attestation format sent when the device does not attest.
	AttestationFormatNone = "none"

	// AttestationFormatAppleAppAttest is the Apple App Attest attestation format.
	AttestationFormatAppleAppAttest = "apple-appattest"

	// AttestationFormatAndroidKey is the Android Key Attestation attestation format.
	AttestationFormatAndroidKey = "android-key"

	// AndroidAttestationStatusURL is the default Google attestation status list location.
	AndroidAttestationStatusURL = "https://android.googleapis.com/attestation/status"
)

// AttestationInput is the attestation material decoded from a registration confirmation.
type AttestationInput struct {
	Format string

	KeyID, AttestationObject string

	ApprovalChain, TransportChain []string

	ApprovalSPKI, TransportSPKI []byte

	Challenge string

	RegistrationString string
}

// AttestationResult is the outcome of a successful attestation verification.
type AttestationResult struct {
	Format string

	Identity, Detail string

	KeyID []byte

	Attested bool
}

// VerifierOption configures a Verifier.
type VerifierOption func(v *Verifier)

// WithClock sets the clock used for certificate validity and status list caching.
func WithClock(c clock.Provider) VerifierOption {
	return func(v *Verifier) {
		v.clock = c
	}
}

// WithHTTPClient sets the HTTP client used to fetch the Android attestation status list.
func WithHTTPClient(client *http.Client) VerifierOption {
	return func(v *Verifier) {
		if client != nil {
			v.revocation.client = client
		}
	}
}

// WithAndroidRoots replaces the pinned Android hardware attestation roots.
func WithAndroidRoots(roots []*x509.Certificate) VerifierOption {
	return func(v *Verifier) {
		v.androidRoots = roots
	}
}

// WithAppleRoot replaces the pinned Apple App Attestation root.
func WithAppleRoot(root *x509.Certificate) VerifierOption {
	return func(v *Verifier) {
		v.appleRoot = root
	}
}

// WithRevocationURL sets the location of the Android attestation status list.
func WithRevocationURL(url string) VerifierOption {
	return func(v *Verifier) {
		v.revocation.url = url
	}
}

type formatVerifier func(ctx context.Context, in AttestationInput) (AttestationResult, error)

// Verifier verifies registration attestations according to the configured policy. It is safe for concurrent use and
// is meant to be long-lived, as it caches the Android attestation status list.
type Verifier struct {
	config       schema.AppAuthenticatorAttestation
	clock        clock.Provider
	androidRoots []*x509.Certificate
	appleRoot    *x509.Certificate
	revocation   *androidRevocation
	formats      map[string]formatVerifier
}

// NewVerifier returns a Verifier for the given attestation configuration.
func NewVerifier(config schema.AppAuthenticatorAttestation, opts ...VerifierOption) *Verifier {
	v := &Verifier{
		config:       config,
		clock:        clock.New(),
		androidRoots: GoogleAttestationRoots(),
		appleRoot:    AppleAppAttestationRoot(),
		revocation: &androidRevocation{
			url:    AndroidAttestationStatusURL,
			client: &http.Client{Timeout: 10 * time.Second},
		},
	}

	v.formats = map[string]formatVerifier{
		AttestationFormatAppleAppAttest: v.verifyAppleAppAttest,
		AttestationFormatAndroidKey:     v.verifyAndroidKey,
	}

	for _, opt := range opts {
		opt(v)
	}

	return v
}

// Verify applies the attestation policy of the configured mode to the input. The error describes the reason for the
// server log only; callers answer every failure with the same error on the wire.
func (v *Verifier) Verify(ctx context.Context, in AttestationInput) (AttestationResult, error) {
	mode := v.mode()

	if in.Format == AttestationFormatNone {
		if mode == schema.AppAuthenticatorAttestationModeRequired {
			return AttestationResult{}, fmt.Errorf("attestation format '%s' is not accepted in mode '%s'", in.Format, mode)
		}

		return AttestationResult{Format: AttestationFormatNone}, nil
	}

	verify, ok := v.formats[in.Format]
	if !ok {
		return AttestationResult{}, fmt.Errorf("attestation format '%s' is not supported", in.Format)
	}

	if mode == schema.AppAuthenticatorAttestationModeNone {
		return AttestationResult{}, fmt.Errorf("attestation format '%s' is not accepted in mode '%s'", in.Format, mode)
	}

	result, err := verify(ctx, in)
	if err != nil {
		return AttestationResult{}, fmt.Errorf("error verifying attestation format '%s': %w", in.Format, err)
	}

	return result, nil
}

func (v *Verifier) mode() string {
	switch v.config.Mode {
	case schema.AppAuthenticatorAttestationModeNone, schema.AppAuthenticatorAttestationModePreferred:
		return v.config.Mode
	default:
		return schema.AppAuthenticatorAttestationModeRequired
	}
}
