package handlers

import (
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// TOTPVerifier is the interface for verifying TOTPs.
type TOTPVerifier interface {
	Verify(token, secret string) (bool, error)
}

// TOTPVerifierImpl the production implementation for TOTP verification.
type TOTPVerifierImpl struct {
	Period uint
	Skew   uint
}

// Verify verifies TOTPs.
func (tv *TOTPVerifierImpl) Verify(token, secret string) (bool, error) {
	opts := totp.ValidateOpts{
		Period:    tv.Period,
		Skew:      tv.Skew,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	}
	return totp.ValidateCustom(token, secret, time.Now().UTC(), opts)
}
