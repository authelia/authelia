// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package configuration

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestShouldEnableAllServerEndpointRateLimitsByDefault(t *testing.T) {
	rt := reflect.TypeFor[schema.ServerEndpointRateLimits]()

	for i := range rt.NumField() {
		name := rt.Field(i).Tag.Get("koanf")

		t.Run(name, func(t *testing.T) {
			assert.Equal(t, true, defaults["server.endpoints.rate_limits."+name+".enable"])
		})
	}
}
