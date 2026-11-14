// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package events

// Event type names.
//
//nolint:gosec // The word "credential" in these event names is not indicative of a hardcoded secret.
const (
	TypeUserPasswordChanged             = "com.authelia.user.password.changed"
	TypeUserPasswordReset               = "com.authelia.user.password.reset"
	TypeUserCredentialTOTPAdded         = "com.authelia.user.credential.totp.added"
	TypeUserCredentialTOTPRemoved       = "com.authelia.user.credential.totp.removed"
	TypeUserCredentialWebAuthnAdded     = "com.authelia.user.credential.webauthn.added"
	TypeUserCredentialWebAuthnRemoved   = "com.authelia.user.credential.webauthn.removed"
	TypeUserIdentityVerificationStarted = "com.authelia.user.identity_verification.started"
	TypeUserSessionElevationRequested   = "com.authelia.user.session.elevation.requested"
	TypeSecurityAuthenticationSucceeded = "com.authelia.security.authentication.succeeded"
	TypeSecurityAuthenticationFailed    = "com.authelia.security.authentication.failed"
	TypeSecurityBanApplied              = "com.authelia.security.ban.applied"
	TypeSecurityBanExpired              = "com.authelia.security.ban.expired"
	TypeSystemStartupCheck              = "com.authelia.system.startup_check"
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
	HeaderDeliveryAttempt = "X-Authelia-Delivery-Attempt"
	HeaderSignature       = "X-Authelia-Signature"
	HeaderEventCount      = "X-Authelia-Event-Count"
)

// Metadata headers from the CloudEvents HTTP protocol binding, which carry the context attributes of a single event.
// Every header name beginning with HeaderPrefixCloudEvents is reserved for them.
const (
	HeaderPrefixCloudEvents = "ce-"

	HeaderCloudEventsSpecVersion     = HeaderPrefixCloudEvents + "specversion"
	HeaderCloudEventsID              = HeaderPrefixCloudEvents + "id"
	HeaderCloudEventsType            = HeaderPrefixCloudEvents + "type"
	HeaderCloudEventsSource          = HeaderPrefixCloudEvents + "source"
	HeaderCloudEventsTime            = HeaderPrefixCloudEvents + "time"
	HeaderCloudEventsDataSchema      = HeaderPrefixCloudEvents + "dataschema"
	HeaderCloudEventsSubject         = HeaderPrefixCloudEvents + "subject"
	HeaderCloudEventsAutheliaVersion = HeaderPrefixCloudEvents + "autheliaversion"

	hexUpper = "0123456789ABCDEF"
)

// Abuse protection headers from the CloudEvents HTTP webhook specification.
const (
	HeaderWebhookRequestOrigin   = "WebHook-Request-Origin"
	HeaderWebhookRequestRate     = "WebHook-Request-Rate"
	HeaderWebhookRequestCallback = "WebHook-Request-Callback"
	HeaderWebhookAllowedOrigin   = "WebHook-Allowed-Origin"
	HeaderWebhookAllowedRate     = "WebHook-Allowed-Rate"

	// WebhookRateUnlimited is the specification's wildcard for an unrestricted rate.
	WebhookRateUnlimited = "*"
)

// TagSensitive is the struct tag marking a field as credential equivalent, which is omitted unless a destination
// disables redaction.
const TagSensitive = "sensitive"

var registry = map[string]Descriptor{}
