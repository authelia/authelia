// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package regulation

import "fmt"

// ErrUserIsBanned user is banned error message.
var ErrUserIsBanned = fmt.Errorf("user is banned")

const (
	// AuthType1FA is the string representing an auth log for first-factor authentication.
	AuthType1FA = "1FA"

	// AuthTypeTOTP is the string representing an auth log for second-factor authentication via TOTP.
	AuthTypeTOTP = "TOTP"

	// AuthTypeWebauthn is the string representing an auth log for second-factor authentication via FIDO2/CTAP2/WebAuthn.
	AuthTypeWebauthn = "Webauthn"

	// AuthTypeDuo is the string representing an auth log for second-factor authentication via DUO.
	AuthTypeDuo = "Duo"
)
