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
	// redisTTLNoKey is the TTL reply for a key which does not exist.
	redisTTLNoKey = time.Duration(-2)

	// redisTTLNoExpiry is the TTL reply for a key which exists but has no associated expiry.
	redisTTLNoExpiry = time.Duration(-1)
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
