package schema

import (
	"crypto/tls"
	"time"
)

// Cache represents the cache backend configuration.
type Cache struct {
	Redis         *RedisCache         `koanf:"redis" yaml:"redis,omitempty" toml:"redis,omitempty" json:"redis,omitempty" jsonschema:"title=Redis Cache" jsonschema_description:"Redis Cache Configuration."`
	RedisSentinel *RedisSentinelCache `koanf:"redis_sentinel" yaml:"redis_sentinel,omitempty" toml:"redis_sentinel,omitempty" json:"redis_sentinel,omitempty" jsonschema:"title=Redis Sentinel Cache" jsonschema_description:"Redis Sentinel Cache Configuration."`
	RedisCluster  *RedisClusterCache  `koanf:"redis_cluster" yaml:"redis_cluster,omitempty" toml:"redis_cluster,omitempty" json:"redis_cluster,omitempty" jsonschema:"title=Redis Cluster Cache" jsonschema_description:"Redis Cluster Cache Configuration."`
}

// RedisCache represents the configuration for a standalone Redis cache backend.
type RedisCache struct {
	Address                    *AddressTCP   `koanf:"address" yaml:"address,omitempty" toml:"address,omitempty" json:"address,omitempty" jsonschema:"title=Address" jsonschema_description:"The address for the Redis server to connect to."`
	Database                   int           `koanf:"database" yaml:"database,omitempty" toml:"database,omitempty" json:"database,omitempty" jsonschema:"title=Database" jsonschema_description:"The database to use for the Redis server."`
	Username                   string        `koanf:"username" yaml:"username,omitempty" toml:"username,omitempty" json:"username,omitempty" jsonschema:"title=Username" jsonschema_description:"The username to use for the Redis server."`
	Password                   string        `koanf:"password" yaml:"password,omitempty" toml:"password,omitempty" json:"password,omitempty" jsonschema:"title=Password" jsonschema_description:"The password to use for the Redis server."`
	TLS                        *TLS          `koanf:"tls" yaml:"tls,omitempty" toml:"tls,omitempty" json:"tls,omitempty" jsonschema:"title=TLS" jsonschema_description:"The TLS configuration for the Redis server."`
	DialTimeout                time.Duration `koanf:"dial_timeout" yaml:"dial_timeout,omitempty" toml:"dial_timeout,omitempty" json:"dial_timeout,omitempty" jsonschema:"default=5 seconds,title=Dial Timeout" jsonschema_description:"The dial timeout for the Redis server."`
	ReadTimeout                time.Duration `koanf:"read_timeout" yaml:"read_timeout,omitempty" toml:"read_timeout,omitempty" json:"read_timeout,omitempty" jsonschema:"default=3 seconds,title=Read Timeout" jsonschema_description:"The read timeout for the Redis server."`
	WriteTimeout               time.Duration `koanf:"write_timeout" yaml:"write_timeout,omitempty" toml:"write_timeout,omitempty" json:"write_timeout,omitempty" jsonschema:"default=3 seconds,title=Write Timeout" jsonschema_description:"The write timeout for the Redis server."`
	ConnectionTimeout          time.Duration `koanf:"connection_timeout" yaml:"connection_timeout,omitempty" toml:"connection_timeout,omitempty" json:"connection_timeout,omitempty" jsonschema:"default=10 seconds,title=Connection Timeout" jsonschema_description:"The connection timeout for the Redis server."`
	IdleTimeout                time.Duration `koanf:"idle_timeout" yaml:"idle_timeout,omitempty" toml:"idle_timeout,omitempty" json:"idle_timeout,omitempty" jsonschema:"default=10 seconds,title=Idle Timeout" jsonschema_description:"The connection idle timeout for the Redis server."`
	PoolTimeout                time.Duration `koanf:"pool_timeout" yaml:"pool_timeout,omitempty" toml:"pool_timeout,omitempty" json:"pool_timeout,omitempty" jsonschema:"default=10 seconds,title=Pool Timeout" jsonschema_description:"The pool timeout for the Redis server."`
	MinimumRetryBackoff        time.Duration `koanf:"minimum_retry_backoff" yaml:"minimum_retry_backoff,omitempty" toml:"minimum_retry_backoff,omitempty" json:"minimum_retry_backoff,omitempty" jsonschema:"default=100 milliseconds,title=Minimum Retry Backoff" jsonschema_description:"The minimum retry backoff to use when connecting to the Redis server."`
	MaximumRetryBackoff        time.Duration `koanf:"maximum_retry_backoff" yaml:"maximum_retry_backoff,omitempty" toml:"maximum_retry_backoff,omitempty" json:"maximum_retry_backoff,omitempty" jsonschema:"default=100 milliseconds,title=Maximum Retry Backoff" jsonschema_description:"The maximum retry backoff to use when connecting to the Redis server."`
	MaximumRetries             int           `koanf:"maximum_retries" yaml:"maximum_retries,omitempty" toml:"maximum_retries,omitempty" json:"maximum_retries,omitempty" jsonschema:"default=3,title=Max Retries" jsonschema_description:"The maximum number of retries to attempt when connecting to the Redis server."`
	PoolSize                   int           `koanf:"pool_size" yaml:"pool_size,omitempty" toml:"pool_size,omitempty" json:"pool_size,omitempty" jsonschema:"default=8,title=Pool Size" jsonschema_description:"The pool size to use when connecting to the Redis server."`
	PoolMinimumIdleConnections int           `koanf:"pool_minimum_idle_connections" yaml:"pool_minimum_idle_connections,omitempty" toml:"pool_minimum_idle_connections,omitempty" json:"pool_minimum_idle_connections,omitempty" jsonschema:"default=10,title=Pool Minimum Idle Connections" jsonschema_description:"The minimum idle connections to keep open with the pool when connecting to the Redis server."`
	PoolMaximumIdleConnections int           `koanf:"pool_maximum_idle_connections" yaml:"pool_maximum_idle_connections,omitempty" toml:"pool_maximum_idle_connections,omitempty" json:"pool_maximum_idle_connections,omitempty" jsonschema:"default=10,title=Pool Maximum Idle Connections" jsonschema_description:"The maximum idle connections to keep open with the pool when connecting to the Redis server."`
	PoolMaximumConnections     int           `koanf:"pool_maximum_connections" yaml:"pool_maximum_connections,omitempty" toml:"pool_maximum_connections,omitempty" json:"pool_maximum_connections,omitempty" jsonschema:"default=10,title=Pool Maximum Connections" jsonschema_description:"The maximum number of connections the pool can have open when connecting to the Redis server."`
	DialerRetries              int           `koanf:"dialer_retries" yaml:"dialer_retries,omitempty" toml:"dialer_retries,omitempty" json:"dialer_retries,omitempty" jsonschema:"default=5,title=Dialer Retries" jsonschema_description:"The maximum number of retry attempts when dialing a connection fails."`
	DialerRetryTimeout         time.Duration `koanf:"dialer_retry_timeout" yaml:"dialer_retry_timeout,omitempty" toml:"dialer_retry_timeout,omitempty" json:"dialer_retry_timeout,omitempty" jsonschema:"default=100 milliseconds,title=Dialer Retry Timeout" jsonschema_description:"The backoff duration between dial retry attempts."`
	MaximumConcurrentDials     int           `koanf:"maximum_concurrent_dials" yaml:"maximum_concurrent_dials,omitempty" toml:"maximum_concurrent_dials,omitempty" json:"maximum_concurrent_dials,omitempty" jsonschema:"title=Maximum Concurrent Dials" jsonschema_description:"The maximum number of concurrent connection creation goroutines. Defaults to the pool size."`
	ConnectionLifetimeJitter   time.Duration `koanf:"connection_lifetime_jitter" yaml:"connection_lifetime_jitter,omitempty" toml:"connection_lifetime_jitter,omitempty" json:"connection_lifetime_jitter,omitempty" jsonschema:"title=Connection Lifetime Jitter" jsonschema_description:"The absolute jitter applied to the connection timeout to prevent all connections expiring simultaneously."`
	ContextTimeoutEnabled      bool          `koanf:"context_timeout_enabled" yaml:"context_timeout_enabled,omitempty" toml:"context_timeout_enabled,omitempty" json:"context_timeout_enabled,omitempty" jsonschema:"default=false,title=Context Timeout Enabled" jsonschema_description:"Enables the client respecting context timeouts and deadlines."`
	PoolFIFO                   bool          `koanf:"pool_fifo" yaml:"pool_fifo,omitempty" toml:"pool_fifo,omitempty" json:"pool_fifo,omitempty" jsonschema:"default=false,title=Pool FIFO" jsonschema_description:"Uses a FIFO connection pool rather than a LIFO connection pool."`
	ReadBufferSize             int           `koanf:"read_buffer_size" yaml:"read_buffer_size,omitempty" toml:"read_buffer_size,omitempty" json:"read_buffer_size,omitempty" jsonschema:"default=32768,title=Read Buffer Size" jsonschema_description:"The size of the read buffer in bytes for each connection."`
	WriteBufferSize            int           `koanf:"write_buffer_size" yaml:"write_buffer_size,omitempty" toml:"write_buffer_size,omitempty" json:"write_buffer_size,omitempty" jsonschema:"default=32768,title=Write Buffer Size" jsonschema_description:"The size of the write buffer in bytes for each connection."`
	FailingTimeout             time.Duration `koanf:"failing_timeout" yaml:"failing_timeout,omitempty" toml:"failing_timeout,omitempty" json:"failing_timeout,omitempty" jsonschema:"default=15 seconds,title=Failing Timeout" jsonschema_description:"The duration a node is avoided for after being marked as failing."`
}

// RedisSentinelCache represents the configuration for a Redis Sentinel cache backend.
type RedisSentinelCache struct {
	MasterName                 string        `koanf:"master_name" yaml:"master_name,omitempty" toml:"master_name,omitempty" json:"master_name,omitempty" jsonschema:"title=Master Name" jsonschema_description:"The name of the sentinel master."`
	SentinelUsername           string        `koanf:"sentinel_username" yaml:"sentinel_username,omitempty" toml:"sentinel_username,omitempty" json:"sentinel_username,omitempty" jsonschema:"title=Sentinel Username" jsonschema_description:"The username for the Redis Sentinel connection."`
	SentinelPassword           string        `koanf:"sentinel_password" yaml:"sentinel_password,omitempty" toml:"sentinel_password,omitempty" json:"sentinel_password,omitempty" jsonschema:"title=Sentinel Password" jsonschema_description:"The password for the Redis Sentinel connection."`
	SentinelMode               string        `koanf:"sentinel_mode" yaml:"sentinel_mode,omitempty" toml:"sentinel_mode,omitempty" json:"sentinel_mode,omitempty" jsonschema:"title=Sentinel Mode,default=failover,enum=failover,enum=cluster" jsonschema_description:"The mode for the Redis Sentinel connection."`
	Addresses                  []*AddressTCP `koanf:"addresses" yaml:"addresses,omitempty" toml:"addresses,omitempty" json:"addresses,omitempty" jsonschema:"title=Addresses" jsonschema_description:"The addresses of the Redis Sentinel nodes to connect to."`
	RouteByLatency             bool          `koanf:"route_by_latency" yaml:"route_by_latency,omitempty" toml:"route_by_latency,omitempty" json:"route_by_latency,omitempty" jsonschema:"title=Route By Latency" jsonschema_description:"Route commands to the node with the lowest latency."`
	RouteRandomly              bool          `koanf:"route_randomly" yaml:"route_randomly,omitempty" toml:"route_randomly,omitempty" json:"route_randomly,omitempty" jsonschema:"title=Route Randomly" jsonschema_description:"Route commands randomly to nodes."`
	Database                   int           `koanf:"database" yaml:"database,omitempty" toml:"database,omitempty" json:"database,omitempty" jsonschema:"title=Database" jsonschema_description:"The database to use for the Redis server."`
	Username                   string        `koanf:"username" yaml:"username,omitempty" toml:"username,omitempty" json:"username,omitempty" jsonschema:"title=Username" jsonschema_description:"The username to use for the Redis server."`
	Password                   string        `koanf:"password" yaml:"password,omitempty" toml:"password,omitempty" json:"password,omitempty" jsonschema:"title=Password" jsonschema_description:"The password to use for the Redis server."`
	TLS                        *TLS          `koanf:"tls" yaml:"tls,omitempty" toml:"tls,omitempty" json:"tls,omitempty" jsonschema:"title=TLS" jsonschema_description:"The TLS configuration for the Redis server."`
	DialTimeout                time.Duration `koanf:"dial_timeout" yaml:"dial_timeout,omitempty" toml:"dial_timeout,omitempty" json:"dial_timeout,omitempty" jsonschema:"default=5 seconds,title=Dial Timeout" jsonschema_description:"The dial timeout for the Redis server."`
	ReadTimeout                time.Duration `koanf:"read_timeout" yaml:"read_timeout,omitempty" toml:"read_timeout,omitempty" json:"read_timeout,omitempty" jsonschema:"default=3 seconds,title=Read Timeout" jsonschema_description:"The read timeout for the Redis server."`
	WriteTimeout               time.Duration `koanf:"write_timeout" yaml:"write_timeout,omitempty" toml:"write_timeout,omitempty" json:"write_timeout,omitempty" jsonschema:"default=3 seconds,title=Write Timeout" jsonschema_description:"The write timeout for the Redis server."`
	ConnectionTimeout          time.Duration `koanf:"connection_timeout" yaml:"connection_timeout,omitempty" toml:"connection_timeout,omitempty" json:"connection_timeout,omitempty" jsonschema:"default=10 seconds,title=Connection Timeout" jsonschema_description:"The connection timeout for the Redis server."`
	IdleTimeout                time.Duration `koanf:"idle_timeout" yaml:"idle_timeout,omitempty" toml:"idle_timeout,omitempty" json:"idle_timeout,omitempty" jsonschema:"default=10 seconds,title=Idle Timeout" jsonschema_description:"The connection idle timeout for the Redis server."`
	PoolTimeout                time.Duration `koanf:"pool_timeout" yaml:"pool_timeout,omitempty" toml:"pool_timeout,omitempty" json:"pool_timeout,omitempty" jsonschema:"default=10 seconds,title=Pool Timeout" jsonschema_description:"The pool timeout for the Redis server."`
	MinimumRetryBackoff        time.Duration `koanf:"minimum_retry_backoff" yaml:"minimum_retry_backoff,omitempty" toml:"minimum_retry_backoff,omitempty" json:"minimum_retry_backoff,omitempty" jsonschema:"default=100 milliseconds,title=Minimum Retry Backoff" jsonschema_description:"The minimum retry backoff to use when connecting to the Redis server."`
	MaximumRetryBackoff        time.Duration `koanf:"maximum_retry_backoff" yaml:"maximum_retry_backoff,omitempty" toml:"maximum_retry_backoff,omitempty" json:"maximum_retry_backoff,omitempty" jsonschema:"default=100 milliseconds,title=Maximum Retry Backoff" jsonschema_description:"The maximum retry backoff to use when connecting to the Redis server."`
	MaximumRetries             int           `koanf:"maximum_retries" yaml:"maximum_retries,omitempty" toml:"maximum_retries,omitempty" json:"maximum_retries,omitempty" jsonschema:"default=3,title=Max Retries" jsonschema_description:"The maximum number of retries to attempt when connecting to the Redis server."`
	PoolSize                   int           `koanf:"pool_size" yaml:"pool_size,omitempty" toml:"pool_size,omitempty" json:"pool_size,omitempty" jsonschema:"default=8,title=Pool Size" jsonschema_description:"The pool size to use when connecting to the Redis server."`
	PoolMinimumIdleConnections int           `koanf:"pool_minimum_idle_connections" yaml:"pool_minimum_idle_connections,omitempty" toml:"pool_minimum_idle_connections,omitempty" json:"pool_minimum_idle_connections,omitempty" jsonschema:"default=10,title=Pool Minimum Idle Connections" jsonschema_description:"The minimum idle connections to keep open with the pool when connecting to the Redis server."`
	PoolMaximumIdleConnections int           `koanf:"pool_maximum_idle_connections" yaml:"pool_maximum_idle_connections,omitempty" toml:"pool_maximum_idle_connections,omitempty" json:"pool_maximum_idle_connections,omitempty" jsonschema:"default=10,title=Pool Maximum Idle Connections" jsonschema_description:"The maximum idle connections to keep open with the pool when connecting to the Redis server."`
	PoolMaximumConnections     int           `koanf:"pool_maximum_connections" yaml:"pool_maximum_connections,omitempty" toml:"pool_maximum_connections,omitempty" json:"pool_maximum_connections,omitempty" jsonschema:"default=10,title=Pool Maximum Connections" jsonschema_description:"The maximum number of connections the pool can have open when connecting to the Redis server."`
	DialerRetries              int           `koanf:"dialer_retries" yaml:"dialer_retries,omitempty" toml:"dialer_retries,omitempty" json:"dialer_retries,omitempty" jsonschema:"default=5,title=Dialer Retries" jsonschema_description:"The maximum number of retry attempts when dialing a connection fails."`
	DialerRetryTimeout         time.Duration `koanf:"dialer_retry_timeout" yaml:"dialer_retry_timeout,omitempty" toml:"dialer_retry_timeout,omitempty" json:"dialer_retry_timeout,omitempty" jsonschema:"default=100 milliseconds,title=Dialer Retry Timeout" jsonschema_description:"The backoff duration between dial retry attempts."`
	MaximumConcurrentDials     int           `koanf:"maximum_concurrent_dials" yaml:"maximum_concurrent_dials,omitempty" toml:"maximum_concurrent_dials,omitempty" json:"maximum_concurrent_dials,omitempty" jsonschema:"title=Maximum Concurrent Dials" jsonschema_description:"The maximum number of concurrent connection creation goroutines. Defaults to the pool size."`
	ConnectionLifetimeJitter   time.Duration `koanf:"connection_lifetime_jitter" yaml:"connection_lifetime_jitter,omitempty" toml:"connection_lifetime_jitter,omitempty" json:"connection_lifetime_jitter,omitempty" jsonschema:"title=Connection Lifetime Jitter" jsonschema_description:"The absolute jitter applied to the connection timeout to prevent all connections expiring simultaneously."`
	ContextTimeoutEnabled      bool          `koanf:"context_timeout_enabled" yaml:"context_timeout_enabled,omitempty" toml:"context_timeout_enabled,omitempty" json:"context_timeout_enabled,omitempty" jsonschema:"default=false,title=Context Timeout Enabled" jsonschema_description:"Enables the client respecting context timeouts and deadlines."`
	PoolFIFO                   bool          `koanf:"pool_fifo" yaml:"pool_fifo,omitempty" toml:"pool_fifo,omitempty" json:"pool_fifo,omitempty" jsonschema:"default=false,title=Pool FIFO" jsonschema_description:"Uses a FIFO connection pool rather than a LIFO connection pool."`
	ReadBufferSize             int           `koanf:"read_buffer_size" yaml:"read_buffer_size,omitempty" toml:"read_buffer_size,omitempty" json:"read_buffer_size,omitempty" jsonschema:"default=32768,title=Read Buffer Size" jsonschema_description:"The size of the read buffer in bytes for each connection."`
	WriteBufferSize            int           `koanf:"write_buffer_size" yaml:"write_buffer_size,omitempty" toml:"write_buffer_size,omitempty" json:"write_buffer_size,omitempty" jsonschema:"default=32768,title=Write Buffer Size" jsonschema_description:"The size of the write buffer in bytes for each connection."`
	FailingTimeout             time.Duration `koanf:"failing_timeout" yaml:"failing_timeout,omitempty" toml:"failing_timeout,omitempty" json:"failing_timeout,omitempty" jsonschema:"default=15 seconds,title=Failing Timeout" jsonschema_description:"The duration a node is avoided for after being marked as failing."`
	ReplicaOnly                bool          `koanf:"replica_only" yaml:"replica_only,omitempty" toml:"replica_only,omitempty" json:"replica_only,omitempty" jsonschema:"default=false,title=Replica Only" jsonschema_description:"Routes all commands to replica nodes."`
	UseDisconnectedReplicas    bool          `koanf:"use_disconnected_replicas" yaml:"use_disconnected_replicas,omitempty" toml:"use_disconnected_replicas,omitempty" json:"use_disconnected_replicas,omitempty" jsonschema:"default=false,title=Use Disconnected Replicas" jsonschema_description:"Allows routing commands to replicas which sentinel reports as disconnected."`
}

// RedisClusterCache represents the configuration for a Redis Cluster cache backend.
type RedisClusterCache struct {
	Addresses                  []*AddressTCP `koanf:"addresses" yaml:"addresses,omitempty" toml:"addresses,omitempty" json:"addresses,omitempty" jsonschema:"title=Addresses" jsonschema_description:"The addresses of the Redis Cluster nodes to connect to."`
	RouteByReplica             bool          `koanf:"route_by_replica" yaml:"route_by_replica,omitempty" toml:"route_by_replica,omitempty" json:"route_by_replica,omitempty" jsonschema:"title=Route By Replica" jsonschema_description:"Route read-only commands to replica nodes."`
	RouteByLatency             bool          `koanf:"route_by_latency" yaml:"route_by_latency,omitempty" toml:"route_by_latency,omitempty" json:"route_by_latency,omitempty" jsonschema:"title=Route By Latency" jsonschema_description:"Route commands to the node with the lowest latency."`
	RouteRandomly              bool          `koanf:"route_randomly" yaml:"route_randomly,omitempty" toml:"route_randomly,omitempty" json:"route_randomly,omitempty" jsonschema:"title=Route Randomly" jsonschema_description:"Route commands randomly to nodes."`
	Username                   string        `koanf:"username" yaml:"username,omitempty" toml:"username,omitempty" json:"username,omitempty" jsonschema:"title=Username" jsonschema_description:"The username to use for the Redis server."`
	Password                   string        `koanf:"password" yaml:"password,omitempty" toml:"password,omitempty" json:"password,omitempty" jsonschema:"title=Password" jsonschema_description:"The password to use for the Redis server."`
	TLS                        *TLS          `koanf:"tls" yaml:"tls,omitempty" toml:"tls,omitempty" json:"tls,omitempty" jsonschema:"title=TLS" jsonschema_description:"The TLS configuration for the Redis server."`
	DialTimeout                time.Duration `koanf:"dial_timeout" yaml:"dial_timeout,omitempty" toml:"dial_timeout,omitempty" json:"dial_timeout,omitempty" jsonschema:"default=5 seconds,title=Dial Timeout" jsonschema_description:"The dial timeout for the Redis server."`
	ReadTimeout                time.Duration `koanf:"read_timeout" yaml:"read_timeout,omitempty" toml:"read_timeout,omitempty" json:"read_timeout,omitempty" jsonschema:"default=3 seconds,title=Read Timeout" jsonschema_description:"The read timeout for the Redis server."`
	WriteTimeout               time.Duration `koanf:"write_timeout" yaml:"write_timeout,omitempty" toml:"write_timeout,omitempty" json:"write_timeout,omitempty" jsonschema:"default=3 seconds,title=Write Timeout" jsonschema_description:"The write timeout for the Redis server."`
	ConnectionTimeout          time.Duration `koanf:"connection_timeout" yaml:"connection_timeout,omitempty" toml:"connection_timeout,omitempty" json:"connection_timeout,omitempty" jsonschema:"default=10 seconds,title=Connection Timeout" jsonschema_description:"The connection timeout for the Redis server."`
	IdleTimeout                time.Duration `koanf:"idle_timeout" yaml:"idle_timeout,omitempty" toml:"idle_timeout,omitempty" json:"idle_timeout,omitempty" jsonschema:"default=10 seconds,title=Idle Timeout" jsonschema_description:"The connection idle timeout for the Redis server."`
	PoolTimeout                time.Duration `koanf:"pool_timeout" yaml:"pool_timeout,omitempty" toml:"pool_timeout,omitempty" json:"pool_timeout,omitempty" jsonschema:"default=10 seconds,title=Pool Timeout" jsonschema_description:"The pool timeout for the Redis server."`
	MinimumRetryBackoff        time.Duration `koanf:"minimum_retry_backoff" yaml:"minimum_retry_backoff,omitempty" toml:"minimum_retry_backoff,omitempty" json:"minimum_retry_backoff,omitempty" jsonschema:"default=100 milliseconds,title=Minimum Retry Backoff" jsonschema_description:"The minimum retry backoff to use when connecting to the Redis server."`
	MaximumRetryBackoff        time.Duration `koanf:"maximum_retry_backoff" yaml:"maximum_retry_backoff,omitempty" toml:"maximum_retry_backoff,omitempty" json:"maximum_retry_backoff,omitempty" jsonschema:"default=100 milliseconds,title=Maximum Retry Backoff" jsonschema_description:"The maximum retry backoff to use when connecting to the Redis server."`
	MaximumRedirects           int           `koanf:"maximum_redirects" yaml:"maximum_redirects,omitempty" toml:"maximum_redirects,omitempty" json:"maximum_redirects,omitempty" jsonschema:"default=3,title=Max Redirects" jsonschema_description:"The maximum number of redirects to follow when connecting to the Redis server."`
	MaximumRetries             int           `koanf:"maximum_retries" yaml:"maximum_retries,omitempty" toml:"maximum_retries,omitempty" json:"maximum_retries,omitempty" jsonschema:"default=3,title=Max Retries" jsonschema_description:"The maximum number of retries before giving up connecting to the Redis server."`
	PoolSize                   int           `koanf:"pool_size" yaml:"pool_size,omitempty" toml:"pool_size,omitempty" json:"pool_size,omitempty" jsonschema:"default=8,title=Pool Size" jsonschema_description:"The pool size to use when connecting to the Redis server."`
	PoolMinimumIdleConnections int           `koanf:"pool_minimum_idle_connections" yaml:"pool_minimum_idle_connections,omitempty" toml:"pool_minimum_idle_connections,omitempty" json:"pool_minimum_idle_connections,omitempty" jsonschema:"default=10,title=Pool Minimum Idle Connections" jsonschema_description:"The minimum idle connections to keep open with the pool when connecting to the Redis server."`
	PoolMaximumIdleConnections int           `koanf:"pool_maximum_idle_connections" yaml:"pool_maximum_idle_connections,omitempty" toml:"pool_maximum_idle_connections,omitempty" json:"pool_maximum_idle_connections,omitempty" jsonschema:"default=10,title=Pool Maximum Idle Connections" jsonschema_description:"The maximum idle connections to keep open with the pool when connecting to the Redis server."`
	PoolMaximumConnections     int           `koanf:"pool_maximum_connections" yaml:"pool_maximum_connections,omitempty" toml:"pool_maximum_connections,omitempty" json:"pool_maximum_connections,omitempty" jsonschema:"default=10,title=Pool Maximum Connections" jsonschema_description:"The maximum number of connections the pool can have open when connecting to the Redis server."`
	DialerRetries              int           `koanf:"dialer_retries" yaml:"dialer_retries,omitempty" toml:"dialer_retries,omitempty" json:"dialer_retries,omitempty" jsonschema:"default=5,title=Dialer Retries" jsonschema_description:"The maximum number of retry attempts when dialing a connection fails."`
	DialerRetryTimeout         time.Duration `koanf:"dialer_retry_timeout" yaml:"dialer_retry_timeout,omitempty" toml:"dialer_retry_timeout,omitempty" json:"dialer_retry_timeout,omitempty" jsonschema:"default=100 milliseconds,title=Dialer Retry Timeout" jsonschema_description:"The backoff duration between dial retry attempts."`
	MaximumConcurrentDials     int           `koanf:"maximum_concurrent_dials" yaml:"maximum_concurrent_dials,omitempty" toml:"maximum_concurrent_dials,omitempty" json:"maximum_concurrent_dials,omitempty" jsonschema:"title=Maximum Concurrent Dials" jsonschema_description:"The maximum number of concurrent connection creation goroutines. Defaults to the pool size."`
	ConnectionLifetimeJitter   time.Duration `koanf:"connection_lifetime_jitter" yaml:"connection_lifetime_jitter,omitempty" toml:"connection_lifetime_jitter,omitempty" json:"connection_lifetime_jitter,omitempty" jsonschema:"title=Connection Lifetime Jitter" jsonschema_description:"The absolute jitter applied to the connection timeout to prevent all connections expiring simultaneously."`
	ContextTimeoutEnabled      bool          `koanf:"context_timeout_enabled" yaml:"context_timeout_enabled,omitempty" toml:"context_timeout_enabled,omitempty" json:"context_timeout_enabled,omitempty" jsonschema:"default=false,title=Context Timeout Enabled" jsonschema_description:"Enables the client respecting context timeouts and deadlines."`
	PoolFIFO                   bool          `koanf:"pool_fifo" yaml:"pool_fifo,omitempty" toml:"pool_fifo,omitempty" json:"pool_fifo,omitempty" jsonschema:"default=false,title=Pool FIFO" jsonschema_description:"Uses a FIFO connection pool rather than a LIFO connection pool."`
	ReadBufferSize             int           `koanf:"read_buffer_size" yaml:"read_buffer_size,omitempty" toml:"read_buffer_size,omitempty" json:"read_buffer_size,omitempty" jsonschema:"default=32768,title=Read Buffer Size" jsonschema_description:"The size of the read buffer in bytes for each connection."`
	WriteBufferSize            int           `koanf:"write_buffer_size" yaml:"write_buffer_size,omitempty" toml:"write_buffer_size,omitempty" json:"write_buffer_size,omitempty" jsonschema:"default=32768,title=Write Buffer Size" jsonschema_description:"The size of the write buffer in bytes for each connection."`
	FailingTimeout             time.Duration `koanf:"failing_timeout" yaml:"failing_timeout,omitempty" toml:"failing_timeout,omitempty" json:"failing_timeout,omitempty" jsonschema:"default=15 seconds,title=Failing Timeout" jsonschema_description:"The duration a node is avoided for after being marked as failing."`
	ClusterStateReloadInterval time.Duration `koanf:"cluster_state_reload_interval" yaml:"cluster_state_reload_interval,omitempty" toml:"cluster_state_reload_interval,omitempty" json:"cluster_state_reload_interval,omitempty" jsonschema:"title=Cluster State Reload Interval" jsonschema_description:"The interval the cluster state is reloaded at."`
}

// DefaultRedisCacheConfiguration is the default redis cache configuration.
var DefaultRedisCacheConfiguration = RedisCache{
	DialTimeout:        time.Second * 5,
	ReadTimeout:        time.Second * 3,
	WriteTimeout:       time.Second * 3,
	MaximumRetries:     3,
	PoolSize:           8,
	DialerRetries:      5,
	DialerRetryTimeout: time.Millisecond * 100,
	ReadBufferSize:     32768,
	WriteBufferSize:    32768,
	FailingTimeout:     time.Second * 15,
	TLS: &TLS{
		MinimumVersion: TLSVersion{Value: tls.VersionTLS12},
	},
}

// DefaultRedisSentinelCacheConfiguration is the default redis sentinel cache configuration.
var DefaultRedisSentinelCacheConfiguration = RedisSentinelCache{
	SentinelMode:       "failover",
	DialTimeout:        time.Second * 5,
	ReadTimeout:        time.Second * 3,
	WriteTimeout:       time.Second * 3,
	MaximumRetries:     3,
	PoolSize:           8,
	DialerRetries:      5,
	DialerRetryTimeout: time.Millisecond * 100,
	ReadBufferSize:     32768,
	WriteBufferSize:    32768,
	FailingTimeout:     time.Second * 15,
	TLS: &TLS{
		MinimumVersion: TLSVersion{Value: tls.VersionTLS12},
	},
}

// DefaultRedisClusterCacheConfiguration is the default redis cluster cache configuration.
var DefaultRedisClusterCacheConfiguration = RedisClusterCache{
	DialTimeout:        time.Second * 5,
	ReadTimeout:        time.Second * 3,
	WriteTimeout:       time.Second * 3,
	MaximumRetries:     3,
	MaximumRedirects:   3,
	PoolSize:           8,
	DialerRetries:      5,
	DialerRetryTimeout: time.Millisecond * 100,
	ReadBufferSize:     32768,
	WriteBufferSize:    32768,
	FailingTimeout:     time.Second * 15,
	TLS: &TLS{
		MinimumVersion: TLSVersion{Value: tls.VersionTLS12},
	},
}

// DefaultRedisCachePort is the default port for a redis server.
const DefaultRedisCachePort = 6379

// DefaultRedisSentinelCachePort is the default port for a redis sentinel node.
const DefaultRedisSentinelCachePort = 26379
