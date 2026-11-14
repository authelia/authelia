// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

// Package cache provides the cache backends which store session records, either in process memory or in Redis.
package cache

import (
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	driverParameterFmtAppName = "authelia-%s"
)

const (
	sessionGarbageCollectionFrequency = time.Minute * 5
)

const (
	redisVariantStandalone = "standalone"
	redisVariantSentinel   = "sentinel"
	redisVariantCluster    = "cluster"
)

const (
	redisScoreMinimum = "-inf"
	redisScanCount    = 100
)

const (
	redisPrefix           = "authelia:"
	redisKeySeparator     = ":"
	redisKeySeparatorSlot = "}"

	redisSessionFieldData = "data"
	redisSessionFieldPID  = "pid"
	redisKeySession       = "session:{"
	redisKeySessionUser   = "session-user:{"
	redisKeySessionPublic = "session-public:{"

	redisPrefixSession       = redisPrefix + redisKeySession
	redisPrefixSessionUser   = redisPrefix + redisKeySessionUser
	redisPrefixSessionPublic = redisPrefix + redisKeySessionPublic
)

const (
	redisScriptSessionSave = `
if redis.call('hexists', KEYS[1], 'moved') == 1 or redis.call('hexists', KEYS[1], 'deleted') == 1 then
	return {0}
end

if ARGV[5] == '1' and redis.call('hget', KEYS[1], 'data') ~= ARGV[6] then
	return {0}
end

local ttl = tonumber(ARGV[2])
local previous = redis.call('hmget', KEYS[1], 'pid', 'username')

redis.call('hset', KEYS[1], 'data', ARGV[1], 'pid', ARGV[3], 'username', ARGV[4])

if ttl > 0 then
	redis.call('pexpire', KEYS[1], ttl)
else
	redis.call('persist', KEYS[1])
end

return {1, previous[1] or '', previous[2] or ''}
`

	redisScriptSessionMove = `
if redis.call('exists', KEYS[1]) == 0 or redis.call('hexists', KEYS[1], 'moved') == 1 or redis.call('hexists', KEYS[1], 'deleted') == 1 then
	return {0}
end

local ttl = tonumber(ARGV[2])
local previous = redis.call('hmget', KEYS[1], 'pid', 'username')

redis.call('del', KEYS[1])
redis.call('hset', KEYS[1], 'moved', ARGV[1])

if ttl > 0 then
	redis.call('pexpire', KEYS[1], ttl)
end

return {1, previous[1] or '', previous[2] or ''}
`

	redisScriptSessionDelete = `
if redis.call('exists', KEYS[1]) == 0 or redis.call('hexists', KEYS[1], 'moved') == 1 or redis.call('hexists', KEYS[1], 'deleted') == 1 then
	return {'', ''}
end

local ttl = redis.call('pttl', KEYS[1])
local previous = redis.call('hmget', KEYS[1], 'pid', 'username')

redis.call('del', KEYS[1])
redis.call('hset', KEYS[1], 'deleted', 1)

if ttl > 0 then
	redis.call('pexpire', KEYS[1], ttl)
end

return {previous[1] or '', previous[2] or ''}
`

	redisScriptSessionGet = `
local data = redis.call('hget', KEYS[1], 'data')

if not data then
	return {}
end

return {data, redis.call('pttl', KEYS[1])}
`

	redisScriptSessionTouch = `
if redis.call('hexists', KEYS[1], 'data') == 0 then
	return 0
end

local ttl = tonumber(ARGV[1])
local current = redis.call('pttl', KEYS[1])

if current < 0 or ttl <= current then
	return 0
end

redis.call('pexpire', KEYS[1], ttl)

return 1
`

	redisScriptSessionTouchPublicID = `
local current = redis.call('pttl', KEYS[1])

if current == -2 then
	redis.call('set', KEYS[1], ARGV[1], 'px', ARGV[2])

	return 1
end

if current == -1 or tonumber(ARGV[2]) <= current or redis.call('get', KEYS[1]) ~= ARGV[1] then
	return 0
end

redis.call('pexpire', KEYS[1], ARGV[2])

return 1
`

	redisScriptSessionTouchUsername = `
local current = redis.call('zscore', KEYS[1], ARGV[1])

if current and tonumber(ARGV[2]) <= tonumber(current) then
	return 0
end

redis.call('zadd', KEYS[1], ARGV[2], ARGV[1])

return 1
`

	redisScriptDeleteIfEqual = `
if redis.call('get', KEYS[1]) == ARGV[1] then
	return redis.call('del', KEYS[1])
end

return 0
`
)

var (
	redisSessionSave   = redis.NewScript(redisScriptSessionSave)
	redisSessionMove   = redis.NewScript(redisScriptSessionMove)
	redisSessionDelete = redis.NewScript(redisScriptSessionDelete)
	redisSessionGet    = redis.NewScript(redisScriptSessionGet)
	redisSessionTouch  = redis.NewScript(redisScriptSessionTouch)
	redisDeleteIfEqual = redis.NewScript(redisScriptDeleteIfEqual)

	redisSessionTouchPublicID = redis.NewScript(redisScriptSessionTouchPublicID)
	redisSessionTouchUsername = redis.NewScript(redisScriptSessionTouchUsername)
)
