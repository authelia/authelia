package session

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"github.com/valyala/bytebufferpool"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// LICENSE: MIT https://github.com/fasthttp/session/blob/master/LICENSE
// SOURCE: https://github.com/fasthttp/session/blob/cd9080042fc350c0b630c401f43e7d5ecee77882/providers/memory/provider.go
// Changes to the original code prior to the first commit were only aesthetic. All other changes are logged via git SCM.

// NewRedisStandaloneStore returns a new RedisStore which connects to a redis standalone instance.
func NewRedisStandaloneStore(config *schema.RedisSessionConfiguration, certPool *x509.CertPool, logger *logrus.Logger) (provider *RedisStore) {
	redis.SetLogger(&redisLogger{logger})

	network := "tcp"

	var addr string

	if config.Port == 0 {
		network = "unix"
		addr = config.Host
	} else {
		addr = fmt.Sprintf("%s:%d", config.Host, config.Port)
	}

	var tlsConfig *tls.Config

	if config.TLS != nil {
		tlsConfig = utils.NewTLSConfig(config.TLS, tls.VersionTLS12, certPool)
	}

	return &RedisStore{
		db: redis.NewClient(&redis.Options{
			Network:      network,
			Addr:         addr,
			Username:     config.Username,
			Password:     config.Password,
			DB:           config.DatabaseIndex,
			PoolSize:     config.MaximumActiveConnections,
			MinIdleConns: config.MinimumIdleConnections,
			IdleTimeout:  300,
			TLSConfig:    tlsConfig,
		}),
		separator:     []byte(redisKeySeparator),
		wildcard:      []byte(redisKeyWildcard),
		prefixSession: []byte(fmt.Sprintf("%s%s%s", redisKeyPrefix, redisKeySeparator, redisKeyPrefixSession)),
		prefixProfile: []byte(fmt.Sprintf("%s%s%s", redisKeyPrefix, redisKeySeparator, redisKeyPrefixProfile)),
	}
}

// NewRedisFailoverStore returns a new RedisStore which connects to a redis failover instance (redis sentinel).
func NewRedisFailoverStore(config *schema.RedisSessionConfiguration, certPool *x509.CertPool, logger *logrus.Logger) (provider *RedisStore) {
	redis.SetLogger(&redisLogger{logger})

	var tlsConfig *tls.Config

	if config.TLS != nil {
		tlsConfig = utils.NewTLSConfig(config.TLS, tls.VersionTLS12, certPool)
	}

	addrs := make([]string, 0)

	if config.Host != "" {
		addrs = append(addrs, fmt.Sprintf("%s:%d", strings.ToLower(config.Host), config.Port))
	}

	for _, node := range config.HighAvailability.Nodes {
		addr := fmt.Sprintf("%s:%d", strings.ToLower(node.Host), node.Port)
		if !utils.IsStringInSlice(addr, addrs) {
			addrs = append(addrs, addr)
		}
	}

	return &RedisStore{
		db: redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:       config.HighAvailability.SentinelName,
			SentinelAddrs:    addrs,
			SentinelPassword: config.HighAvailability.SentinelPassword,
			RouteByLatency:   config.HighAvailability.RouteByLatency,
			RouteRandomly:    config.HighAvailability.RouteRandomly,
			Username:         config.Username,
			Password:         config.Password,
			DB:               config.DatabaseIndex,
			PoolSize:         config.MaximumActiveConnections,
			MinIdleConns:     config.MinimumIdleConnections,
			IdleTimeout:      300,
			TLSConfig:        tlsConfig,
		}),
		separator:      []byte(redisKeySeparator),
		wildcard:       []byte(redisKeyWildcard),
		prefixSession:  []byte(fmt.Sprintf("%s%s%s", redisKeyPrefix, redisKeySeparator, redisKeyPrefixSession)),
		prefixSessions: []byte(fmt.Sprintf("%s%s%ss", redisKeyPrefix, redisKeySeparator, redisKeyPrefixSession)),
		prefixProfile:  []byte(fmt.Sprintf("%s%s%s", redisKeyPrefix, redisKeySeparator, redisKeyPrefixProfile)),
	}
}

// RedisStore is a session store inside redis.
type RedisStore struct {
	db redis.Cmdable

	separator      []byte
	wildcard       []byte
	prefixSession  []byte
	prefixSessions []byte
	prefixProfile  []byte
}

func (s RedisStore) getProfileKey(uid []byte) (finalKey string) {
	key := bytebufferpool.Get()
	key.Set(s.prefixSession)

	_, _ = key.Write(s.separator)
	_, _ = key.Write(uid)

	finalKey = key.String()
	bytebufferpool.Put(key)

	return finalKey
}

func (s RedisStore) getSessionsKey(uid []byte) (finalKey string) {
	key := bytebufferpool.Get()
	key.Set(s.prefixSession)

	_, _ = key.Write(s.separator)
	_, _ = key.Write(uid)

	finalKey = key.String()
	bytebufferpool.Put(key)

	return finalKey
}

func (s RedisStore) getSessionKey(id []byte) (finalKey string) {
	key := bytebufferpool.Get()
	key.Set(s.prefixSession)

	_, _ = key.Write(s.separator)
	_, _ = key.Write(id)

	finalKey = key.String()
	bytebufferpool.Put(key)

	return finalKey
}

// Get returns the data of the given session id.
func (s *RedisStore) Get(id []byte) (data []byte, err error) {
	key := s.getSessionKey(id)

	data, err = s.db.Get(context.Background(), key).Bytes()
	if err != nil && err != redis.Nil {
		return nil, err
	}

	return data, nil
}

// Save the session data and expiration from the given session id.
func (s *RedisStore) Save(id, data []byte, expiration time.Duration) (err error) {
	key := s.getSessionKey(id)

	return s.db.Set(context.Background(), key, data, expiration).Err()
}

// Regenerate updates the session id and expiration with the new session id of the given current session id.
func (s *RedisStore) Regenerate(id, newID []byte, expiration time.Duration) (err error) {
	key := s.getSessionKey(id)
	newKey := s.getSessionKey(newID)

	exists, err := s.db.Exists(context.Background(), key).Result()
	if err != nil {
		return err
	}

	if exists > 0 {
		if err = s.db.Rename(context.Background(), key, newKey).Err(); err != nil {
			return err
		}

		if err = s.db.Expire(context.Background(), newKey, expiration).Err(); err != nil {
			return err
		}
	}

	return nil
}

// Destroy destroys the session from the given id.
func (s *RedisStore) Destroy(id []byte) error {
	key := s.getSessionKey(id)

	return s.db.Del(context.Background(), key).Err()
}

// Count returns the count of stored sessions.
func (s *RedisStore) Count() int {
	reply, err := s.db.Keys(context.Background(), s.getSessionKey([]byte("*"))).Result()
	if err != nil {
		return 0
	}

	return len(reply)
}

// NeedGC indicates if the GC needs to be run.
func (s *RedisStore) NeedGC() bool {
	return false
}

// GC destroys the expired sessions.
func (s *RedisStore) GC() error {
	return nil
}
