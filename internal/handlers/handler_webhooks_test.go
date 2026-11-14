// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/events"
	"github.com/authelia/authelia/v4/internal/mocks"
)

func TestWebhookConfirm(t *testing.T) {
	testCases := []struct {
		name     string
		emitter  events.Emitter
		uri      string
		rate     string
		expected int
		have     confirmation
	}{
		{
			"ShouldConfirmWithTheKeyAndRate",
			&confirmer{name: "admin-api", key: "abc"},
			"/api/webhooks/confirm?id=admin-api&key=abc",
			"120",
			fasthttp.StatusOK,
			confirmation{"admin-api", "abc", "120"},
		},
		{
			"ShouldRejectAnIncorrectKey",
			&confirmer{name: "admin-api", key: "abc"},
			"/api/webhooks/confirm?id=admin-api&key=incorrect",
			"",
			fasthttp.StatusNotFound,
			confirmation{"admin-api", "incorrect", ""},
		},
		{
			"ShouldRejectWhenTheEmitterCanNotConfirm",
			events.NewNoOpEmitter(),
			"/api/webhooks/confirm?id=admin-api&key=abc",
			"",
			fasthttp.StatusNotFound,
			confirmation{},
		},
	}

	for _, method := range []string{fasthttp.MethodGet, fasthttp.MethodPost} {
		for _, tc := range testCases {
			t.Run(method+"/"+tc.name, func(t *testing.T) {
				mock := mocks.NewMockAutheliaCtx(t)

				defer mock.Close()

				if c, ok := tc.emitter.(*confirmer); ok {
					c.have = confirmation{}
				}

				mock.Ctx.Providers.Events = tc.emitter
				mock.Ctx.Request.Header.SetMethod(method)
				mock.Ctx.Request.SetRequestURI(tc.uri)

				if tc.rate != "" {
					mock.Ctx.Request.Header.Set(events.HeaderWebhookAllowedRate, tc.rate)
				}

				WebhookConfirm(mock.Ctx)

				assert.Equal(t, tc.expected, mock.Ctx.Response.StatusCode())

				if c, ok := tc.emitter.(*confirmer); ok {
					assert.Equal(t, tc.have, c.have)
				}
			})
		}
	}
}

type confirmer struct {
	name, key string
	have      confirmation
}

func (c *confirmer) Emit(_ context.Context, _ *events.Event) {}

func (c *confirmer) Confirm(name, key, rate string) bool {
	c.have = confirmation{name, key, rate}

	return name == c.name && key == c.key
}
