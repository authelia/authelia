// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package externalidentity

import (
	"time"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/session"
)

type authenticationMethodsReferencePolicy struct {
	trust    bool
	override bool
	defaults []string
}

func newAuthenticationMethodsReferencePolicy(config schema.AuthenticationBackendExternalIdentityProviderAMR) authenticationMethodsReferencePolicy {
	return authenticationMethodsReferencePolicy{
		trust:    config.Trust,
		override: config.Override,
		defaults: config.Default,
	}
}

func newAssertionlessAuthenticationMethodsReferencePolicy(config schema.AuthenticationBackendExternalIdentityProviderAMR) authenticationMethodsReferencePolicy {
	defaults := config.Default

	if len(defaults) == 0 {
		defaults = passwordAuthenticationMethodsReference()
	}

	return authenticationMethodsReferencePolicy{override: true, defaults: defaults}
}

func passwordAuthenticationMethodsReference() []string {
	userSession := session.UserSession{}

	userSession.SetOneFactorPassword(time.Time{}, &authentication.UserDetails{}, false)

	return userSession.AuthenticationMethodRefs.MarshalRFC8176()
}

func (p authenticationMethodsReferencePolicy) resolve(asserted []string) []string {
	switch {
	case p.override, len(asserted) == 0:
		return p.defaults
	case p.trust:
		return asserted
	default:
		return nil
	}
}
