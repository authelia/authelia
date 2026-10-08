// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/logging"
)

func TestAllowedHosts(t *testing.T) {
	testCases := []struct {
		name         string
		have         string
		allowed      []string
		expected     int
		expectedBody string
	}{
		{"ShouldRespondOKWhenNotConfigured", "authelia", nil, fasthttp.StatusOK, "next"},
		{"ShouldRespondOKWhenHostMatches", "authelia", []string{"authelia"}, fasthttp.StatusOK, "next"},
		{"ShouldRespondNotFoundWhenHostNotMatch", "auth.phishingdomain.com", []string{"authelia"}, fasthttp.StatusNotFound, "404 Not Found"},
		{"ShouldRespondOKWhenHostMatchesMulti", "authelia", []string{"127.0.0.1", "authelia"}, fasthttp.StatusOK, "next"},
		{"ShouldRespondOKWhenHostMatchesDuplicates", "authelia", []string{"authelia", "127.0.0.1", "authelia"}, fasthttp.StatusOK, "next"},
		{"ShouldRespondNotFoundWhenHostMatchesWithoutPort", "authelia:9091", []string{"authelia"}, fasthttp.StatusNotFound, "404 Not Found"},
		{"ShouldRespondOKWhenHostMatchesWithPort", "authelia:9091", []string{"authelia:9091"}, fasthttp.StatusOK, "next"},
		{"ShouldRespondOKWhenConfiguredHostUppercase", "authelia:9091", []string{"AUTHELIA:9091"}, fasthttp.StatusOK, "next"},
		{"ShouldRespondOKWhenRequestHostUppercase", "Auth.Example.com", []string{"auth.example.com"}, fasthttp.StatusOK, "next"},
		{"ShouldRespondOKWhenDuplicatesDifferByCase", "authelia", []string{"Authelia", "authelia"}, fasthttp.StatusOK, "next"},
		{"ShouldRespondNotFoundWhenPortDiffers", "AUTHELIA:9092", []string{"Authelia:9091"}, fasthttp.StatusNotFound, "404 Not Found"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			next := func(ctx *fasthttp.RequestCtx) {
				ctx.SetBodyString("next")
			}

			handler := Wrap(AllowedHosts(tc.allowed), next)

			ctx := &fasthttp.RequestCtx{}

			ctx.Request.SetHost(tc.have)

			handler(ctx)

			assert.Equal(t, tc.expected, ctx.Response.StatusCode())
			assert.Equal(t, tc.expectedBody, string(ctx.Response.Body()))
		})
	}
}

func TestAllowedHostsShouldLogRejectedRequests(t *testing.T) {
	logger := logging.Logger()

	level := logger.GetLevel()

	logger.SetLevel(logrus.DebugLevel)

	hook := test.NewLocal(logger)

	t.Cleanup(func() {
		logger.SetLevel(level)
		logger.ReplaceHooks(logrus.LevelHooks{})
	})

	handler := Wrap(AllowedHosts([]string{"authelia"}), func(ctx *fasthttp.RequestCtx) {})

	ctx := &fasthttp.RequestCtx{}

	ctx.Request.SetRequestURI("/api/health")
	ctx.Request.SetHost("authelia")

	handler(ctx)

	assert.Len(t, hook.AllEntries(), 0)

	ctx = &fasthttp.RequestCtx{}

	ctx.Request.SetRequestURI("/api/health")
	ctx.Request.SetHost("auth.phishingdomain.com")

	handler(ctx)

	entries := hook.AllEntries()

	require.Len(t, entries, 1)

	assert.Equal(t, logrus.DebugLevel, entries[0].Level)
	assert.Equal(t, "Request rejected as the host is not one of the allowed hosts", entries[0].Message)
	assert.Equal(t, "auth.phishingdomain.com", entries[0].Data[logging.FieldHost])
	assert.Equal(t, "/api/health", entries[0].Data[logging.FieldPath])
}
