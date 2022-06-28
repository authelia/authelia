package model

import (
	"regexp"
)

const (
	errFmtValueNil           = "cannot value model type '%T' with value nil to driver.Value"
	errFmtScanNil            = "cannot scan model type '%T' from value nil: type doesn't support nil values"
	errFmtScanInvalidType    = "cannot scan model type '%T' from type '%T' with value '%v'"
	errFmtScanInvalidTypeErr = "cannot scan model type '%T' from type '%T' with value '%v': %w"
)

const (
	// SecondFactorMethodTOTP method using Time-Based One-Time Password applications like Google Authenticator.
	SecondFactorMethodTOTP = "totp"

	// SecondFactorMethodWebauthn method using Webauthn devices like YubiKey's.
	SecondFactorMethodWebauthn = "webauthn"

	// SecondFactorMethodDuo method using Duo application to receive push notifications.
	SecondFactorMethodDuo = "mobile_push"
)

var reSemanticVersion = regexp.MustCompile(`^v?(?P<Major>\d+)\.(?P<Minor>\d+)\.(?P<Patch>\d+)(\-(?P<PreRelease>[a-zA-Z0-9-]+(\.[a-zA-Z0-9-]+)*))?(\+(?P<Metadata>[a-zA-Z0-9-]+(\.[a-zA-Z0-9-]+)*))?$`)

const (
	semverRegexpGroupPreRelease = "PreRelease"
)
