// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"github.com/redis/go-redis/v9"
)

var mockRedisScripts = []*redis.Script{redisSessionSave, redisSessionMove, redisSessionDelete, redisDeleteIfEqual}
