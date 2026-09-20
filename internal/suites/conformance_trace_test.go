// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConformanceTrace(t *testing.T) {
	t.Run("ShouldRecordNothingUnlessDebugging", func(t *testing.T) {
		trace := NewConformanceTrace(false)

		trace.Logf("driving leg %d", 0)

		assert.Empty(t, trace.Lines())
	})

	t.Run("ShouldRecordEachLineStampedWithTheTime", func(t *testing.T) {
		trace := NewConformanceTrace(true)

		trace.Logf("driving leg %d", 0)
		trace.Logf("finished leg %d", 0)

		lines := trace.Lines()

		require.Len(t, lines, 2)
		assert.Regexp(t, regexp.MustCompile(`^\d{2}:\d{2}:\d{2}\.\d{3} driving leg 0$`), lines[0])
		assert.Regexp(t, regexp.MustCompile(`^\d{2}:\d{2}:\d{2}\.\d{3} finished leg 0$`), lines[1])
	})

	t.Run("ShouldBeSafeWithoutATrace", func(t *testing.T) {
		var trace *ConformanceTrace

		assert.NotPanics(t, func() { trace.Logf("driving leg %d", 0) })
		assert.Empty(t, trace.Lines())
	})
}

func TestSuiteDebug(t *testing.T) {
	t.Setenv("SUITE_DEBUG", "true")
	assert.True(t, suiteDebug())

	t.Setenv("SUITE_DEBUG", "")
	assert.False(t, suiteDebug())
}

func TestConformanceRunner_TracesEachModuleOnlyWhenDebugging(t *testing.T) {
	for _, debug := range []bool{false, true} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/api/runner":
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"id":"m1"}`))
			case "/api/runner/m1/wait-state":
				_, _ = w.Write([]byte(`{"state":"FINISHED"}`))
			case "/api/info/m1":
				_, _ = w.Write([]byte(`{"_id":"m1","status":"FINISHED","result":"PASSED"}`))
			default:
				t.Errorf("unexpected path %s", r.URL.Path)
			}
		}))

		client, err := NewConformanceClient(server.URL)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

		runner := NewConformanceRunner(client, nil, "plan1", []ConformancePlanModule{{TestModule: "oidcc-server"}})
		runner.debug = debug

		var outcomes []ConformanceOutcome

		for outcome := range runner.Run(ctx) {
			outcomes = append(outcomes, outcome)
		}

		cancel()
		server.Close()

		require.Len(t, outcomes, 1)

		if debug {
			require.NotEmpty(t, outcomes[0].Trace, "a debugging run records what the module did")
			assert.Contains(t, outcomes[0].Trace[0], "created as 'm1' in state 'FINISHED'")
		} else {
			assert.Empty(t, outcomes[0].Trace, "a run which is not debugging records nothing")
		}
	}
}
