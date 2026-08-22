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
		ConnMaxLifetime:       config.ConnectionTimeout,
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
		ConnMaxLifetime:         config.ConnectionTimeout,
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
		ConnMaxLifetime:            config.ConnectionTimeout,
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

// getFailingTimeoutSeconds converts the configured duration into the whole seconds the client expects. A duration
// below a second which isn't zero is raised to one second so a deliberate value is never silently discarded.
func getFailingTimeoutSeconds(timeout time.Duration) (seconds int) {
	if timeout <= 0 {
		return 0
	}

	if seconds = int(timeout.Seconds()); seconds == 0 {
		return 1
	}

	return seconds
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

// StartupCheck implements the Provider interface, pinging the server to confirm it is reachable.
func (r *Redis) StartupCheck() (err error) {
	return r.client.Ping(context.Background()).Err()
}

// SessionGet implements the Provider interface.
func (r *Redis) SessionGet(ctx context.Context, issuer, id string) (record session.Record, err error) {
	var data []byte

	if data, err = r.client.Get(ctx, getSessionKey(issuer, id)).Bytes(); err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	if len(data) == 0 {
		return nil, nil
	}

	return session.NewRecord(id, data), nil
}

// SessionGetByPublicID implements the Provider interface.
func (r *Redis) SessionGetByPublicID(ctx context.Context, issuer, pid string) (record session.Record, err error) {
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

// SessionSave implements the Provider interface.
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

// SessionSaveData updates the session data. Every key which refers to the session has its expiry refreshed alongside it,
// as each is a distinct key with an independent TTL which would otherwise lapse while the session is still alive.
func (r *Redis) SessionSaveData(ctx context.Context, issuer, id, pid, username string, expiration time.Duration, data []byte) (err error) {
	return r.SessionSave(ctx, issuer, id, pid, username, expiration, data)
}

// SessionDelete implements the Provider interface.
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

// SessionChangeID moves a session to a new id. The data is written rather than the old key being renamed, as the caller
// reseals the session against the id it is stored under and the two must be updated together. A session which no longer
// exists is not recreated.
func (r *Redis) SessionChangeID(ctx context.Context, issuer, oldID, id, pid, username string, expiration time.Duration, data []byte) (err error) {
	keys := []string{getSessionKey(issuer, oldID), getSessionKey(issuer, id), getSessionPublicKey(issuer, pid)}
	args := []any{data, getSessionExpirationMilliseconds(expiration), id}

	if username != "" {
		keys = append(keys, getSessionUserKey(issuer, username))
		args = append(args, oldID, getSessionScore(expiration), id)
	}

	if err = sessionChangeIDScript.Run(ctx, r.client, keys, args...).Err(); err != nil {
		return err
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

var sessionChangeIDScript = redis.NewScript(`
local ttl = tonumber(ARGV[2])

if redis.call('exists', KEYS[1]) == 0 then
	return 0
end

redis.call('del', KEYS[1])

if ttl > 0 then
	redis.call('set', KEYS[2], ARGV[1], 'px', ttl)
	redis.call('set', KEYS[3], ARGV[3], 'px', ttl)
else
	redis.call('set', KEYS[2], ARGV[1])
	redis.call('set', KEYS[3], ARGV[3])
end

if #KEYS > 3 then
	redis.call('zrem', KEYS[4], ARGV[4])
	redis.call('zadd', KEYS[4], ARGV[5], ARGV[6])
end

return 1
`)

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

func getClientName() (name string) {
	return fmt.Sprintf(driverParameterFmtAppName, strings.Split(utils.Version(), " ")[0])
}
