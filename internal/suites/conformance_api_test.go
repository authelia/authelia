// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/oidc/conformance"
)

const conformanceTestScreenshot = "data:image/png;base64,iVBORw0KGgo="

func TestConformanceClient_CreatePlan(t *testing.T) {
	var query string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/plan", r.URL.Path)

		query = r.URL.RawQuery

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"name":"oidcc-basic-certification-test-plan","id":"plan1","modules":[{"testModule":"oidcc-server","variant":{"response_type":"code"}}]}`))
	}))

	defer server.Close()

	client, err := NewConformanceClient(server.URL)
	require.NoError(t, err)

	plan := &conformance.Plan{Alias: "alias"}

	created, err := client.CreatePlan(context.Background(), "oidcc-basic-certification-test-plan", &conformance.PlanVariant{ServerMetadata: "discovery"}, plan)
	require.NoError(t, err)

	assert.Equal(t, "plan1", created.ID)
	require.Len(t, created.Modules, 1)
	assert.Equal(t, "oidcc-server", created.Modules[0].TestModule)
	assert.Equal(t, "code", created.Modules[0].Variant["response_type"])
	assert.Contains(t, query, "planName=oidcc-basic-certification-test-plan")
	assert.Contains(t, query, "variant=")
}

func TestConformanceClient_CreatePlanRejectsAResponseWithoutAnIdentifier(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"name":"oidcc-basic-certification-test-plan","modules":[{"testModule":"oidcc-server"}]}`))
	}))

	defer server.Close()

	client, err := NewConformanceClient(server.URL)
	require.NoError(t, err)

	_, err = client.CreatePlan(context.Background(), "oidcc-basic-certification-test-plan", nil, &conformance.Plan{Alias: "the-alias"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "the-alias")
	assert.Contains(t, err.Error(), "returned no plan identifier")
}

func TestConformanceClient_WaitStateReturnsPersistedStatusWhenTestNoLongerRunning(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/runner/abc/wait-state":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"test not found"}`))
		case "/api/info/abc":
			_, _ = w.Write([]byte(`{"_id":"abc","testName":"oidcc-server","status":"FINISHED","result":"PASSED"}`))
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))

	defer server.Close()

	client, err := NewConformanceClient(server.URL)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	state, err := client.WaitState(ctx, "abc", "FINISHED")
	require.NoError(t, err)
	assert.Equal(t, "FINISHED", state)
}

func TestConformanceClient_WaitStateRetriesOnTimeout(t *testing.T) {
	var calls int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/runner/abc/wait-state", r.URL.Path)

		calls++

		if calls == 1 {
			_, _ = w.Write([]byte(`{"timeout":true}`))

			return
		}

		_, _ = w.Write([]byte(`{"state":"WAITING"}`))
	}))

	defer server.Close()

	client, err := NewConformanceClient(server.URL)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	state, err := client.WaitState(ctx, "abc", "WAITING")
	require.NoError(t, err)
	assert.Equal(t, "WAITING", state)
	assert.Equal(t, 2, calls)
}

func TestConformanceClient_WaitStateRetriesOnTransientError(t *testing.T) {
	var calls int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/runner/abc/wait-state":
			calls++

			if calls == 1 {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error":"internal error"}`))

				return
			}

			_, _ = w.Write([]byte(`{"state":"FINISHED"}`))
		case "/api/info/abc":
			t.Errorf("unexpected call to %s: a transient 5xx must not fall back to Info", r.URL.Path)
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))

	defer server.Close()

	client, err := NewConformanceClient(server.URL)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	state, err := client.WaitState(ctx, "abc", "FINISHED")
	require.NoError(t, err)
	assert.Equal(t, "FINISHED", state)
	assert.Equal(t, 2, calls)
}

func TestConformanceClient_WaitStateReturnsFramedErrorOnPersistentTransientFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/runner/abc/wait-state", r.URL.Path)

		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal error"}`))
	}))

	defer server.Close()

	client, err := NewConformanceClient(server.URL)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*750)
	defer cancel()

	start := time.Now()

	_, err = client.WaitState(ctx, "abc", "FINISHED")

	elapsed := time.Since(start)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "abc")
	assert.Contains(t, err.Error(), "FINISHED")
	assert.Less(t, elapsed, time.Second*5, "WaitState should return promptly once ctx expires rather than hang")
}

func TestConformanceClient_CreateTestReportsAliasCollision(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"error":"alias already in use"}`))
	}))

	defer server.Close()

	client, err := NewConformanceClient(server.URL)
	require.NoError(t, err)

	_, err = client.CreateTest(context.Background(), "plan1", "oidcc-server", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "alias already in use")
}

func TestConformanceClient_CreateTestRejectsAResponseWithoutAnIdentifier(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"name":"oidcc-server"}`))
	}))

	defer server.Close()

	client, err := NewConformanceClient(server.URL)
	require.NoError(t, err)

	id, err := client.CreateTest(context.Background(), "plan1", "oidcc-server", nil)
	require.Error(t, err)
	assert.Empty(t, id)
	assert.Contains(t, err.Error(), "oidcc-server")
	assert.Contains(t, err.Error(), "no test identifier")
}

func TestConformanceClient_WaitStateFailsFastOnAClientError(t *testing.T) {
	for _, status := range []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			var calls atomic.Int32

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				w.WriteHeader(status)
			}))

			defer server.Close()

			client, err := NewConformanceClient(server.URL)
			require.NoError(t, err)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()

			_, err = client.WaitState(ctx, "m1", "FINISHED")
			require.Error(t, err)
			require.NoError(t, ctx.Err(), "a client error is not going to change on retry, so it must not wait out the context")

			var statusErr *conformanceUnexpectedStatusError

			require.ErrorAs(t, err, &statusErr)
			assert.Equal(t, status, statusErr.Status)
			assert.Contains(t, err.Error(), "module 'm1'")
			assert.Equal(t, int32(1), calls.Load())
		})
	}
}

func TestConformanceClient_LogExposesUnfilledPlaceholders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/log/abc", r.URL.Path)

		_, _ = w.Write([]byte(`[{"msg":"a"},{"msg":"upload me","upload":"ph1","result":"REVIEW"}]`))
	}))

	defer server.Close()

	client, err := NewConformanceClient(server.URL)
	require.NoError(t, err)

	entries, err := client.Log(context.Background(), "abc")
	require.NoError(t, err)
	require.Len(t, entries, 2)
	assert.Equal(t, "", entries[0].Upload)
	assert.Equal(t, "ph1", entries[1].Upload)
}

func TestConformanceClient_UploadPlaceholder(t *testing.T) {
	var body []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/log/abc/images/ph1", r.URL.Path)

		body, _ = io.ReadAll(r.Body)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))

	defer server.Close()

	client, err := NewConformanceClient(server.URL)
	require.NoError(t, err)

	require.NoError(t, client.UploadPlaceholder(context.Background(), "abc", "ph1", conformanceTestScreenshot))
	assert.Equal(t, conformanceTestScreenshot, string(body))
}

func TestConformanceClient_CreatePlanRejectsAnEmptyPlanName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("the client must not post a nameless plan, got %s", r.URL.String())
	}))

	defer server.Close()

	client, err := NewConformanceClient(server.URL)
	require.NoError(t, err)

	_, err = client.CreatePlan(context.Background(), "", nil, &conformance.Plan{Alias: "alias"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "alias")
	assert.Contains(t, err.Error(), "plan name is empty")
}

func TestConformanceClient_WaitStateFailsFastOnAnUnrecognizedBody(t *testing.T) {
	testCases := []struct {
		name     string
		body     string
		expected string
	}{
		{"ShouldFailOnAnEmptyJSONObject", `{}`, "neither a state nor a timeout"},
		{"ShouldFailOnAFalseTimeout", `{"timeout":false}`, "neither a state nor a timeout"},
		{"ShouldFailOnASchemaChange", `{"status":"FINISHED"}`, "neither a state nor a timeout"},
		{"ShouldFailOnAMalformedBody", `<html>gateway</html>`, "is not JSON"},
		{"ShouldFailOnAnEmptyBody", ``, "is not JSON"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var calls int

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/runner/abc/wait-state", r.URL.Path)

				calls++

				_, _ = w.Write([]byte(tc.body))
			}))

			defer server.Close()

			client, err := NewConformanceClient(server.URL)
			require.NoError(t, err)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()

			start := time.Now()

			_, err = client.WaitState(ctx, "abc", "FINISHED")

			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expected)
			assert.Contains(t, err.Error(), "abc")
			assert.Equal(t, 1, calls, "an unrecognized body must not be retried")
			assert.Less(t, time.Since(start), time.Second, "an unrecognized body must fail fast rather than spin")
		})
	}
}

func TestNewConformanceClientRejectsAnInvalidURL(t *testing.T) {
	_, err := NewConformanceClient("not a url")

	assert.ErrorContains(t, err, "error parsing conformance base url")
}

func TestConformanceClient_WaitReady(t *testing.T) {
	t.Run("ShouldReturnOnceTheServerAnswers", func(t *testing.T) {
		var calls atomic.Int32

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/plan", r.URL.Path)
			assert.Equal(t, "1", r.URL.Query().Get("length"))

			if calls.Add(1) == 1 {
				w.WriteHeader(http.StatusServiceUnavailable)

				return
			}

			_, _ = w.Write([]byte(`{"data":[]}`))
		}))

		defer server.Close()

		client, err := NewConformanceClient(server.URL)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		require.NoError(t, client.WaitReady(ctx))
		assert.Equal(t, int32(2), calls.Load(), "the server is asked again after it could not answer")
	})

	t.Run("ShouldGiveUpWhenTheContextExpires", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		}))

		defer server.Close()

		client, err := NewConformanceClient(server.URL)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
		defer cancel()

		err = client.WaitReady(ctx)

		assert.ErrorContains(t, err, "conformance suite was not ready")
		assert.ErrorContains(t, err, "expected status 200 but got 503")
	})
}

func TestConformanceClient_Requests(t *testing.T) {
	var (
		mu       sync.Mutex
		requests []string
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()

		requests = append(requests, r.Method+" "+r.URL.RequestURI())
		mu.Unlock()

		switch r.URL.Path {
		case "/api/runner":
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"m1"}`))
		case "/api/runner/m1":
			_, _ = w.Write([]byte(`{}`))
		case "/api/runner/browser/m1/visit":
			w.WriteHeader(http.StatusNoContent)
		case "/api/plan/exporthtml/p1":
			_, _ = w.Write([]byte("PK archive"))
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))

	defer server.Close()

	client, err := NewConformanceClient(server.URL)
	require.NoError(t, err)

	ctx := context.Background()

	id, err := client.CreateTest(ctx, "p1", "oidcc-server", map[string]string{"response_type": "code"})
	require.NoError(t, err)
	assert.Equal(t, "m1", id)

	require.NoError(t, client.StartTest(ctx, "m1"))
	require.NoError(t, client.Visit(ctx, "m1", "https://login.example.com:8080/api/oidc/authorization?client_id=a&state=b"))

	archive := filepath.Join(t.TempDir(), "plans", "conformance-basic.zip")

	require.NoError(t, client.ExportPlanHTML(ctx, "p1", archive))

	data, err := os.ReadFile(archive)
	require.NoError(t, err)
	assert.Equal(t, "PK archive", string(data))

	mu.Lock()
	defer mu.Unlock()

	assert.Equal(t, []string{
		"POST /api/runner?plan=p1&test=oidcc-server&variant=%7B%22response_type%22%3A%22code%22%7D",
		"POST /api/runner/m1",
		"POST /api/runner/browser/m1/visit?url=https%3A%2F%2Flogin.example.com%3A8080%2Fapi%2Foidc%2Fauthorization%3Fclient_id%3Da%26state%3Db",
		"GET /api/plan/exporthtml/p1",
	}, requests)
}

func TestConformanceClient_ReportsServerErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"broken"}`))
	}))

	defer server.Close()

	client, err := NewConformanceClient(server.URL)
	require.NoError(t, err)

	ctx := context.Background()

	_, err = client.CreatePlan(ctx, "oidcc-basic-certification-test-plan", nil, &conformance.Plan{Alias: "alias"})
	assert.ErrorContains(t, err, "broken")

	_, err = client.Info(ctx, "m1")
	assert.ErrorContains(t, err, "broken")

	_, err = client.BrowserStatus(ctx, "m1")
	assert.ErrorContains(t, err, "broken")

	_, err = client.Log(ctx, "m1")
	assert.ErrorContains(t, err, "broken")

	err = client.ExportPlanHTML(ctx, "p1", filepath.Join(t.TempDir(), "conformance-basic.zip"))
	assert.EqualError(t, err, "unexpected status 500 exporting plan 'p1'")
}

func TestConformanceClient_ExportPlanHTMLReportsADirectoryItCannotCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("PK archive"))
	}))

	defer server.Close()

	client, err := NewConformanceClient(server.URL)
	require.NoError(t, err)

	file := filepath.Join(t.TempDir(), "file")
	require.NoError(t, os.WriteFile(file, nil, 0600))

	assert.Error(t, client.ExportPlanHTML(context.Background(), "p1", filepath.Join(file, "conformance-basic.zip")))
}

func TestConformanceClient_LogRejectsAMalformedEntry(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"msg":5}]`))
	}))

	defer server.Close()

	client, err := NewConformanceClient(server.URL)
	require.NoError(t, err)

	_, err = client.Log(context.Background(), "m1")
	assert.Error(t, err)
}

func TestConformanceClient_WaitStateFromInfo(t *testing.T) {
	testCases := []struct {
		name     string
		info     func(w http.ResponseWriter)
		expected string
	}{
		{
			"ShouldReportAPersistedStatusWhichIsNotWaitedFor",
			func(w http.ResponseWriter) { _, _ = w.Write([]byte(`{"_id":"m1","status":"INTERRUPTED"}`)) },
			"module 'm1' is no longer running and its persisted status is 'INTERRUPTED', which is not one of [FINISHED]",
		},
		{
			"ShouldReportThatThePersistedStatusCouldNotBeRead",
			func(w http.ResponseWriter) { w.WriteHeader(http.StatusInternalServerError) },
			"expected status 200 but got 500",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/api/runner/m1/wait-state":
					w.WriteHeader(http.StatusNotFound)
				case "/api/info/m1":
					tc.info(w)
				default:
					t.Errorf("unexpected path %s", r.URL.Path)
				}
			}))

			defer server.Close()

			client, err := NewConformanceClient(server.URL)
			require.NoError(t, err)

			_, err = client.WaitState(context.Background(), "m1", "FINISHED")

			assert.ErrorContains(t, err, tc.expected)
		})
	}
}

func TestConformanceClient_WaitStateStopsWhenTheContextExpiresBetweenPolls(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Millisecond * 20)

		_, _ = w.Write([]byte(`{"timeout":true}`))
	}))

	defer server.Close()

	client, err := NewConformanceClient(server.URL)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	_, err = client.WaitState(ctx, "m1", "FINISHED")

	assert.ErrorContains(t, err, "module 'm1' did not reach one of [FINISHED]")
}

func TestConformanceClient_ExportPlanHTMLReportsWhatFailed(t *testing.T) {
	t.Run("ShouldReportAServerWhichCannotBeReached", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

		client, err := NewConformanceClient(server.URL)
		require.NoError(t, err)

		server.Close()

		assert.Error(t, client.ExportPlanHTML(context.Background(), "p1", filepath.Join(t.TempDir(), "conformance-basic.zip")))
	})

	t.Run("ShouldReportAnArchiveWhichCannotBeCreated", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("PK archive"))
		}))

		defer server.Close()

		client, err := NewConformanceClient(server.URL)
		require.NoError(t, err)

		assert.Error(t, client.ExportPlanHTML(context.Background(), "p1", t.TempDir()), "the path is a directory")
	})
}

func TestConformanceClient_WaitStateFromInfoStopsWhenTheContextExpires(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/runner/m1/wait-state" {
			w.WriteHeader(http.StatusNotFound)

			return
		}

		time.Sleep(time.Millisecond * 200)

		_, _ = w.Write([]byte(`{"_id":"m1","status":"FINISHED"}`))
	}))

	defer server.Close()

	client, err := NewConformanceClient(server.URL)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	_, err = client.WaitState(ctx, "m1", "FINISHED")

	assert.ErrorContains(t, err, "module 'm1' did not reach one of [FINISHED]")
}
