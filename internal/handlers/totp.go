package handlers

import (
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// TOTPVerifier is the interface for verifying TOTPs.
type TOTPVerifier interface {
	Verify(token, secret, algorithm string) (bool, error)
}

// TOTPVerifierImpl the production implementation for TOTP verification.
type TOTPVerifierImpl struct {
	Period uint
	Skew   uint
}

// Verify verifies TOTPs.
func (tv *TOTPVerifierImpl) Verify(token, secret, algorithm string) (bool, error) {
	algo := otp.AlgorithmSHA512
	if algorithm == "sha1" {
		algo = otp.AlgorithmSHA1
	}

	opts := totp.ValidateOpts{
		Period:    tv.Period,
		Skew:      tv.Skew,
		Digits:    otp.DigitsSix,
		Algorithm: algo,
	}

	return totp.ValidateCustom(token, secret, time.Now().UTC(), opts)
}
