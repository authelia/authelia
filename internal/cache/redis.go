package cache

import (
	"bytes"
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

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

func NewRedisCluster(config *schema.RedisClusterCache, rootCAs *x509.CertPool) (r *Redis) {
	addreses := make([]string, len(config.Addresses))

	for i, address := range config.Addresses {
		addreses[i] = address.NetworkAddress()
	}

	options := &redis.ClusterOptions{
		Addrs:                      addreses,
		ClientName:                 fmt.Sprintf(driverParameterFmtAppName, utils.Version()),
		MaxRedirects:               config.MaximumRedirects,
		ReadOnly:                   config.RouteByReplica,
		RouteByLatency:             config.RouteByLatency,
		RouteRandomly:              config.RouteRandomly,
		Protocol:                   3,
		Username:                   config.Username,
		Password:                   config.Password,
		MaxRetries:                 config.MaximumRedirects,
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

func NewRedisSentinel(config *schema.RedisSentinelCache, rootCAs *x509.CertPool) *Redis {
	addreses := make([]string, len(config.Addresses))

	for i, address := range config.Addresses {
		addreses[i] = address.NetworkAddress()
	}

	options := &redis.FailoverOptions{
		MasterName:            config.MasterName,
		SentinelAddrs:         addreses,
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

	if config.RouteByReplica {
		client = redis.NewFailoverClusterClient(options)
	} else {
		client = redis.NewFailoverClient(options)
	}

	return NewRedis(client, "sentinel")
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

func (r *Redis) SessionGet(ctx context.Context, id, issuer string) (data []byte, err error) {
	if data, err = r.client.Get(ctx, getSessionKey(id, issuer)).Bytes(); err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	return data, nil
}

func (r *Redis) SessionGetByPublicID(ctx context.Context, pid, issuer string) (data []byte, err error) {
	var id string

	if id, err = r.client.Get(ctx, getPublicKey(pid, issuer)).Result(); err != nil {
		return nil, err
	}

	return r.SessionGet(ctx, id, issuer)
}

func (r *Redis) SessionGetIDsByUsername(ctx context.Context, username, issuer string) (ids []string, err error) {
	return r.client.SMembers(ctx, getUserKey(username, issuer)).Result()
}

func (r *Redis) SessionSave(ctx context.Context, id, pubid, username, issuer string, expiration time.Duration, data []byte) (err error) {
	pipe := r.client.TxPipeline()

	var sadd *redis.IntCmd

	set := pipe.Set(ctx, getSessionKey(id, issuer), data, expiration)
	setpub := pipe.Set(ctx, getPublicKey(pubid, issuer), id, expiration)

	if username != "" {
		sadd = pipe.SAdd(ctx, getUserKey(username, issuer), id)
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

	if username == "" {
		if err = sadd.Err(); err != nil {
			return err
		}
	}

	return nil
}

func (r *Redis) SessionSetUsername(ctx context.Context, id, username, issuer string) (err error) {
	return r.client.SAdd(ctx, getUserKey(username, issuer), id).Err()
}

func (r *Redis) SessionSaveData(ctx context.Context, id, issuer string, expiration time.Duration, data []byte) (err error) {
	return r.client.Set(ctx, getSessionKey(id, issuer), data, expiration).Err()
}

func (r *Redis) SessionDelete(ctx context.Context, id, pubid, username, issuer string) (err error) {
	pipe := r.client.TxPipeline()

	del := pipe.Del(ctx, getSessionKey(id, issuer))
	delpub := pipe.Del(ctx, getPublicKey(pubid, issuer))
	srem := pipe.SRem(ctx, getUserKey(username, issuer))

	if _, err = pipe.Exec(ctx); err != nil {
		return err
	}

	if err = del.Err(); err != nil {
		return err
	}

	if err = delpub.Err(); err != nil {
		return err
	}

	if err = srem.Err(); err != nil {
		return err
	}

	return nil
}

func (r *Redis) SessionChangeID(ctx context.Context, oldID, id, pubid, username, issuer string, expiration time.Duration) (err error) {
	oldKey := getSessionKey(oldID, issuer)

	exists, err := r.client.Exists(ctx, oldKey).Result()
	if err != nil {
		return err
	}

	if exists > 0 {
		key := getSessionKey(id, issuer)
		userkey := getUserKey(username, issuer)

		pipe := r.client.TxPipeline()

		rename := pipe.Rename(ctx, oldKey, key)
		expire := pipe.Expire(ctx, key, expiration)
		set := pipe.Set(ctx, getPublicKey(pubid, issuer), id, expiration)
		srem := pipe.SRem(ctx, userkey, oldID)
		sadd := pipe.SAdd(ctx, userkey, id)

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

		if err = srem.Err(); err != nil {
			return err
		}

		if err = sadd.Err(); err != nil {
			return err
		}
	}

	return nil
}

func (r *Redis) SessionGarbageCollection(ctx context.Context) (err error) {
	return nil
}

func (r *Redis) SessionGarbageCollectionRequired(ctx context.Context) (required bool) {
	return false
}

func getUserKey(username, issuer string) (key string) {
	buf := bytes.NewBuffer(nil)

	buf.WriteString("authelia:")
	buf.WriteString("user-session:")
	buf.WriteString(issuer)
	buf.WriteString(":")
	buf.WriteString(username)

	return buf.String()
}
func getPublicKey(pid, issuer string) (key string) {
	buf := bytes.NewBuffer(nil)

	buf.WriteString("authelia:")
	buf.WriteString("public-session:")
	buf.WriteString(issuer)
	buf.WriteString(":")
	buf.WriteString(pid)

	return buf.String()
}

func getSessionKey(id, issuer string) (key string) {
	buf := bytes.NewBuffer(nil)

	buf.WriteString("authelia:")
	buf.WriteString("session:")
	buf.WriteString(issuer)
	buf.WriteString(":")
	buf.WriteString(id)

	return buf.String()
}
