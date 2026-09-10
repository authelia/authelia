// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/storage"
)

type healthTestStorage struct {
	storage.Provider

	err   error
	calls int
}

func (h *healthTestStorage) StartupCheck() (err error) {
	h.calls++

	return h.err
}

func TestHealthVerboseGET(t *testing.T) {
	testCases := []struct {
		name           string
		config         schema.ServerEndpointHealth
		err            error
		expectedStatus int
		expectedState  string
		expectedDetail bool
	}{
		{
			"ShouldReportOKWhenTheProviderPasses",
			schema.ServerEndpointHealth{Providers: []string{schema.ProviderNameStorage}},
			nil,
			fasthttp.StatusOK,
			"ok",
			false,
		},
		{
			"ShouldReportServiceUnavailableWhenTheProviderFails",
			schema.ServerEndpointHealth{Providers: []string{schema.ProviderNameStorage}},
			errors.New("connection refused"),
			fasthttp.StatusServiceUnavailable,
			"error",
			false,
		},
		{
			"ShouldWithholdTheErrorMessageByDefault",
			schema.ServerEndpointHealth{Providers: []string{schema.ProviderNameStorage}},
			errors.New("connection refused to postgres://user@db.internal:5432"),
			fasthttp.StatusServiceUnavailable,
			"error",
			false,
		},
		{
			"ShouldIncludeTheErrorMessageWhenDetailedIsEnabled",
			schema.ServerEndpointHealth{Providers: []string{schema.ProviderNameStorage}, Detailed: true},
			errors.New("connection refused to postgres://user@db.internal:5432"),
			fasthttp.StatusServiceUnavailable,
			"error",
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			mock.Ctx.Providers.StorageProvider = &healthTestStorage{err: tc.err}

			HealthVerboseGET(tc.config)(mock.Ctx)

			assert.Equal(t, tc.expectedStatus, mock.Ctx.Response.StatusCode())

			body := HealthVerboseResponse{}
			require.NoError(t, json.Unmarshal(mock.Ctx.Response.Body(), &body))

			require.Contains(t, body.Providers, schema.ProviderNameStorage)
			assert.Equal(t, tc.expectedState, body.Providers[schema.ProviderNameStorage].Status)
			assert.False(t, body.Cached)

			if tc.expectedDetail {
				assert.Equal(t, tc.err.Error(), body.Providers[schema.ProviderNameStorage].Error)
			} else {
				assert.Empty(t, body.Providers[schema.ProviderNameStorage].Error)
			}
		})
	}
}

func TestHealthVerboseGETShouldReuseTheCachedResult(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	provider := &healthTestStorage{}
	mock.Ctx.Providers.StorageProvider = provider

	handler := HealthVerboseGET(schema.ServerEndpointHealth{
		Providers: []string{schema.ProviderNameStorage},
		Cache:     time.Minute,
	})

	handler(mock.Ctx)
	assert.Equal(t, 1, provider.calls)

	first := HealthVerboseResponse{}
	require.NoError(t, json.Unmarshal(mock.Ctx.Response.Body(), &first))
	assert.False(t, first.Cached)

	mock.Ctx.Response.Reset()

	handler(mock.Ctx)
	assert.Equal(t, 1, provider.calls, "the second request inside the cache duration must not re-probe")

	second := HealthVerboseResponse{}
	require.NoError(t, json.Unmarshal(mock.Ctx.Response.Body(), &second))
	assert.True(t, second.Cached)
	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
}

func TestHealthVerboseGETShouldNotCacheWhenDisabled(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	provider := &healthTestStorage{}
	mock.Ctx.Providers.StorageProvider = provider

	handler := HealthVerboseGET(schema.ServerEndpointHealth{Providers: []string{schema.ProviderNameStorage}})

	handler(mock.Ctx)
	mock.Ctx.Response.Reset()
	handler(mock.Ctx)

	assert.Equal(t, 2, provider.calls)
}
