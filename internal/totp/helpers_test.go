package totp

import (
	"testing"

	"github.com/authelia/otp"
	"github.com/stretchr/testify/assert"
)

func TestOTPStringToAlgo(t *testing.T) {
	assert.Equal(t, otp.AlgorithmSHA1, otpStringToAlgo("SHA1"))
	assert.Equal(t, otp.AlgorithmSHA256, otpStringToAlgo("SHA256"))
	assert.Equal(t, otp.AlgorithmSHA512, otpStringToAlgo("SHA512"))
	assert.Equal(t, otp.AlgorithmSHA1, otpStringToAlgo(""))
}
