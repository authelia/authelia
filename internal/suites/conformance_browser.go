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

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	log "github.com/sirupsen/logrus"
)

// ConformancePageState is what the browser is looking at part way through a conformance module's authorization flow.
type ConformancePageState int

const (
	// ConformancePageUnknown is a page the driver has no action for.
	ConformancePageUnknown ConformancePageState = iota

	// ConformancePageFirstFactor is Authelia's first factor sign in form.
	ConformancePageFirstFactor

	// ConformancePageConsent is Authelia's OpenID Connect 1.0 consent decision stage.
	ConformancePageConsent

	// ConformancePageAutheliaError is Authelia reporting an error rather than presenting a stage.
	ConformancePageAutheliaError

	// ConformancePageCallback is the conformance suite's own callback, which ends a leg of the flow.
	ConformancePageCallback
)

const (
	conformanceSelectorFirstFactor             = "#first-factor-stage"
	conformanceSelectorConsent                 = "#openid-consent-decision-stage"
	conformanceSelectorConsentAccept           = "#openid-consent-accept"
	conformanceSelectorConsentReauthentication = "#openid-consent-prompt-login"
	conformanceSelectorConsentPassword         = "#openid-consent-prompt-login #password-textfield"
	conformanceSelectorConsentAuthenticate     = "#openid-consent-authenticate"
	conformanceSelectorUsername                = "#username-textfield"
	conformanceSelectorPassword                = "#password-textfield"
	conformanceSelectorSignIn                  = "#sign-in-button"
	conformanceCallbackPathPrefix              = "/test/a/"
	conformanceSelectorAutheliaError           = `.notification[data-type="error"]`
	conformanceSelectorCompletionError         = `[data-testid="openid-completion-outcome"][data-outcome="error"]`
	conformanceElementTimeout                  = time.Second * 2
	conformanceDriveInterval                   = time.Millisecond * 250
	conformanceSettleTimeout                   = time.Second * 30
	conformanceLegPatience                     = time.Second * 30
	conformancePageTimeout                     = time.Second * 30
	conformanceSignInAttempts                  = 3
	conformanceConsentAttempts                 = 3
)

// ConformanceClassifyPage decides what the browser is looking at. It takes the observations rather than the page so
// that the decision is testable without a browser. The callback wins over every stage selector, because a stage left
// in the DOM of the document being navigated away from would otherwise be read as the current page.
//
// The callback is recognized by the path alone. The authorization request carries the callback in its redirect_uri
// parameter, so matching anywhere in the URL would end the leg while the browser was still on Authelia.
//
// The error toast is ranked below both stage selectors rather than above them. It is not a terminal error page: it is
// a portalled, auto dismissing overlay (web/src/components/UI/Toast.tsx) which co-exists with whatever stage is on
// screen, so a transient toast raised while the sign in form is up would otherwise end the leg on a page the driver
// could have worked. A toast with no stage beneath it is still treated as terminal, which is the case the ranking is
// there to catch.
func ConformanceClassifyPage(uri string, hasFirstFactor, hasConsent, hasError bool) ConformancePageState {
	switch {
	case conformanceIsCallback(uri):
		return ConformancePageCallback
	case hasConsent:
		return ConformancePageConsent
	case hasFirstFactor:
		return ConformancePageFirstFactor
	case hasError:
		return ConformancePageAutheliaError
	default:
		return ConformancePageUnknown
	}
}

func conformanceIsCallback(uri string) bool {
	u, err := url.Parse(uri)
	if err != nil {
		return false
	}

	return strings.HasPrefix(u.Path, conformanceCallbackPathPrefix)
}

// ConformanceLeg is what the driver saw over the course of one browser leg of a module's authorization flow. It is
// recorded as the leg is driven rather than probed afterwards: [ConformanceBrowser.Drive] only returns once the flow
// has left Authelia for the conformance suite's callback, so anything asked of the live DOM at that point is a
// question about the callback page and not about what Authelia did. Being a plain value also makes the module
// overrides which consume it testable without a browser.
type ConformanceLeg struct {
	// Index is the zero based position of this leg among the module's browser legs. Modules such as
	// oidcc-prompt-login perform two authorization round trips, and only the second one carries the parameter under
	// test, so an expectation about re-authentication belongs to Index 1 rather than to every leg.
	Index int

	// FirstFactor is whether Authelia presented its first factor sign in form at any point during the leg.
	FirstFactor bool

	// Consent is whether Authelia presented its OpenID Connect 1.0 consent decision stage at any point during the leg.
	Consent bool

	// Reauthentication is whether Authelia asked for the password again during the leg. This is the signal a
	// prompt=login or max_age expectation turns on, not FirstFactor: Authelia asks on the consent decision form
	// rather than by returning to the sign in page.
	Reauthentication bool

	// AutheliaError is whether the leg ended on Authelia reporting an error rather than reaching the callback.
	AutheliaError bool

	// ErrorURL is the page the leg ended on when AutheliaError is set. Authelia's completion view carries the error in
	// its query, so this is what describes the error to a report.
	ErrorURL string
}

// ConformanceBrowser is one conformance plan's isolated browser context. Every method returns an error rather than
// failing a test, because the driver runs on a goroutine which is not the test's.
type ConformanceBrowser struct {
	session   *RodSession
	incognito *rod.Browser
	page      *rod.Page
	username  string
	password  string
}

// NewConformanceBrowser returns a ConformanceBrowser with its own cookie jar.
func NewConformanceBrowser(session *RodSession) (browser *ConformanceBrowser, err error) {
	incognito, err := session.WebDriver.Incognito()
	if err != nil {
		return nil, fmt.Errorf("error creating an isolated browser context: %w", err)
	}

	page, err := incognito.Page(proto.TargetCreateTarget{})
	if err != nil {
		return nil, fmt.Errorf("error creating a tab: %w", err)
	}

	return &ConformanceBrowser{
		session:   session,
		incognito: incognito,
		page:      page,
		username:  testUsername,
		password:  testPassword,
	}, nil
}

// Close disposes of the browser context.
func (b *ConformanceBrowser) Close() {
	if b.incognito == nil {
		return
	}

	_ = b.incognito.Close()

	b.incognito = nil
}

// ClearCookies discards every cookie in this context, for the modules whose instructions are to remove any cookies
// received from the provider before proceeding.
func (b *ConformanceBrowser) ClearCookies() error {
	return proto.NetworkClearBrowserCookies{}.Call(b.page.Timeout(conformancePageTimeout))
}

func (b *ConformanceBrowser) has(selector string) bool {
	has, _, err := b.page.Timeout(conformanceElementTimeout).Has(selector)

	return err == nil && has
}

// hasAutheliaError reports whether Authelia is reporting an error: either the error toast, or the consent completion
// view it redirects to when a request is rejected before it can be sent back to the client, such as one with an
// unregistered redirect_uri.
func (b *ConformanceBrowser) hasAutheliaError() bool {
	return b.has(conformanceSelectorAutheliaError) || b.has(conformanceSelectorCompletionError)
}

func (b *ConformanceBrowser) url() string {
	info, err := b.page.Timeout(conformancePageTimeout).Info()
	if err != nil {
		return ""
	}

	return info.URL
}

// title returns the page's title, which is the only description available of a page the classifier has no action for.
func (b *ConformanceBrowser) title() string {
	info, err := b.page.Timeout(conformancePageTimeout).Info()
	if err != nil {
		return ""
	}

	return info.Title
}

// settle waits, bounded by conformanceSettleTimeout, for either selector to leave the page or the URL to move on from
// startURL, so a click or form submission's effect has a chance to land before the next reclassification acts on the
// same page again. It never outlives ctx, and gives up silently on timeout: the caller's loop re-acts in that case,
// which is the same behavior as if this guard were not here.
// settle waits for a submitted form to clear, and fails if it does not.
//
// Falling through to let the caller act again would mean submitting the same form twice, which is the one thing this
// driver must not do: each submission carries the consent session's flow back to Authelia, and the second response to
// an already answered session is rejected. Waiting longer and then failing turns a slow backend into a slow pass or a
// legible failure, instead of into a server_error attributed to the provider.
func (b *ConformanceBrowser) settle(ctx context.Context, selector, startURL string) error {
	deadline := time.Now().Add(conformanceSettleTimeout)

	for time.Now().Before(deadline) {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if !b.has(selector) || b.url() != startURL {
			return nil
		}

		time.Sleep(conformanceDriveInterval)
	}

	return fmt.Errorf("the form at '%s' was submitted but '%s' was still on screen %s later", startURL, selector, conformanceSettleTimeout)
}

// Drive navigates to uri and works the flow through to the conformance suite's callback. It is deliberately reactive:
// it acts on whatever page appears rather than on a script per module, so that a module added by a future conformance
// suite release runs instead of failing.
//
// It returns what it saw along the way as a [ConformanceLeg] with the given index, so that a module override can be
// asserted against what Authelia actually presented during the leg rather than against the callback page the browser
// is sitting on by the time this returns. The record starts empty on every call, so an observation never leaks from
// one leg into the next.
func (b *ConformanceBrowser) Drive(ctx context.Context, index int, uri string) (leg ConformanceLeg, err error) {
	leg = ConformanceLeg{Index: index}

	if err = b.session.doNavigate(b.page.Timeout(conformancePageTimeout), uri); err != nil {
		return leg, fmt.Errorf("error navigating to '%s': %w", uri, err)
	}

	attempts, consents := 0, 0

	progress := newConformanceProgress(conformanceLegPatience, time.Now())

	for {
		if ctx.Err() != nil {
			return leg, fmt.Errorf("the flow beginning at '%s' did not reach the conformance callback, the last page was '%s': %w", uri, b.url(), ctx.Err())
		}

		pageURL := b.url()

		switch ConformanceClassifyPage(pageURL, b.has(conformanceSelectorFirstFactor), b.has(conformanceSelectorConsent), b.hasAutheliaError()) {
		case ConformancePageCallback:
			return leg, nil
		case ConformancePageFirstFactor:
			leg.FirstFactor = true

			if attempts++; attempts > conformanceSignInAttempts {
				return leg, fmt.Errorf("the sign in form at '%s' was still present after %d attempts", pageURL, conformanceSignInAttempts)
			}

			if err = b.submitSignIn(ctx, pageURL, attempts); err != nil {
				return leg, err
			}

			progress.observe(time.Now(), pageURL, true)
		case ConformancePageConsent:
			leg.Consent = true

			if consents++; consents > conformanceConsentAttempts {
				return leg, fmt.Errorf("the consent decision form at '%s' was still present after %d attempts", pageURL, conformanceConsentAttempts)
			}

			var reauthentication bool

			if reauthentication, err = b.submitConsent(ctx, pageURL, consents); err != nil {
				return leg, err
			}

			leg.Reauthentication = leg.Reauthentication || reauthentication

			progress.observe(time.Now(), pageURL, true)
		case ConformancePageAutheliaError:
			// An error page can be the point of a module, so it ends the leg without failing. The module's own result
			// decides whether it was expected.
			leg.AutheliaError, leg.ErrorURL = true, pageURL

			return leg, nil
		case ConformancePageUnknown:
			// Failing here rather than waiting out the module's whole context is what keeps one unrecognized page
			// from costing five minutes, and a systemic problem from costing that once per module.
			if progress.observe(time.Now(), pageURL, false) {
				return leg, fmt.Errorf("the flow beginning at '%s' stopped at '%s' (%q), a page the driver has no action for, and it did not change for %s",
					uri, pageURL, b.title(), conformanceLegPatience)
			}

			time.Sleep(conformanceDriveInterval)
		}
	}
}

func (b *ConformanceBrowser) signIn() (err error) {
	if err = b.input(conformanceSelectorUsername, b.username); err != nil {
		return err
	}

	if err = b.input(conformanceSelectorPassword, b.password); err != nil {
		return err
	}

	return b.click(conformanceSelectorSignIn)
}

func (b *ConformanceBrowser) input(selector, value string) (err error) {
	element, err := b.page.Timeout(elementLocateTimeout).Element(selector)
	if err != nil {
		return fmt.Errorf("error locating '%s': %w", selector, err)
	}

	if err = element.SelectAllText(); err != nil {
		return fmt.Errorf("error selecting the text of '%s': %w", selector, err)
	}

	if err = element.Input(value); err != nil {
		return fmt.Errorf("error typing into '%s': %w", selector, err)
	}

	return nil
}

func (b *ConformanceBrowser) click(selector string) (err error) {
	element, err := b.page.Timeout(elementLocateTimeout).Element(selector)
	if err != nil {
		return fmt.Errorf("error locating '%s': %w", selector, err)
	}

	if err = element.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("error clicking '%s': %w", selector, err)
	}

	return nil
}

// conformanceProgress decides when a leg has stopped getting anywhere.
//
// It is deliberately separate from the driver so the decision can be tested without a browser: the driver supplies
// observations, this decides whether to keep waiting. Progress means either a page the driver acted on or a change of
// URL; anything else is the page sitting still, and patience for that is finite.
type conformanceProgress struct {
	patience time.Duration
	url      string
	deadline time.Time
}

// newConformanceProgress returns a conformanceProgress with its patience running from now.
func newConformanceProgress(patience time.Duration, now time.Time) *conformanceProgress {
	return &conformanceProgress{patience: patience, deadline: now.Add(patience)}
}

// observe records what the driver saw and reports whether the leg has stalled. A recognized page or a change of URL is
// progress and restarts the patience; only an unrecognized page that is not moving can exhaust it.
func (p *conformanceProgress) observe(now time.Time, url string, recognized bool) (stalled bool) {
	if url != p.url {
		p.url, p.deadline = url, now.Add(p.patience)

		return false
	}

	if recognized {
		p.deadline = now.Add(p.patience)

		return false
	}

	return now.After(p.deadline)
}

// submitSignIn fills the sign in form and waits for it to clear.
//
// The form is submitted exactly once per call and the wait is not allowed to fall through: a second submission sends
// the same consent flow back to Authelia, and the second response to an already answered consent session is rejected
// as a server_error which looks like the provider's fault.
func (b *ConformanceBrowser) submitSignIn(ctx context.Context, pageURL string, attempt int) (err error) {
	log.Debugf("Conformance driver signing in at '%s' (attempt %d)", pageURL, attempt)

	if err = b.signIn(); err != nil {
		return fmt.Errorf("error signing in at '%s': %w", pageURL, err)
	}

	if err = b.settle(ctx, conformanceSelectorFirstFactor, pageURL); err != nil {
		return fmt.Errorf("error signing in at '%s': %w", pageURL, err)
	}

	return nil
}

// submitConsent works the consent decision form through to the end and reports whether it asked for the password.
//
// The form has two steps behind one stage id and one URL. It arrives on its decision step; accepting either finishes
// the flow or reveals the re-authentication step in place. Only after that second step is submitted does the page
// leave. Each step is submitted exactly once, which matters because every submission that reaches Authelia responds to
// the same consent session, and the second response to an answered session is rejected.
func (b *ConformanceBrowser) submitConsent(ctx context.Context, pageURL string, attempt int) (reauthentication bool, err error) {
	log.Debugf("Conformance driver accepting the consent decision form at '%s' (attempt %d)", pageURL, attempt)

	if err = b.click(conformanceSelectorConsentAccept); err != nil {
		return false, fmt.Errorf("error accepting consent at '%s': %w", pageURL, err)
	}

	if reauthentication, err = b.awaitConsentAccepted(ctx, pageURL); err != nil {
		return false, err
	}

	if !reauthentication {
		return false, nil
	}

	log.Debugf("Conformance driver answering the consent form's re-authentication step at '%s'", pageURL)

	if err = b.input(conformanceSelectorConsentPassword, b.password); err != nil {
		return true, fmt.Errorf("error answering the re-authentication step at '%s': %w", pageURL, err)
	}

	if err = b.click(conformanceSelectorConsentAuthenticate); err != nil {
		return true, fmt.Errorf("error answering the re-authentication step at '%s': %w", pageURL, err)
	}

	if err = b.settle(ctx, conformanceSelectorConsentReauthentication, pageURL); err != nil {
		return true, fmt.Errorf("error answering the re-authentication step at '%s': %w", pageURL, err)
	}

	return true, nil
}

// awaitConsentAccepted waits for accepting the decision form to take effect, and reports whether what it revealed was
// the re-authentication step rather than the end of the flow.
func (b *ConformanceBrowser) awaitConsentAccepted(ctx context.Context, pageURL string) (reauthentication bool, err error) {
	deadline := time.Now().Add(conformanceSettleTimeout)

	for time.Now().Before(deadline) {
		if ctx.Err() != nil {
			return false, ctx.Err()
		}

		reauthentication, done := ConformanceConsentProgress(b.has(conformanceSelectorConsentReauthentication), b.has(conformanceSelectorConsent), b.url() != pageURL)
		if done {
			return reauthentication, nil
		}

		time.Sleep(conformanceDriveInterval)
	}

	return false, fmt.Errorf("the consent decision form at '%s' was accepted but neither left nor asked for the password %s later", pageURL, conformanceSettleTimeout)
}

// ConformanceConsentProgress reports what accepting the consent decision achieved.
//
// The decision form starts in its decision step showing accept and deny (DecisionFormView.tsx, useState<Step> is
// "decision"). Accepting is what reveals the re-authentication step when the flow requires a fresh login: handleAccept
// switches the step in place, without a request and without leaving the page, so the stage and the URL are unchanged
// either side of it. Accepting therefore has two legitimate outcomes, and only one of them looks like the form going
// away.
func ConformanceConsentProgress(hasReauthentication, hasStage, urlChanged bool) (reauthentication, done bool) {
	switch {
	case hasReauthentication:
		return true, true
	case !hasStage || urlChanged:
		return false, true
	default:
		return false, false
	}
}
