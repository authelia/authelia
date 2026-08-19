package suites

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/cdp"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

const (
	waitElementsAttempts = 10
	elementRetryInterval = time.Millisecond * 50
	setupTestTimeout     = time.Second * 30
)

// RodSession binding a chrome session with devtool protocol.
type RodSession struct {
	Launcher  *launcher.Launcher
	WebDriver *rod.Browser

	RodSuiteCredentialsProvider

	shared   bool
	contexts []*rod.Browser
}

type sharedBrowser struct {
	launcher *launcher.Launcher
	browser  *rod.Browser
}

var (
	sharedBrowsersMutex sync.Mutex
	sharedBrowsers      = map[string]*sharedBrowser{}
)

type RodSessionCredentials struct {
	TOTP *OptionsTOTP
}

type RodSessionOpts struct {
	proxy           string
	provider        RodSuiteCredentialsProvider
	disableDevtools bool
}

type RodSessionOpt func(opts *RodSessionOpts) (err error)

func RodSessionWithProxy(proxy string) RodSessionOpt {
	return func(opts *RodSessionOpts) (err error) {
		opts.proxy = proxy

		return nil
	}
}

func RodSessionWithCredentials(provider RodSuiteCredentialsProvider) RodSessionOpt {
	return func(opts *RodSessionOpts) (err error) {
		opts.provider = provider

		return nil
	}
}

// RodSessionWithoutDevtools disables auto-opened devtools so visual snapshots don't
// capture Chrome's device-emulation overlay.
func RodSessionWithoutDevtools() RodSessionOpt {
	return func(opts *RodSessionOpts) (err error) {
		opts.disableDevtools = true

		return nil
	}
}

func NewRodSession(options ...RodSessionOpt) (session *RodSession, err error) {
	opts := &RodSessionOpts{}

	for _, option := range options {
		if err = option(opts); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	if opts.provider == nil {
		opts.provider = NewRodSuiteCredentials()
	}

	if opts.disableDevtools {
		var shared *sharedBrowser

		if shared, err = newBrowser(opts); err != nil {
			return nil, err
		}

		return &RodSession{
			Launcher:                    shared.launcher,
			WebDriver:                   shared.browser,
			RodSuiteCredentialsProvider: opts.provider,
		}, nil
	}

	sharedBrowsersMutex.Lock()
	defer sharedBrowsersMutex.Unlock()

	shared, ok := sharedBrowsers[opts.proxy]

	if !ok {
		if shared, err = newBrowser(opts); err != nil {
			return nil, err
		}

		sharedBrowsers[opts.proxy] = shared
	}

	return &RodSession{
		Launcher:                    shared.launcher,
		WebDriver:                   shared.browser,
		RodSuiteCredentialsProvider: opts.provider,
		shared:                      true,
	}, nil
}

func newBrowser(opts *RodSessionOpts) (shared *sharedBrowser, err error) {
	var browserPath string

	if browserPath, err = GetBrowserPath(); err != nil {
		return nil, err
	}

	headless := false
	trace := true
	motion := 0 * time.Second

	if os.Getenv("HEADLESS") != "" {
		headless = true
		trace = false
		motion = 0 * time.Second
	}

	l := launcher.New().
		Bin(browserPath).
		Proxy(opts.proxy).
		Headless(headless).
		Devtools(!headless && !opts.disableDevtools)

	if opts.disableDevtools {
		l.Set("font-render-hinting", "none")
		l.Set("disable-lcd-text")
		l.Set("disable-font-subpixel-positioning")
		l.Set("force-color-profile", "srgb")
	}

	url := l.MustLaunch()

	browser := rod.New().
		ControlURL(url).
		Trace(trace).
		SlowMotion(motion).
		MustConnect()

	browser.MustIgnoreCertErrors(true)

	return &sharedBrowser{launcher: l, browser: browser}, nil
}

func closeSharedBrowsers() {
	sharedBrowsersMutex.Lock()
	defer sharedBrowsersMutex.Unlock()

	for key, shared := range sharedBrowsers {
		if err := shared.browser.Close(); err != nil {
			log.Warnf("Error closing the shared browser: %v", err)
		}

		shared.launcher.Cleanup()

		delete(sharedBrowsers, key)
	}
}

// StartRod create a rod/chromedp session.
func StartRod() (*RodSession, error) {
	return NewRodSession()
}

// Stop stop the rod/chromedp session.
func (rs *RodSession) Stop() error {
	for _, incognito := range rs.contexts {
		if err := incognito.Close(); err != nil {
			log.Warnf("Error disposing a browser context: %v", err)
		}
	}

	rs.contexts = nil

	if rs.shared {
		return nil
	}

	if err := rs.WebDriver.Close(); err != nil {
		return err
	}

	rs.Launcher.Cleanup()

	return nil
}

// CheckElementExistsLocatedBySelector reports whether at least one element matching the CSS
// selector currently exists in the DOM.
func (rs *RodSession) CheckElementExistsLocatedBySelector(t *testing.T, page *rod.Page, selector string) bool {
	exists, _, err := page.Has(selector)
	require.NoError(t, err)

	return exists
}

// CheckElementExistsLocatedByID checks the existence of an element located by an id.
func (rs *RodSession) CheckElementExistsLocatedByID(t *testing.T, page *rod.Page, cssSelector string) bool {
	return rs.CheckElementExistsLocatedBySelector(t, page, "#"+cssSelector)
}

// WaitElementLocatedBySelector waits for an element matching the CSS selector to appear in the DOM.
func (rs *RodSession) WaitElementLocatedBySelector(t *testing.T, page *rod.Page, selector string) *rod.Element {
	e, err := page.Element(selector)
	require.NoError(t, err)
	require.NotNil(t, e)

	return e
}

// WaitElementLocatedByClassName waits for an element located by class name.
func (rs *RodSession) WaitElementLocatedByClassName(t *testing.T, page *rod.Page, className string) *rod.Element {
	return rs.WaitElementLocatedBySelector(t, page, "."+className)
}

// WaitElementLocatedByID waits for an element located by an id.
func (rs *RodSession) WaitElementLocatedByID(t *testing.T, page *rod.Page, cssSelector string) *rod.Element {
	return rs.WaitElementLocatedBySelector(t, page, "#"+cssSelector)
}

// ClickElementLocatedBySelector clicks the element matching the CSS selector, locating it again and
// retrying while the click cannot be performed. Locating an element and acting on it are two round
// trips, and the document can move on in between: a re-render leaves the handle pointing at a node no
// longer in the document, and an element mid transition has no visible shape, sits outside the
// viewport, or is disabled and so takes no pointer events. rod retries none of these, so a single
// attempt reports the transient state as a failure rather than waiting for it to pass.
func (rs *RodSession) ClickElementLocatedBySelector(t *testing.T, page *rod.Page, selector string) {
	require.NoError(t, rs.retryElementAction(page, selector, func(element *rod.Element) error {
		return element.Click("left", 1)
	}))
}

// ClickElementLocatedByID clicks the element located by an id, retrying while the click cannot be
// performed.
func (rs *RodSession) ClickElementLocatedByID(t *testing.T, page *rod.Page, cssSelector string) {
	rs.ClickElementLocatedBySelector(t, page, "#"+cssSelector)
}

// TypeElementLocatedByID types into the element located by an id, retrying on the same terms as
// ClickElementLocatedByID.
func (rs *RodSession) TypeElementLocatedByID(t *testing.T, page *rod.Page, cssSelector, value string) {
	require.NoError(t, rs.retryElementAction(page, "#"+cssSelector, func(element *rod.Element) error {
		return element.Type(rs.toInputs(value)...)
	}))
}

// retryElementAction locates the element and performs the action, repeating both until the action
// succeeds, reports something other than a transient state, or the page context expires.
func (rs *RodSession) retryElementAction(page *rod.Page, selector string, action func(element *rod.Element) error) (err error) {
	ctx := page.GetContext()

	for {
		var element *rod.Element

		if element, err = page.Element(selector); err != nil {
			return err
		}

		if err = action(element); err == nil || !isTransientElementError(err) {
			return err
		}

		select {
		case <-ctx.Done():
			return err
		case <-time.After(elementRetryInterval):
		}
	}
}

// isTransientElementError reports whether the error describes a state the element is expected to leave
// on its own: a node replaced by a re-render, or one that is not yet interactable because it is
// animating in, scrolled out of view, covered, or disabled.
func isTransientElementError(err error) bool {
	var notInteractable *rod.NotInteractableError

	return isDetachedNodeError(err) || errors.As(err, &notInteractable)
}

func isDetachedNodeError(err error) bool {
	var e *cdp.Error

	return errors.As(err, &e) && strings.Contains(e.Message, "detached")
}

// WaitElementsLocatedBySelector waits for at least one element matching the CSS selector to
// appear, then returns all current matches.
func (rs *RodSession) WaitElementsLocatedBySelector(t *testing.T, page *rod.Page, selector string) rod.Elements {
	var (
		elements rod.Elements
		err      error
	)

	for range waitElementsAttempts {
		if _, err = page.Element(selector); err != nil {
			break
		}

		if elements, err = page.Elements(selector); err != nil {
			break
		}

		if len(elements) != 0 {
			return elements
		}
	}

	require.NoError(t, err)
	require.NotEmpty(t, elements)

	return elements
}

// WaitElementsLocatedByID waits for at least one element matching the id to appear, then returns
// all current matches.
func (rs *RodSession) WaitElementsLocatedByID(t *testing.T, page *rod.Page, cssSelector string) rod.Elements {
	return rs.WaitElementsLocatedBySelector(t, page, "#"+cssSelector)
}

// WaitForVisualStable blocks until document.fonts.ready resolves and all in-flight images
// have settled. Used as the sync point for visual snapshot tests so layouts don't shift
// mid-capture due to a late font swap or image load.
func (rs *RodSession) WaitForVisualStable(t *testing.T, page *rod.Page) {
	_, err := page.Eval(`async () => {
		await document.fonts.ready;
		await Promise.all(
			Array.from(document.images).map(img =>
				img.complete ? null : new Promise(resolve => {
					img.addEventListener('load', resolve, { once: true });
					img.addEventListener('error', resolve, { once: true });
				})
			)
		);
		return true;
	}`)
	require.NoError(t, err)
}

// SetColorScheme overrides the page's prefers-color-scheme media feature. Call before
// navigation so the initial render picks up the override.
func (rs *RodSession) SetColorScheme(t *testing.T, page *rod.Page, scheme string) {
	err := proto.EmulationSetEmulatedMedia{
		Features: []*proto.EmulationMediaFeature{
			{Name: "prefers-color-scheme", Value: scheme},
		},
	}.Call(page)
	require.NoError(t, err)
}

// FullPageScreenshot captures a PNG of the full scrollable page with scrollbars hidden
// so width deltas don't flap between runs.
func (rs *RodSession) FullPageScreenshot(t *testing.T, page *rod.Page) []byte {
	_, err := page.Eval(`() => new Promise(resolve => {
		const style = document.createElement('style');
		style.textContent = 'html { scrollbar-width: none; } html::-webkit-scrollbar { display: none; }';
		document.head.appendChild(style);
		requestAnimationFrame(() => resolve(true));
	})`)
	require.NoError(t, err)

	screenshot, err := page.Screenshot(true, &proto.PageCaptureScreenshot{
		Format: proto.PageCaptureScreenshotFormatPng,
	})
	require.NoError(t, err)

	return screenshot
}

// DoAndWaitForNavigation runs action and blocks until the next main-frame navigation fires,
// returning the destination URL from the CDP event.
func (rs *RodSession) DoAndWaitForNavigation(ctx context.Context, page *rod.Page, action func()) string {
	var destURL string

	wait := page.Context(ctx).EachEvent(func(e *proto.PageFrameNavigated) bool {
		if e.Frame.ParentID != "" {
			return false
		}

		destURL = e.Frame.URL

		return true
	})

	action()
	wait()

	return destURL
}

func (rs *RodSession) waitBodyContains(t *testing.T, page *rod.Page, pattern string) {
	text, err := page.MustElementR("body", pattern).Text()
	require.NoError(t, err)
	require.NotNil(t, text)

	if strings.Contains(text, pattern) {
		err = nil
	} else {
		err = fmt.Errorf("body does not contain pattern: %s", pattern)
	}

	require.NoError(t, err)
}

func (rs *RodSession) toInputs(in string) (out []input.Key) {
	out = make([]input.Key, len(in))

	for i, c := range in {
		out[i] = input.Key(c)
	}

	return out
}
