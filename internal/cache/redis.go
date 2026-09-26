// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"bytes"
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/utils"
)

// The NewRedisStandalone creates a Redis Provider that uses a standalone redis or redis compliant server.
func NewRedisStandalone(config *schema.RedisCache, rootCAs *x509.CertPool) *Redis {
	options := &redis.Options{
		Network:               config.Address.Network(),
		Addr:                  config.Address.NetworkAddress(),
		ClientName:            getClientName(),
		Protocol:              3,
		Username:              config.Username,
		Password:              config.Password,
		DB:                    config.Database,
		MaxRetries:            config.MaximumRetries,
		MinRetryBackoff:       config.MinimumRetryBackoff,
		MaxRetryBackoff:       config.MaximumRetryBackoff,
		DialTimeout:           config.DialTimeout,
		DialerRetries:         config.DialerRetries,
		DialerRetryTimeout:    config.DialerRetryTimeout,
		ReadTimeout:           config.ReadTimeout,
		WriteTimeout:          config.WriteTimeout,
		ContextTimeoutEnabled: config.ContextTimeoutEnabled,
		ReadBufferSize:        config.ReadBufferSize,
		WriteBufferSize:       config.WriteBufferSize,
		PoolFIFO:              config.PoolFIFO,
		PoolSize:              config.PoolSize,
		MaxConcurrentDials:    config.MaximumConcurrentDials,
		PoolTimeout:           config.PoolTimeout,
		MinIdleConns:          config.PoolMinimumIdleConnections,
		MaxIdleConns:          config.PoolMaximumIdleConnections,
		MaxActiveConns:        config.PoolMaximumConnections,
		ConnMaxIdleTime:       config.IdleTimeout,
		ConnMaxLifetime:       config.ConnectionLifetime,
		ConnMaxLifetimeJitter: config.ConnectionLifetimeJitter,
		TLSConfig:             utils.NewTLSConfig(config.TLS, rootCAs),
		FailingTimeoutSeconds: getFailingTimeoutSeconds(config.FailingTimeout),
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
		MasterName:              config.MasterName,
		SentinelAddrs:           addresses,
		ClientName:              getClientName(),
		SentinelUsername:        config.SentinelUsername,
		SentinelPassword:        config.SentinelPassword,
		RouteByLatency:          config.RouteByLatency,
		RouteRandomly:           config.RouteRandomly,
		ReplicaOnly:             config.ReplicaOnly,
		UseDisconnectedReplicas: config.UseDisconnectedReplicas,
		Protocol:                3,
		Username:                config.Username,
		Password:                config.Password,
		DB:                      config.Database,
		MaxRetries:              config.MaximumRetries,
		MinRetryBackoff:         config.MinimumRetryBackoff,
		MaxRetryBackoff:         config.MaximumRetryBackoff,
		DialTimeout:             config.DialTimeout,
		DialerRetries:           config.DialerRetries,
		DialerRetryTimeout:      config.DialerRetryTimeout,
		ReadTimeout:             config.ReadTimeout,
		WriteTimeout:            config.WriteTimeout,
		ContextTimeoutEnabled:   config.ContextTimeoutEnabled,
		ReadBufferSize:          config.ReadBufferSize,
		WriteBufferSize:         config.WriteBufferSize,
		PoolFIFO:                config.PoolFIFO,
		PoolSize:                config.PoolSize,
		MaxConcurrentDials:      config.MaximumConcurrentDials,
		PoolTimeout:             config.PoolTimeout,
		MinIdleConns:            config.PoolMinimumIdleConnections,
		MaxIdleConns:            config.PoolMaximumIdleConnections,
		MaxActiveConns:          config.PoolMaximumConnections,
		ConnMaxIdleTime:         config.IdleTimeout,
		ConnMaxLifetime:         config.ConnectionLifetime,
		ConnMaxLifetimeJitter:   config.ConnectionLifetimeJitter,
		TLSConfig:               utils.NewTLSConfig(config.TLS, rootCAs),
		FailingTimeoutSeconds:   getFailingTimeoutSeconds(config.FailingTimeout),
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
		ClientName:                 getClientName(),
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
		DialerRetries:              config.DialerRetries,
		DialerRetryTimeout:         config.DialerRetryTimeout,
		ReadTimeout:                config.ReadTimeout,
		WriteTimeout:               config.WriteTimeout,
		ContextTimeoutEnabled:      config.ContextTimeoutEnabled,
		MaxConcurrentDials:         config.MaximumConcurrentDials,
		PoolFIFO:                   config.PoolFIFO,
		PoolSize:                   config.PoolSize,
		PoolTimeout:                config.PoolTimeout,
		MinIdleConns:               config.PoolMinimumIdleConnections,
		MaxIdleConns:               config.PoolMaximumIdleConnections,
		MaxActiveConns:             config.PoolMaximumConnections,
		ConnMaxIdleTime:            config.IdleTimeout,
		ConnMaxLifetime:            config.ConnectionLifetime,
		ConnMaxLifetimeJitter:      config.ConnectionLifetimeJitter,
		ReadBufferSize:             config.ReadBufferSize,
		WriteBufferSize:            config.WriteBufferSize,
		TLSConfig:                  utils.NewTLSConfig(config.TLS, rootCAs),
		DisableRoutingPolicies:     false,
		FailingTimeoutSeconds:      getFailingTimeoutSeconds(config.FailingTimeout),
		ClusterStateReloadInterval: config.ClusterStateReloadInterval,
	}

	return NewRedis(redis.NewClusterClient(options), "cluster")
}

// NewRedis returns a new Redis Provider for the given client, where the variant records which of the standalone,
// sentinel, or cluster deployments it was built for.
func NewRedis(client redis.Cmdable, variant string) *Redis {
	return &Redis{
		client:  client,
		variant: variant,
	}
}

// Redis is a Provider which stores sessions in Redis.
type Redis struct {
	client  redis.Cmdable
	variant string
}

// Variant returns which of the standalone, sentinel, or cluster deployments the Redis Provider was built for.
func (r *Redis) Variant() string {
	return r.variant
}

// StartupCheck implements the Provider interface, pinging the server to confirm it is reachable.
func (r *Redis) StartupCheck() (err error) {
	return r.client.Ping(context.Background()).Err()
}

// SessionGet implements the Provider interface.
func (r *Redis) SessionGet(ctx context.Context, issuer, id string) (record session.Record, err error) {
	var data []byte

	if data, err = r.client.HGet(ctx, getSessionKey(issuer, id), redisSessionFieldData).Bytes(); err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	if len(data) == 0 {
		return nil, nil
	}

	return session.NewRecord(id, data), nil
}

// SessionGetByPublicID implements the Provider interface. The public id index is maintained separately from the session
// it refers to, so it's only trusted when that session still records the public id.
func (r *Redis) SessionGetByPublicID(ctx context.Context, issuer, pid string) (record session.Record, err error) {
	var id string

	if id, err = r.client.Get(ctx, getSessionPublicKey(issuer, pid)).Result(); err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}

		return nil, err
	}

	var values []any

	if values, err = r.client.HMGet(ctx, getSessionKey(issuer, id), redisSessionFieldData, redisSessionFieldPID).Result(); err != nil {
		return nil, err
	}

	data, recorded := getRedisString(values, 0), getRedisString(values, 1)

	if len(data) == 0 || recorded != pid {
		return nil, nil
	}

	return session.NewRecord(id, []byte(data)), nil
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

// SessionSave implements the Provider interface. The session is saved by a script so that the save is discarded
// atomically when the session was moved to a new id after the caller retrieved it, as saving it would restore the
// session under the id it was moved from. The indexes which refer to the session are updated once it's saved.
func (r *Redis) SessionSave(ctx context.Context, issuer, id, pid, username string, expiration time.Duration, data []byte) (err error) {
	var previousPID, previousUsername string

	if previousPID, previousUsername, err = r.sessionSave(ctx, issuer, id, pid, username, expiration, data); err != nil {
		return err
	}

	if err = r.sessionIndex(ctx, issuer, id, pid, username, expiration); err != nil {
		return err
	}

	if previousPID != "" && previousPID != pid {
		if err = r.sessionRetirePublicID(ctx, issuer, id, previousPID); err != nil {
			return err
		}
	}

	if previousUsername != "" && previousUsername != username {
		if err = r.sessionRetireUsername(ctx, issuer, id, previousUsername); err != nil {
			return err
		}
	}

	return nil
}

// SessionSaveData updates the session data. Every key which refers to the session has its expiry refreshed alongside it,
// as each is a distinct key with an independent TTL which would otherwise lapse while the session is still alive.
func (r *Redis) SessionSaveData(ctx context.Context, issuer, id, pid, username string, expiration time.Duration, data []byte) (err error) {
	return r.SessionSave(ctx, issuer, id, pid, username, expiration, data)
}

// SessionDelete implements the Provider interface. The indexes named by the caller are retired, falling back to those the
// session recorded when the caller doesn't name them.
func (r *Redis) SessionDelete(ctx context.Context, issuer, id, pid, username string) (err error) {
	var (
		result                        []any
		previousPID, previousUsername string
	)

	if result, err = redisSessionDelete.Run(ctx, r.client, []string{getSessionKey(issuer, id)}).Slice(); err != nil {
		return err
	}

	previousPID, previousUsername = getRedisString(result, 0), getRedisString(result, 1)

	if pid == "" {
		pid = previousPID
	}

	if username == "" {
		username = previousUsername
	}

	if pid != "" {
		if err = r.sessionRetirePublicID(ctx, issuer, id, pid); err != nil {
			return err
		}
	}

	if username != "" {
		if err = r.sessionRetireUsername(ctx, issuer, id, username); err != nil {
			return err
		}
	}

	return nil
}

// SessionChangeID moves a session to a new id. The session at the old id is atomically replaced with a marker recording
// the new id, which discards any later save of the old id, before the session is written to the new id. The data is
// written rather than the old key being renamed, as the caller reseals the session against the id it is stored under,
// and the old and new keys are in different slots of a cluster. A session which no longer exists is not recreated.
func (r *Redis) SessionChangeID(ctx context.Context, issuer, oldID, id, pid, username string, expiration time.Duration, data []byte) (err error) {
	var (
		result                        []any
		previousPID, previousUsername string
	)

	if result, err = redisSessionMove.Run(ctx, r.client, []string{getSessionKey(issuer, oldID)}, id, getSessionExpirationMilliseconds(expiration)).Slice(); err != nil {
		return err
	}

	if getRedisInt(result, 0) == 0 {
		return nil
	}

	previousPID, previousUsername = getRedisString(result, 1), getRedisString(result, 2)

	if _, _, err = r.sessionSave(ctx, issuer, id, pid, username, expiration, data); err != nil {
		return err
	}

	if err = r.sessionIndex(ctx, issuer, id, pid, username, expiration); err != nil {
		return err
	}

	// The public id index of the session now refers to the new id, so it's only retired when the session had another.
	if previousPID != "" && previousPID != pid {
		if err = r.sessionRetirePublicID(ctx, issuer, oldID, previousPID); err != nil {
			return err
		}
	}

	if previousUsername != "" {
		if err = r.sessionRetireUsername(ctx, issuer, oldID, previousUsername); err != nil {
			return err
		}
	}

	return nil
}

func (r *Redis) sessionSave(ctx context.Context, issuer, id, pid, username string, expiration time.Duration, data []byte) (previousPID, previousUsername string, err error) {
	var result []any

	if result, err = redisSessionSave.Run(ctx, r.client, []string{getSessionKey(issuer, id)}, data, getSessionExpirationMilliseconds(expiration), pid, username).Slice(); err != nil {
		return "", "", err
	}

	if getRedisInt(result, 0) == 0 {
		return "", "", session.ErrSessionSuperseded
	}

	return getRedisString(result, 1), getRedisString(result, 2), nil
}

func (r *Redis) sessionIndex(ctx context.Context, issuer, id, pid, username string, expiration time.Duration) (err error) {
	if pid == "" && username == "" {
		return nil
	}

	if _, err = r.client.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		if pid != "" {
			pipe.Set(ctx, getSessionPublicKey(issuer, pid), id, getSessionIndexExpiration(expiration))
		}

		if username != "" {
			pipe.ZAdd(ctx, getSessionUserKey(issuer, username), redis.Z{Score: getSessionScore(expiration), Member: id})
		}

		return nil
	}); err != nil {
		return fmt.Errorf("error updating the session indexes: %w", err)
	}

	return nil
}

func (r *Redis) sessionRetirePublicID(ctx context.Context, issuer, id, pid string) (err error) {
	if err = redisDeleteIfEqual.Run(ctx, r.client, []string{getSessionPublicKey(issuer, pid)}, id).Err(); err != nil {
		return fmt.Errorf("error removing the session public id index: %w", err)
	}

	return nil
}

func (r *Redis) sessionRetireUsername(ctx context.Context, issuer, id, username string) (err error) {
	if err = r.client.ZRem(ctx, getSessionUserKey(issuer, username), id).Err(); err != nil {
		return fmt.Errorf("error removing the session from the username index: %w", err)
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

func getSessionExpirationMilliseconds(expiration time.Duration) (milliseconds int64) {
	if expiration <= 0 {
		return 0
	}

	if milliseconds = expiration.Milliseconds(); milliseconds < 1 {
		return 1
	}

	return milliseconds
}

func getSessionScore(expiration time.Duration) (score float64) {
	if expiration <= 0 {
		return math.Inf(1)
	}

	return float64(time.Now().Add(expiration).Unix())
}

func getSessionScoreNow() (score string) {
	return strconv.FormatInt(time.Now().Unix(), 10)
}

func getSessionIndexExpiration(expiration time.Duration) time.Duration {
	if expiration <= 0 {
		return 0
	}

	return time.Duration(getSessionExpirationMilliseconds(expiration)) * time.Millisecond
}

func getSessionUserKey(issuer, username string) (key string) {
	return getSessionSlotKey(redisPrefixSessionUser, issuer, username)
}

func getSessionPublicKey(issuer, pid string) (key string) {
	return getSessionSlotKey(redisPrefixSessionPublic, issuer, pid)
}

func getSessionKey(issuer, id string) (key string) {
	return getSessionSlotKey(redisPrefixSession, issuer, id)
}

func getSessionSlotKey(prefix, issuer, value string) (key string) {
	buf := bytes.NewBuffer(nil)

	buf.WriteString(prefix)
	buf.WriteString(issuer)
	buf.WriteString(redisKeySeparator)
	buf.WriteString(value)
	buf.WriteString(redisKeySeparatorSlot)

	return buf.String()
}

func getRedisString(values []any, i int) string {
	if i >= len(values) {
		return ""
	}

	value, _ := values[i].(string)

	return value
}

func getRedisInt(values []any, i int) int64 {
	if i >= len(values) {
		return 0
	}

	value, _ := values[i].(int64)

	return value
}

func getClientName() (name string) {
	return fmt.Sprintf(driverParameterFmtAppName, strings.Split(utils.Version(), " ")[0])
}

func getFailingTimeoutSeconds(timeout time.Duration) (seconds int) {
	if timeout <= 0 {
		return 0
	}

	if seconds = int(timeout.Seconds()); seconds == 0 {
		return 1
	}

	return seconds
}
