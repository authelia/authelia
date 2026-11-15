// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/cache"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestNewProvidersShouldConfigureTheCache(t *testing.T) {
	address := func(value string) *schema.AddressTCP {
		a, err := schema.NewAddressFromNetworkValuesDefault(value, 0, schema.AddressSchemeTCP, schema.AddressSchemeUnix)
		require.NoError(t, err)

		return &schema.AddressTCP{Address: *a}
	}

	testCases := []struct {
		name     string
		have     schema.Cache
		expected string
		variant  string
	}{
		{
			"ShouldUseMemoryByDefault",
			schema.Cache{},
			"*cache.Memory",
			"",
		},
		{
			"ShouldUseRedis",
			schema.Cache{Redis: &schema.RedisCache{Address: address("tcp://redis.example.com:6379")}},
			"*cache.Redis",
			"standalone",
		},
		{
			"ShouldUseRedisSentinel",
			schema.Cache{RedisSentinel: &schema.RedisSentinelCache{MasterName: "authelia", Addresses: []*schema.AddressTCP{address("tcp://sentinel.example.com:26379")}}},
			"*cache.Redis",
			"sentinel",
		},
		{
			"ShouldUseRedisCluster",
			schema.Cache{RedisCluster: &schema.RedisClusterCache{Addresses: []*schema.AddressTCP{address("tcp://node.example.com:6379")}}},
			"*cache.Redis",
			"cluster",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			providers, _, errs := NewProviders(&schema.Configuration{Cache: tc.have}, nil)

			assert.Len(t, errs, 0)
			assert.Equal(t, tc.expected, fmt.Sprintf("%T", providers.Cache))

			if tc.variant == "" {
				return
			}

			provider, ok := providers.Cache.(*cache.Redis)
			require.True(t, ok)

			assert.Equal(t, tc.variant, provider.Variant())
		})
	}
}
