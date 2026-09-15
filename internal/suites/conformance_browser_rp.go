// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	conformanceRelyingPartySecondFactor  = "#second-factor-stage"
	conformanceRelyingPartyAuthenticated = "#authenticated-view"
	conformanceRelyingPartyDocumentMark  = "__autheliaConformanceRelyingPartyLeg"
	conformanceRelyingPartyMailTimeout   = time.Second * 30
)

// ConformanceRelyingPartyLeg is what the driver saw when it signed in to Authelia with an external provider.
type ConformanceRelyingPartyLeg struct {
	// Callback is the redirect URI request which delivered the authorization response to Authelia.
	Callback string

	// Location is where Authelia redirected the browser once it handled the callback.
	Location string

	// Accepted is whether Authelia accepted the identity the provider asserted: when signing in, whether the account
	// the identity is linked to was signed in, and when linking, whether the identity was linked to the account.
	Accepted bool

	// Proposed is whether Authelia offered to link the identity the provider asserted, which is what it does for a
	// browser without a session once it has validated an identity which is not linked to an account.
	Proposed bool

	// Notified is whether Authelia showed an error notification after it rejected the identity.
	Notified bool
}

// ConformanceRelyingPartyAuthenticated reports whether the state the portal reported after the callback is the account
// the identity is linked to being signed in. Every leg starts without a session, so any authenticated state was
// established by the callback.
func ConformanceRelyingPartyAuthenticated(username string, level int, expected string) bool {
	return expected != "" && username == expected && level >= 1
}

// ConformanceRelyingPartyProposed reports whether location is the portal offering to link the identity of provider.
func ConformanceRelyingPartyProposed(location, provider string) bool {
	u, err := url.Parse(location)
	if err != nil {
		return false
	}

	return provider != "" && u.Path == externalIdentityLinkPath && u.Query().Get(externalIdentityLinkProviderQueryArg) == provider
}

// ConformanceRelyingPartyRejected reports whether location is Authelia reporting that it rejected the identity of a
// provider, which it does on the sign in page for a browser without a session and on the Linked Accounts page for one
// which is signed in.
func ConformanceRelyingPartyRejected(location string) bool {
	u, err := url.Parse(location)
	if err != nil {
		return false
	}

	return u.Query().Get(externalIdentityErrorQueryArg) == "true"
}

// SignInWithProvider signs in to Authelia with the external provider from the portal, starting without a session, and
// reports whether the account the identity is linked to was signed in. The provider is the conformance suite, which
// authorizes without any interaction, so there is nothing to drive between the portal and the callback.
func (b *ConformanceBrowser) SignInWithProvider(ctx context.Context, portal, provider, username string) (leg ConformanceRelyingPartyLeg, err error) {
	selector, err := b.startSignIn(ctx, portal, provider)
	if err != nil {
		return leg, err
	}

	if leg.Callback, leg.Location, err = b.followCallback(ctx, provider, func() error { return b.session.doClick(b.page, selector) }); err != nil {
		return leg, err
	}

	if !b.awaitRedirectedPortal(ctx, portal, conformanceSelectorFirstFactor, conformanceRelyingPartySecondFactor, conformanceRelyingPartyAuthenticated) {
		log.Warnf("Conformance driver did not see the portal page '%s' render after the callback redirect, the last page was '%s'", leg.Location, b.url())
	}

	stateUsername, level, err := b.session.doGetSessionState(b.page)
	if err != nil {
		return leg, fmt.Errorf("error reading the session state after the callback redirect: %w", err)
	}

	leg.Accepted = ConformanceRelyingPartyAuthenticated(stateUsername, level, username)
	leg.Proposed = ConformanceRelyingPartyProposed(leg.Location, provider)

	if !leg.Accepted && !leg.Proposed {
		leg.Notified = b.session.doAwaitElement(ctx, b.page, conformanceSelectorAutheliaError) == nil
	}

	return leg, nil
}

// StartWithProvider starts a sign in to Authelia with the external provider from the portal, starting without a
// session, with the same request the sign in button makes, but does not follow the authorization request to the
// provider. A module which is only started is decided by the requests Authelia makes to start the flow, and finishes
// once it has seen them, so following the authorization request would only send a request to a module which has already
// finished.
func (b *ConformanceBrowser) StartWithProvider(ctx context.Context, portal, provider string) (err error) {
	if _, err = b.startSignIn(ctx, portal, provider); err != nil {
		return err
	}

	const script = `(path) => fetch(path, {method: 'POST', credentials: 'same-origin', headers: {'Content-Type': 'application/json'}, body: JSON.stringify({keepMeLoggedIn: false})}).then((response) => response.status)`

	result, err := b.page.Context(ctx).Timeout(pageActionTimeout).Eval(script, portal+fmt.Sprintf(externalIdentityStartPathFmt, provider))
	if err != nil {
		return fmt.Errorf("error starting the sign in with the '%s' provider: %w", provider, err)
	}

	log.Debugf("Conformance driver started the sign in with the '%s' provider, which responded with the status %d", provider, result.Value.Int())

	return nil
}

// LinkWithProvider links the identity of the external provider to the account through the portal, the way a user
// does: it signs in to the account, starts the link from the Linked Accounts settings page, and accepts the proposed
// link, completing the identity verification accepting it requires. It reports whether the identity was linked.
func (b *ConformanceBrowser) LinkWithProvider(ctx context.Context, portal, provider, username, email string) (leg ConformanceRelyingPartyLeg, err error) {
	if err = b.signInWithPassword(ctx, portal, username); err != nil {
		return leg, fmt.Errorf("error signing in to link the '%s' provider: %w", provider, err)
	}

	if err = b.session.doOpen(ctx, b.page, portal+externalIdentityLinkedAccountsPath); err != nil {
		return leg, err
	}

	start := fmt.Sprintf(externalIdentityLinkStartSelectorFmt, provider)

	if err = b.session.doAwaitEnabled(b.page, start); err != nil {
		return leg, fmt.Errorf("error starting the link with the '%s' provider: %w", provider, err)
	}

	if leg.Callback, leg.Location, err = b.followCallback(ctx, provider, func() error { return b.session.doClick(b.page, start) }); err != nil {
		return leg, err
	}

	if !isURLPath(leg.Location, externalIdentityLinkedAccountsPath) || ConformanceRelyingPartyRejected(leg.Location) || !b.awaitRedirectedPortal(ctx, portal, externalIdentityLinkPendingSelector) {
		leg.Notified = b.session.doAwaitElement(ctx, b.page, conformanceSelectorAutheliaError) == nil

		return leg, nil
	}

	if err = b.session.doClick(b.page, externalIdentityLinkAcceptSelector); err != nil {
		return leg, fmt.Errorf("error accepting the link with the '%s' provider: %w", provider, err)
	}

	if err = b.session.doAwaitElement(ctx, b.page, identityVerificationDialogSelector); err != nil {
		return leg, fmt.Errorf("error verifying the identity to link the '%s' provider: %w", provider, err)
	}

	code, err := doAwaitOneTimeCodeForRecipient(ctx, email, conformanceRelyingPartyMailTimeout)
	if err != nil {
		return leg, fmt.Errorf("error verifying the identity to link the '%s' provider: %w", provider, err)
	}

	if err = b.session.doInput(b.page, identityVerificationCodeSelector, code); err != nil {
		return leg, fmt.Errorf("error verifying the identity to link the '%s' provider: %w", provider, err)
	}

	if err = b.session.doClick(b.page, identityVerificationVerifySelector); err != nil {
		return leg, fmt.Errorf("error verifying the identity to link the '%s' provider: %w", provider, err)
	}

	if err = b.session.doAwaitElement(ctx, b.page, externalIdentityLinkListedSelector); err != nil {
		return leg, fmt.Errorf("error confirming the link with the '%s' provider: %w", provider, err)
	}

	leg.Accepted = true

	return leg, nil
}

func (b *ConformanceBrowser) startSignIn(ctx context.Context, portal, provider string) (selector string, err error) {
	if err = b.startWithoutSession(ctx, portal+"/"); err != nil {
		return "", err
	}

	if err = b.session.doAwaitElement(ctx, b.page, conformanceSelectorFirstFactor); err != nil {
		return "", fmt.Errorf("error starting the sign in with the '%s' provider: %w", provider, err)
	}

	selector = fmt.Sprintf(externalIdentitySignInSelectorFmt, provider)

	if err = b.session.doAwaitEnabled(b.page, selector); err != nil {
		return "", fmt.Errorf("error starting the sign in with the '%s' provider from the page '%s': %w", provider, b.url(), err)
	}

	return selector, nil
}

func (b *ConformanceBrowser) signInWithPassword(ctx context.Context, portal, username string) (err error) {
	if err = b.startWithoutSession(ctx, portal+"/"); err != nil {
		return err
	}

	if err = b.session.doAwaitElement(ctx, b.page, conformanceSelectorFirstFactor); err != nil {
		return err
	}

	b.username, b.password = username, testPassword

	if _, err = b.submitSignIn(ctx, b.url(), 1); err != nil {
		return err
	}

	return b.session.doAwaitSignedIn(ctx, b.page, username)
}

// startWithoutSession opens the portal with every cookie discarded, and returns an error when the portal still reports
// a signed in account, as a leg which starts with the session a previous leg established does not test anything.
func (b *ConformanceBrowser) startWithoutSession(ctx context.Context, uri string) (err error) {
	if err = b.ClearCookies(); err != nil {
		return fmt.Errorf("error discarding the cookies of the browser context: %w", err)
	}

	if err = b.session.doOpen(ctx, b.page, uri); err != nil {
		return err
	}

	username, level, err := b.session.doGetSessionState(b.page)
	if err != nil {
		return fmt.Errorf("error reading the session state after discarding the cookies: %w", err)
	}

	if username != "" || level != 0 {
		return fmt.Errorf("the portal still reported the account '%s' signed in at level %d after discarding the cookies", username, level)
	}

	return nil
}

func (b *ConformanceBrowser) followCallback(ctx context.Context, provider string, action func() error) (callback, location string, err error) {
	if err = b.session.doMarkDocument(b.page, conformanceRelyingPartyDocumentMark); err != nil {
		log.Warnf("Conformance driver could not mark the document at '%s': %v", b.url(), err)
	}

	var requested bool

	if callback, location, requested, err = b.session.doAwaitExternalIdentityCallback(ctx, b.page, provider, conformanceLegPatience, action); err != nil {
		return "", "", fmt.Errorf("the '%s' provider's callback was not redirected by Authelia (callback requested %t), the last page was '%s': %w", provider, requested, b.url(), err)
	}

	return callback, location, nil
}

func (b *ConformanceBrowser) awaitRedirectedPortal(ctx context.Context, portal string, selectors ...string) bool {
	deadline := time.Now().Add(conformanceSettleTimeout)

	for time.Now().Before(deadline) && ctx.Err() == nil {
		if marked, ok := b.session.isMarkedDocument(b.page, conformanceRelyingPartyDocumentMark); ok && !marked && strings.HasPrefix(b.url(), portal) {
			for _, selector := range selectors {
				if b.has(selector) {
					return true
				}
			}
		}

		time.Sleep(conformanceDriveInterval)
	}

	return false
}
