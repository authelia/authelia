package suites

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/launcher"
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
	proxy    string
	provider RodSuiteCredentialsProvider
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
		Devtools(true)
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

// CheckElementExistsLocatedByID checks the existence of an element located by an id.
func (rs *RodSession) CheckElementExistsLocatedByID(t *testing.T, page *rod.Page, cssSelector string) bool {
	b, _, err := page.Has("#" + cssSelector)
	require.NoError(t, err)

	return b
}

// WaitElementLocatedByClassName wait an element is located by class name.
func (rs *RodSession) WaitElementLocatedByClassName(t *testing.T, page *rod.Page, className string) *rod.Element {
	e, err := page.Element("." + className)
	require.NoError(t, err)
	require.NotNil(t, e)

	return e
}

// WaitElementLocatedByID waits for an element located by an id.
func (rs *RodSession) WaitElementLocatedByID(t *testing.T, page *rod.Page, cssSelector string) *rod.Element {
	e, err := page.Element("#" + cssSelector)
	require.NoError(t, err)
	require.NotNil(t, e)

	return e
}

// WaitElementsLocatedByID waits for an elements located by an id.
func (rs *RodSession) WaitElementsLocatedByID(t *testing.T, page *rod.Page, cssSelector string) rod.Elements {
	e, err := page.Elements("#" + cssSelector)
	require.NoError(t, err)
	require.NotNil(t, e)

	return e
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
