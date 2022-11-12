package oidc

import (
	"errors"

	"github.com/ory/fosite"
)

var errPasswordsDoNotMatch = errors.New("the passwords don't match")

var (
	ErrIssuerCouldNotDerive        = fosite.ErrServerError.WithHint("Could not safely derive the issuer.")
	ErrSubjectCouldNotLookup       = fosite.ErrServerError.WithHint("Could not lookup user subject.")
	ErrConsentCouldNotPerform      = fosite.ErrServerError.WithHint("Could not perform consent.")
	ErrConsentCouldNotGenerate     = fosite.ErrServerError.WithHint("Could not generate the consent session.")
	ErrConsentCouldNotSave         = fosite.ErrServerError.WithHint("Could not save the consent session.")
	ErrConsentCouldNotLookup       = fosite.ErrServerError.WithHint("Failed to lookup the consent session.")
	ErrConsentMalformedChallengeID = fosite.ErrServerError.WithHint("Malformed consent session challenge ID.")
)
