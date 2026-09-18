// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package events

// Event type names.
//
//nolint:gosec // The word "credential" in these event names is not indicative of a hardcoded secret.
const (
	TypeUserPasswordChanged             = "user.password.changed"
	TypeUserPasswordReset               = "user.password.reset"
	TypeUserCredentialTOTPAdded         = "user.credential.totp.added"
	TypeUserCredentialTOTPRemoved       = "user.credential.totp.removed"
	TypeUserCredentialWebAuthnAdded     = "user.credential.webauthn.added"
	TypeUserCredentialWebAuthnRemoved   = "user.credential.webauthn.removed"
	TypeUserIdentityVerificationStarted = "user.identity_verification.started"
	TypeUserSessionElevationRequested   = "user.session.elevation.requested"
	TypeSecurityAuthenticationSucceeded = "security.authentication.succeeded"
	TypeSecurityAuthenticationFailed    = "security.authentication.failed"
	TypeSecurityBanApplied              = "security.ban.applied"
	TypeSecurityBanExpired              = "security.ban.expired"
	TypeSystemStartupCheck              = "system.startup_check"
)

// Authentication stages and methods.
const (
	StageFirstFactor  = "first_factor"
	StageSecondFactor = "second_factor"

	MethodPassword = "password"
	MethodTOTP     = "totp"
	MethodWebAuthn = "webauthn"
	MethodDuo      = "duo"
)

// Ban target types.
const (
	TargetTypeUser = "user"
	TargetTypeIP   = "ip"
)

// Authentication failure reasons.
//
//nolint:gosec // The word "credential" in this constant name is not indicative of a hardcoded secret.
const (
	ReasonInvalidCredentials = "invalid_credentials"
	ReasonUserNotFound       = "user_not_found"
	ReasonBanned             = "banned"
	ReasonInternalError      = "internal_error"
)

// Signature algorithms.
const (
	SignatureAlgorithmSHA256 = "sha256"
	SignatureAlgorithmSHA512 = "sha512"
)

// Envelope and schema constants.
const (
	SpecVersion            = "1.0"
	ContractVersion        = "1"
	DataContentType        = "application/json"
	ContentTypeHeader      = "application/cloudevents+json; charset=utf-8"
	ContentTypeBatchHeader = "application/cloudevents-batch+json; charset=utf-8"

	schemaBaseURL = "https://www.authelia.com/schemas/webhooks/v" + ContractVersion + "/"
)

// Header names.
const (
	HeaderContractVersion = "X-Authelia-Webhook-Version"
	HeaderEventType       = "X-Authelia-Event-Type"
	HeaderEventID         = "X-Authelia-Event-Id"
	HeaderDeliveryAttempt = "X-Authelia-Delivery-Attempt"
	HeaderSignature       = "X-Authelia-Signature"
	HeaderEventCount      = "X-Authelia-Event-Count"
)

// Abuse protection headers from the CloudEvents HTTP webhook specification.
const (
	HeaderWebhookRequestOrigin = "WebHook-Request-Origin"
	HeaderWebhookRequestRate   = "WebHook-Request-Rate"
	HeaderWebhookAllowedOrigin = "WebHook-Allowed-Origin"
	HeaderWebhookAllowedRate   = "WebHook-Allowed-Rate"

	// WebhookRateUnlimited is the specification's wildcard for an unrestricted rate.
	WebhookRateUnlimited = "*"
)

// TagSensitive is the struct tag marking a field as credential equivalent, which is omitted unless a destination
// disables redaction.
const TagSensitive = "sensitive"
