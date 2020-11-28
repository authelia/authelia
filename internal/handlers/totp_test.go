package handlers

import (
	"testing"

	"github.com/pquerna/otp"
	"github.com/stretchr/testify/assert"
)

func TestShouldGetCorrectTOTPAlgo(t *testing.T) {
	algo, err := AlgorithmStringToOTPAlgorithm("md5")

	assert.NoError(t, err)
	assert.Equal(t, algo, otp.AlgorithmMD5)

	algo, err = AlgorithmStringToOTPAlgorithm("sha1")

	assert.NoError(t, err)
	assert.Equal(t, algo, otp.AlgorithmSHA1)

	algo, err = AlgorithmStringToOTPAlgorithm("sha256")

	assert.NoError(t, err)
	assert.Equal(t, algo, otp.AlgorithmSHA256)

	algo, err = AlgorithmStringToOTPAlgorithm("sha512")

	assert.NoError(t, err)
	assert.Equal(t, algo, otp.AlgorithmSHA512)
}

func TestShouldReturnErrorAndSHA1OnInvalidAlgorithm(t *testing.T) {
	algo, err := AlgorithmStringToOTPAlgorithm("aes")

	assert.Equal(t, algo, otp.AlgorithmSHA1)
	assert.EqualError(t, err, "unknown OTP algorithm")
}
