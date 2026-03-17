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
	// sessionGarbageCollectionFrequency is the frequency expired sessions are collected at.
	sessionGarbageCollectionFrequency = time.Minute * 5
)

const (
	// redisScoreMinimum is the lower bound of a ZRANGEBYSCORE style range, which is unbounded.
	redisScoreMinimum = "-inf"

	// redisScanCount is the number of keys hinted per SCAN iteration during garbage collection.
	redisScanCount = 100
)

const (
	redisPrefix           = "authelia:"
	redisKeySeparator     = ":"
	redisKeySeparatorSlot = "}" + redisKeySeparator

	redisSessionFieldData = "data"
	redisKeySession       = "session:{"
	redisKeySessionUser   = "session-user:{"
	redisKeySessionPublic = "session-public:{"

	redisPrefixSession       = redisPrefix + redisKeySession
	redisPrefixSessionUser   = redisPrefix + redisKeySessionUser
	redisPrefixSessionPublic = redisPrefix + redisKeySessionPublic
)

// The scripts below are run by the Redis provider. Each is given the keys it operates on, which must be passed as keys
// rather than derived within the script so that a cluster client can route them, and the values it needs as arguments.
// The prefixes are passed as arguments so a key discovered within the script, such as the public id recorded against a
// session, can be turned into the key which holds it.
const (
	redisScriptSessionSave = `
local ttl = tonumber(ARGV[2])
local current = redis.call('get', KEYS[2])

if current and current ~= ARGV[3] then
	return 0
end

local previous = redis.call('hmget', KEYS[1], 'pid', 'username')

if previous[1] and previous[1] ~= ARGV[4] and redis.call('get', ARGV[7] .. previous[1]) == ARGV[3] then
	redis.call('del', ARGV[7] .. previous[1])
end

if previous[2] and previous[2] ~= '' and previous[2] ~= ARGV[5] then
	redis.call('zrem', ARGV[8] .. previous[2], ARGV[3])
end

redis.call('hset', KEYS[1], 'data', ARGV[1], 'pid', ARGV[4], 'username', ARGV[5])

if ttl > 0 then
	redis.call('pexpire', KEYS[1], ttl)
	redis.call('set', KEYS[2], ARGV[3], 'px', ttl)
else
	redis.call('persist', KEYS[1])
	redis.call('set', KEYS[2], ARGV[3])
end

if #KEYS > 2 then
	redis.call('zadd', KEYS[3], ARGV[6], ARGV[3])
end

return 1
`

	redisScriptSessionDelete = `
local previous = redis.call('hmget', KEYS[1], 'pid', 'username')

redis.call('del', KEYS[1])

if ARGV[2] ~= '' then
	if redis.call('get', ARGV[4] .. ARGV[2]) == ARGV[1] then
		redis.call('del', ARGV[4] .. ARGV[2])
	end
elseif previous[1] and previous[1] ~= '' and redis.call('get', ARGV[4] .. previous[1]) == ARGV[1] then
	redis.call('del', ARGV[4] .. previous[1])
end

local username = ARGV[3]

if username == '' and previous[2] then
	username = previous[2]
end

if username ~= '' then
	redis.call('zrem', ARGV[5] .. username, ARGV[1])
end

return 1
`

	redisScriptSessionChangeID = `
local ttl = tonumber(ARGV[2])

if redis.call('exists', KEYS[1]) == 0 then
	return 0
end

local previous = redis.call('hmget', KEYS[1], 'pid', 'username')

redis.call('del', KEYS[1])

if previous[1] and previous[1] ~= ARGV[4] and redis.call('get', ARGV[8] .. previous[1]) == ARGV[7] then
	redis.call('del', ARGV[8] .. previous[1])
end

if previous[2] and previous[2] ~= '' then
	redis.call('zrem', ARGV[9] .. previous[2], ARGV[7])
end

redis.call('hset', KEYS[2], 'data', ARGV[1], 'pid', ARGV[4], 'username', ARGV[5])

if ttl > 0 then
	redis.call('pexpire', KEYS[2], ttl)
	redis.call('set', KEYS[3], ARGV[3], 'px', ttl)
else
	redis.call('persist', KEYS[2])
	redis.call('set', KEYS[3], ARGV[3])
end

if #KEYS > 3 then
	redis.call('zadd', KEYS[4], ARGV[6], ARGV[3])
end

return 1
`
)

var (
	redisSessionSave     = redis.NewScript(redisScriptSessionSave)
	redisSessionDelete   = redis.NewScript(redisScriptSessionDelete)
	redisSessionChangeID = redis.NewScript(redisScriptSessionChangeID)
)
