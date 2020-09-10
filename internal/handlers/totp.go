package handlers

import (
	"errors"
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
func (tv *TOTPVerifierImpl) Verify(token, secret, algo string) (bool, error) {
	algorithm, _ := AlgorithmStringToOTPAlgorithm(algo)

	opts := totp.ValidateOpts{
		Period:    tv.Period,
		Skew:      tv.Skew,
		Digits:    otp.DigitsSix,
		Algorithm: algorithm,
	}

	return totp.ValidateCustom(token, secret, time.Now().UTC(), opts)
}

// AlgorithmStringToOTPAlgorithm converts a string into a valid OTP algorithm.
func AlgorithmStringToOTPAlgorithm(algo string) (algorithm otp.Algorithm, err error) {
	switch algo {
	case "md5":
		return otp.AlgorithmMD5, nil
	case "sha1":
		return otp.AlgorithmSHA1, nil
	case "sha256":
		return otp.AlgorithmSHA256, nil
	case "sha512":
		return otp.AlgorithmSHA512, nil
	default:
		return otp.AlgorithmSHA1, errors.New("unknown OTP algorithm")
	}
}
