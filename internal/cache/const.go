package cache

import (
	"time"
)

const (
	driverParameterFmtAppName = "authelia %s"
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
	redisKeySession       = "session:{"
	redisKeySessionUser   = "session-user:{"
	redisKeySessionPublic = "session-public:{"

	redisPrefixSession       = redisPrefix + redisKeySession
	redisPrefixSessionUser   = redisPrefix + redisKeySessionUser
	redisPrefixSessionPublic = redisPrefix + redisKeySessionPublic
)
