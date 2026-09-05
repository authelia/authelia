package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/mocks"
)

func TestHandle1FAResponse(t *testing.T) {
	testCases := []struct {
		name      string
		targetURI string
		method    string
		expected  string
		expectedf func(t *testing.T, mock *mocks.MockAutheliaCtx)
	}{
		{
			"ShouldRedirectWithoutRequestMethod",
			"https://app.example.com/",
			"",
			`{"status":"OK","data":{"redirect":"https://app.example.com/"}}`,
			nil,
		},
		{
			"ShouldRedirectWithUppercaseRequestMethod",
			"https://app.example.com/",
			fasthttp.MethodGet,
			`{"status":"OK","data":{"redirect":"https://app.example.com/"}}`,
			nil,
		},
		{
			"ShouldRedirectWithLowercaseRequestMethod",
			"https://app.example.com/",
			"get",
			`{"status":"OK","data":{"redirect":"https://app.example.com/"}}`,
			nil,
		},
		{
			"ShouldPreserveFragment",
			"https://app.example.com/#/dashboard",
			fasthttp.MethodGet,
			`{"status":"OK","data":{"redirect":"https://app.example.com/#/dashboard"}}`,
			nil,
		},
		{
			"ShouldPreserveQuery",
			"https://app.example.com/?a=1&b=2",
			fasthttp.MethodGet,
			`{"status":"OK","data":{"redirect":"https://app.example.com/?a=1\u0026b=2"}}`,
			nil,
		},
		{
			"ShouldNotRedirectToUnsafeTargetURI",
			"https://evil.example.org/",
			fasthttp.MethodGet,
			`{"status":"OK"}`,
			nil,
		},
		{
			"ShouldNotRedirectToTargetURIRequiringTwoFactor",
			"https://two-factor.example.com/",
			fasthttp.MethodGet,
			`{"status":"OK"}`,
			nil,
		},
		{
			"ShouldNotRedirectWithoutTargetURI",
			"",
			fasthttp.MethodGet,
			`{"status":"OK"}`,
			nil,
		},
		{
			"ShouldHandleInvalidRequestMethod",
			"https://app.example.com/",
			"GET1",
			`{"status":"KO","message":"Authentication failed. Check your credentials."}`,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred parsing the target URL 'https://app.example.com/'", "method header with value 'GET1' has invalid characters")
			},
		},
		{
			"ShouldHandleInvalidTargetURI",
			"notaurl",
			fasthttp.MethodGet,
			`{"status":"KO","message":"Authentication failed. Check your credentials."}`,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred parsing the target URL 'notaurl'", "error occurred parsing object url: parse \"notaurl\": invalid URI for request")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			Handle1FAResponse(mock.Ctx, tc.targetURI, tc.method, testUsername, nil)

			assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
			assert.Equal(t, tc.expected, string(mock.Ctx.Response.Body()))

			if tc.expectedf != nil {
				tc.expectedf(t, mock)
			}
		})
	}
}
