package totp

import (
	"github.com/authelia/otp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func otpStringToAlgo(in string) (algorithm otp.Algorithm) {
	switch in {
	case schema.TOTPAlgorithmSHA1:
		return otp.AlgorithmSHA1
	case schema.TOTPAlgorithmSHA256:
		return otp.AlgorithmSHA256
	case schema.TOTPAlgorithmSHA512:
		return otp.AlgorithmSHA512
	default:
		return otp.AlgorithmSHA1
	}
}
