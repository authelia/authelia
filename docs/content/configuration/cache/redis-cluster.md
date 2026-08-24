---
title: "Redis Cluster"
description: "Redis Cluster Cache Configuration"
summary: "Configuring the Redis Cluster Cache."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 106400
toc: true
aliases: []
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

This is a cache provider. Not configuring a cache provider leaves Authelia
[stateful](../../overview/authorization/statelessness.md). It's important in highly available scenarios to configure
a cache and we highly recommend it in production environments.

This provider connects to a [redis] Cluster, which shards the keyspace across multiple masters each with their own
replicas.

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
cache:
  redis_cluster:
    addresses:
      - 'tcp://redis1:6379'
      - 'tcp://redis2:6379'
      - 'tcp://redis3:6379'
    route_by_replica: false
    route_by_latency: false
    route_randomly: false
    maximum_redirects: 3
    cluster_state_reload_interval: '10 seconds'
    username: 'authelia'
    password: 'authelia'
    dial_timeout: '5 seconds'
    read_timeout: '3 seconds'
    write_timeout: '3 seconds'
    connection_timeout: '0'
    idle_timeout: '30 minutes'
    pool_timeout: '0'
    minimum_retry_backoff: '8 milliseconds'
    maximum_retry_backoff: '512 milliseconds'
    maximum_retries: 3
    pool_size: 8
    pool_minimum_idle_connections: 0
    pool_maximum_idle_connections: 0
    pool_maximum_connections: 0
    dialer_retries: 5
    dialer_retry_timeout: '100 milliseconds'
    maximum_concurrent_dials: 0
    connection_lifetime_jitter: '0'
    context_timeout_enabled: false
    pool_fifo: false
    read_buffer_size: 32768
    write_buffer_size: 32768
    failing_timeout: '15 seconds'
    tls:
      server_name: 'myredis.{{< sitevar name="domain" nojs="example.com" >}}'
      skip_verify: false
      minimum_version: 'TLS1.2'
      maximum_version: 'TLS1.3'
      certificate_chain: |
        -----BEGIN CERTIFICATE-----
        ...
        -----END CERTIFICATE-----
        -----BEGIN CERTIFICATE-----
        ...
        -----END CERTIFICATE-----
      private_key: |
        -----BEGIN PRIVATE KEY-----
        ...
        -----END PRIVATE KEY-----
```

## Options

This section describes the individual configuration options.

### addresses

{{< confkey type="list(string)" syntax="address" required="yes" >}}

Configures the addresses of the [redis] Cluster nodes used to discover the rest of the cluster. Each address is a
connector and the scheme must either be the `unix` scheme or one of the `tcp` schemes. Unlike some other addresses in
Authelia a port must be explicitly included when using a `tcp` scheme as no port is assumed on your behalf.

You do not need to list every node in the cluster, however listing more than one avoids a single point of failure during
the initial discovery.

__Examples:__

```yaml {title="configuration.yml"}
cache:
  redis_cluster:
    addresses:
      - 'tcp://127.0.0.1:6379'
      - 'tcp://[fd00:1111:2222:3333::1]:6379'
      - 'unix:///var/run/redis.sock'
```

### route_by_replica

{{< confkey type="boolean" default="false" required="no" >}}

Enables routing read-only commands to replica nodes rather than sending every command to the master which owns the slot.

### route_by_latency

{{< confkey type="boolean" default="false" required="no" >}}

Routes read-only commands to the master or replica with the lowest latency. Enabling this implicitly enables
[route_by_replica](#route_by_replica).

### route_randomly

{{< confkey type="boolean" default="false" required="no" >}}

Routes read-only commands to a randomly selected master or replica. Enabling this implicitly enables
[route_by_replica](#route_by_replica).

### maximum_redirects

{{< confkey type="integer" default="3" required="no" >}}

The maximum number of times a command is retried in response to a `MOVED` or `ASK` redirect before the error is
returned. A value of 0 or less is replaced with the default.

### cluster_state_reload_interval

{{< confkey type="string,integer" syntax="duration" default="10 seconds" required="no" >}}

How often the cached view of the cluster slot topology is reloaded. Lowering this makes the client notice resharding and
failover sooner at the cost of additional `CLUSTER SLOTS` traffic.

### username

{{< confkey type="string" required="no" >}}

The username for [redis authentication](https://redis.io/commands/auth). Only supported in [redis] 6.0+, and [redis]
currently offers backwards compatibility with password-only auth. You probably do not need to set this unless you went
through the process of setting up [redis ACLs](https://redis.io/topics/acl).

### password

{{< confkey type="string" required="no" secret="yes" >}}

The password for [redis authentication](https://redis.io/commands/auth).

It's __strongly recommended__ this is a
[Random Alphanumeric String](../../reference/guides/generating-secure-values.md#generating-a-random-alphanumeric-string) with 64 or more
characters and the user password is changed to this value.

### dial_timeout

{{< confkey type="string,integer" syntax="duration" default="5 seconds" required="no" >}}

The maximum amount of time to wait while establishing a new connection to a cluster node before giving up. A value of 0
or less is replaced with the default.

### read_timeout

{{< confkey type="string,integer" syntax="duration" default="3 seconds" required="no" >}}

The maximum amount of time to wait for a socket read before the command fails. A value of 0 or less is replaced with the
default, so reads always have a deadline.

### write_timeout

{{< confkey type="string,integer" syntax="duration" default="3 seconds" required="no" >}}

The maximum amount of time to wait for a socket write before the command fails. A value of 0 or less is replaced with
the default, so writes always have a deadline.

### connection_timeout

{{< confkey type="string,integer" syntax="duration" default="unlimited" required="no" >}}

The maximum lifetime of a connection, after which it's closed and replaced with a new one. This is useful for
periodically rebalancing connections after the cluster topology changes. When set to 0 connections are never closed
because of their age.

### idle_timeout

{{< confkey type="string,integer" syntax="duration" default="30 minutes" required="no" >}}

The maximum amount of time a connection may sit idle in the pool before it's closed. This should be lower than the
timeout configured on the [redis] nodes themselves so idle connections are reaped by Authelia rather than dropped by the
server.

### pool_timeout

{{< confkey type="string,integer" syntax="duration" default="read_timeout + 1 second" required="no" >}}

The maximum amount of time to wait for a free connection when every connection in the pool is busy, after which the
command fails.

### minimum_retry_backoff

{{< confkey type="string,integer" syntax="duration" default="8 milliseconds" required="no" >}}

The lower bound of the backoff period between retries of a failed command. See [maximum_retries](#maximum_retries).

### maximum_retry_backoff

{{< confkey type="string,integer" syntax="duration" default="512 milliseconds" required="no" >}}

The upper bound of the backoff period between retries of a failed command. See [maximum_retries](#maximum_retries).

### maximum_retries

{{< confkey type="integer" default="3" required="no" >}}

The maximum number of times a failed command is retried before the error is returned. A value of 0 or less is replaced
with the default, so retries can't be disabled entirely. This is distinct from
[maximum_redirects](#maximum_redirects) which governs slot redirects rather than errors.

### pool_size

{{< confkey type="integer" default="8" required="no" >}}

The base number of socket connections held open per cluster node. Additional connections are opened in excess of this
value when every connection is busy; use [pool_maximum_connections](#pool_maximum_connections) to place a hard limit on
the pool.

### pool_minimum_idle_connections

{{< confkey type="integer" default="0" required="no" >}}

The minimum number of idle connections kept open in the pool. This is useful when establishing a new connection is slow
and you want to avoid paying that cost on the request path.

### pool_maximum_idle_connections

{{< confkey type="integer" default="0" required="no" >}}

The maximum number of idle connections retained in the pool. When set to 0 idle connections are not closed on the basis
of their number, only on the basis of [idle_timeout](#idle_timeout).

### pool_maximum_connections

{{< confkey type="integer" default="0" required="no" >}}

The maximum number of connections allocated by the pool at any given time. When set to 0 the number of connections is
unlimited. When the limit is reached, further commands block until a connection is released back into the pool.

### dialer_retries

{{< confkey type="integer" default="5" required="no" >}}

The maximum number of attempts made when dialing a new connection fails. A value of 0 or less is replaced with the
default.

### dialer_retry_timeout

{{< confkey type="string,integer" syntax="duration" default="100 milliseconds" required="no" >}}

The backoff period between attempts to dial a new connection. See [dialer_retries](#dialer_retries).

### maximum_concurrent_dials

{{< confkey type="integer" default="pool_size" required="no" >}}

The maximum number of connections which may be established concurrently. When set to 0 or less this defaults to
[pool_size](#pool_size), and any value above [pool_size](#pool_size) is capped at [pool_size](#pool_size).

### connection_lifetime_jitter

{{< confkey type="string,integer" syntax="duration" default="0" required="no" >}}

The maximum random duration subtracted from the lifetime of each connection, which prevents every connection in the pool
expiring at the same moment. This has no effect unless [connection_timeout](#connection_timeout) is configured, and
values greater than [connection_timeout](#connection_timeout) are capped at that value.

### context_timeout_enabled

{{< confkey type="boolean" default="false" required="no" >}}

Enables the client honoring the deadlines and cancellations of the context a command is executed with, in addition to
the configured read and write timeouts.

### pool_fifo

{{< confkey type="boolean" default="false" required="no" >}}

Uses a FIFO connection pool rather than a LIFO connection pool. FIFO has slightly higher overhead than LIFO but closes
idle connections faster, which reduces the size of the pool.

### read_buffer_size

{{< confkey type="integer" default="32768" required="no" >}}

The size in bytes of the read buffer allocated for each connection. A value of 0 or less is replaced with the default.
Smaller buffers reduce memory usage for large pools at the cost of throughput.

### write_buffer_size

{{< confkey type="integer" default="32768" required="no" >}}

The size in bytes of the write buffer allocated for each connection. A value of 0 or less is replaced with the default.
Smaller buffers reduce memory usage for large pools at the cost of throughput.

### failing_timeout

{{< confkey type="string,integer" syntax="duration" default="15 seconds" required="no" >}}

The duration a cluster node is avoided for after it has been marked as failing. The client only supports whole seconds
so this value is truncated to a whole number of seconds, and a non-zero value below one second is raised to one second
so a deliberate value is never silently discarded.

### tls

{{< confkey type="structure" structure="tls" required="no" >}}

If defined enables connecting over a TLS socket and additionally controls the TLS connection
verification parameters for the [redis] nodes.

By default Authelia uses the system certificate trust for TLS certificate verification of TLS connections and the
[certificates_directory](../miscellaneous/introduction.md#certificates_directory) global option can be used to augment
this.

## Notes

There is no `database` option for this provider. [redis] Cluster does not support the `SELECT` command and every key
lives in database 0.

[redis]: https://redis.io
