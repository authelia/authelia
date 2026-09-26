// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"context"
	"errors"
	"math"
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
		Name        string
		Username    string
		Expiration  time.Duration
		Results     []any
		PipelineErr error
		RemoveErr   error
		Error       string
		Assert      func(t *testing.T, client *mockRedisCmdable)
	}{
		{
			"ShouldSaveTheSessionThenIndexIt",
			"john",
			time.Hour,
			nil,
			nil,
			nil,
			"",
			func(t *testing.T, client *mockRedisCmdable) {
				require.Len(t, client.evals, 1)

				assert.Equal(t, mockRedisEval{script: redisSessionSave, keys: []string{keySession}, args: []any{[]byte("data"), int64(3600000), "pid", "john"}}, client.evals[0])
				assert.Equal(t, []mockRedisSet{{key: keyPublic, value: "id", expiration: time.Hour}}, client.sets)

				require.Len(t, client.added[keyUser], 1)
				assert.Equal(t, "id", client.added[keyUser][0].Member)
				assert.InDelta(t, float64(time.Now().Add(time.Hour).Unix()), client.added[keyUser][0].Score, 2)

				assert.Empty(t, client.removed)
			},
		},
		{
			"ShouldOmitTheUsernameIndexForAnAnonymousSession",
			"",
			time.Hour,
			nil,
			nil,
			nil,
			"",
			func(t *testing.T, client *mockRedisCmdable) {
				assert.Equal(t, []any{[]byte("data"), int64(3600000), "pid", ""}, client.evals[0].args)
				assert.Len(t, client.sets, 1)
				assert.Empty(t, client.added)
			},
		},
		{
			"ShouldNotExpireKeysWhenTheExpirationIsNotPositive",
			"john",
			0,
			nil,
			nil,
			nil,
			"",
			func(t *testing.T, client *mockRedisCmdable) {
				assert.Equal(t, int64(0), client.evals[0].args[1])
				assert.Equal(t, time.Duration(0), client.sets[0].expiration)
				assert.True(t, math.IsInf(client.added[keyUser][0].Score, 1))
			},
		},
		{
			"ShouldRetireThePreviousPublicIDAndUsername",
			"john",
			time.Hour,
			[]any{[]any{int64(1), "oldpid", "jane"}},
			nil,
			nil,
			"",
			func(t *testing.T, client *mockRedisCmdable) {
				require.Len(t, client.evals, 2)

				assert.Equal(t, mockRedisEval{script: redisDeleteIfEqual, keys: []string{getSessionPublicKey("example.com", "oldpid")}, args: []any{"id"}}, client.evals[1])
				assert.Equal(t, map[string][]any{getSessionUserKey("example.com", "jane"): {"id"}}, client.removed)
			},
		},
		{
			"ShouldNotRetireAnUnchangedPublicIDOrUsername",
			"john",
			time.Hour,
			[]any{[]any{int64(1), "pid", "john"}},
			nil,
			nil,
			"",
			func(t *testing.T, client *mockRedisCmdable) {
				assert.Len(t, client.evals, 1)
				assert.Empty(t, client.removed)
			},
		},
		{
			"ShouldReturnSupersededWithoutIndexingWhenTheSessionWasMoved",
			"john",
			time.Hour,
			[]any{[]any{int64(0)}},
			nil,
			nil,
			session.ErrSessionSuperseded.Error(),
			func(t *testing.T, client *mockRedisCmdable) {
				assert.Len(t, client.evals, 1)
				assert.Empty(t, client.sets)
				assert.Empty(t, client.added)
			},
		},
		{
			"ShouldReturnErrorOnFailure",
			"john",
			time.Hour,
			[]any{errors.New("connection refused")},
			nil,
			nil,
			"connection refused",
			func(t *testing.T, client *mockRedisCmdable) {
				assert.Empty(t, client.sets)
			},
		},
		{
			"ShouldReturnErrorWhenIndexingFails",
			"john",
			time.Hour,
			nil,
			errors.New("connection refused"),
			nil,
			"error updating the session indexes: connection refused",
			nil,
		},
		{
			"ShouldReturnErrorWhenRetiringThePreviousPublicIDFails",
			"john",
			time.Hour,
			[]any{[]any{int64(1), "oldpid", "john"}, errors.New("connection refused")},
			nil,
			nil,
			"error removing the session public id index: connection refused",
			nil,
		},
		{
			"ShouldReturnErrorWhenRetiringThePreviousUsernameFails",
			"john",
			time.Hour,
			[]any{[]any{int64(1), "pid", "jane"}},
			nil,
			errors.New("connection refused"),
			"error removing the session from the username index: connection refused",
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
				client := &mockRedisCmdable{evalResults: append([]any(nil), tc.Results...), pipelineErr: tc.PipelineErr, removeErr: tc.RemoveErr}
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

				if tc.Assert != nil {
					tc.Assert(t, client)
				}
			})
		}
	}
}

func TestRedis_SessionSaveShouldReturnSupersededWhenTheScriptDiscardsTheSave(t *testing.T) {
	client := &mockRedisCmdable{evalResults: []any{[]any{int64(0)}, []any{int64(0)}}}
	provider := NewRedis(client, "standalone")

	assert.ErrorIs(t, provider.SessionSave(context.Background(), "example.com", "id", "pid", "john", time.Hour, []byte("stale")), session.ErrSessionSuperseded)
	assert.ErrorIs(t, provider.SessionSaveData(context.Background(), "example.com", "id", "pid", "john", time.Hour, []byte("stale")), session.ErrSessionSuperseded)
}

func TestRedis_SessionDelete(t *testing.T) {
	testCases := []struct {
		Name             string
		PublicID         string
		Username         string
		Results          []any
		Error            string
		ExpectedPublicID string
		ExpectedUsername string
	}{
		{"ShouldDeleteSessionAndProvidedLookups", "pid", "john", []any{[]any{"recorded", "jane"}}, "", "pid", "john"},
		{"ShouldRecoverTheLookupsWhenTheyAreNotProvided", "", "", []any{[]any{"recorded", "jane"}}, "", "recorded", "jane"},
		{"ShouldNotRetireLookupsWhichAreNeitherProvidedNorRecorded", "", "", []any{[]any{"", ""}}, "", "", ""},
		{"ShouldReturnErrorOnFailure", "pid", "john", []any{errors.New("connection refused")}, "connection refused", "", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			client := &mockRedisCmdable{evalResults: tc.Results}

			err := NewRedis(client, "standalone").SessionDelete(context.Background(), "example.com", "id", tc.PublicID, tc.Username)

			require.NotEmpty(t, client.evals)
			assert.Equal(t, mockRedisEval{script: redisSessionDelete, keys: []string{getSessionKey("example.com", "id")}}, client.evals[0])

			if tc.Error != "" {
				assert.EqualError(t, err, tc.Error)

				return
			}

			require.NoError(t, err)

			if tc.ExpectedPublicID == "" {
				assert.Len(t, client.evals, 1)
			} else {
				require.Len(t, client.evals, 2)
				assert.Equal(t, mockRedisEval{script: redisDeleteIfEqual, keys: []string{getSessionPublicKey("example.com", tc.ExpectedPublicID)}, args: []any{"id"}}, client.evals[1])
			}

			if tc.ExpectedUsername == "" {
				assert.Empty(t, client.removed)
			} else {
				assert.Equal(t, map[string][]any{getSessionUserKey("example.com", tc.ExpectedUsername): {"id"}}, client.removed)
			}
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
