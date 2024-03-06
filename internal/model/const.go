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

	// SecondFactorMethodWebAuthn method using WebAuthn credentials like YubiKey's.
	SecondFactorMethodWebAuthn = "webauthn"

	// SecondFactorMethodDuo method using Duo application to receive push notifications.
	SecondFactorMethodDuo = "mobile_push"
)

var (
	reSemanticVersion = regexp.MustCompile(`^v?(?P<Major>0|[1-9]\d*)\.(?P<Minor>0|[1-9]\d*)\.(?P<Patch>0|[1-9]\d*)(?:-(?P<PreRelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<Metadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)
	reToken64         = regexp.MustCompile(`^[a-zA-Z0-9_.~+/=-]+$`)
)

const (
	semverRegexpGroupPreRelease = "PreRelease"
)

const (
	FormatJSONSchemaIdentifier         = "https://www.authelia.com/schemas/%s/json-schema/%s.json"
	FormatJSONSchemaYAMLLanguageServer = "# yaml-language-server: $schema=" + FormatJSONSchemaIdentifier
)
