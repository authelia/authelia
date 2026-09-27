// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConformanceRunner_ReportsTheStepWhichFailed(t *testing.T) {
	created := conformanceReply(http.StatusCreated, `{"id":"m1"}`)

	testCases := []struct {
		name     string
		routes   map[string]http.HandlerFunc
		expected string
		id       string
	}{
		{
			"ShouldReportAModuleWhichCouldNotBeCreated",
			map[string]http.HandlerFunc{"/api/runner": conformanceReply(http.StatusInternalServerError, "broken")},
			"error creating module 'oidcc-server'",
			"",
		},
		{
			"ShouldReportAModuleWhichCouldNotBeWaitedOn",
			map[string]http.HandlerFunc{
				"/api/runner":               created,
				"/api/runner/m1/wait-state": conformanceReply(http.StatusBadRequest, "bad"),
			},
			"expected status 200 but got 400",
			"m1",
		},
		{
			"ShouldReportAConfiguredModuleWhichCouldNotBeStarted",
			map[string]http.HandlerFunc{
				"/api/runner":               created,
				"/api/runner/m1/wait-state": conformanceReply(http.StatusOK, `{"state":"CONFIGURED"}`),
				"POST /api/runner/m1":       conformanceReply(http.StatusInternalServerError, "broken"),
			},
			"expected status 200 but got 500",
			"m1",
		},
		{
			"ShouldReportAModuleWhoseStatusCouldNotBeRead",
			map[string]http.HandlerFunc{
				"/api/runner":               created,
				"/api/runner/m1/wait-state": conformanceReply(http.StatusOK, `{"state":"FINISHED"}`),
				"/api/info/m1":              conformanceReply(http.StatusInternalServerError, "broken"),
			},
			"expected status 200 but got 500",
			"m1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			outcome := conformanceRunOne(t, newConformanceFakeClient(t, tc.routes))

			require.Error(t, outcome.Err)
			assert.ErrorContains(t, outcome.Err, tc.expected)
			assert.Equal(t, tc.id, outcome.ID)
		})
	}
}

func TestConformanceRunner_StartsAConfiguredModule(t *testing.T) {
	var started atomic.Bool

	client := newConformanceFakeClient(t, map[string]http.HandlerFunc{
		"/api/runner": conformanceReply(http.StatusCreated, `{"id":"m1"}`),
		"/api/runner/m1/wait-state": func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.Query().Get("states"), conformanceStatusConfigured) {
				_, _ = w.Write([]byte(`{"state":"CONFIGURED"}`))

				return
			}

			_, _ = w.Write([]byte(`{"state":"FINISHED"}`))
		},
		"POST /api/runner/m1": func(w http.ResponseWriter, _ *http.Request) {
			started.Store(true)

			_, _ = w.Write([]byte(`{}`))
		},
		"/api/info/m1": conformanceReply(http.StatusOK, `{"_id":"m1","status":"FINISHED","result":"PASSED"}`),
	})

	outcome := conformanceRunOne(t, client)

	require.NoError(t, outcome.Err)
	assert.True(t, started.Load())
	assert.Equal(t, "PASSED", outcome.Result)
}

func TestConformanceRunner_InteractReportsTheCallWhichFailed(t *testing.T) {
	status := conformanceReply(http.StatusOK, `{"urls":[]}`)
	log := conformanceReply(http.StatusOK, `[]`)

	testCases := []struct {
		name     string
		routes   map[string]http.HandlerFunc
		expected string
	}{
		{
			"ShouldReportTheBrowserStatus",
			map[string]http.HandlerFunc{"/api/runner/browser/m1": conformanceReply(http.StatusInternalServerError, "status broken")},
			"status broken",
		},
		{
			"ShouldReportTheLog",
			map[string]http.HandlerFunc{
				"/api/runner/browser/m1": status,
				"/api/log/m1":            conformanceReply(http.StatusInternalServerError, "log broken"),
			},
			"log broken",
		},
		{
			"ShouldReportTheStatus",
			map[string]http.HandlerFunc{
				"/api/runner/browser/m1": status,
				"/api/log/m1":            log,
				"/api/info/m1":           conformanceReply(http.StatusInternalServerError, "info broken"),
			},
			"info broken",
		},
		{
			"ShouldReportTheWaitForAModuleBetweenLegs",
			map[string]http.HandlerFunc{
				"/api/runner/browser/m1":    status,
				"/api/log/m1":               log,
				"/api/info/m1":              conformanceReply(http.StatusOK, `{"_id":"m1","status":"RUNNING"}`),
				"/api/runner/m1/wait-state": conformanceReply(http.StatusBadRequest, "wait broken"),
			},
			"wait broken",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runner := NewConformanceRunner(newConformanceFakeClient(t, tc.routes), &ConformanceBrowser{}, "plan1", nil)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()

			assert.ErrorContains(t, runner.interact(ctx, "m1", ConformanceOverride{}), tc.expected)
		})
	}
}

func TestConformanceRunner_FillPlaceholdersReportsAnUploadWhichFailed(t *testing.T) {
	client := newConformanceFakeClient(t, map[string]http.HandlerFunc{
		"/api/log/m1":            conformanceReply(http.StatusOK, `[{"msg":"upload","upload":"ph1","result":"REVIEW"}]`),
		"/api/log/m1/images/ph1": conformanceReply(http.StatusInternalServerError, "upload broken"),
	})

	filled, pending, err := NewConformanceRunner(client, nil, "plan1", nil).fillPlaceholders(context.Background(), "m1", conformanceTestScreenshot)

	assert.ErrorContains(t, err, "upload broken")
	assert.False(t, filled)
	assert.False(t, pending)
}

func TestConformanceRunner_InteractStopsWhenTheContextHasExpired(t *testing.T) {
	runner := NewConformanceRunner(nil, &ConformanceBrowser{}, "plan1", nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	assert.ErrorContains(t, runner.interact(ctx, "m1", ConformanceOverride{}), "module 'm1' did not leave the WAITING state")
}

func TestConformanceRunner_InteractPollsAWaitingModuleAgain(t *testing.T) {
	var infos atomic.Int32

	client := newConformanceFakeClient(t, map[string]http.HandlerFunc{
		"/api/runner/browser/m1": conformanceReply(http.StatusOK, `{"urls":[]}`),
		"/api/log/m1":            conformanceReply(http.StatusOK, `[]`),
		"/api/info/m1": func(w http.ResponseWriter, _ *http.Request) {
			if infos.Add(1) == 1 {
				_, _ = w.Write([]byte(`{"_id":"m1","status":"WAITING"}`))

				return
			}

			_, _ = w.Write([]byte(`{"_id":"m1","status":"FINISHED"}`))
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	require.NoError(t, NewConformanceRunner(client, &ConformanceBrowser{}, "plan1", nil).interact(ctx, "m1", ConformanceOverride{}))
	assert.Equal(t, int32(2), infos.Load())
}

func TestConformanceRunner_ModuleState(t *testing.T) {
	t.Run("ShouldSettleAfterAnEvidenceImage", func(t *testing.T) {
		client := newConformanceFakeClient(t, map[string]http.HandlerFunc{
			"/api/info/m1": conformanceReply(http.StatusOK, `{"_id":"m1","status":"FINISHED"}`),
		})

		runner := NewConformanceRunner(client, nil, "plan1", nil)
		runner.trace = NewConformanceTrace(true)

		state, err := runner.moduleState(context.Background(), "m1", true)

		require.NoError(t, err)
		assert.Equal(t, conformanceStatusFinished, state)
		require.Len(t, runner.trace.Lines(), 1)
		assert.Contains(t, runner.trace.Lines()[0], "Module 'm1' is in state 'FINISHED' after its evidence image was uploaded")
	})

	t.Run("ShouldReportAStatusWhichCouldNotBeReadWhileSettling", func(t *testing.T) {
		client := newConformanceFakeClient(t, map[string]http.HandlerFunc{
			"/api/info/m1": conformanceReply(http.StatusInternalServerError, "broken"),
		})

		_, err := NewConformanceRunner(client, nil, "plan1", nil).moduleState(context.Background(), "m1", true)

		assert.ErrorContains(t, err, "while waiting for it to take up its evidence image")
	})
}

func TestConformanceDescribeErrorPageFallsBackForAnUnparsableURL(t *testing.T) {
	assert.Equal(t, "%zz", conformanceDescribeErrorPage("%zz"))
}

func TestConformanceRunner_HandsTheModuleTraceToItsBrowser(t *testing.T) {
	client := newConformanceFakeClient(t, map[string]http.HandlerFunc{
		"/api/runner":               conformanceReply(http.StatusCreated, `{"id":"m1"}`),
		"/api/runner/m1/wait-state": conformanceReply(http.StatusOK, `{"state":"FINISHED"}`),
		"/api/info/m1":              conformanceReply(http.StatusOK, `{"_id":"m1","status":"FINISHED","result":"PASSED"}`),
	})

	browser := &ConformanceBrowser{}
	runner := NewConformanceRunner(client, browser, "plan1", []ConformancePlanModule{{TestModule: "oidcc-server"}})
	runner.debug = true

	for outcome := range runner.Run(context.Background()) {
		require.NoError(t, outcome.Err)
	}

	assert.Same(t, runner.trace, browser.trace, "what the driver does is recorded in the module's trace")
}

func newConformanceFakeClient(t *testing.T, routes map[string]http.HandlerFunc) *ConformanceClient {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route, ok := routes[r.Method+" "+r.URL.Path]
		if !ok {
			route, ok = routes[r.URL.Path]
		}

		if !ok {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)

			return
		}

		route(w, r)
	}))

	t.Cleanup(server.Close)

	client, err := NewConformanceClient(server.URL)
	require.NoError(t, err)

	return client
}

func conformanceReply(status int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
}

func conformanceRunOne(t *testing.T, client *ConformanceClient) ConformanceOutcome {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	var outcomes []ConformanceOutcome

	for outcome := range NewConformanceRunner(client, nil, "plan1", []ConformancePlanModule{{TestModule: "oidcc-server"}}).Run(ctx) {
		outcomes = append(outcomes, outcome)
	}

	require.Len(t, outcomes, 1)

	return outcomes[0]
}
