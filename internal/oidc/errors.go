package oidc

import (
	"errors"

	"github.com/ory/fosite"
)

var errPasswordsDoNotMatch = errors.New("The provided client secret did not match the registered client secret.")

var (
	// ErrIssuerCouldNotDerive is sent when the issuer couldn't be determined from the headers.
	ErrIssuerCouldNotDerive = fosite.ErrServerError.WithHint("Could not safely derive the issuer.")

	// ErrSubjectCouldNotLookup is sent when the Subject Identifier for a user couldn't be generated or obtained from the database.
	ErrSubjectCouldNotLookup = fosite.ErrServerError.WithHint("Could not lookup user subject.")

	// ErrConsentCouldNotPerform is sent when the Consent Session couldn't be performed for varying reasons.
	ErrConsentCouldNotPerform = fosite.ErrServerError.WithHint("Could not perform consent.")

	// ErrConsentCouldNotGenerate is sent when the Consent Session failed to be generated for some reason, usually a failed UUIDv4 generation.
	ErrConsentCouldNotGenerate = fosite.ErrServerError.WithHint("Could not generate the consent session.")

	// ErrConsentCouldNotSave is sent when the Consent Session couldn't be saved to the database.
	ErrConsentCouldNotSave = fosite.ErrServerError.WithHint("Could not save the consent session.")

	// ErrConsentCouldNotLookup is sent when the Consent ID is not a known UUID.
	ErrConsentCouldNotLookup = fosite.ErrServerError.WithHint("Failed to lookup the consent session.")

	// ErrConsentMalformedChallengeID is sent when the Consent ID is not a UUID.
	ErrConsentMalformedChallengeID = fosite.ErrServerError.WithHint("Malformed consent session challenge ID.")

	// ErrPAREnforcedClientMissingPAR is sent when a client has EnforcePAR configured but the Authorization Request was not Pushed.
	ErrPAREnforcedClientMissingPAR = fosite.ErrInvalidRequest.WithHint("Pushed Authorization Requests are enforced for this client but no such request was sent.")
)
