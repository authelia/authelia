package suites

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/stretchr/testify/require"
)

// RodSession binding a chrome session with devtool protocol.
type RodSession struct {
	Launcher  *launcher.Launcher
	WebDriver *rod.Browser
}

// StartRodWithProxy create a rod/chromedp session.
func StartRodWithProxy(proxy string) (*RodSession, error) {
	browserPath := os.Getenv("BROWSER_PATH")
	if browserPath == "" {
		browserPath = "/usr/bin/chromium-browser"
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
		Proxy(proxy).
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
		Launcher:  l,
		WebDriver: browser,
	}, nil
}

// StartRod create a rod/chromedp session.
func StartRod() (*RodSession, error) {
	return StartRodWithProxy("")
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
