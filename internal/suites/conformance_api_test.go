// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/oidc/conformance"
)

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
		require.Equal(t, "/api/runner/abc/wait-state", r.URL.Path)

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
		require.Equal(t, "/api/runner/abc/wait-state", r.URL.Path)

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

func TestConformanceClient_LogExposesUnfilledPlaceholders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/log/abc", r.URL.Path)

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
		require.Equal(t, "/api/log/abc/images/ph1", r.URL.Path)

		body, _ = io.ReadAll(r.Body)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))

	defer server.Close()

	client, err := NewConformanceClient(server.URL)
	require.NoError(t, err)

	require.NoError(t, client.UploadPlaceholder(context.Background(), "abc", "ph1", conformancePlaceholderImage))
	assert.Equal(t, conformancePlaceholderImage, string(body))
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
				require.Equal(t, "/api/runner/abc/wait-state", r.URL.Path)

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
