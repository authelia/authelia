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
	elementActionTimeout  = time.Second * 10
	elementAttemptTimeout = time.Second
	elementRetryInterval  = time.Millisecond * 50
	elementStateTimeout   = time.Second * 5
	pageRenderTimeout     = time.Second * 5
	setupTestTimeout      = time.Second * 30
	waitElementsAttempts  = 10
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

// RodSessionCredentials represents the credentials available to a RodSession.
type RodSessionCredentials struct {
	TOTP *OptionsTOTP
}

// RodSessionOpts represents the options used to create a RodSession.
type RodSessionOpts struct {
	proxy           string
	provider        RodSuiteCredentialsProvider
	disableDevtools bool
}

// RodSessionOpt is a function which configures the RodSessionOpts.
type RodSessionOpt func(opts *RodSessionOpts) (err error)

// RodSessionWithProxy returns a RodSessionOpt which sets the proxy.
func RodSessionWithProxy(proxy string) RodSessionOpt {
	return func(opts *RodSessionOpts) (err error) {
		opts.proxy = proxy

		return nil
	}
}

// RodSessionWithCredentials returns a RodSessionOpt which sets the credentials provider.
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

// NewRodSession returns a new RodSession given the provided options.
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

	// Chrome resolves the suite domains from these rules rather than from /etc/hosts, so a browser launched by one
	// suite reaches that suite's network even while another suite on this machine owns the names in /etc/hosts. The
	// proxy hostnames the NetworkACL suite dials are resolved the same way.
	l.Set("host-resolver-rules", HostResolverRules())

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

// errElementValueNotSet reports that the value did not reach the element, which typing does without
// error when the keys are dispatched at a node the document has already replaced.
var errElementValueNotSet = errors.New("element value was not set")

// elementState describes why an element could not be acted on. rod names the condition it stopped at
// but not the cause, and the failure screenshot is taken once the page has settled, by which time the
// cause has usually gone.
const elementState = `(selector) => {
	const element = document.querySelector(selector);

	if (!element) {
		return 'not in the document';
	}

	const style = getComputedStyle(element);
	const rect = element.getBoundingClientRect();
	const center = document.elementFromPoint(rect.left + rect.width / 2, rect.top + rect.height / 2);

	const describe = (node) => {
		if (node === null) {
			return 'nothing';
		}

		if (node === element) {
			return 'itself';
		}

		const slot = node.getAttribute('data-slot');

		return (node.id ? '#' + node.id : node.tagName.toLowerCase()) + (slot ? '[' + slot + ']' : '');
	};

	return JSON.stringify({
		disabled: element.disabled === true,
		inert: element.closest('[inert], [data-base-ui-inert]') !== null,
		pointerEvents: style.pointerEvents,
		visibility: style.visibility,
		display: style.display,
		opacity: style.opacity,
		value: element.value === undefined ? null : element.value,
		rect: [Math.round(rect.x), Math.round(rect.y), Math.round(rect.width), Math.round(rect.height)],
		viewport: [innerWidth, innerHeight],
		topmostAtCenter: describe(center),
	});
}`

// ClickElementLocatedBySelector clicks the element matching the CSS selector, locating it again and
// retrying while the click cannot be performed.
func (rs *RodSession) ClickElementLocatedBySelector(t *testing.T, page *rod.Page, selector string) {
	rs.doElementAction(t, page, selector, func(element *rod.Element) error {
		return element.Click("left", 1)
	})
}

// ClickElementLocatedByID clicks the element located by an id, retrying while the click cannot be
// performed.
func (rs *RodSession) ClickElementLocatedByID(t *testing.T, page *rod.Page, cssSelector string) {
	rs.ClickElementLocatedBySelector(t, page, "#"+cssSelector)
}

// TypeElementLocatedByID types into the element located by an id, retrying on the same terms as
// ClickElementLocatedByID.
func (rs *RodSession) TypeElementLocatedByID(t *testing.T, page *rod.Page, cssSelector, value string) {
	rs.doElementAction(t, page, "#"+cssSelector, func(element *rod.Element) error {
		// An attempt that failed partway through leaves what it managed to type behind, so the field is
		// replaced rather than appended to and the value is confirmed before the attempt is accepted.
		if err := element.SelectAllText(); err != nil {
			return err
		}

		if err := element.Type(rs.toInputs(value)...); err != nil {
			return err
		}

		property, err := element.Property("value")
		if err != nil {
			return err
		}

		if property.Str() != value {
			return errElementValueNotSet
		}

		return nil
	})
}

func (rs *RodSession) waitElementTextIs(t *testing.T, page *rod.Page, selector, expected string) {
	rs.waitElementText(t, page, selector, expected, true)
}

func (rs *RodSession) waitElementTextIsNot(t *testing.T, page *rod.Page, selector, unexpected string) {
	rs.waitElementText(t, page, selector, unexpected, false)
}

func (rs *RodSession) waitElementText(t *testing.T, page *rod.Page, selector, value string, equal bool) {
	ctx, cancel := context.WithTimeout(page.GetContext(), elementActionTimeout)

	defer cancel()

	bounded := page.Context(ctx)
	observed := "unavailable"

	for {
		if element, err := bounded.Element(selector); err == nil {
			// rod reports the text as rendered, which is the value the assertions read: the components
			// uppercase their labels through a style the text in the document does not account for.
			if text, errText := element.Text(); errText == nil {
				observed = text

				if (text == value) == equal {
					return
				}
			}
		}

		select {
		case <-ctx.Done():
			if equal {
				require.Failf(t, "Element did not take the expected text",
					"selector '%s' expected '%s', observed '%s'", selector, value, observed)
			} else {
				require.Failf(t, "Element kept the text it was expected to leave",
					"selector '%s' was still '%s'", selector, value)
			}

			return
		case <-time.After(elementRetryInterval):
		}
	}
}

func (rs *RodSession) doElementAction(t *testing.T, page *rod.Page, selector string, action func(element *rod.Element) error) {
	ctx, cancel := context.WithTimeout(page.GetContext(), elementActionTimeout)

	defer cancel()

	err := retryElementAction(page.Context(ctx), selector, action)
	if err == nil {
		return
	}

	require.Failf(t, "Element action did not succeed",
		"selector '%s' gave up after %s: %v\nelement state: %s",
		selector, elementActionTimeout, err, describeElement(page, selector))
}

func retryElementAction(page *rod.Page, selector string, action func(element *rod.Element) error) error {
	ctx := page.GetContext()

	// The state the element was last in says why it could not be acted on; the expiry that ends the
	// attempt says only that it never left that state, so the former is reported in preference.
	var last error

	for {
		element, err := page.Element(selector)
		if err != nil {
			return preferElementError(ctx, err, last)
		}

		// Each attempt is capped separately. rod retries a covered element for as long as its context
		// allows, which would spend the whole budget hovering one handle: the element is re-located
		// instead, so a node the document replaced while it was covered is not held on to.
		attempt, cancel := context.WithTimeout(ctx, elementAttemptTimeout)
		err = action(element.Context(attempt))
		expired := attempt.Err() != nil

		cancel()

		if err == nil {
			return nil
		}

		if !expired && !isTransientElementError(err) {
			return err
		}

		// An attempt cut short reports its own expiry rather than the element, so it must not displace
		// the state a previous attempt saw the element in.
		if !expired {
			last = err
		}

		select {
		case <-ctx.Done():
			return preferElementError(ctx, err, last)
		case <-time.After(elementRetryInterval):
		}
	}
}

func preferElementError(ctx context.Context, err, last error) error {
	if last != nil && ctx.Err() != nil {
		return last
	}

	return err
}

func describeElement(page *rod.Page, selector string) string {
	// A context of its own, so the state still reports once the one the action ran under has expired.
	result, err := page.Context(context.Background()).Timeout(elementStateTimeout).Eval(elementState, selector)
	if err != nil {
		return fmt.Sprintf("unavailable: %v", err)
	}

	return result.Value.Str()
}

func isTransientElementError(err error) bool {
	var notInteractable *rod.NotInteractableError

	return isDetachedNodeError(err) || errors.Is(err, errElementValueNotSet) || errors.As(err, &notInteractable)
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
