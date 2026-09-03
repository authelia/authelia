// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"context"
	"errors"
	"fmt"

	"github.com/authelia/authelia/v4/internal/oidc/conformance"
)

// NewConformanceRelyingPartyRunner returns a ConformanceRunner for one Relying Party plan. Its modules are driven by
// signing in to Authelia with the external provider identified by provider, which is the conformance suite itself.
func NewConformanceRelyingPartyRunner(client *ConformanceClient, browser *ConformanceBrowser, planID string, modules []ConformancePlanModule, provider string) *ConformanceRunner {
	runner := NewConformanceRunner(client, browser, planID, modules)

	runner.relyingParty = &conformanceRelyingParty{provider: provider, user: conformanceRelyingPartyUsers[provider]}

	return runner
}

func (r *ConformanceRunner) signInWithProvider(ctx context.Context, id, module string) (err error) {
	if r.browser == nil {
		return errors.New("the module needs a browser but the runner has none")
	}

	rp := r.relyingParty

	if rp.user.username == "" {
		return fmt.Errorf("the '%s' provider has no account to link its identity to", rp.provider)
	}

	var leg ConformanceRelyingPartyLeg

	if !rp.linked && conformanceRelyingPartyExpectations[module] {
		r.trace.Logf("Module '%s' linking the '%s' provider to the account '%s'", id, rp.provider, rp.user.username)

		if leg, err = r.browser.LinkWithProvider(ctx, LoginBaseURL, rp.provider, rp.user.username, rp.user.email); err != nil {
			return err
		}

		rp.linked = leg.Accepted
	} else {
		r.trace.Logf("Module '%s' signing in with the '%s' provider", id, rp.provider)

		if leg, err = r.browser.SignInWithProvider(ctx, LoginBaseURL, rp.provider, rp.user.username); err != nil {
			return err
		}
	}

	r.trace.Logf("Module '%s' callback '%s' redirected to '%s', accepted %t", id, leg.Callback, leg.Location, leg.Accepted)

	return ConformanceAssertRelyingPartyLeg(module, leg)
}

// ConformanceAssertRelyingPartyLeg returns an error when Authelia's treatment of the identity asserted by module is
// not the expected one.
func ConformanceAssertRelyingPartyLeg(module string, leg ConformanceRelyingPartyLeg) error {
	if !leg.Accepted && !leg.Notified {
		return fmt.Errorf("expected Authelia to show an error notification when it rejected the identity asserted by the provider but it did not, redirecting to '%s'", leg.Location)
	}

	expected, ok := conformanceRelyingPartyExpectations[module]

	switch {
	case !ok || expected == leg.Accepted:
		return nil
	case expected:
		return fmt.Errorf("expected Authelia to accept the identity asserted by the provider but it rejected it, redirecting to '%s'", leg.Location)
	default:
		return fmt.Errorf("expected Authelia to reject the identity asserted by the provider but it accepted it, redirecting to '%s'", leg.Location)
	}
}

var (
	conformanceRelyingPartyExpectations = map[string]bool{
		"oidcc-client-test":                        true,
		"oidcc-client-test-invalid-iss":            false,
		"oidcc-client-test-missing-sub":            false,
		"oidcc-client-test-invalid-aud":            false,
		"oidcc-client-test-missing-iat":            false,
		"oidcc-client-test-kid-absent-single-jwks": true,
		"oidcc-client-test-idtoken-sig-rs256":      true,
		"oidcc-client-test-idtoken-sig-none":       false,
		"oidcc-client-test-invalid-sig-rs256":      false,
		"oidcc-client-test-userinfo-invalid-sub":   false,
		"oidcc-client-test-nonce-invalid":          false,
		"oidcc-client-test-scope-userinfo-claims":  true,
		"oidcc-client-test-client-secret-basic":    true,
	}

	conformanceRelyingPartyUsers = map[string]conformanceRelyingPartyUser{
		"conformance-" + conformance.NameRelyingPartyBasic:         {username: "harry", email: "harry.potter@authelia.com"},
		"conformance-" + conformance.NameRelyingPartyBasicFormPost: {username: "bob", email: "bob.dylan@authelia.com"},
	}
)

type conformanceRelyingPartyUser struct {
	username string
	email    string
}

type conformanceRelyingParty struct {
	provider string
	user     conformanceRelyingPartyUser
	linked   bool
}
