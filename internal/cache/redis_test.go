package cache

import (
	"context"
	"errors"
	"math"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/session"
)

func TestRedis_SessionGet(t *testing.T) {
	testCases := []struct {
		Name     string
		Values   map[string]string
		Err      error
		Expected session.Record
		Error    string
	}{
		{"ShouldReturnRecordWhenPresent", map[string]string{getSessionKey("example.com", "id"): "data"}, nil, session.NewRecord("id", []byte("data")), ""},
		{"ShouldReturnNoRecordWhenMissing", nil, nil, nil, ""},
		{"ShouldReturnErrorOnFailure", nil, errors.New("connection refused"), nil, "connection refused"},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			provider := NewRedis(&mockRedisCmdable{values: tc.Values, err: tc.Err}, "standalone")

			record, err := provider.SessionGet(context.Background(), "example.com", "id")

			if tc.Error == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.Error)
			}

			assert.Equal(t, tc.Expected, record)
		})
	}
}

func TestRedis_SessionGetByPublicID(t *testing.T) {
	testCases := []struct {
		Name     string
		Values   map[string]string
		Err      error
		Expected session.Record
		Error    string
	}{
		{
			"ShouldResolveThroughToTheSession",
			map[string]string{
				getSessionPublicKey("example.com", "pid"): "id",
				getSessionKey("example.com", "id"):        "data",
			},
			nil, session.NewRecord("id", []byte("data")), "",
		},
		{"ShouldReturnNoRecordWhenPublicIDMissing", nil, nil, nil, ""},
		{
			"ShouldReturnNoRecordWhenSessionMissing",
			map[string]string{getSessionPublicKey("example.com", "pid"): "id"},
			nil, nil, "",
		},
		{"ShouldReturnErrorOnFailure", nil, errors.New("connection refused"), nil, "connection refused"},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			provider := NewRedis(&mockRedisCmdable{values: tc.Values, err: tc.Err}, "standalone")

			record, err := provider.SessionGetByPublicID(context.Background(), "example.com", "pid")

			if tc.Error == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.Error)
			}

			assert.Equal(t, tc.Expected, record)
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

func TestRedis_SessionChangeID(t *testing.T) {
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
			"ShouldMoveSessionAndUsernameIndexInASingleScript",
			"john",
			time.Hour,
			nil,
			"",
			[]string{
				getSessionKey("example.com", "old"),
				getSessionKey("example.com", "new"),
				getSessionPublicKey("example.com", "pid"),
				getSessionUserKey("example.com", "john"),
			},
			func(t *testing.T, args []any) {
				require.Len(t, args, 6)
				assert.Equal(t, []byte("resealed"), args[0])
				assert.Equal(t, int64(3600000), args[1])
				assert.Equal(t, "new", args[2])
				assert.Equal(t, "old", args[3])
				assert.InDelta(t, float64(time.Now().Add(time.Hour).Unix()), args[4], 2)
				assert.Equal(t, "new", args[5])
			},
		},
		{
			"ShouldOmitTheUsernameIndexForAnAnonymousSession",
			"",
			time.Hour,
			nil,
			"",
			[]string{
				getSessionKey("example.com", "old"),
				getSessionKey("example.com", "new"),
				getSessionPublicKey("example.com", "pid"),
			},
			func(t *testing.T, args []any) {
				require.Len(t, args, 3)
				assert.Equal(t, []byte("resealed"), args[0])
				assert.Equal(t, int64(3600000), args[1])
				assert.Equal(t, "new", args[2])
			},
		},
		{
			"ShouldNotExpireKeysWhenTheExpirationIsNotPositive",
			"john",
			0,
			nil,
			"",
			[]string{
				getSessionKey("example.com", "old"),
				getSessionKey("example.com", "new"),
				getSessionPublicKey("example.com", "pid"),
				getSessionUserKey("example.com", "john"),
			},
			func(t *testing.T, args []any) {
				require.Len(t, args, 6)
				assert.Equal(t, int64(0), args[1])
				assert.True(t, math.IsInf(args[4].(float64), 1))
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
		t.Run(tc.Name, func(t *testing.T) {
			client := &mockRedisCmdable{err: tc.Err}

			err := NewRedis(client, "standalone").SessionChangeID(context.Background(), "example.com", "old", "new", "pid", tc.Username, tc.Expiration, []byte("resealed"))

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

func TestRedis_SessionChangeIDKeysShareAClusterSlot(t *testing.T) {
	client := &mockRedisCmdable{}

	require.NoError(t, NewRedis(client, "standalone").SessionChangeID(context.Background(), "example.com", "old", "new", "pid", "john", time.Hour, []byte("resealed")))
	require.Len(t, client.evalKeys, 4)

	for _, key := range client.evalKeys {
		assert.Equal(t, "example.com", key[strings.Index(key, "{")+1:strings.Index(key, "}")])
	}
}

func TestRedis_SessionExpirationMilliseconds(t *testing.T) {
	testCases := []struct {
		Name       string
		Expiration time.Duration
		Expected   int64
	}{
		{"ShouldConvertAPositiveExpiration", time.Hour, 3600000},
		{"ShouldNotExpireWhenZero", 0, 0},
		{"ShouldNotExpireWhenNegative", -time.Hour, 0},
		{"ShouldRaiseASubMillisecondExpiration", time.Microsecond, 1},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			assert.Equal(t, tc.Expected, getSessionExpirationMilliseconds(tc.Expiration))
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
	evalKeys  []string
	evalArgs  []any
}

func (m *mockRedisCmdable) EvalSha(ctx context.Context, sha1 string, keys []string, args ...any) *redis.Cmd {
	m.evalKeys = keys
	m.evalArgs = args

	cmd := redis.NewCmd(ctx, "evalsha", sha1)

	if m.err != nil {
		cmd.SetErr(m.err)
	}

	return cmd
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
