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
if redis.call('hexists', KEYS[1], 'moved') == 1 then
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
if redis.call('exists', KEYS[1]) == 0 or redis.call('hexists', KEYS[1], 'moved') == 1 then
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
local previous = redis.call('hmget', KEYS[1], 'pid', 'username')

redis.call('del', KEYS[1])

return {previous[1] or '', previous[2] or ''}
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
	redisDeleteIfEqual = redis.NewScript(redisScriptDeleteIfEqual)
)
