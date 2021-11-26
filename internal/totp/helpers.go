package totp

import (
	"github.com/pquerna/otp"
)

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
