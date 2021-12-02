package regulation

import "fmt"

// ErrUserIsBanned user is banned error message.
var ErrUserIsBanned = fmt.Errorf("user is banned")

const (
	// AuthType1FA is the string representing an auth log for first-factor authentication.
	AuthType1FA = "1FA"

	// AuthTypeTOTP is the string representing an auth log for second-factor authentication via TOTP.
	AuthTypeTOTP = "TOTP"

	// AuthTypeU2F is the string representing an auth log for second-factor authentication via FIDO/CTAP1/U2F.
	AuthTypeU2F = "U2F"

	// AuthTypeWebAuthn is the string representing an auth log for second-factor authentication via FIDO2/CTAP2/WebAuthn.
	// TODO: Add WebAuthn.

	// AuthTypeDuo is the string representing an auth log for second-factor authentication via DUO.
	AuthTypeDuo = "Duo"
)
