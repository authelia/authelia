// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConformanceBrowser_Drive(t *testing.T) {
	driver := newConformanceTestBrowser(t)
	portal := conformanceFakePortal(t)

	drive := func(t *testing.T, index int, path string) ConformanceLeg {
		t.Helper()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()

		leg, err := driver.Drive(ctx, index, portal.URL+path)
		require.NoError(t, err)

		return leg
	}

	t.Run("ShouldSignInThenAnswerTheConsentFormsReauthentication", func(t *testing.T) {
		leg := drive(t, 1, "/authorize-reauthenticate")

		assert.Equal(t, 1, leg.Index)
		assert.True(t, leg.FirstFactor)
		assert.True(t, leg.Consent)
		assert.True(t, leg.Reauthentication)
		assert.False(t, leg.AutheliaError)
		assert.True(t, strings.HasPrefix(leg.LoginScreenshot, "data:image/png;base64,"), "the password step is captured")
		assert.Empty(t, leg.ErrorScreenshot)
	})

	t.Run("ShouldAcceptConsentWhichAsksForNoPassword", func(t *testing.T) {
		leg := drive(t, 0, "/consent")

		assert.True(t, leg.Consent)
		assert.False(t, leg.Reauthentication)
		assert.False(t, leg.FirstFactor)
		assert.Empty(t, leg.LoginScreenshot)
	})

	t.Run("ShouldDriveALegWhichStartsOnTheLastLegsCallback", func(t *testing.T) {
		drive(t, 0, "/consent")

		leg := drive(t, 1, "/error")

		assert.True(t, leg.AutheliaError)
	})

	t.Run("ShouldEndTheLegOnAnAutheliaErrorPage", func(t *testing.T) {
		leg := drive(t, 0, "/error")

		assert.True(t, leg.AutheliaError)
		assert.Equal(t, portal.URL+"/error", leg.ErrorURL)
		assert.True(t, strings.HasPrefix(leg.ErrorScreenshot, "data:image/png;base64,"), "the error page is captured")
	})

	t.Run("ShouldHandCoverageOverAsThePageIsLeft", func(t *testing.T) {
		before, _ := os.ReadDir(coverageDir)

		drive(t, 0, "/coverage")

		assert.Eventually(t, func() bool {
			after, _ := os.ReadDir(coverageDir)

			return len(after) > len(before)
		}, time.Second*5, time.Millisecond*50)
	})

	t.Run("ShouldClearTheContextsCookies", func(t *testing.T) {
		drive(t, 0, "/authorize")

		cookies, err := driver.page.Cookies(nil)
		require.NoError(t, err)
		require.NotEmpty(t, cookies)

		require.NoError(t, driver.ClearCookies())

		cookies, err = driver.page.Cookies(nil)
		require.NoError(t, err)
		assert.Empty(t, cookies)
	})

	t.Run("ShouldReportANavigationWhichFailed", func(t *testing.T) {
		_, err := driver.Drive(context.Background(), 0, "http://127.0.0.1:1/")

		assert.ErrorContains(t, err, "error navigating to 'http://127.0.0.1:1/'")
	})

	t.Run("ShouldStopWhenTheContextHasExpired", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := driver.Drive(ctx, 0, portal.URL+"/error")

		assert.ErrorContains(t, err, "did not reach the conformance callback")
	})
}

func TestConformanceRunner_VisitURLsDrivesEachLeg(t *testing.T) {
	driver := newConformanceTestBrowser(t)
	portal := conformanceFakePortal(t)

	var (
		mu      sync.Mutex
		visited []string
	)

	visit := func(status int) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()

			visited = append(visited, r.URL.Query().Get("url"))
			mu.Unlock()

			w.WriteHeader(status)
		}
	}

	t.Run("ShouldVisitALegOnceItIsDriven", func(t *testing.T) {
		client := newConformanceFakeClient(t, map[string]http.HandlerFunc{"/api/runner/browser/m1/visit": visit(http.StatusNoContent)})

		runner := NewConformanceRunner(client, driver, "plan1", nil)
		runner.trace = NewConformanceTrace(true)

		legs := &conformanceLegs{visited: map[string]bool{}}

		progressed, err := runner.visitURLs(context.Background(), "m1", ConformanceOverride{Screenshot: ConformanceScreenshotErrorPage}, legs, []string{portal.URL + "/error"})

		require.NoError(t, err)
		assert.True(t, progressed)
		assert.Equal(t, portal.URL+"/error", legs.errorURL)
		assert.True(t, strings.HasPrefix(legs.screenshot, "data:image/png;base64,"))
		assert.Equal(t, []string{portal.URL + "/error"}, visited)
		assert.Len(t, runner.trace.Lines(), 2, "the leg is traced as it starts and as it ends")
	})

	t.Run("ShouldNotVisitALegWhoseAssertionFailed", func(t *testing.T) {
		client := newConformanceFakeClient(t, map[string]http.HandlerFunc{})

		legs := &conformanceLegs{visited: map[string]bool{}}

		_, err := NewConformanceRunner(client, driver, "plan1", nil).visitURLs(context.Background(), "m1", ConformanceOverride{Assert: conformanceAssertErrorPage}, legs, []string{portal.URL + "/consent"})

		assert.EqualError(t, err, "expected Authelia to show an error page rather than redirect to the client but it did not")
	})

	t.Run("ShouldReportAVisitWhichFailed", func(t *testing.T) {
		client := newConformanceFakeClient(t, map[string]http.HandlerFunc{"/api/runner/browser/m1/visit": visit(http.StatusInternalServerError)})

		legs := &conformanceLegs{visited: map[string]bool{}}

		_, err := NewConformanceRunner(client, driver, "plan1", nil).visitURLs(context.Background(), "m1", ConformanceOverride{}, legs, []string{portal.URL + "/consent-reauthenticate"})

		assert.ErrorContains(t, err, "expected status 204 but got 500")
	})

	t.Run("ShouldReportALegWhichCouldNotBeDriven", func(t *testing.T) {
		client := newConformanceFakeClient(t, map[string]http.HandlerFunc{})

		legs := &conformanceLegs{visited: map[string]bool{}}

		_, err := NewConformanceRunner(client, driver, "plan1", nil).visitURLs(context.Background(), "m1", ConformanceOverride{}, legs, []string{"http://127.0.0.1:1/"})

		assert.ErrorContains(t, err, "error navigating")
	})
}

func TestConformanceBrowser_AfterClose(t *testing.T) {
	driver := newConformanceTestBrowser(t)

	driver.Close()

	assert.Empty(t, driver.url())
	assert.Empty(t, driver.title())
	assert.False(t, driver.has(conformanceSelectorFirstFactor))
	assert.Empty(t, driver.screenshot())

	assert.NotPanics(t, driver.Close, "closing twice is harmless")
}

func TestConformanceRunner_WithABrowser(t *testing.T) {
	driver := newConformanceTestBrowser(t)

	t.Run("ShouldClearCookiesForAModuleWhichAsksForIt", func(t *testing.T) {
		client := newConformanceFakeClient(t, map[string]http.HandlerFunc{
			"/api/runner":               conformanceReply(http.StatusCreated, `{"id":"m1"}`),
			"/api/runner/m1/wait-state": conformanceReply(http.StatusOK, `{"state":"FINISHED"}`),
			"/api/info/m1":              conformanceReply(http.StatusOK, `{"_id":"m1","status":"FINISHED","result":"PASSED"}`),
		})

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		var outcomes []ConformanceOutcome

		for outcome := range NewConformanceRunner(client, driver, "plan1", []ConformancePlanModule{{TestModule: "oidcc-display-page"}}).Run(ctx) {
			outcomes = append(outcomes, outcome)
		}

		require.Len(t, outcomes, 1)
		require.NoError(t, outcomes[0].Err)
		assert.Equal(t, "PASSED", outcomes[0].Result)
	})

	t.Run("ShouldReportALegWhichCouldNotBeDriven", func(t *testing.T) {
		client := newConformanceFakeClient(t, map[string]http.HandlerFunc{
			"/api/runner/browser/m1": conformanceReply(http.StatusOK, `{"urls":["http://127.0.0.1:1/"]}`),
		})

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		err := NewConformanceRunner(client, driver, "plan1", nil).interact(ctx, "m1", ConformanceOverride{})

		assert.ErrorContains(t, err, "error navigating to 'http://127.0.0.1:1/'")
	})
}

func conformanceFakePortal(t *testing.T) *httptest.Server {
	t.Helper()

	const callback = "/test/a/alias/callback?code=abc"

	page := func(w http.ResponseWriter, body string) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = fmt.Fprintf(w, "<!DOCTYPE html><html><head><title>Login - Authelia</title></head><body>%s</body></html>", body)
	}

	mux := http.NewServeMux()

	for _, variant := range []struct {
		suffix           string
		reauthentication bool
	}{
		{"", false},
		{"-reauthenticate", true},
	} {
		mux.HandleFunc("/authorize"+variant.suffix, func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/login"+variant.suffix, http.StatusFound)
		})

		mux.HandleFunc("/login"+variant.suffix, func(w http.ResponseWriter, _ *http.Request) {
			page(w, fmt.Sprintf(`<div id="first-factor-stage">
<input id="username-textfield"><input id="password-textfield" type="password">
<button id="sign-in-button" onclick="signIn()">Sign in</button></div>
<script>
document.cookie = 'authelia_session=session; path=/';

function signIn() {
  if (document.getElementById('username-textfield').value === %q &&
    document.getElementById('password-textfield').value === %q) location.href = %q;
}
</script>`, testUsername, testPassword, "/consent"+variant.suffix))
		})

		mux.HandleFunc("/consent"+variant.suffix, func(w http.ResponseWriter, _ *http.Request) {
			page(w, fmt.Sprintf(`<div id="openid-consent-decision-stage">
<button id="openid-consent-accept" onclick="accept()">Accept</button></div>
<script>
function accept() {
  if (%t) {
    const stage = document.getElementById('openid-consent-decision-stage');
    stage.innerHTML = '<div id="openid-consent-prompt-login"><input id="password-textfield" type="password"></div>' +
      '<button id="openid-consent-authenticate" onclick="authenticate()">Authenticate</button>';

    return;
  }

  location.href = %q;
}

function authenticate() {
  if (document.querySelector('#openid-consent-prompt-login #password-textfield').value === %q) location.href = %q;
}
</script>`, variant.reauthentication, callback, testPassword, callback))
		})
	}

	mux.HandleFunc("/error", func(w http.ResponseWriter, _ *http.Request) {
		page(w, `<div data-testid="openid-completion-outcome" data-outcome="error">An error occurred processing the request</div>`)
	})

	mux.HandleFunc("/coverage", func(w http.ResponseWriter, _ *http.Request) {
		page(w, fmt.Sprintf(`<script>window.__coverage__ = {"/node/src/app/src/index.tsx": {"s": {"0": 1}}};
setTimeout(() => { location.href = %q; }, 100);</script>`, callback))
	})

	mux.HandleFunc("/test/a/alias/callback", func(w http.ResponseWriter, _ *http.Request) {
		page(w, "callback")
	})

	server := httptest.NewServer(mux)

	t.Cleanup(server.Close)

	return server
}

func newConformanceTestBrowser(t *testing.T) *ConformanceBrowser {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	preserveCoverageDir(t)

	path, err := GetBrowserPath()
	require.NoError(t, err)

	l := launcher.New().Bin(path).Headless(true)
	t.Cleanup(l.Cleanup)

	browser := rod.New().ControlURL(l.MustLaunch()).MustConnect()

	t.Cleanup(func() { _ = browser.Close() })

	driver, err := NewConformanceBrowser(&RodSession{WebDriver: browser})
	require.NoError(t, err)

	t.Cleanup(driver.Close)

	return driver
}

func preserveCoverageDir(t *testing.T) {
	t.Helper()

	existed := map[string]bool{}

	entries, err := os.ReadDir(coverageDir)
	dirExisted := err == nil

	for _, entry := range entries {
		existed[entry.Name()] = true
	}

	t.Cleanup(func() {
		entries, _ := os.ReadDir(coverageDir)

		for _, entry := range entries {
			if !existed[entry.Name()] {
				_ = os.Remove(filepath.Join(coverageDir, entry.Name()))
			}
		}

		if !dirExisted {
			_ = os.Remove(coverageDir)
		}
	})
}
