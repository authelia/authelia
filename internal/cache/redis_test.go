package cache

import (
	"context"
	"errors"
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedis_SessionGet(t *testing.T) {
	testCases := []struct {
		Name     string
		Values   map[string]string
		Err      error
		Expected []byte
		Error    string
	}{
		{"ShouldReturnDataWhenPresent", map[string]string{getSessionKey("example.com", "id"): "data"}, nil, []byte("data"), ""},
		{"ShouldReturnNoDataWhenMissing", nil, nil, nil, ""},
		{"ShouldReturnErrorOnFailure", nil, errors.New("connection refused"), nil, "connection refused"},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			provider := NewRedis(&mockRedisCmdable{values: tc.Values, err: tc.Err}, "standalone")

			data, err := provider.SessionGet(context.Background(), "example.com", "id")

			if tc.Error == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.Expected, data)
			} else {
				assert.EqualError(t, err, tc.Error)
				assert.Nil(t, data)
			}
		})
	}
}

func TestRedis_SessionGetByPublicID(t *testing.T) {
	testCases := []struct {
		Name     string
		Values   map[string]string
		Err      error
		Expected []byte
		Error    string
	}{
		{
			"ShouldResolveThroughToTheSession",
			map[string]string{
				getSessionPublicKey("example.com", "pid"): "id",
				getSessionKey("example.com", "id"):        "data",
			},
			nil, []byte("data"), "",
		},
		{"ShouldReturnNoDataWhenPublicIDMissing", nil, nil, nil, ""},
		{
			"ShouldReturnNoDataWhenSessionMissing",
			map[string]string{getSessionPublicKey("example.com", "pid"): "id"},
			nil, nil, "",
		},
		{"ShouldReturnErrorOnFailure", nil, errors.New("connection refused"), nil, "connection refused"},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			provider := NewRedis(&mockRedisCmdable{values: tc.Values, err: tc.Err}, "standalone")

			data, err := provider.SessionGetByPublicID(context.Background(), "example.com", "pid")

			if tc.Error == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.Expected, data)
			} else {
				assert.EqualError(t, err, tc.Error)
				assert.Nil(t, data)
			}
		})
	}
}

func TestRedis_SessionKeys(t *testing.T) {
	testCases := []struct {
		Name     string
		Actual   string
		Expected string
	}{
		{"ShouldBuildSessionKey", getSessionKey("example.com", "id"), "authelia:session:{example.com}:id"},
		{"ShouldBuildPublicKey", getSessionPublicKey("example.com", "pid"), "authelia:session-public:{example.com}:pid"},
		{"ShouldBuildUserKey", getSessionUserKey("example.com", "john"), "authelia:session-user:{example.com}:john"},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			assert.Equal(t, tc.Expected, tc.Actual)
		})
	}
}

func TestRedis_SessionScore(t *testing.T) {
	testCases := []struct {
		Name       string
		Expiration time.Duration
		Never      bool
	}{
		{"ShouldScoreByExpiryWhenPositive", time.Hour, false},
		{"ShouldNeverExpireWhenZero", 0, true},
		{"ShouldNeverExpireWhenNegative", -time.Hour, true},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			score := getSessionScore(tc.Expiration)

			if tc.Never {
				assert.True(t, math.IsInf(score, 1))
			} else {
				assert.InDelta(t, float64(time.Now().Add(tc.Expiration).Unix()), score, 2)
			}
		})
	}
}

func TestRedis_SessionGetIDsByUsername(t *testing.T) {
	t.Run("ShouldPruneExpiredMembersBeforeReading", func(t *testing.T) {
		client := &mockRedisCmdable{pipeliner: &mockRedisPipeliner{members: []string{"id1", "id2"}, pruned: map[string]string{}}}

		ids, err := NewRedis(client, "standalone").SessionGetIDsByUsername(context.Background(), "example.com", "john")

		require.NoError(t, err)
		assert.Equal(t, []string{"id1", "id2"}, ids)

		pruned, ok := client.pipeliner.pruned[getSessionUserKey("example.com", "john")]
		require.True(t, ok)

		score, err := strconv.ParseInt(pruned, 10, 64)
		require.NoError(t, err)
		assert.InDelta(t, time.Now().Unix(), score, 2)
	})

	t.Run("ShouldNotQueryForAnonymousSession", func(t *testing.T) {
		client := &mockRedisCmdable{}

		ids, err := NewRedis(client, "standalone").SessionGetIDsByUsername(context.Background(), "example.com", "")

		require.NoError(t, err)
		assert.Nil(t, ids)
	})
}

func TestRedis_SessionGarbageCollection(t *testing.T) {
	client := &mockRedisCmdable{
		scanned: []string{
			getSessionUserKey("example.com", "john"),
			getSessionUserKey("example.com", "jane"),
		},
		pruned: map[string]string{},
	}

	provider := NewRedis(client, "standalone")

	assert.Equal(t, sessionGarbageCollectionFrequency, provider.SessionGarbageCollectionFrequency(context.Background()))

	require.NoError(t, provider.SessionGarbageCollection(context.Background()))

	assert.Len(t, client.pruned, 2)
	assert.Contains(t, client.pruned, getSessionUserKey("example.com", "john"))
	assert.Contains(t, client.pruned, getSessionUserKey("example.com", "jane"))
}

type mockRedisCmdable struct {
	redis.Cmdable

	values    map[string]string
	err       error
	ttl       time.Duration
	added     map[string][]redis.Z
	scanned   []string
	pruned    map[string]string
	pipeliner *mockRedisPipeliner
}

func (m *mockRedisCmdable) TTL(ctx context.Context, key string) *redis.DurationCmd {
	cmd := redis.NewDurationCmd(ctx, time.Second, "ttl", key)
	cmd.SetVal(m.ttl)

	return cmd
}

func (m *mockRedisCmdable) ZAdd(ctx context.Context, key string, members ...redis.Z) *redis.IntCmd {
	m.added[key] = append(m.added[key], members...)

	return redis.NewIntCmd(ctx, "zadd", key)
}

func (m *mockRedisCmdable) Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd {
	cmd := redis.NewScanCmd(ctx, nil, "scan", cursor, match, count)
	cmd.SetVal(m.scanned, 0)

	return cmd
}

func (m *mockRedisCmdable) ZRemRangeByScore(ctx context.Context, key, min, max string) *redis.IntCmd {
	m.pruned[key] = max

	return redis.NewIntCmd(ctx, "zremrangebyscore", key)
}

func (m *mockRedisCmdable) TxPipeline() redis.Pipeliner {
	return m.pipeliner
}

type mockRedisPipeliner struct {
	redis.Pipeliner

	members []string
	pruned  map[string]string
	err     error
}

func (m *mockRedisPipeliner) ZRemRangeByScore(ctx context.Context, key, min, max string) *redis.IntCmd {
	m.pruned[key] = max

	return redis.NewIntCmd(ctx, "zremrangebyscore", key)
}

func (m *mockRedisPipeliner) ZRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	cmd := redis.NewStringSliceCmd(ctx, "zrange", key, start, stop)
	cmd.SetVal(m.members)

	return cmd
}

func (m *mockRedisPipeliner) Exec(ctx context.Context) (cmds []redis.Cmder, err error) {
	return nil, m.err
}

func (m *mockRedisCmdable) Get(ctx context.Context, key string) *redis.StringCmd {
	cmd := redis.NewStringCmd(ctx, "get", key)

	switch value, ok := m.values[key]; {
	case m.err != nil:
		cmd.SetErr(m.err)
	case ok:
		cmd.SetVal(value)
	default:
		cmd.SetErr(redis.Nil)
	}

	return cmd
}

func TestGetFailingTimeoutSeconds(t *testing.T) {
	testCases := []struct {
		Name     string
		Have     time.Duration
		Expected int
	}{
		{"ShouldReturnZeroWhenUnset", 0, 0},
		{"ShouldReturnZeroWhenNegative", -time.Second, 0},
		{"ShouldReturnWholeSeconds", time.Second * 15, 15},
		{"ShouldTruncateToWholeSeconds", time.Millisecond * 2500, 2},
		{"ShouldRaiseSubSecondToOne", time.Millisecond * 100, 1},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			assert.Equal(t, tc.Expected, getFailingTimeoutSeconds(tc.Have))
		})
	}
}
