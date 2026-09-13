// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"context"
	"encoding/base64"
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
	conformanceDriveInterval                   = time.Millisecond * 250
	conformanceSettleTimeout                   = time.Second * 30
	conformanceLegPatience                     = time.Second * 30
	conformancePageTimeout                     = time.Second * 30
	conformanceSignInAttempts                  = 3
	conformanceConsentAttempts                 = 3
	conformanceScreenshotLimit                 = 500 * 1024
	conformanceCoverageBinding                 = "__autheliaConformanceCoverage"
	conformanceCoverageScript                  = `addEventListener('beforeunload', () => { if (window.__coverage__) window.` + conformanceCoverageBinding + `(JSON.stringify(window.__coverage__)); });`
)

// ConformanceClassifyPage decides what the browser is looking at. It takes the observations rather than the page so
// that the decision is testable without a browser. The callback wins over every stage selector, because a stage left
// in the DOM of the document being navigated away from would otherwise be read as the current page.
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

	// LoginScreenshot is the last page during the leg which asked for the password, captured with it entered and
	// before it was submitted: the consent form's re-authentication step, or the sign in form.
	LoginScreenshot string

	// ErrorScreenshot is the Authelia error page the leg ended on.
	ErrorScreenshot string
}

// ConformanceBrowser is one conformance plan's isolated browser context. Every method returns an error rather than
// failing a test, because the driver runs on a goroutine which is not the test's.
type ConformanceBrowser struct {
	session   *RodSession
	incognito *rod.Browser
	page      *rod.Page
	username  string
	password  string

	trace        *ConformanceTrace
	stopCoverage func()
}

// NewConformanceBrowser returns a ConformanceBrowser with its own cookie jar.
func NewConformanceBrowser(session *RodSession) (browser *ConformanceBrowser, err error) {
	incognito, err := session.WebDriver.Incognito()
	if err != nil {
		return nil, fmt.Errorf("error creating an isolated browser context: %w", err)
	}

	page, err := incognito.Page(proto.TargetCreateTarget{})
	if err != nil {
		_ = incognito.Close()

		return nil, fmt.Errorf("error creating a tab: %w", err)
	}

	browser = &ConformanceBrowser{
		session:   session,
		incognito: incognito,
		page:      page,
		username:  testUsername,
		password:  testPassword,
	}

	if err = browser.collectCoverage(); err != nil {
		_ = incognito.Close()

		return nil, fmt.Errorf("error arranging for frontend coverage to be collected: %w", err)
	}

	return browser, nil
}

func (b *ConformanceBrowser) collectCoverage() (err error) {
	if err = (proto.RuntimeAddBinding{Name: conformanceCoverageBinding}).Call(b.page); err != nil {
		return err
	}

	if _, err = b.page.EvalOnNewDocument(conformanceCoverageScript); err != nil {
		return err
	}

	page, cancel := b.page.WithCancel()

	wait := page.EachEvent(func(e *proto.RuntimeBindingCalled) {
		if e.Name != conformanceCoverageBinding {
			return
		}

		if err := writeCoverage(coverageDir, e.Payload); err != nil {
			log.Errorf("Error writing coverage: %v", err)
		}
	})

	done := make(chan struct{})

	go func() {
		wait()
		close(done)
	}()

	b.stopCoverage = func() {
		cancel()
		<-done
	}

	return nil
}

// Close disposes of the browser context. The document the browser ends on is never left, so its coverage is collected
// directly, and any coverage still being written is finished before the context goes.
func (b *ConformanceBrowser) Close() {
	if b.incognito == nil {
		return
	}

	b.session.collectCoverage(b.page.Timeout(conformancePageTimeout))

	if b.stopCoverage != nil {
		b.stopCoverage()
	}

	_ = b.incognito.Close()

	b.incognito = nil
}

// ClearCookies discards every cookie in this context, for the modules whose instructions are to remove any cookies
// received from the provider before proceeding.
func (b *ConformanceBrowser) ClearCookies() error {
	return b.session.doClearCookies(b.page)
}

func (b *ConformanceBrowser) has(selector string) bool {
	return b.session.hasElement(b.page, selector)
}

func (b *ConformanceBrowser) hasAutheliaError() bool {
	return b.has(conformanceSelectorAutheliaError) || b.has(conformanceSelectorCompletionError)
}

func (b *ConformanceBrowser) url() string {
	return b.session.pageURL(b.page)
}

func (b *ConformanceBrowser) title() string {
	return b.session.pageTitle(b.page)
}

// Drive navigates to uri and works the flow through to the conformance suite's callback. It is deliberately reactive:
// it acts on whatever page appears rather than on a script per module, so that a module added by a future conformance
// suite release runs instead of failing.
func (b *ConformanceBrowser) Drive(ctx context.Context, index int, uri string) (leg ConformanceLeg, err error) {
	leg = ConformanceLeg{Index: index}

	loaded := b.page.Timeout(conformancePageTimeout).WaitNavigation(proto.PageLifecycleEventNameDOMContentLoaded)

	if err = b.session.doNavigate(b.page.Timeout(conformancePageTimeout), uri); err != nil {
		return leg, fmt.Errorf("error navigating to '%s': %w", uri, err)
	}

	loaded()

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

			var screenshot string

			if screenshot, err = b.submitSignIn(ctx, pageURL, attempts); err != nil {
				return leg, err
			}

			leg.LoginScreenshot = conformanceLatest(leg.LoginScreenshot, screenshot)

			progress.observe(time.Now(), pageURL, true)
		case ConformancePageConsent:
			leg.Consent = true

			if consents++; consents > conformanceConsentAttempts {
				return leg, fmt.Errorf("the consent decision form at '%s' was still present after %d attempts", pageURL, conformanceConsentAttempts)
			}

			var (
				reauthentication bool
				screenshot       string
			)

			if reauthentication, screenshot, err = b.submitConsent(ctx, pageURL, consents); err != nil {
				return leg, err
			}

			leg.Reauthentication = leg.Reauthentication || reauthentication
			leg.LoginScreenshot = conformanceLatest(leg.LoginScreenshot, screenshot)

			progress.observe(time.Now(), pageURL, true)
		case ConformancePageAutheliaError:
			leg.AutheliaError, leg.ErrorURL, leg.ErrorScreenshot = true, pageURL, b.screenshot()

			return leg, nil
		case ConformancePageUnknown:
			if progress.observe(time.Now(), pageURL, false) {
				return leg, fmt.Errorf("the flow beginning at '%s' stopped at '%s' (%q), a page the driver has no action for, and it did not change for %s",
					uri, pageURL, b.title(), conformanceLegPatience)
			}

			time.Sleep(conformanceDriveInterval)
		}
	}
}

func (b *ConformanceBrowser) signIn() (screenshot string, err error) {
	if err = b.session.doInput(b.page, conformanceSelectorUsername, b.username); err != nil {
		return "", err
	}

	if err = b.session.doInput(b.page, conformanceSelectorPassword, b.password); err != nil {
		return "", err
	}

	screenshot = b.screenshot()

	return screenshot, b.session.doClick(b.page, conformanceSelectorSignIn)
}

func (b *ConformanceBrowser) screenshot() string {
	uri, err := conformanceScreenshotDataURI(func(format proto.PageCaptureScreenshotFormat, quality int) ([]byte, error) {
		req := &proto.PageCaptureScreenshot{Format: format}

		if quality != 0 {
			req.Quality = &quality
		}

		return b.page.Timeout(conformancePageTimeout).Screenshot(false, req)
	})
	if err != nil {
		log.Warnf("Conformance driver could not capture a screenshot of '%s': %v", b.url(), err)

		return ""
	}

	return uri
}

func conformanceScreenshotDataURI(capture func(format proto.PageCaptureScreenshotFormat, quality int) ([]byte, error)) (uri string, err error) {
	for _, attempt := range []struct {
		format  proto.PageCaptureScreenshotFormat
		quality int
		mime    string
	}{
		{proto.PageCaptureScreenshotFormatPng, 0, "image/png"},
		{proto.PageCaptureScreenshotFormatJpeg, 80, "image/jpeg"},
		{proto.PageCaptureScreenshotFormatJpeg, 50, "image/jpeg"},
	} {
		var data []byte

		if data, err = capture(attempt.format, attempt.quality); err != nil {
			return "", err
		}

		if len(data) <= conformanceScreenshotLimit {
			return "data:" + attempt.mime + ";base64," + base64.StdEncoding.EncodeToString(data), nil
		}
	}

	return "", fmt.Errorf("the screenshot does not fit the conformance suite's %dKB upload limit in any format", conformanceScreenshotLimit/1024)
}

func conformanceLatest(current, next string) string {
	if next == "" {
		return current
	}

	return next
}

type conformanceProgress struct {
	patience time.Duration
	url      string
	deadline time.Time
}

func newConformanceProgress(patience time.Duration, now time.Time) *conformanceProgress {
	return &conformanceProgress{patience: patience, deadline: now.Add(patience)}
}

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

func (b *ConformanceBrowser) submitSignIn(ctx context.Context, pageURL string, attempt int) (screenshot string, err error) {
	b.trace.Logf("Signing in at '%s' (attempt %d)", pageURL, attempt)

	if screenshot, err = b.signIn(); err != nil {
		return "", fmt.Errorf("error signing in at '%s': %w", pageURL, err)
	}

	if err = b.session.doAwaitSubmitted(ctx, b.page, conformanceSelectorFirstFactor, pageURL); err != nil {
		return "", fmt.Errorf("error signing in at '%s': %w", pageURL, err)
	}

	return screenshot, nil
}

func (b *ConformanceBrowser) submitConsent(ctx context.Context, pageURL string, attempt int) (reauthentication bool, screenshot string, err error) {
	b.trace.Logf("Accepting the consent decision form at '%s' (attempt %d)", pageURL, attempt)

	if err = b.session.doClick(b.page, conformanceSelectorConsentAccept); err != nil {
		return false, "", fmt.Errorf("error accepting consent at '%s': %w", pageURL, err)
	}

	if reauthentication, err = b.awaitConsentAccepted(ctx, pageURL); err != nil {
		return false, "", err
	}

	if !reauthentication {
		return false, "", nil
	}

	b.trace.Logf("Answering the consent form's re-authentication step at '%s'", pageURL)

	if err = b.session.doInput(b.page, conformanceSelectorConsentPassword, b.password); err != nil {
		return true, "", fmt.Errorf("error answering the re-authentication step at '%s': %w", pageURL, err)
	}

	screenshot = b.screenshot()

	if err = b.session.doClick(b.page, conformanceSelectorConsentAuthenticate); err != nil {
		return true, "", fmt.Errorf("error answering the re-authentication step at '%s': %w", pageURL, err)
	}

	if err = b.session.doAwaitSubmitted(ctx, b.page, conformanceSelectorConsentReauthentication, pageURL); err != nil {
		return true, "", fmt.Errorf("error answering the re-authentication step at '%s': %w", pageURL, err)
	}

	return true, screenshot, nil
}

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
