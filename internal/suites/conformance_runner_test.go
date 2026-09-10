// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
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

		filled, err := runner.fillPlaceholders(ctx, "m1")
		require.NoError(t, err)

		assert.True(t, filled)
		assert.Equal(t, 1, uploadCount)
		assert.Equal(t, "ph1", uploadedID)
		assert.Equal(t, conformancePlaceholderImage, string(uploadedBody))
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

		filled, err := runner.fillPlaceholders(ctx, "m1")
		require.NoError(t, err)

		assert.False(t, filled)
		assert.False(t, uploadCalled)
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
			// The plan's browser context may already hold a session from an earlier module, so the first
			// authorization is not expected to prompt and says nothing about the parameter under test.
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
			// The route Authelia actually takes: prompt=login redirects to the consent decision endpoint and asks
			// for the password there, so the first factor stage is never raised.
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
			// This is the case the previous contract could not express: an assertion which cannot observe a prompt
			// cannot fail, and a test which cannot fail is not a test.
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

func TestConformanceOverrides_AreWiredToTheReauthenticationAssertions(t *testing.T) {
	// The wiring is what decides which module each expectation is applied to, and it is not otherwise exercised by
	// anything which runs without a conformance server.
	for module, expected := range map[string]ConformanceLeg{
		"oidcc-prompt-login":  {Index: conformanceReauthenticationLeg, FirstFactor: false},
		"oidcc-max-age-1":     {Index: conformanceReauthenticationLeg, FirstFactor: false},
		"oidcc-max-age-10000": {Index: conformanceReauthenticationLeg, FirstFactor: true},
	} {
		override, ok := conformanceOverrides[module]

		require.Truef(t, ok, "the '%s' module has no override", module)
		require.NotNilf(t, override.Assert, "the '%s' module has no assertion", module)

		assert.Errorf(t, override.Assert(expected), "the '%s' module's assertion cannot fail", module)
	}
}

func TestConformanceRunner_VisitURLsDoesNotRedriveOrRecountAVisitedURL(t *testing.T) {
	// A module's browser status repeats the URLs it has already published on every poll, so the leg index only stays
	// meaningful if a URL which has already been driven neither advances the count nor reaches the browser again -
	// there is no browser or client on this runner, so either would panic rather than pass.
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

// TestConformanceOverrides_ClearCookiesMatchesTheModulesThatRequireIt pins the cookie-clearing set to the modules
// whose upstream summaries ask for it.
//
// A module missing from this set does not fail loudly. It inherits whatever session the plan's browser is carrying
// from the module before it, and most of them tolerate that -- oidcc-prompt-none-not-logged-in does not, because
// prompt=none against a live session returns a code where the module is asserting an error. An entry present without
// an upstream basis is the same problem in reverse: it makes the table look considered when it is guessing.
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
