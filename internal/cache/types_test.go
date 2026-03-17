// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"time"

	"github.com/redis/go-redis/v9"
)

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
