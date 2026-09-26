// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

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
		Hashes   map[string]map[string]string
		Err      error
		Expected session.Record
		Error    string
	}{
		{
			"ShouldResolveThroughToTheSession",
			map[string]string{getSessionPublicKey("example.com", "pid"): "id"},
			map[string]map[string]string{getSessionKey("example.com", "id"): {"data": "data", "pid": "pid"}},
			nil, session.NewRecord("id", []byte("data")), "",
		},
		{"ShouldReturnNoRecordWhenPublicIDMissing", nil, nil, nil, nil, ""},
		{
			"ShouldReturnNoRecordWhenSessionMissing",
			map[string]string{getSessionPublicKey("example.com", "pid"): "id"},
			nil, nil, nil, "",
		},
		{
			"ShouldReturnNoRecordWhenTheSessionRecordsAnotherPublicID",
			map[string]string{getSessionPublicKey("example.com", "pid"): "id"},
			map[string]map[string]string{getSessionKey("example.com", "id"): {"data": "data", "pid": "another"}},
			nil, nil, "",
		},
		{
			"ShouldReturnNoRecordWhenTheSessionWasMoved",
			map[string]string{getSessionPublicKey("example.com", "pid"): "id"},
			map[string]map[string]string{getSessionKey("example.com", "id"): {"moved": "new"}},
			nil, nil, "",
		},
		{"ShouldReturnErrorOnFailure", nil, nil, errors.New("connection refused"), nil, "connection refused"},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			provider := NewRedis(&mockRedisCmdable{values: tc.Values, hashes: tc.Hashes, err: tc.Err}, "standalone")

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
		{"ShouldBuildSessionKey", getSessionKey("example.com", "id"), "authelia:session:{example.com:id}"},
		{"ShouldBuildPublicKey", getSessionPublicKey("example.com", "pid"), "authelia:session-public:{example.com:pid}"},
		{"ShouldBuildUserKey", getSessionUserKey("example.com", "john"), "authelia:session-user:{example.com:john}"},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			assert.Equal(t, tc.Expected, tc.Actual)
		})
	}
}

func TestRedis_SessionKeysShouldBeTaggedByTheirOwnValue(t *testing.T) {
	for _, key := range []string{getSessionKey("example.com", "id"), getSessionPublicKey("example.com", "id"), getSessionUserKey("example.com", "id")} {
		assert.Equal(t, "example.com:id", getTestHashTag(key))
	}

	assert.NotEqual(t, getTestHashTag(getSessionKey("example.com", "one")), getTestHashTag(getSessionKey("example.com", "two")))
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
	keyOld, keyNew := getSessionKey("example.com", "old"), getSessionKey("example.com", "new")

	testCases := []struct {
		Name       string
		Username   string
		Expiration time.Duration
		Results    []any
		Error      string
		Assert     func(t *testing.T, client *mockRedisCmdable)
	}{
		{
			"ShouldMarkTheOldSessionMovedThenSaveAndIndexTheNewSession",
			"john",
			time.Hour,
			[]any{[]any{int64(1), "pid", "john"}},
			"",
			func(t *testing.T, client *mockRedisCmdable) {
				require.Len(t, client.evals, 2)

				assert.Equal(t, mockRedisEval{script: redisSessionMove, keys: []string{keyOld}, args: []any{"new", int64(3600000)}}, client.evals[0])
				assert.Equal(t, mockRedisEval{script: redisSessionSave, keys: []string{keyNew}, args: []any{[]byte("resealed"), int64(3600000), "pid", "john"}}, client.evals[1])

				assert.Equal(t, []mockRedisSet{{key: getSessionPublicKey("example.com", "pid"), value: "new", expiration: time.Hour}}, client.sets)
				require.Len(t, client.added[getSessionUserKey("example.com", "john")], 1)
				assert.Equal(t, "new", client.added[getSessionUserKey("example.com", "john")][0].Member)
				assert.Equal(t, map[string][]any{getSessionUserKey("example.com", "john"): {"old"}}, client.removed)
			},
		},
		{
			"ShouldRetireThePublicIDOfTheOldSessionWhenItDiffers",
			"john",
			time.Hour,
			[]any{[]any{int64(1), "oldpid", "john"}},
			"",
			func(t *testing.T, client *mockRedisCmdable) {
				require.Len(t, client.evals, 3)

				assert.Equal(t, mockRedisEval{script: redisDeleteIfEqual, keys: []string{getSessionPublicKey("example.com", "oldpid")}, args: []any{"old"}}, client.evals[2])
			},
		},
		{
			"ShouldOmitTheUsernameIndexForAnAnonymousSession",
			"",
			time.Hour,
			[]any{[]any{int64(1), "pid", ""}},
			"",
			func(t *testing.T, client *mockRedisCmdable) {
				require.Len(t, client.evals, 2)

				assert.Empty(t, client.added)
				assert.Empty(t, client.removed)
			},
		},
		{
			"ShouldNotExpireKeysWhenTheExpirationIsNotPositive",
			"john",
			0,
			[]any{[]any{int64(1), "pid", "john"}},
			"",
			func(t *testing.T, client *mockRedisCmdable) {
				assert.Equal(t, []any{"new", int64(0)}, client.evals[0].args)
				assert.Equal(t, time.Duration(0), client.sets[0].expiration)
				assert.True(t, math.IsInf(client.added[getSessionUserKey("example.com", "john")][0].Score, 1))
			},
		},
		{
			"ShouldNotRecreateASessionWhichNoLongerExists",
			"john",
			time.Hour,
			[]any{[]any{int64(0)}},
			"",
			func(t *testing.T, client *mockRedisCmdable) {
				require.Len(t, client.evals, 1)

				assert.Empty(t, client.sets)
				assert.Empty(t, client.added)
			},
		},
		{
			"ShouldReturnErrorWhenTheMoveFails",
			"john",
			time.Hour,
			[]any{errors.New("connection refused")},
			"connection refused",
			func(t *testing.T, client *mockRedisCmdable) {
				require.Len(t, client.evals, 1)
			},
		},
		{
			"ShouldReturnErrorWhenTheSaveFails",
			"john",
			time.Hour,
			[]any{[]any{int64(1), "pid", "john"}, errors.New("connection refused")},
			"connection refused",
			func(t *testing.T, client *mockRedisCmdable) {
				require.Len(t, client.evals, 2)

				assert.Empty(t, client.sets)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			client := &mockRedisCmdable{evalResults: tc.Results}

			err := NewRedis(client, "standalone").SessionChangeID(context.Background(), "example.com", "old", "new", "pid", tc.Username, tc.Expiration, []byte("resealed"))

			if tc.Error == "" {
				require.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.Error)
			}

			tc.Assert(t, client)
		})
	}
}

func TestRedis_SessionScriptsShouldOperateOnASingleKey(t *testing.T) {
	ctx := context.Background()
	client := &mockRedisCmdable{evalResults: []any{[]any{int64(1), "oldpid", "jane"}, []any{int64(1), "oldpid", "jane"}}}
	provider := NewRedis(client, "standalone")

	require.NoError(t, provider.SessionSave(ctx, "example.com", "id", "pid", "john", time.Hour, []byte("data")))
	require.NoError(t, provider.SessionChangeID(ctx, "example.com", "id", "new", "pid", "john", time.Hour, []byte("data")))
	require.NoError(t, provider.SessionDelete(ctx, "example.com", "new", "", ""))

	require.NotEmpty(t, client.evals)

	for _, eval := range client.evals {
		assert.Len(t, eval.keys, 1)
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

func getTestHashTag(key string) string {
	return key[strings.Index(key, "{")+1 : strings.Index(key, "}")]
}

type mockRedisCmdable struct {
	redis.Cmdable

	values    map[string]string
	hashes    map[string]map[string]string
	err       error
	ttl       time.Duration
	added     map[string][]redis.Z
	scanned   []string
	pruned    map[string]string
	pruneErr  error
	pipeliner *mockRedisPipeliner

	evals       []mockRedisEval
	evalResults []any

	sets        []mockRedisSet
	removed     map[string][]any
	pipelineErr error
	removeErr   error
}

type mockRedisEval struct {
	script *redis.Script
	keys   []string
	args   []any
}

type mockRedisSet struct {
	key        string
	value      any
	expiration time.Duration
}

var mockRedisScripts = []*redis.Script{redisSessionSave, redisSessionMove, redisSessionDelete, redisDeleteIfEqual}

func (m *mockRedisCmdable) EvalSha(ctx context.Context, sha1 string, keys []string, args ...any) *redis.Cmd {
	eval := mockRedisEval{keys: keys, args: args}

	for _, script := range mockRedisScripts {
		if script.Hash() == sha1 {
			eval.script = script
		}
	}

	m.evals = append(m.evals, eval)

	cmd := redis.NewCmd(ctx, "evalsha", sha1)

	switch {
	case m.err != nil:
		cmd.SetErr(m.err)
	case len(m.evalResults) != 0:
		result := m.evalResults[0]
		m.evalResults = m.evalResults[1:]

		if err, ok := result.(error); ok {
			cmd.SetErr(err)
		} else {
			cmd.SetVal(result)
		}
	default:
		cmd.SetVal([]any{int64(1), "", ""})
	}

	return cmd
}

func (m *mockRedisCmdable) Pipelined(ctx context.Context, fn func(redis.Pipeliner) error) ([]redis.Cmder, error) {
	if err := fn(&mockRedisIndexPipeliner{parent: m}); err != nil {
		return nil, err
	}

	return nil, m.pipelineErr
}

func (m *mockRedisCmdable) ZRem(ctx context.Context, key string, members ...any) *redis.IntCmd {
	if m.removed == nil {
		m.removed = map[string][]any{}
	}

	m.removed[key] = append(m.removed[key], members...)

	cmd := redis.NewIntCmd(ctx, "zrem", key)

	if m.removeErr != nil {
		cmd.SetErr(m.removeErr)
	}

	return cmd
}

func (m *mockRedisCmdable) HMGet(ctx context.Context, key string, fields ...string) *redis.SliceCmd {
	cmd := redis.NewSliceCmd(ctx, "hmget", key)

	if m.err != nil {
		cmd.SetErr(m.err)

		return cmd
	}

	values := make([]any, len(fields))

	for i, field := range fields {
		if value, ok := m.hashes[key][field]; ok {
			values[i] = value
		}
	}

	cmd.SetVal(values)

	return cmd
}

type mockRedisIndexPipeliner struct {
	redis.Pipeliner

	parent *mockRedisCmdable
}

func (p *mockRedisIndexPipeliner) Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd {
	p.parent.sets = append(p.parent.sets, mockRedisSet{key: key, value: value, expiration: expiration})

	return redis.NewStatusCmd(ctx, "set", key)
}

func (p *mockRedisIndexPipeliner) ZAdd(ctx context.Context, key string, members ...redis.Z) *redis.IntCmd {
	if p.parent.added == nil {
		p.parent.added = map[string][]redis.Z{}
	}

	p.parent.added[key] = append(p.parent.added[key], members...)

	return redis.NewIntCmd(ctx, "zadd", key)
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

	cmd := redis.NewIntCmd(ctx, "zremrangebyscore", key)

	if m.pruneErr != nil {
		cmd.SetErr(m.pruneErr)
	}

	return cmd
}

func (m *mockRedisCmdable) TxPipeline() redis.Pipeliner {
	return m.pipeliner
}

func (m *mockRedisCmdable) HGet(ctx context.Context, key, field string) *redis.StringCmd {
	return m.Get(ctx, key)
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
