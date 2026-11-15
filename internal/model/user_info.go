// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"github.com/authelia/authelia/v4/internal/utils"
)

// UserInfo represents the user information required by the web UI.
type UserInfo struct {
	// The users display name.
	DisplayName string `db:"-" json:"display_name"`

	// The users email address.
	Emails []string `db:"-" json:"emails"`

	// The preferred 2FA method.
	Method string `db:"second_factor_method" json:"method" valid:"required"`

	// True if a TOTP device has been registered.
	HasTOTP bool `db:"has_totp" json:"has_totp" valid:"required"`

	// True if a WebAuthn credential has been registered.
	HasWebAuthn bool `db:"has_webauthn" json:"has_webauthn" valid:"required"`

	// True if a duo device has been configured as the preferred.
	HasDuo bool `db:"has_duo" json:"has_duo" valid:"required"`

	// True if the user has at least one unused recovery code.
	HasRecoveryCodes bool `db:"-" json:"has_recovery_codes"`

	// True if the user has 2 or fewer unused recovery codes (used by the frontend to show a low-codes nudge).
	LowRecoveryCodes bool `db:"-" json:"low_recovery_codes"`
}

// SetDefaultPreferred2FAMethod configures the default method based on what is configured as available and the users available methods.
func (i *UserInfo) SetDefaultPreferred2FAMethod(methods []string, fallback string) (changed bool) {
	if len(methods) == 0 {
		// No point attempting to change the method if no methods are available.
		return false
	}

	before := i.Method

	totp, webauthn, duo := utils.IsStringInSlice(SecondFactorMethodTOTP, methods), utils.IsStringInSlice(SecondFactorMethodWebAuthn, methods), utils.IsStringInSlice(SecondFactorMethodDuo, methods)

	if i.Method == "" && utils.IsStringInSlice(fallback, methods) {
		i.Method = fallback
	} else if i.Method != "" && !utils.IsStringInSlice(i.Method, methods) {
		i.Method = ""
	}

	if i.Method == "" {
		i.setMethod(totp, webauthn, duo, methods, fallback)
	}

	return before != i.Method
}

func (i *UserInfo) setMethod(totp, webauthn, duo bool, methods []string, fallback string) {
	switch {
	case i.HasTOTP && totp:
		i.Method = SecondFactorMethodTOTP
	case i.HasWebAuthn && webauthn:
		i.Method = SecondFactorMethodWebAuthn
	case i.HasDuo && duo:
		i.Method = SecondFactorMethodDuo
	case fallback != "" && utils.IsStringInSlice(fallback, methods):
		i.Method = fallback
	case totp:
		i.Method = SecondFactorMethodTOTP
	case webauthn:
		i.Method = SecondFactorMethodWebAuthn
	case duo:
		i.Method = SecondFactorMethodDuo
	}
}
