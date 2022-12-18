package schema

import (
	"time"
)

type IdentityValidation struct {
	ResetPassword          ResetPasswordIdentityValidation          `koanf:"reset_password"`
	CredentialRegistration CredentialRegistrationIdentityValidation `koanf:"credential_registration"`
}

type ResetPasswordIdentityValidation struct {
	EmailExpiration time.Duration `koanf:"email_expiration"`
}

type CredentialRegistrationIdentityValidation struct {
	EmailExpiration     time.Duration `koanf:"email_expiration"`
	ElevationExpiration time.Duration `koanf:"elevation_expiration"`
	OTPCharacters       int           `koanf:"otp_characters"`
	Skip2FA             bool          `koanf:"skip_2fa"`
}
