---
layout: default
title: Redis
parent: Session
grand_parent: Configuration
nav_order: 1
---

# Redis

This is a session provider. By default Authelia uses an in-memory provider. Not configuring redis leaves Authelia
[stateful](../../features/statelessness.md). It's important in highly available scenarios to configure this option and
we highly recommend it in production environments. It requires you setup [redis] as well.

## Configuration

```yaml
session:
  redis:
    host: 127.0.0.1
    port: 6379
    username: authelia
    password: authelia
    database_index: 0
    maximum_active_connections: 8
    minimum_idle_connections: 0
    tls:
      server_name: myredis.example.com
      skip_verify: false
      minimum_version: TLS1.2
    high_availability:
      sentinel_name: mysentinel
      sentinel_password: sentinel_specific_pass
      nodes:
        - host: sentinel-node1
          port: 26379
        - host: sentinel-node2
          port: 26379
      route_by_latency: false
      route_randomly: false
```

## Options

### host

The [redis] host or unix socket path. If utilising an IPv6 literal address it must be enclosed by square brackets and
quoted:
```yaml
host: "[fd00:1111:2222:3333::1]"
```

### port

The port [redis] is listening on.

### username

The username for [redis authentication](https://redis.io/commands/auth). Only supported in [redis] 6.0+, and [redis]
currently offers backwards compatibility with password-only auth. You probably do not need to set this unless you went
through the process of setting up [redis ACLs](https://redis.io/topics/acl).

### password

The password for [redis authentication](https://redis.io/commands/auth).

### database_index

The index number of the [redis] database, the same value as specified with the redis SELECT command.

### maximum_active_connections

The maximum connections open to [redis] at the same time.

### minimum_idle_connections

The minimum number of [redis] connections to keep open as long as they don't exceed the maximum active connections. This
is useful if there are long delays in establishing connections.

### tls

If defined enables [redis] over TLS, and additionally controls the TLS connection validation process. You can see how to 
configure the tls section [here](../index.md#tls-configuration).

### high_availability

When defining this session it enables [redis sentinel] connections. It's possible in
the future we may add [redis cluster](https://redis.io/topics/cluster-tutorial).

#### sentinel_name

The [redis sentinel] master name. This is defined in your [redis sentinel] configuration, it is not a hostname. This
must be defined currently for a high availability configuration.

#### sentinel_password

The password for the [redis sentinel] connection. A [redis sentinel] username is not supported at this time due to the 
upstream library not supporting it. 

#### nodes

A list of [redis sentinel] nodes to load balance over. This list is added to the host in the [redis] section above. It 
is required you either define the [redis] host or one [redis sentinel] node. The [redis] host must be a [redis sentinel]
host, not a regular one. The individual [redis] hosts are determined using [redis sentinel] commands.

Each node has a host and port configuration. Example:

```yaml
- host: redis-sentinel-0
  port: 26379
```

#### route_by_latency

Prioritizes low latency [redis sentinel] nodes when set to true.

#### route_randomly

Randomly chooses [redis sentinel] nodes when set to true.

[redis]: https://redis.io
[redis sentinel]: https://redis.io/topics/sentinel