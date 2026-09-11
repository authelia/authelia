// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConformanceRunner_ReportsOutcomesInPlanOrder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch { //nolint:staticcheck // a tagged switch would obscure the multi-condition cases below.
		case r.URL.Path == "/api/runner":
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"` + r.URL.Query().Get("test") + `-id"}`)) //nolint:gosec // test double echoing a query param set by this same test, not attacker controlled.
		case r.URL.Path == "/api/runner/oidcc-server-id/wait-state",
			r.URL.Path == "/api/runner/oidcc-scope-address-id/wait-state":
			_, _ = w.Write([]byte(`{"state":"FINISHED"}`))
		case r.URL.Path == "/api/info/oidcc-server-id":
			_, _ = w.Write([]byte(`{"_id":"oidcc-server-id","status":"FINISHED","result":"PASSED"}`))
		case r.URL.Path == "/api/info/oidcc-scope-address-id":
			_, _ = w.Write([]byte(`{"_id":"oidcc-scope-address-id","status":"FINISHED","result":"WARNING"}`))
		case r.URL.Path == "/api/log/oidcc-server-id", r.URL.Path == "/api/log/oidcc-scope-address-id":
			_, _ = w.Write([]byte(`[]`))
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))

	defer server.Close()

	client, err := NewConformanceClient(server.URL)
	require.NoError(t, err)

	modules := []ConformancePlanModule{{TestModule: "oidcc-server"}, {TestModule: "oidcc-scope-address"}}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	var outcomes []ConformanceOutcome

	for outcome := range NewConformanceRunner(client, nil, "plan1", modules).Run(ctx) {
		outcomes = append(outcomes, outcome)
	}

	require.Len(t, outcomes, 2)

	assert.Equal(t, "Server", outcomes[0].Name)
	assert.Equal(t, "PASSED", outcomes[0].Result)
	require.NoError(t, outcomes[0].Err)

	assert.Equal(t, "ScopeAddress", outcomes[1].Name)
	assert.Equal(t, "WARNING", outcomes[1].Result)
}

func TestConformanceRunner_SkipsUnattendedModules(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("the runner must not contact the server for a skipped module, got %s", r.URL.Path)
	}))

	defer server.Close()

	client, err := NewConformanceClient(server.URL)
	require.NoError(t, err)

	modules := []ConformancePlanModule{{TestModule: "oidcc-server-rotate-keys"}}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	var outcomes []ConformanceOutcome

	for outcome := range NewConformanceRunner(client, nil, "plan1", modules).Run(ctx) {
		outcomes = append(outcomes, outcome)
	}

	require.Len(t, outcomes, 1)
	assert.True(t, outcomes[0].Skipped)
	assert.Equal(t, "ServerRotateKeys", outcomes[0].Name)
	assert.NotEmpty(t, outcomes[0].Reason)
}

func TestConformanceRunner_FillPlaceholdersUploadsOnlyUnfilledEntries(t *testing.T) {
	t.Run("UploadsOnlyTheEntryWithAnUnfilledPlaceholder", func(t *testing.T) {
		var (
			uploadCount  int
			uploadedID   string
			uploadedBody []byte
		)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/api/log/m1":
				_, _ = w.Write([]byte(`[{"msg":"info","result":"INFO"},{"msg":"upload","upload":"ph1","result":"REVIEW"}]`))
			case "/api/log/m1/images/ph1":
				uploadCount++
				uploadedID = "ph1"
				uploadedBody, _ = io.ReadAll(r.Body)

				_, _ = w.Write([]byte(`{}`))
			default:
				t.Errorf("unexpected path %s", r.URL.Path)
			}
		}))

		defer server.Close()

		client, err := NewConformanceClient(server.URL)
		require.NoError(t, err)

		runner := NewConformanceRunner(client, nil, "plan1", nil)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		filled, pending, err := runner.fillPlaceholders(ctx, "m1", conformanceTestScreenshot)
		require.NoError(t, err)

		assert.True(t, filled)
		assert.False(t, pending)
		assert.Equal(t, 1, uploadCount)
		assert.Equal(t, "ph1", uploadedID)
		assert.Equal(t, conformanceTestScreenshot, string(uploadedBody))
	})

	t.Run("UploadsNothingWhenNoPlaceholderIsUnfilled", func(t *testing.T) {
		var uploadCalled bool

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/api/log/m1":
				_, _ = w.Write([]byte(`[{"msg":"info","result":"INFO"}]`))
			case "/api/log/m1/images/ph1":
				uploadCalled = true

				_, _ = w.Write([]byte(`{}`))
			default:
				t.Errorf("unexpected path %s", r.URL.Path)
			}
		}))

		defer server.Close()

		client, err := NewConformanceClient(server.URL)
		require.NoError(t, err)

		runner := NewConformanceRunner(client, nil, "plan1", nil)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		filled, pending, err := runner.fillPlaceholders(ctx, "m1", conformanceTestScreenshot)
		require.NoError(t, err)

		assert.False(t, filled)
		assert.False(t, pending)
		assert.False(t, uploadCalled)
	})

	t.Run("WaitsForTheScreenshotBeforeUploading", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/api/log/m1":
				_, _ = w.Write([]byte(`[{"msg":"upload","upload":"ph1","result":"REVIEW"}]`))
			default:
				t.Errorf("unexpected path %s", r.URL.Path)
			}
		}))

		defer server.Close()

		client, err := NewConformanceClient(server.URL)
		require.NoError(t, err)

		runner := NewConformanceRunner(client, nil, "plan1", nil)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		filled, pending, err := runner.fillPlaceholders(ctx, "m1", "")
		require.NoError(t, err)

		assert.False(t, filled)
		assert.True(t, pending)
	})
}

func TestConformanceOverrideAssertions(t *testing.T) {
	testCases := []struct {
		name     string
		assert   func(leg ConformanceLeg) error
		leg      ConformanceLeg
		expected string
	}{
		{
			"ShouldIgnoreTheFirstLegWhenReauthenticationIsExpected",
			conformanceAssertReauthentication,
			ConformanceLeg{Index: 0, FirstFactor: false},
			"",
		},
		{
			"ShouldIgnoreTheFirstLegEvenWhenItPrompted",
			conformanceAssertReauthentication,
			ConformanceLeg{Index: 0, FirstFactor: true},
			"",
		},
		{
			"ShouldPassWhenTheSecondLegPrompted",
			conformanceAssertReauthentication,
			ConformanceLeg{Index: 1, FirstFactor: true},
			"",
		},
		{
			"ShouldFailWhenTheSecondLegDidNotPrompt",
			conformanceAssertReauthentication,
			ConformanceLeg{Index: 1, FirstFactor: false},
			"expected Authelia to ask for the password again on the second authorization but it did not",
		},
		{
			"ShouldPassWhenTheSecondLegReauthenticatedOnTheConsentForm",
			conformanceAssertReauthentication,
			ConformanceLeg{Index: 1, Consent: true, Reauthentication: true},
			"",
		},
		{
			"ShouldFailWhenTheSecondLegShowedConsentWithoutAskingForThePassword",
			conformanceAssertReauthentication,
			ConformanceLeg{Index: 1, Consent: true, Reauthentication: false},
			"expected Authelia to ask for the password again on the second authorization but it did not",
		},
		{
			"ShouldFailWhenTheSecondLegReauthenticatedOnTheConsentFormButShouldNotHave",
			conformanceAssertNoReauthentication,
			ConformanceLeg{Index: 1, Consent: true, Reauthentication: true},
			"expected Authelia to reuse the existing session on the second authorization but it asked for the password again",
		},
		{
			"ShouldIgnoreTheFirstLegWhenNoReauthenticationIsExpected",
			conformanceAssertNoReauthentication,
			ConformanceLeg{Index: 0, FirstFactor: true},
			"",
		},
		{
			"ShouldPassWhenTheSecondLegDidNotPrompt",
			conformanceAssertNoReauthentication,
			ConformanceLeg{Index: 1, FirstFactor: false},
			"",
		},
		{
			"ShouldFailWhenTheSecondLegPrompted",
			conformanceAssertNoReauthentication,
			ConformanceLeg{Index: 1, FirstFactor: true},
			"expected Authelia to reuse the existing session on the second authorization but it asked for the password again",
		},
		{
			"ShouldNotBeConfusedByConsentOnTheSecondLeg",
			conformanceAssertNoReauthentication,
			ConformanceLeg{Index: 1, FirstFactor: false, Consent: true},
			"",
		},
		{
			"ShouldPassWhenTheLegEndedOnAnAutheliaErrorPage",
			conformanceAssertErrorPage,
			ConformanceLeg{Index: 0, AutheliaError: true},
			"",
		},
		{
			"ShouldFailWhenTheLegReachedTheCallbackInsteadOfAnErrorPage",
			conformanceAssertErrorPage,
			ConformanceLeg{Index: 0},
			"expected Authelia to show an error page rather than redirect to the client but it did not",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.assert(tc.leg)

			if tc.expected == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expected)
			}
		})
	}
}

func TestConformanceOverrides_AreWiredToTheirAssertions(t *testing.T) {
	for module, expected := range map[string]ConformanceLeg{
		"oidcc-prompt-login":  {Index: conformanceReauthenticationLeg, FirstFactor: false},
		"oidcc-max-age-1":     {Index: conformanceReauthenticationLeg, FirstFactor: false},
		"oidcc-max-age-10000": {Index: conformanceReauthenticationLeg, FirstFactor: true},

		"oidcc-ensure-registered-redirect-uri": {Index: 0},
	} {
		override, ok := conformanceOverrides[module]

		require.Truef(t, ok, "the '%s' module has no override", module)
		require.NotNilf(t, override.Assert, "the '%s' module has no assertion", module)

		assert.Errorf(t, override.Assert(expected), "the '%s' module's assertion cannot fail", module)
	}
}

func TestConformanceRunner_VisitURLsDoesNotRedriveOrRecountAVisitedURL(t *testing.T) {
	legs := &conformanceLegs{visited: map[string]bool{"https://conformance.example.com/first": true}, count: 1}

	override := ConformanceOverride{Assert: func(leg ConformanceLeg) error {
		return fmt.Errorf("the assertion must not run for an already driven leg, got %+v", leg)
	}}

	runner := NewConformanceRunner(nil, nil, "plan1", nil)

	progressed, err := runner.visitURLs(context.Background(), "m1", override, legs, []string{"https://conformance.example.com/first"})
	require.NoError(t, err)

	assert.False(t, progressed)
	assert.Equal(t, 1, legs.count)
}

func TestConformanceOverrides_ClearCookiesMatchesTheModulesThatRequireIt(t *testing.T) {
	expected := []string{
		"oidcc-display-page",
		"oidcc-display-popup",
		"oidcc-login-hint",
		"oidcc-prompt-none-not-logged-in",
		"oidcc-ui-locales",
	}

	var actual []string

	for module, override := range conformanceOverrides {
		if override.ClearCookies {
			actual = append(actual, module)
		}
	}

	sort.Strings(actual)

	assert.Equal(t, expected, actual)
}

func TestConformanceRunner_TreatsAnInterruptedModuleAsFinished(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/runner":
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"m1"}`))
		case "/api/runner/m1/wait-state":
			if !strings.Contains(r.URL.Query().Get("states"), conformanceStatusInterrupted) {
				time.Sleep(time.Millisecond * 10)

				_, _ = w.Write([]byte(`{"timeout":true}`))

				return
			}

			_, _ = w.Write([]byte(`{"state":"INTERRUPTED"}`))
		case "/api/info/m1":
			_, _ = w.Write([]byte(`{"_id":"m1","status":"INTERRUPTED","result":"FAILED"}`))
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))

	defer server.Close()

	client, err := NewConformanceClient(server.URL)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	var outcomes []ConformanceOutcome

	for outcome := range NewConformanceRunner(client, nil, "plan1", []ConformancePlanModule{{TestModule: "oidcc-max-age-1"}}).Run(ctx) {
		outcomes = append(outcomes, outcome)
	}

	require.Len(t, outcomes, 1)
	require.NoError(t, outcomes[0].Err)

	assert.Equal(t, conformanceStatusInterrupted, outcomes[0].Status)
	assert.Equal(t, "FAILED", outcomes[0].Result)
}

func TestConformanceRunner_InteractKeepsDrivingAModuleWhichIsRunningBetweenLegs(t *testing.T) {
	var (
		mu                  sync.Mutex
		browserPolls, infos int
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		switch r.URL.Path {
		case "/api/runner/browser/m1":
			browserPolls++

			_, _ = w.Write([]byte(`{"urls":[]}`))
		case "/api/log/m1":
			_, _ = w.Write([]byte(`[]`))
		case "/api/info/m1":
			infos++

			if infos == 1 {
				_, _ = w.Write([]byte(`{"_id":"m1","status":"RUNNING"}`))

				return
			}

			_, _ = w.Write([]byte(`{"_id":"m1","status":"FINISHED"}`))
		case "/api/runner/m1/wait-state":
			assert.Equal(t, "WAITING,FINISHED,INTERRUPTED", r.URL.Query().Get("states"))

			_, _ = w.Write([]byte(`{"state":"WAITING"}`))
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))

	defer server.Close()

	client, err := NewConformanceClient(server.URL)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	runner := NewConformanceRunner(client, &ConformanceBrowser{}, "plan1", nil)

	require.NoError(t, runner.interact(ctx, "m1", ConformanceOverride{}))

	mu.Lock()
	defer mu.Unlock()

	assert.Equal(t, 2, browserPolls, "the browser status must be read again once the module is back in WAITING")
}

func TestConformanceRunner_ReportsTheStatusOfAModuleWhichRanOutOfTime(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/runner":
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"m1"}`))
		case "/api/runner/m1/wait-state":
			if strings.Contains(r.URL.Query().Get("states"), conformanceStatusConfigured) {
				_, _ = w.Write([]byte(`{"state":"FINISHED"}`))

				return
			}

			time.Sleep(time.Millisecond * 10)

			_, _ = w.Write([]byte(`{"timeout":true}`))
		case "/api/info/m1":
			_, _ = w.Write([]byte(`{"_id":"m1","status":"WAITING"}`))
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))

	defer server.Close()

	client, err := NewConformanceClient(server.URL)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var outcomes []ConformanceOutcome

	for outcome := range NewConformanceRunner(client, nil, "plan1", []ConformancePlanModule{{TestModule: "oidcc-server"}}).Run(ctx) {
		outcomes = append(outcomes, outcome)
	}

	require.Len(t, outcomes, 1)
	require.Error(t, outcomes[0].Err)

	assert.Equal(t, conformanceStatusWaiting, outcomes[0].Status)
}

func TestConformanceLegs_Stalled(t *testing.T) {
	t.Run("ShouldNameAutheliasErrorWhenTheLastLegEndedOnItsErrorPage", func(t *testing.T) {
		legs := &conformanceLegs{errorURL: "https://login.example.com:8080/consent/completion?error=invalid_request_object&error_description=The+request+parameter+contains+an+invalid+Request+Object.&error_hint=Could+not+be+validated.&error_debug=Expected+typ+JWT."}

		err := legs.stalled("m1")

		var stall *ConformanceErrorPageStallError

		require.ErrorAs(t, err, &stall)
		assert.EqualError(t, err, "module 'm1' is waiting for the flow to return to the client, but Authelia ended it on an error page and the module asked for no evidence image of it: "+
			"error 'invalid_request_object', description 'The request parameter contains an invalid Request Object.', hint 'Could not be validated.', debug 'Expected typ JWT.'")
	})

	t.Run("ShouldFallBackToTheURLWhenTheErrorPageCarriesNoParameters", func(t *testing.T) {
		legs := &conformanceLegs{errorURL: "https://login.example.com:8080/"}

		assert.EqualError(t, legs.stalled("m1"), "module 'm1' is waiting for the flow to return to the client, but Authelia ended it on an error page and the module asked for no evidence image of it: https://login.example.com:8080/")
	})

	t.Run("ShouldReportAPlainStallOtherwise", func(t *testing.T) {
		legs := &conformanceLegs{}

		err := legs.stalled("m1")

		var stall *ConformanceErrorPageStallError

		assert.False(t, errors.As(err, &stall))
		assert.EqualError(t, err, "module 'm1' stayed in the WAITING state with nothing left to visit or fill")
	})
}

func TestConformanceOverride_ScreenshotOf(t *testing.T) {
	leg := func(index int) ConformanceLeg {
		return ConformanceLeg{Index: index, LoginScreenshot: fmt.Sprintf("login-%d", index), ErrorScreenshot: fmt.Sprintf("error-%d", index)}
	}

	reauthentication := ConformanceOverride{Screenshot: ConformanceScreenshotReauthentication}

	assert.Empty(t, reauthentication.screenshotOf(leg(0)))
	assert.Equal(t, "login-1", reauthentication.screenshotOf(leg(1)))

	errorPage := ConformanceOverride{Screenshot: ConformanceScreenshotErrorPage}

	assert.Equal(t, "error-0", errorPage.screenshotOf(leg(0)))
	assert.Empty(t, errorPage.screenshotOf(ConformanceLeg{Index: 0, LoginScreenshot: "login-0"}))

	assert.Empty(t, ConformanceOverride{}.screenshotOf(leg(1)), "a module with no screenshot kind is given none")
}

func TestConformanceOverrides_ScreenshotEachPlaceholderModule(t *testing.T) {
	expected := map[string]ConformanceScreenshot{
		"oidcc-prompt-login": ConformanceScreenshotReauthentication,
		"oidcc-max-age-1":    ConformanceScreenshotReauthentication,

		"oidcc-ensure-registered-redirect-uri":          ConformanceScreenshotErrorPage,
		"oidcc-ensure-request-object-with-redirect-uri": ConformanceScreenshotErrorPage,
	}

	actual := map[string]ConformanceScreenshot{}

	for module, override := range conformanceOverrides {
		if override.Screenshot != ConformanceScreenshotNone {
			actual[module] = override.Screenshot
		}
	}

	assert.Equal(t, expected, actual)
}

func TestConformanceOverrides_RequestObjectModules(t *testing.T) {
	_, ok := conformanceOverrides["oidcc-unsigned-request-object-supported-correctly-or-rejected-as-unsupported"]
	assert.False(t, ok)

	override := conformanceOverrides["oidcc-ensure-request-object-with-redirect-uri"]

	assert.Nil(t, override.Assert, "either outcome is accepted by the module")
	assert.Equal(t, "error", override.screenshotOf(ConformanceLeg{AutheliaError: true, ErrorScreenshot: "error"}))
	assert.Empty(t, override.screenshotOf(ConformanceLeg{}), "a leg which reached the client has no error page to capture")
}

func TestConformanceLegs_StalledOnAPlaceholderWithoutAScreenshot(t *testing.T) {
	legs := &conformanceLegs{pendingPlaceholder: true, errorURL: "https://login.example.com:8080/consent/completion?error=invalid_request"}

	err := legs.stalled("m1")

	var stall *ConformanceErrorPageStallError

	assert.False(t, errors.As(err, &stall), "a module waiting on its placeholder is not stranded for want of one")
	assert.EqualError(t, err, "module 'm1' is waiting for an evidence image, but no leg showed the page it asks for")
}

func TestConformanceRunner_LetsAFailedModuleFinishBeforeMovingOn(t *testing.T) {
	testCases := []struct {
		name     string
		settle   string
		expected string
	}{
		{"ShouldWaitForAModuleWhichIsFinishing", `{"state":"FINISHED"}`, "FINISHED"},
		{"ShouldNotWaitOnAModuleWhichIsWaitingOnTheSuite", `{"state":"WAITING"}`, "WAITING"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				mu      sync.Mutex
				settled []string
			)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mu.Lock()
				defer mu.Unlock()

				switch r.URL.Path {
				case "/api/runner":
					w.WriteHeader(http.StatusCreated)
					_, _ = w.Write([]byte(`{"id":"m1"}`))
				case "/api/runner/m1/wait-state":
					states := r.URL.Query().Get("states")

					if states == "CONFIGURED,WAITING,FINISHED,INTERRUPTED" && len(settled) == 0 && !strings.Contains(r.URL.RawQuery, "settle") {
						settled = append(settled, "create:"+states)

						_, _ = w.Write([]byte(`{"state":"WAITING"}`))

						return
					}

					settled = append(settled, "settle:"+states)

					_, _ = w.Write([]byte(tc.settle))
				case "/api/info/m1":
					_, _ = w.Write([]byte(`{"_id":"m1","status":"` + tc.expected + `","result":"FAILED"}`))
				default:
					t.Errorf("unexpected path %s", r.URL.Path)
				}
			}))

			defer server.Close()

			client, err := NewConformanceClient(server.URL)
			require.NoError(t, err)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()

			var outcomes []ConformanceOutcome

			for outcome := range NewConformanceRunner(client, nil, "plan1", []ConformancePlanModule{{TestModule: "oidcc-server"}}).Run(ctx) {
				outcomes = append(outcomes, outcome)
			}

			require.Len(t, outcomes, 1)
			require.Error(t, outcomes[0].Err)
			assert.Equal(t, tc.expected, outcomes[0].Status)

			mu.Lock()
			defer mu.Unlock()

			assert.Equal(t, []string{"create:CONFIGURED,WAITING,FINISHED,INTERRUPTED", "settle:CONFIGURED,WAITING,FINISHED,INTERRUPTED"}, settled)
		})
	}
}
