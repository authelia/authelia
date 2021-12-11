package models

// UserInfo represents the user information required by the web UI.
type UserInfo struct {
	// The users display name.
	DisplayName string `db:"-" json:"display_name"`

	// The preferred 2FA method.
	Method string `db:"second_factor_method" json:"method" valid:"required"`

	// True if a TOTP device has been registered.
	HasTOTP bool `db:"has_totp" json:"has_totp" valid:"required"`

	// True if a Webauthn device has been registered.
	HasWebauthn bool `db:"has_webauthn" json:"has_webauthn" valid:"required"`

	// True if a duo device has been configured as the preferred.
	HasDuo bool `db:"has_duo" json:"has_duo" valid:"required"`
}
