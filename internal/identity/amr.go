// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"github.com/authelia/authelia/v4/internal/configuration/schema"
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
