package cache

import (
	"bytes"
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// The NewRedisStandalone creates a Redis Provider that uses a standalone redis or redis compliant server.
func NewRedisStandalone(config *schema.RedisCache, rootCAs *x509.CertPool) *Redis {
	options := &redis.Options{
		Network:               config.Address.Network(),
		Addr:                  config.Address.NetworkAddress(),
		ClientName:            fmt.Sprintf(driverParameterFmtAppName, utils.Version()),
		Protocol:              3,
		Username:              config.Username,
		Password:              config.Password,
		DB:                    config.Database,
		MaxRetries:            config.MaximumRetries,
		MinRetryBackoff:       config.MinimumRetryBackoff,
		MaxRetryBackoff:       config.MaximumRetryBackoff,
		DialTimeout:           config.DialTimeout,
		DialerRetries:         0,
		DialerRetryTimeout:    0,
		ReadTimeout:           config.ReadTimeout,
		WriteTimeout:          config.WriteTimeout,
		ContextTimeoutEnabled: false,
		ReadBufferSize:        0,
		WriteBufferSize:       0,
		PoolFIFO:              false,
		PoolSize:              config.PoolSize,
		MaxConcurrentDials:    0,
		PoolTimeout:           config.PoolTimeout,
		MinIdleConns:          config.PoolMinimumIdleConnections,
		MaxIdleConns:          config.PoolMaximumIdleConnections,
		MaxActiveConns:        config.PoolMaximumConnections,
		ConnMaxIdleTime:       config.IdleTimeout,
		ConnMaxLifetime:       config.ConnectionTimeout,
		ConnMaxLifetimeJitter: 0,
		TLSConfig:             utils.NewTLSConfig(config.TLS, rootCAs),
		FailingTimeoutSeconds: 0,
	}

	return NewRedis(redis.NewClient(options), "standalone")
}

// The NewRedisSentinel creates a Redis Provider that uses a redis or redis compliant server which supports sentinel.
func NewRedisSentinel(config *schema.RedisSentinelCache, rootCAs *x509.CertPool) *Redis {
	addresses := make([]string, len(config.Addresses))

	for i, address := range config.Addresses {
		addresses[i] = address.NetworkAddress()
	}

	options := &redis.FailoverOptions{
		MasterName:            config.MasterName,
		SentinelAddrs:         addresses,
		ClientName:            fmt.Sprintf(driverParameterFmtAppName, utils.Version()),
		SentinelUsername:      config.SentinelUsername,
		SentinelPassword:      config.SentinelPassword,
		RouteByLatency:        config.RouteByLatency,
		RouteRandomly:         config.RouteRandomly,
		Protocol:              3,
		Username:              config.Username,
		Password:              config.Password,
		DB:                    config.Database,
		MaxRetries:            config.MaximumRetries,
		MinRetryBackoff:       config.MinimumRetryBackoff,
		MaxRetryBackoff:       config.MaximumRetryBackoff,
		DialTimeout:           config.DialTimeout,
		DialerRetries:         0,
		DialerRetryTimeout:    0,
		ReadTimeout:           config.ReadTimeout,
		WriteTimeout:          config.WriteTimeout,
		ContextTimeoutEnabled: false,
		ReadBufferSize:        0,
		WriteBufferSize:       0,
		PoolFIFO:              false,
		PoolSize:              config.PoolSize,
		MaxConcurrentDials:    0,
		PoolTimeout:           config.PoolTimeout,
		MinIdleConns:          config.PoolMinimumIdleConnections,
		MaxIdleConns:          config.PoolMaximumIdleConnections,
		MaxActiveConns:        config.PoolMaximumConnections,
		ConnMaxIdleTime:       config.IdleTimeout,
		ConnMaxLifetime:       config.ConnectionTimeout,
		ConnMaxLifetimeJitter: 0,
		TLSConfig:             utils.NewTLSConfig(config.TLS, rootCAs),
		FailingTimeoutSeconds: 0,
	}

	var client redis.Cmdable

	switch config.SentinelMode {
	case "cluster":
		client = redis.NewFailoverClusterClient(options)
	default:
		client = redis.NewFailoverClient(options)
	}

	return NewRedis(client, "sentinel")
}

// The NewRedisCluster creates a Redis Provider that uses a redis or redis compliant server which supports clustering.
func NewRedisCluster(config *schema.RedisClusterCache, rootCAs *x509.CertPool) (r *Redis) {
	addresses := make([]string, len(config.Addresses))

	for i, address := range config.Addresses {
		addresses[i] = address.NetworkAddress()
	}

	options := &redis.ClusterOptions{
		Addrs:                      addresses,
		ClientName:                 fmt.Sprintf(driverParameterFmtAppName, utils.Version()),
		MaxRedirects:               config.MaximumRedirects,
		ReadOnly:                   config.RouteByReplica,
		RouteByLatency:             config.RouteByLatency,
		RouteRandomly:              config.RouteRandomly,
		Protocol:                   3,
		Username:                   config.Username,
		Password:                   config.Password,
		MaxRetries:                 config.MaximumRetries,
		MinRetryBackoff:            config.MinimumRetryBackoff,
		MaxRetryBackoff:            config.MaximumRetryBackoff,
		DialTimeout:                config.DialTimeout,
		DialerRetries:              0,
		DialerRetryTimeout:         0,
		ReadTimeout:                config.ReadTimeout,
		WriteTimeout:               config.WriteTimeout,
		ContextTimeoutEnabled:      false,
		MaxConcurrentDials:         0,
		PoolFIFO:                   false,
		PoolSize:                   config.PoolSize,
		PoolTimeout:                config.PoolTimeout,
		MinIdleConns:               config.PoolMinimumIdleConnections,
		MaxIdleConns:               config.PoolMaximumIdleConnections,
		MaxActiveConns:             config.PoolMaximumConnections,
		ConnMaxIdleTime:            config.IdleTimeout,
		ConnMaxLifetime:            config.ConnectionTimeout,
		ConnMaxLifetimeJitter:      0,
		ReadBufferSize:             0,
		WriteBufferSize:            0,
		TLSConfig:                  utils.NewTLSConfig(config.TLS, rootCAs),
		DisableRoutingPolicies:     false,
		FailingTimeoutSeconds:      0,
		MaintNotificationsConfig:   nil,
		ShardPicker:                nil,
		ClusterStateReloadInterval: 0,
	}

	return NewRedis(redis.NewClusterClient(options), "cluster")
}

func NewRedis(client redis.Cmdable, variant string) *Redis {
	return &Redis{
		client:  client,
		variant: variant,
	}
}

type Redis struct {
	client  redis.Cmdable
	variant string
}

func (r *Redis) StartupCheck() (err error) {
	return r.client.Ping(context.Background()).Err()
}

func (r *Redis) SessionGet(ctx context.Context, issuer, id string) (data []byte, err error) {
	if data, err = r.client.Get(ctx, getSessionKey(issuer, id)).Bytes(); err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	return data, nil
}

func (r *Redis) SessionGetByPublicID(ctx context.Context, issuer, pid string) (data []byte, err error) {
	var id string

	if id, err = r.client.Get(ctx, getSessionPublicKey(issuer, pid)).Result(); err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}

		return nil, err
	}

	return r.SessionGet(ctx, issuer, id)
}

// SessionGetIDsByUsername returns the signatures of every unexpired session belonging to the given username and issuer.
// Expired members are dropped from the index before it is read, as a sorted set has no per member expiry of its own.
func (r *Redis) SessionGetIDsByUsername(ctx context.Context, issuer, username string) (ids []string, err error) {
	if username == "" {
		return nil, nil
	}

	key := getSessionUserKey(issuer, username)

	pipe := r.client.TxPipeline()

	pipe.ZRemRangeByScore(ctx, key, redisScoreMinimum, getSessionScoreNow())

	zrange := pipe.ZRange(ctx, key, 0, -1)

	if _, err = pipe.Exec(ctx); err != nil {
		return nil, err
	}

	return zrange.Result()
}

func (r *Redis) SessionSave(ctx context.Context, issuer, id, pid, username string, expiration time.Duration, data []byte) (err error) {
	pipe := r.client.TxPipeline()

	var zadd *redis.IntCmd

	set := pipe.Set(ctx, getSessionKey(issuer, id), data, expiration)
	setpub := pipe.Set(ctx, getSessionPublicKey(issuer, pid), id, expiration)

	if username != "" {
		zadd = pipe.ZAdd(ctx, getSessionUserKey(issuer, username), redis.Z{Score: getSessionScore(expiration), Member: id})
	}

	if _, err = pipe.Exec(ctx); err != nil {
		return
	}

	if err = set.Err(); err != nil {
		return err
	}

	if err = setpub.Err(); err != nil {
		return err
	}

	if username != "" {
		if err = zadd.Err(); err != nil {
			return err
		}
	}

	return nil
}

// SessionSetUsername links an existing session to a specific user. The expiry recorded in the index is derived from the
// remaining TTL of the session itself, as the caller doesn't supply one.
func (r *Redis) SessionSetUsername(ctx context.Context, issuer, id, username string) (err error) {
	if username == "" {
		return nil
	}

	var ttl time.Duration

	if ttl, err = r.client.TTL(ctx, getSessionKey(issuer, id)).Result(); err != nil {
		return err
	}

	switch ttl {
	case redisTTLNoKey:
		// The session no longer exists, so there is nothing to index.
		return nil
	case redisTTLNoExpiry:
		ttl = 0
	}

	return r.client.ZAdd(ctx, getSessionUserKey(issuer, username), redis.Z{Score: getSessionScore(ttl), Member: id}).Err()
}

// SessionSaveData updates the session data. Every key which refers to the session has its expiry refreshed alongside it,
// as each is a distinct key with an independent TTL which would otherwise lapse while the session is still alive.
func (r *Redis) SessionSaveData(ctx context.Context, issuer, id, pid, username string, expiration time.Duration, data []byte) (err error) {
	return r.SessionSave(ctx, issuer, id, pid, username, expiration, data)
}

func (r *Redis) SessionDelete(ctx context.Context, issuer, id, pid, username string) (err error) {
	pipe := r.client.TxPipeline()

	var zrem *redis.IntCmd

	del := pipe.Del(ctx, getSessionKey(issuer, id))
	delpub := pipe.Del(ctx, getSessionPublicKey(issuer, pid))

	if username != "" {
		zrem = pipe.ZRem(ctx, getSessionUserKey(issuer, username), id)
	}

	if _, err = pipe.Exec(ctx); err != nil {
		return err
	}

	if err = del.Err(); err != nil {
		return err
	}

	if err = delpub.Err(); err != nil {
		return err
	}

	if username != "" {
		if err = zrem.Err(); err != nil {
			return err
		}
	}

	return nil
}

func (r *Redis) SessionChangeID(ctx context.Context, issuer, oldID, id, pid, username string, expiration time.Duration) (err error) {
	oldKey := getSessionKey(issuer, oldID)

	exists, err := r.client.Exists(ctx, oldKey).Result()
	if err != nil {
		return err
	}

	if exists > 0 {
		key := getSessionKey(issuer, id)
		userkey := getSessionUserKey(issuer, username)

		pipe := r.client.TxPipeline()

		var (
			zrem *redis.IntCmd
			zadd *redis.IntCmd
		)

		rename := pipe.Rename(ctx, oldKey, key)
		expire := pipe.Expire(ctx, key, expiration)

		set := pipe.Set(ctx, getSessionPublicKey(issuer, pid), id, expiration)

		if username != "" {
			zrem = pipe.ZRem(ctx, userkey, oldID)
			zadd = pipe.ZAdd(ctx, userkey, redis.Z{Score: getSessionScore(expiration), Member: id})
		}

		if _, err = pipe.Exec(ctx); err != nil {
			return err
		}

		if err = rename.Err(); err != nil {
			return err
		}

		if err = expire.Err(); err != nil {
			return err
		}

		if err = set.Err(); err != nil {
			return err
		}

		if username != "" {
			if err = zrem.Err(); err != nil {
				return err
			}

			if err = zadd.Err(); err != nil {
				return err
			}
		}
	}

	return nil
}

// SessionGarbageCollectionFrequency returns the frequency the username indexes are pruned at. The session and public id
// keys expire themselves, but a sorted set has no per member expiry so its expired members are only removed when the
// index is read or collected.
func (r *Redis) SessionGarbageCollectionFrequency(ctx context.Context) (frequency time.Duration) {
	return sessionGarbageCollectionFrequency
}

// SessionGarbageCollection removes expired members from every username index.
func (r *Redis) SessionGarbageCollection(ctx context.Context) (err error) {
	score := getSessionScoreNow()

	// A cluster client scans a single node, so each master is visited individually to cover the whole keyspace.
	if cluster, ok := r.client.(*redis.ClusterClient); ok {
		return cluster.ForEachMaster(ctx, func(ctx context.Context, client *redis.Client) error {
			return r.sessionGarbageCollection(ctx, client, score)
		})
	}

	return r.sessionGarbageCollection(ctx, r.client, score)
}

func (r *Redis) sessionGarbageCollection(ctx context.Context, client redis.Cmdable, score string) (err error) {
	iter := client.Scan(ctx, 0, redisPrefixSessionUser+"*", redisScanCount).Iterator()

	for iter.Next(ctx) {
		if err = client.ZRemRangeByScore(ctx, iter.Val(), redisScoreMinimum, score).Err(); err != nil {
			return fmt.Errorf("error removing expired sessions from the user index '%s': %w", iter.Val(), err)
		}
	}

	return iter.Err()
}

// getSessionScore converts a relative expiration into the sorted set score the session is indexed under, which is the unix
// time it expires at. A non-positive expiration never expires and is scored so that it is never pruned.
func getSessionScore(expiration time.Duration) (score float64) {
	if expiration <= 0 {
		return math.Inf(1)
	}

	return float64(time.Now().Add(expiration).Unix())
}

// getSessionScoreNow returns the current time as the upper bound of a ZRANGEBYSCORE style range, which is inclusive.
func getSessionScoreNow() (score string) {
	return strconv.FormatInt(time.Now().Unix(), 10)
}

func getSessionUserKey(issuer, username string) (key string) {
	buf := bytes.NewBuffer(nil)

	buf.WriteString(redisPrefixSessionUser)
	buf.WriteString(issuer)
	buf.WriteString(redisKeySeparatorSlot)
	buf.WriteString(username)

	return buf.String()
}

func getSessionPublicKey(issuer, pid string) (key string) {
	buf := bytes.NewBuffer(nil)

	buf.WriteString(redisPrefixSessionPublic)
	buf.WriteString(issuer)
	buf.WriteString(redisKeySeparatorSlot)
	buf.WriteString(pid)

	return buf.String()
}

func getSessionKey(issuer, id string) (key string) {
	buf := bytes.NewBuffer(nil)

	buf.WriteString(redisPrefixSession)
	buf.WriteString(issuer)
	buf.WriteString(redisKeySeparatorSlot)
	buf.WriteString(id)

	return buf.String()
}
