// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"context"
	"fmt"
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
	conformanceSelectorConsentAuthenticate = "#openid-consent-authenticate"
	conformanceSelectorUsername      = "#username-textfield"
	conformanceSelectorPassword      = "#password-textfield"
	conformanceSelectorSignIn        = "#sign-in-button"
	conformanceCallbackPathFragment  = "/test/a/"
	conformanceSelectorAutheliaError = `.notification[data-type="error"]`
	conformanceElementTimeout        = time.Second * 2
	conformanceDriveInterval         = time.Millisecond * 250
	conformanceSettleTimeout   = time.Second * 30
	conformanceLegPatience     = time.Second * 30
	conformancePageTimeout     = time.Second * 30
	conformanceSignInAttempts  = 3
	conformanceConsentAttempts = 3
)

// ConformanceClassifyPage decides what the browser is looking at. It takes the observations rather than the page so
// that the decision is testable without a browser. The callback wins over every stage selector, because a stage left
// in the DOM of the document being navigated away from would otherwise be read as the current page.
//
// The error toast is ranked below both stage selectors rather than above them. It is not a terminal error page: it is
// a portalled, auto dismissing overlay (web/src/components/UI/Toast.tsx) which co-exists with whatever stage is on
// screen, so a transient toast raised while the sign in form is up would otherwise end the leg on a page the driver
// could have worked. A toast with no stage beneath it is still treated as terminal, which is the case the ranking is
// there to catch.
func ConformanceClassifyPage(uri string, hasFirstFactor, hasConsent, hasError bool) ConformancePageState {
	switch {
	case strings.Contains(uri, conformanceCallbackPathFragment):
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

		switch ConformanceClassifyPage(pageURL, b.has(conformanceSelectorFirstFactor), b.has(conformanceSelectorConsent), b.has(conformanceSelectorAutheliaError)) {
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
			leg.AutheliaError = true

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

// consent accepts the decision form, first supplying the password when Authelia is asking for it again, and reports
// whether it was asked.
func (b *ConformanceBrowser) consent() (reauthentication bool, err error) {
	if reauthentication = b.has(conformanceSelectorConsentReauthentication); reauthentication {
		if err = b.input(conformanceSelectorConsentPassword, b.password); err != nil {
			return true, err
		}

		// Submitting re-authentication is a separate decision from granting consent, and the form offers only one of
		// them at a time. Whichever step follows this one comes back around the driver's loop.
		return true, b.click(conformanceSelectorConsentAuthenticate)
	}

	return false, b.click(conformanceSelectorConsentAccept)
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

// submitConsent submits the consent decision form and waits for it to clear, reporting whether the form asked for the
// password again. It carries the same single-submission guarantee as submitSignIn, and for the same reason.
func (b *ConformanceBrowser) submitConsent(ctx context.Context, pageURL string, attempt int) (reauthentication bool, err error) {
	log.Debugf("Conformance driver submitting the consent decision form at '%s' (attempt %d)", pageURL, attempt)

	if reauthentication, err = b.consent(); err != nil {
		return reauthentication, fmt.Errorf("error accepting consent at '%s': %w", pageURL, err)
	}

	if err = b.settle(ctx, conformanceSelectorConsent, pageURL); err != nil {
		return reauthentication, fmt.Errorf("error submitting the consent decision form at '%s': %w", pageURL, err)
	}

	return reauthentication, nil
}
