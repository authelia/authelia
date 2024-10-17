package regulation

import "fmt"

// ErrUserIsBanned user is banned error message.
var ErrUserIsBanned = fmt.Errorf("user is banned")

const (
	// AuthType1FA is the string representing an auth log for first-factor authentication.
	AuthType1FA = "1FA"

	// AuthTypeTOTP is the string representing an auth log for second-factor authentication via TOTP.
	AuthTypeTOTP = "TOTP"

	// AuthTypeWebAuthn is the string representing an auth log for second-factor authentication via FIDO2/CTAP2/WebAuthn.
	AuthTypeWebAuthn = "WebAuthn"

	// AuthTypeDuo is the string representing an auth log for second-factor authentication via DUO.
	AuthTypeDuo = "Duo"
)

const (
	modeBoth        = "both"
	typeUser        = "user"
	typeIP          = "ip"
	fieldBanType    = "ban_type"
	fieldUsername   = "username"
	fieldRecordType = "record_type"
)
