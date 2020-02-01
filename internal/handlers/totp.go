package handlers

import (
	"github.com/pquerna/otp/totp"
)

type TOTPVerifier interface {
	Verify(token, secret string) bool
}

type TOTPVerifierImpl struct{}

func (tv *TOTPVerifierImpl) Verify(token, secret string) bool {
	return totp.Validate(token, secret)
}
