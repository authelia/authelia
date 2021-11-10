package handlers

import (
	"errors"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/authelia/authelia/v4/internal/models"
)

// TOTPVerifier is the interface for verifying TOTPs.
type TOTPVerifier interface {
	Verify(config *models.TOTPConfiguration, token string) (bool, error)
}

// TOTPVerifierImpl the production implementation for TOTP verification.
type TOTPVerifierImpl struct {
	Period uint
	Skew   uint
}

// Verify verifies TOTPs.
func (tv *TOTPVerifierImpl) Verify(config *models.TOTPConfiguration, token string) (bool, error) {
	if config == nil {
		return false, errors.New("config not provided")
	}

	opts := totp.ValidateOpts{
		Period:    uint(config.Period),
		Skew:      tv.Skew,
		Digits:    otp.Digits(config.Digits),
		Algorithm: otpStringToAlgo(config.Algorithm),
	}

	return totp.ValidateCustom(token, string(config.Secret), time.Now().UTC(), opts)
}

func otpAlgoToString(algorithm otp.Algorithm) (out string) {
	switch algorithm {
	case otp.AlgorithmSHA1:
		return totpAlgoSHA1
	case otp.AlgorithmSHA256:
		return totpAlgoSHA256
	case otp.AlgorithmSHA512:
		return totpAlgoSHA512
	default:
		return ""
	}
}

func otpStringToAlgo(in string) (algorithm otp.Algorithm) {
	switch in {
	case totpAlgoSHA1:
		return otp.AlgorithmSHA1
	case totpAlgoSHA256:
		return otp.AlgorithmSHA256
	case totpAlgoSHA512:
		return otp.AlgorithmSHA512
	default:
		return otp.AlgorithmSHA1
	}
}
