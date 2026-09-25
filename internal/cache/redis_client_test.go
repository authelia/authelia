// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"context"
	"errors"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/session"
)

func TestNewRedisStandalone(t *testing.T) {
	provider := NewRedisStandalone(&schema.RedisCache{
		Address:        mustCacheAddressTCP(t, "tcp://127.0.0.1:6379"),
		Username:       "user",
		Password:       "pass",
		Database:       2,
		PoolSize:       7,
		FailingTimeout: time.Second * 15,
		IdleTimeout:    time.Minute * 5,
	}, nil)

	require.NotNil(t, provider)
	assert.Equal(t, "standalone", provider.variant)

	client, ok := provider.client.(*redis.Client)
	require.True(t, ok)

	defer client.Close()

	options := client.Options()

	assert.Equal(t, "tcp", options.Network)
	assert.Equal(t, "127.0.0.1:6379", options.Addr)
	assert.Equal(t, getClientName(), options.ClientName)
	assert.Equal(t, 3, options.Protocol)
	assert.Equal(t, "user", options.Username)
	assert.Equal(t, "pass", options.Password)
	assert.Equal(t, 2, options.DB)
	assert.Equal(t, 7, options.PoolSize)
	assert.Equal(t, 15, options.FailingTimeoutSeconds)
	assert.Equal(t, time.Minute*5, options.ConnMaxIdleTime)
}

func TestNewRedisSentinel(t *testing.T) {
	testCases := []struct {
		Name   string
		Mode   string
		Assert func(t *testing.T, client redis.Cmdable)
	}{
		{
			"ShouldUseFailoverClientInFailoverMode",
			"failover",
			func(t *testing.T, client redis.Cmdable) {
				c, ok := client.(*redis.Client)
				require.True(t, ok)

				defer c.Close()

				assert.Equal(t, 2, c.Options().DB)
				assert.Equal(t, 3, c.Options().Protocol)
				assert.Equal(t, time.Minute*5, c.Options().ConnMaxIdleTime)
			},
		},
		{
			"ShouldUseFailoverClientWhenModeIsUnset",
			"",
			func(t *testing.T, client redis.Cmdable) {
				c, ok := client.(*redis.Client)
				require.True(t, ok)

				require.NoError(t, c.Close())
			},
		},
		{
			"ShouldUseFailoverClusterClientInClusterMode",
			"cluster",
			func(t *testing.T, client redis.Cmdable) {
				c, ok := client.(*redis.ClusterClient)
				require.True(t, ok)

				defer c.Close()

				assert.True(t, c.Options().RouteByLatency)
				assert.True(t, c.Options().RouteRandomly)
				assert.Equal(t, 3, c.Options().Protocol)
				assert.Equal(t, time.Minute*5, c.Options().ConnMaxIdleTime)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			config := &schema.RedisSentinelCache{
				MasterName:   "mymaster",
				SentinelMode: tc.Mode,
				Addresses: []*schema.AddressTCP{
					mustCacheAddressTCP(t, "tcp://127.0.0.1:26379"),
					mustCacheAddressTCP(t, "tcp://127.0.0.2:26379"),
				},
				Database:    2,
				IdleTimeout: time.Minute * 5,
			}

			if tc.Mode == "cluster" {
				config.RouteByLatency = true
				config.RouteRandomly = true
			}

			provider := NewRedisSentinel(config, nil)

			require.NotNil(t, provider)
			assert.Equal(t, "sentinel", provider.variant)

			tc.Assert(t, provider.client)
		})
	}
}

func TestNewRedisCluster(t *testing.T) {
	provider := NewRedisCluster(&schema.RedisClusterCache{
		Addresses: []*schema.AddressTCP{
			mustCacheAddressTCP(t, "tcp://127.0.0.1:7000"),
			mustCacheAddressTCP(t, "tcp://127.0.0.2:7001"),
		},
		Username:                   "user",
		Password:                   "pass",
		MaximumRedirects:           5,
		RouteByReplica:             true,
		PoolSize:                   9,
		FailingTimeout:             time.Second * 20,
		ClusterStateReloadInterval: time.Second * 30,
		IdleTimeout:                time.Minute * 5,
	}, nil)

	require.NotNil(t, provider)
	assert.Equal(t, "cluster", provider.variant)

	client, ok := provider.client.(*redis.ClusterClient)
	require.True(t, ok)

	defer client.Close()

	options := client.Options()

	assert.Equal(t, []string{"127.0.0.1:7000", "127.0.0.2:7001"}, options.Addrs)
	assert.Equal(t, getClientName(), options.ClientName)
	assert.Equal(t, 3, options.Protocol)
	assert.Equal(t, "user", options.Username)
	assert.Equal(t, "pass", options.Password)
	assert.Equal(t, 5, options.MaxRedirects)
	assert.True(t, options.ReadOnly)
	assert.Equal(t, 9, options.PoolSize)
	assert.Equal(t, 20, options.FailingTimeoutSeconds)
	assert.Equal(t, time.Second*30, options.ClusterStateReloadInterval)
	assert.Equal(t, time.Minute*5, options.ConnMaxIdleTime)
}

func TestRedis_StartupCheck(t *testing.T) {
	testCases := []struct {
		Name  string
		Err   error
		Error string
	}{
		{"ShouldSucceedWhenReachable", nil, ""},
		{"ShouldReturnErrorWhenUnreachable", errors.New("connection refused"), "connection refused"},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			err := NewRedis(&mockRedisCmdable{err: tc.Err}, "standalone").StartupCheck()

			if tc.Error == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.Error)
			}
		})
	}
}

func TestRedis_SessionSave(t *testing.T) {
	keySession, keyPublic, keyUser := getSessionKey("example.com", "id"), getSessionPublicKey("example.com", "pid"), getSessionUserKey("example.com", "john")

	testCases := []struct {
		Name         string
		Username     string
		Expiration   time.Duration
		Err          error
		Error        string
		ExpectedKeys []string
		Assert       func(t *testing.T, args []any)
	}{
		{
			"ShouldSaveSessionAndLookupsInASingleScript",
			"john",
			time.Hour,
			nil,
			"",
			[]string{keySession, keyPublic, keyUser},
			func(t *testing.T, args []any) {
				require.Len(t, args, 8)
				assert.Equal(t, []byte("data"), args[0])
				assert.Equal(t, int64(3600000), args[1])
				assert.Equal(t, "id", args[2])
				assert.Equal(t, "pid", args[3])
				assert.Equal(t, "john", args[4])
				assert.InDelta(t, float64(time.Now().Add(time.Hour).Unix()), args[5], 2)
				assert.Equal(t, getSessionPublicKey("example.com", ""), args[6])
				assert.Equal(t, getSessionUserKey("example.com", ""), args[7])
			},
		},
		{
			"ShouldOmitTheUsernameIndexForAnAnonymousSession",
			"",
			time.Hour,
			nil,
			"",
			[]string{keySession, keyPublic},
			func(t *testing.T, args []any) {
				require.Len(t, args, 8)
				assert.Equal(t, []byte("data"), args[0])
				assert.Equal(t, int64(3600000), args[1])
				assert.Equal(t, "id", args[2])
				assert.Equal(t, "pid", args[3])
				assert.Equal(t, "", args[4])
			},
		},
		{
			"ShouldNotExpireKeysWhenTheExpirationIsNotPositive",
			"john",
			0,
			nil,
			"",
			[]string{keySession, keyPublic, keyUser},
			func(t *testing.T, args []any) {
				require.Len(t, args, 8)
				assert.Equal(t, int64(0), args[1])
				assert.True(t, math.IsInf(args[5].(float64), 1))
			},
		},
		{
			"ShouldReturnErrorOnFailure",
			"john",
			time.Hour,
			errors.New("connection refused"),
			"connection refused",
			nil,
			nil,
		},
	}

	for _, tc := range testCases {
		for _, data := range []bool{false, true} {
			name := tc.Name
			if data {
				name += "WithSaveData"
			}

			t.Run(name, func(t *testing.T) {
				client := &mockRedisCmdable{err: tc.Err}
				provider := NewRedis(client, "standalone")

				var err error

				if data {
					err = provider.SessionSaveData(context.Background(), "example.com", "id", "pid", tc.Username, tc.Expiration, []byte("data"))
				} else {
					err = provider.SessionSave(context.Background(), "example.com", "id", "pid", tc.Username, tc.Expiration, []byte("data"))
				}

				if tc.Error == "" {
					require.NoError(t, err)
				} else {
					assert.EqualError(t, err, tc.Error)
				}

				if tc.ExpectedKeys != nil {
					assert.Equal(t, tc.ExpectedKeys, client.evalKeys)
				}

				if tc.Assert != nil {
					tc.Assert(t, client.evalArgs)
				}
			})
		}
	}
}

func TestRedis_SessionSaveShouldReturnSupersededWhenTheScriptDiscardsTheSave(t *testing.T) {
	client := &mockRedisCmdable{evalVal: int64(0)}
	provider := NewRedis(client, "standalone")

	assert.ErrorIs(t, provider.SessionSave(context.Background(), "example.com", "id", "pid", "john", time.Hour, []byte("stale")), session.ErrSessionSuperseded)
	assert.ErrorIs(t, provider.SessionSaveData(context.Background(), "example.com", "id", "pid", "john", time.Hour, []byte("stale")), session.ErrSessionSuperseded)
}

func TestRedis_SessionSaveKeysShareAClusterSlot(t *testing.T) {
	client := &mockRedisCmdable{}

	require.NoError(t, NewRedis(client, "standalone").SessionSave(context.Background(), "example.com", "id", "pid", "john", time.Hour, []byte("data")))
	require.Len(t, client.evalKeys, 3)

	for _, key := range client.evalKeys {
		assert.Equal(t, "example.com", key[strings.Index(key, "{")+1:strings.Index(key, "}")])
	}
}

func TestRedis_SessionDelete(t *testing.T) {
	testCases := []struct {
		Name     string
		PublicID string
		Username string
		Err      error
		Error    string
	}{
		{"ShouldDeleteSessionAndProvidedLookups", "pid", "john", nil, ""},
		{"ShouldRecoverTheLookupsWhenTheyAreNotProvided", "", "", nil, ""},
		{"ShouldReturnErrorOnFailure", "pid", "john", errors.New("connection refused"), "connection refused"},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			client := &mockRedisCmdable{err: tc.Err}

			err := NewRedis(client, "standalone").SessionDelete(context.Background(), "example.com", "id", tc.PublicID, tc.Username)

			if tc.Error != "" {
				assert.EqualError(t, err, tc.Error)

				return
			}

			require.NoError(t, err)

			assert.Equal(t, []string{getSessionKey("example.com", "id")}, client.evalKeys)
			assert.Equal(t, []any{"id", tc.PublicID, tc.Username, getSessionPublicKey("example.com", ""), getSessionUserKey("example.com", "")}, client.evalArgs)
		})
	}
}

func TestRedis_SessionGetIDsByUsernameShouldReturnErrorOnExec(t *testing.T) {
	client := &mockRedisCmdable{pipeliner: &mockRedisPipeliner{pruned: map[string]string{}, err: errors.New("exec failed")}}

	ids, err := NewRedis(client, "standalone").SessionGetIDsByUsername(context.Background(), "example.com", "john")

	assert.EqualError(t, err, "exec failed")
	assert.Nil(t, ids)
}

func TestRedis_SessionGarbageCollectionShouldReturnErrorOnPrune(t *testing.T) {
	key := getSessionUserKey("example.com", "john")

	client := &mockRedisCmdable{
		scanned:  []string{key, getSessionUserKey("example.com", "jane")},
		pruned:   map[string]string{},
		pruneErr: errors.New("prune failed"),
	}

	err := NewRedis(client, "standalone").SessionGarbageCollection(context.Background())

	assert.EqualError(t, err, "error removing expired sessions from the user index '"+key+"': prune failed")
	assert.Len(t, client.pruned, 1)
}

func TestRedis_SessionGarbageCollectionShouldVisitClusterMasters(t *testing.T) {
	provider := NewRedisCluster(&schema.RedisClusterCache{
		Addresses:      []*schema.AddressTCP{mustCacheAddressTCP(t, "tcp://127.0.0.1:1")},
		DialTimeout:    time.Millisecond * 100,
		MaximumRetries: -1,
	}, nil)

	client, ok := provider.client.(*redis.ClusterClient)
	require.True(t, ok)

	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	assert.Error(t, provider.SessionGarbageCollection(ctx))
}

func (m *mockRedisCmdable) Ping(ctx context.Context) *redis.StatusCmd {
	cmd := redis.NewStatusCmd(ctx, "ping")

	if m.err != nil {
		cmd.SetErr(m.err)
	}

	return cmd
}

func mustCacheAddressTCP(t *testing.T, value string) *schema.AddressTCP {
	t.Helper()

	address, err := schema.NewAddressFromNetworkValuesDefault(value, 0, schema.AddressSchemeTCP, schema.AddressSchemeUnix)
	require.NoError(t, err)

	return &schema.AddressTCP{Address: *address}
}
