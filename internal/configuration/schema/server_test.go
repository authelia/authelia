// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerEndpointRateLimitsEndpoints(t *testing.T) {
	limits := &ServerEndpointRateLimits{}

	endpoints := limits.Endpoints()

	rv := reflect.ValueOf(limits).Elem()
	rt := rv.Type()

	require.Len(t, endpoints, rt.NumField())

	for i := range rt.NumField() {
		field := rt.Field(i)

		t.Run(field.Name, func(t *testing.T) {
			assert.Equal(t, field.Tag.Get("koanf"), endpoints[i].Name)
			assert.Same(t, rv.Field(i).Addr().Interface().(*ServerEndpointRateLimit), endpoints[i].Config)
		})
	}
}

func TestServerEndpointRateLimitsEndpointsWeight(t *testing.T) {
	config := &Configuration{
		IdentityValidation: IdentityValidation{
			ElevatedSession: IdentityValidationElevatedSession{
				CodeLifespan:      time.Minute,
				ElevationLifespan: time.Hour,
			},
		},
	}

	testCases := []struct {
		name     string
		expected time.Duration
	}{
		{"session_elevation_start", time.Minute},
		{"session_elevation_finish", time.Hour},
	}

	limits := &ServerEndpointRateLimits{}

	weighted := map[string]ServerEndpointRateLimitEndpoint{}

	for _, endpoint := range limits.Endpoints() {
		if endpoint.Weight != nil {
			weighted[endpoint.Name] = endpoint
		}
	}

	require.Len(t, weighted, len(testCases))

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			endpoint, ok := weighted[tc.name]

			require.True(t, ok)
			assert.Equal(t, tc.expected, endpoint.Weight(config))
		})
	}
}
