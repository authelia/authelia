package oidc

import (
	"errors"

	oauthelia2 "authelia.com/provider/oauth2"
)

var (
	errClientSecretMismatch = errors.New("The provided client secret did not match the registered client secret.")
)

var (
	// ErrSubjectCouldNotLookup is sent when the Subject Identifier for a user couldn't be generated or obtained from the database.
	ErrSubjectCouldNotLookup = oauthelia2.ErrServerError.WithHint("Could not lookup user subject.")

	// ErrConsentCouldNotPerform is sent when the Consent Session couldn't be performed for varying reasons.
	ErrConsentCouldNotPerform = oauthelia2.ErrServerError.WithHint("Could not perform consent.")

	// ErrConsentCouldNotGenerate is sent when the Consent Session failed to be generated for some reason, usually a failed UUIDv4 generation.
	ErrConsentCouldNotGenerate = oauthelia2.ErrServerError.WithHint("Could not generate the consent session.")

	// ErrConsentCouldNotSave is sent when the Consent Session couldn't be saved to the database.
	ErrConsentCouldNotSave = oauthelia2.ErrServerError.WithHint("Could not save the consent session.")

	// ErrConsentCouldNotLookup is sent when the Consent ID is not a known UUID.
	ErrConsentCouldNotLookup = oauthelia2.ErrServerError.WithHint("Failed to lookup the consent session.")

	// ErrConsentMalformedChallengeID is sent when the Consent ID is not a UUID.
	ErrConsentMalformedChallengeID = oauthelia2.ErrServerError.WithHint("Malformed consent session challenge ID.")

	ErrClientAuthorizationUserAccessDenied = oauthelia2.ErrAccessDenied.WithHint("The user was denied access to this client.")
)
