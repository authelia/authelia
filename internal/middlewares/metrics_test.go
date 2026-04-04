package middlewares

import (
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

func TestNewMetricsRequest(t *testing.T) {
	testCases := []struct {
		name       string
		method     string
		statusCode int
	}{
		{"ShouldRecordGET200", fasthttp.MethodGet, fasthttp.StatusOK},
		{"ShouldRecordPOST401", fasthttp.MethodPost, fasthttp.StatusUnauthorized},
		{"ShouldRecordDELETE404", fasthttp.MethodDelete, fasthttp.StatusNotFound},
		{"ShouldRecordPUT500", fasthttp.MethodPut, fasthttp.StatusInternalServerError},
	}

	t.Run("ShouldReturnNilWhenRecorderNil", func(t *testing.T) {
		assert.Nil(t, NewMetricsRequest(nil))
	})

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := &mockMetricsRecorder{}

			middleware := NewMetricsRequest(recorder)

			require.NotNil(t, middleware)

			nextCalled := false

			handler := middleware(func(ctx *fasthttp.RequestCtx) {
				nextCalled = true

				ctx.Response.SetStatusCode(tc.statusCode)
			})

			ctx := newRequestCtx(tc.method, 0)
			handler(ctx)

			assert.True(t, nextCalled)
			require.Len(t, recorder.requestCalls, 1)
			assert.Equal(t, strconv.Itoa(tc.statusCode), recorder.requestCalls[0].statusCode)
			assert.Equal(t, tc.method, recorder.requestCalls[0].requestMethod)
			assert.True(t, recorder.requestCalls[0].elapsed >= 0)
		})
	}
}

func TestNewMetricsRequestOpenIDConnect(t *testing.T) {
	testCases := []struct {
		name             string
		endpoint         string
		statusCode       int
		expectedEndpoint string
	}{
		{"ShouldRecordTokenEndpoint", "token", fasthttp.StatusOK, "token"},
		{"ShouldRecordWithHyphenReplacement", "pushed-authorization-request", fasthttp.StatusOK, "pushed_authorization_request"},
		{"ShouldRecordUserinfoEndpoint", "userinfo", fasthttp.StatusForbidden, "userinfo"},
	}

	t.Run("ShouldReturnNilWhenRecorderNil", func(t *testing.T) {
		assert.Nil(t, NewMetricsRequestOpenIDConnect(nil, "token"))
	})

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := &mockMetricsRecorder{}

			middleware := NewMetricsRequestOpenIDConnect(recorder, tc.endpoint)

			require.NotNil(t, middleware)

			nextCalled := false

			handler := middleware(func(ctx *fasthttp.RequestCtx) {
				nextCalled = true

				ctx.Response.SetStatusCode(tc.statusCode)
			})

			ctx := newRequestCtx(fasthttp.MethodPost, 0)
			handler(ctx)

			assert.True(t, nextCalled)
			require.Len(t, recorder.oidcCalls, 1)
			assert.Equal(t, tc.expectedEndpoint, recorder.oidcCalls[0].endpoint)
			assert.True(t, recorder.oidcCalls[0].elapsed >= 0)
		})
	}
}

func TestNewMetricsAuthzRequest(t *testing.T) {
	testCases := []struct {
		name       string
		statusCode int
	}{
		{"ShouldRecordAuthz200", fasthttp.StatusOK},
		{"ShouldRecordAuthz401", fasthttp.StatusUnauthorized},
		{"ShouldRecordAuthz403", fasthttp.StatusForbidden},
	}

	t.Run("ShouldReturnNilWhenRecorderNil", func(t *testing.T) {
		assert.Nil(t, NewMetricsAuthzRequest(nil))
	})

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := &mockMetricsRecorder{}

			middleware := NewMetricsAuthzRequest(recorder)

			require.NotNil(t, middleware)

			nextCalled := false

			handler := middleware(func(ctx *fasthttp.RequestCtx) {
				nextCalled = true

				ctx.Response.SetStatusCode(tc.statusCode)
			})

			ctx := newRequestCtx(fasthttp.MethodGet, 0)
			handler(ctx)

			assert.True(t, nextCalled)
			require.Len(t, recorder.authzCalls, 1)
			assert.Equal(t, strconv.Itoa(tc.statusCode), recorder.authzCalls[0].statusCode)
		})
	}
}

type mockMetricsRecorder struct {
	requestCalls []mockMetricsRequestCall
	oidcCalls    []mockMetricsOIDCCall
	authzCalls   []mockMetricsAuthzCall
	authDurCalls []mockMetricsAuthDurCall
}

type mockMetricsRequestCall struct {
	statusCode    string
	requestMethod string
	elapsed       time.Duration
}

type mockMetricsOIDCCall struct {
	endpoint   string
	statusCode string
	elapsed    time.Duration
}

type mockMetricsAuthzCall struct {
	statusCode string
}

type mockMetricsAuthDurCall struct {
	success bool
	elapsed time.Duration
}

func (m *mockMetricsRecorder) RecordRequest(statusCode, requestMethod string, elapsed time.Duration) {
	m.requestCalls = append(m.requestCalls, mockMetricsRequestCall{statusCode, requestMethod, elapsed})
}

func (m *mockMetricsRecorder) RecordRequestOpenIDConnect(endpoint, statusCode string, elapsed time.Duration) {
	m.oidcCalls = append(m.oidcCalls, mockMetricsOIDCCall{endpoint, statusCode, elapsed})
}

func (m *mockMetricsRecorder) RecordAuthz(statusCode string) {
	m.authzCalls = append(m.authzCalls, mockMetricsAuthzCall{statusCode})
}

func (m *mockMetricsRecorder) RecordAuthenticationDuration(success bool, elapsed time.Duration) {
	m.authDurCalls = append(m.authDurCalls, mockMetricsAuthDurCall{success, elapsed})
}

func newRequestCtx(method string, statusCode int) *fasthttp.RequestCtx {
	var (
		ctx fasthttp.RequestCtx
		req fasthttp.Request
	)

	req.Header.SetMethod(method)

	ctx.Init(&req, &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}, nil)

	ctx.Response.SetStatusCode(statusCode)

	return &ctx
}
