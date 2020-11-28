package handlers

import (
	"errors"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/authelia/authelia/internal/configuration/schema"
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
func (tv *TOTPVerifierImpl) Verify(token, secret, algorithmStr string) (bool, error) {
	algorithm, _ := AlgorithmStringToOTPAlgorithm(algorithmStr)

	opts := totp.ValidateOpts{
		Period:    tv.Period,
		Skew:      tv.Skew,
		Digits:    otp.DigitsSix,
		Algorithm: algorithm,
	}

	return totp.ValidateCustom(token, secret, time.Now().UTC(), opts)
}

// AlgorithmStringToOTPAlgorithm converts a string into a valid OTP algorithm.
func AlgorithmStringToOTPAlgorithm(algorithmStr string) (algorithm otp.Algorithm, err error) {
	switch algorithmStr {
	case schema.MD5:
		return otp.AlgorithmMD5, nil
	case schema.SHA1:
		return otp.AlgorithmSHA1, nil
	case schema.SHA256:
		return otp.AlgorithmSHA256, nil
	case schema.SHA512:
		return otp.AlgorithmSHA512, nil
	default:
		return otp.AlgorithmSHA1, errors.New("unknown OTP algorithm")
	}
}
