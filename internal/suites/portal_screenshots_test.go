// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

//go:build !externalsuites
// +build !externalsuites

package suites

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-rod/rod"
	"github.com/stretchr/testify/require"
)

// The portal screenshots published in the README and on the documentation site. Each theme is
// captured at portalScreenshotWidth by portalScreenshotHeight and emitted at
// portalScreenshotScale, giving a 1350x2100 image.
//
// The height is what the layout needs rather than a frame chosen for its proportions: the sign in
// column runs to 624 pixels at this width, so the brand footer falls past the fold on anything
// shorter than 640 and the taller frame leaves it room to grow.
const (
	portalScreenshotWidth  = 450
	portalScreenshotHeight = 700
	portalScreenshotScale  = 3

	portalScreenshotDir = "docs/static/images"
	portalScreenshotEnv = "AUTHELIA_PORTAL_SCREENSHOTS"
)

// TestPortalScreenshots regenerates the portal screenshots from a running DuoPush suite, which is
// the suite that offers all three second factor methods. It rewrites tracked files, so it only runs
// when portalScreenshotEnv is set:
//
//	authelia-scripts suites setup DuoPush
//	AUTHELIA_PORTAL_SCREENSHOTS=true go test -count=1 -v ./internal/suites -run '^TestPortalScreenshots$'
//
// The sign in pages are reproducible byte for byte. The method selection dialogs are not, because
// the time-based one-time password icon draws the portion of the period that remains, so those two
// images differ by the handful of pixels that wedge covers from one run to the next.
func TestPortalScreenshots(t *testing.T) {
	if os.Getenv(portalScreenshotEnv) == "" {
		t.Skipf("skipping portal screenshot generation, set %s to regenerate", portalScreenshotEnv)
	}

	session, err := NewRodSession(RodSessionWithoutDevtools())
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, session.Stop())
	})

	cwd, err := os.Getwd()
	require.NoError(t, err)

	dir := filepath.Join(strings.TrimSuffix(cwd, "/internal/suites"), portalScreenshotDir)

	for _, theme := range []string{"light", "dark"} {
		t.Run(theme, func(t *testing.T) {
			session.doCapturePortalScreenshots(t, dir, theme)
		})
	}
}

// doCapturePortalScreenshots writes the sign in page and the second factor method selection dialog
// for a single theme.
func (rs *RodSession) doCapturePortalScreenshots(t *testing.T, dir, theme string) {
	page := rs.doCreateTab(t, LoginBaseURL)
	defer page.MustClose()

	page.MustSetViewport(portalScreenshotWidth, portalScreenshotHeight, portalScreenshotScale, false)

	// The stored theme takes precedence over the one the server supplies, so the capture does not
	// depend on the suite's configured theme.
	_, err := page.Eval(fmt.Sprintf(`() => localStorage.setItem("theme.name", %q)`, theme))
	require.NoError(t, err)

	rs.doVisitLoginPage(t, page, BaseDomain, "")
	rs.WaitElementLocatedByID(t, page, "username-textfield")
	rs.WaitForVisualStable(t, page)

	// The portal focuses the username field on load, which would leave a focus ring on the image.
	_, err = page.Eval(`() => document.activeElement instanceof HTMLElement && document.activeElement.blur()`)
	require.NoError(t, err)

	rs.doWaitForAnimations(t, page)

	doWritePortalScreenshot(t, dir, fmt.Sprintf("%s.png", theme), rs.ViewportScreenshot(t, page))

	rs.doFillLoginPageAndClick(t, page, testUsername, testPassword, false)
	rs.ClickElementLocatedByID(t, page, "methods-button")
	rs.WaitElementLocatedByID(t, page, "methods-dialog")
	rs.WaitForVisualStable(t, page)
	rs.doWaitForAnimations(t, page)

	doWritePortalScreenshot(t, dir, fmt.Sprintf("2fa-methods-%s.png", theme), rs.ViewportScreenshot(t, page))
}

// doWaitForAnimations blocks until every running transition and animation has finished, so a dialog
// is captured fully opened rather than part way through its opening transition.
func (rs *RodSession) doWaitForAnimations(t *testing.T, page *rod.Page) {
	_, err := page.Eval(`async () => {
		await new Promise(resolve => requestAnimationFrame(resolve));
		await Promise.all(document.getAnimations().map(animation => animation.finished.catch(() => {})));
		await new Promise(resolve => requestAnimationFrame(resolve));
		return true;
	}`)
	require.NoError(t, err)
}

func doWritePortalScreenshot(t *testing.T, dir, name string, screenshot []byte) {
	path := filepath.Join(dir, name)

	require.NoError(t, os.WriteFile(path, screenshot, 0600))

	t.Logf("wrote %s (%d bytes)", path, len(screenshot))
}
