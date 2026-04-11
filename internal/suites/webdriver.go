package suites

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/stretchr/testify/require"
)

// RodSession binding a chrome session with devtool protocol.
type RodSession struct {
	Launcher  *launcher.Launcher
	WebDriver *rod.Browser

	RodSuiteCredentialsProvider
}

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

// RodSessionWithoutDevtools disables Chrome's --auto-open-devtools-for-tabs flag. With
// devtools attached, Chrome draws a device-emulation badge overlay whenever
// setDeviceMetricsOverride is active, which gets captured in full-page screenshots and
// breaks visual snapshot comparisons.
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
		Devtools(!opts.disableDevtools)

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

	return &RodSession{
		Launcher:                    l,
		WebDriver:                   browser,
		RodSuiteCredentialsProvider: opts.provider,
	}, nil
}

// StartRod create a rod/chromedp session.
func StartRod() (*RodSession, error) {
	return NewRodSession()
}

// Stop stop the rod/chromedp session.
func (rs *RodSession) Stop() error {
	err := rs.WebDriver.Close()
	if err != nil {
		return err
	}

	rs.Launcher.Cleanup()

	return err
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

// WaitElementsLocatedBySelector waits for at least one element matching the CSS selector to
// appear, then returns all current matches.
func (rs *RodSession) WaitElementsLocatedBySelector(t *testing.T, page *rod.Page, selector string) rod.Elements {
	_, err := page.Element(selector)
	require.NoError(t, err)

	elements, err := page.Elements(selector)
	require.NoError(t, err)
	require.NotEmpty(t, elements)

	return elements
}

// WaitElementsLocatedByID waits for an elements located by an id.
func (rs *RodSession) WaitElementsLocatedByID(t *testing.T, page *rod.Page, cssSelector string) rod.Elements {
	e, err := page.Elements("#" + cssSelector)
	require.NoError(t, err)
	require.NotNil(t, e)

	return e
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

// SetColorScheme overrides the page's prefers-color-scheme media feature so that pages with
// adaptive theming render in a deterministic mode regardless of the host's system preference.
// Must be called before navigating to the target URL so the page's initial render picks up
// the override.
func (rs *RodSession) SetColorScheme(t *testing.T, page *rod.Page, scheme string) {
	err := proto.EmulationSetEmulatedMedia{
		Features: []*proto.EmulationMediaFeature{
			{Name: "prefers-color-scheme", Value: scheme},
		},
	}.Call(page)
	require.NoError(t, err)
}

// FullPageScreenshot captures a PNG screenshot of the full scrollable page. Scrollbars are
// hidden before capture so a ~15px width delta from scrollbar presence does not make
// dimensions flap between runs.
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
// returning the destination URL from the CDP event (more race-free than reading page.Info()
// immediately after a cross-document navigation).
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
