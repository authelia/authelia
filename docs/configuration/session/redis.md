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
      # If `sentinel_username` is supplied, Authelia will connect using ACL-based
      # authentication. Otherwise, it will use traditional `requirepass` auth.
      sentinel_username: sentinel_user
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
<div markdown="1">
type: string
{: .label .label-config .label-purple }
required: yes
{: .label .label-config .label-red }
</div>

The [redis] host or unix socket path. If utilising an IPv6 literal address it must be enclosed by square brackets and
quoted:
```yaml
host: "[fd00:1111:2222:3333::1]"
```

### port
<div markdown="1">
type: integer
{: .label .label-config .label-purple }
default: 6379
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

The port [redis] is listening on.

### username
<div markdown="1">
type: string
{: .label .label-config .label-purple }
required: no
{: .label .label-config .label-green }
</div>

The username for [redis authentication](https://redis.io/commands/auth). Only supported in [redis] 6.0+, and [redis]
currently offers backwards compatibility with password-only auth. You probably do not need to set this unless you went
through the process of setting up [redis ACLs](https://redis.io/topics/acl).

### password
<div markdown="1">
type: string
{: .label .label-config .label-purple }
required: no
{: .label .label-config .label-green }
</div>

The password for [redis authentication](https://redis.io/commands/auth).

### database_index
<div markdown="1">
type: integer
{: .label .label-config .label-purple }
default: 0
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

The index number of the [redis] database, the same value as specified with the redis SELECT command.

### maximum_active_connections
<div markdown="1">
type: integer
{: .label .label-config .label-purple }
default: 8
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

The maximum connections open to [redis] at the same time.

### minimum_idle_connections
<div markdown="1">
type: integer
{: .label .label-config .label-purple }
default: 0
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

The minimum number of [redis] connections to keep open as long as they don't exceed the maximum active connections. This
is useful if there are long delays in establishing connections.

### tls

If defined enables [redis] over TLS, and additionally controls the TLS connection validation process. You can see how to
configure the tls section [here](../index.md#tls-configuration).

### high_availability

When defining this session it enables [redis sentinel] connections. It's possible in
the future we may add [redis cluster](https://redis.io/topics/cluster-tutorial).

#### sentinel_name
<div markdown="1">
type: string
{: .label .label-config .label-purple }
required: yes
{: .label .label-config .label-red }
</div>

The [redis sentinel] master name. This is defined in your [redis sentinel] configuration, it is not a hostname. This
must be defined currently for a high availability configuration.

#### sentinel_username
<div markdown="1">
type: string
{: .label .label-config .label-purple }
required: no
{: .label .label-config .label-green }
</div>

The username for the [redis sentinel] connection. If this is provided, it will be used along with the sentinel_password
for ACL-based authentication to the Redis Sentinel. If only a password is provided, the [redis sentinel] connection will
be authenticated with traditional requirepass authentication.

#### sentinel_password
<div markdown="1">
type: string
{: .label .label-config .label-purple }
required: no (yes if sentinel_username is supplied)
{: .label .label-config .label-green }
</div>

The password for the [redis sentinel] connection. If specified with sentinel_username, configures Authelia to
authenticate to the Redis Sentinel with ACL-based authentication. Otherwise, this is used for requirepass
authentication.

#### nodes

A list of [redis sentinel] nodes to load balance over. This list is added to the host in the [redis] section above. It
is required you either define the [redis] host or one [redis sentinel] node. The [redis] host must be a [redis sentinel]
host, not a regular one. The individual [redis] hosts are determined using [redis sentinel] commands.

Each node has a host and port configuration. Example:

```yaml
- host: redis-sentinel-0
  port: 26379
```

##### host
<div markdown="1">
type: boolean
{: .label .label-config .label-purple }
default: false
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

The host of this [redis sentinel] node.

##### port
<div markdown="1">
type: integer
{: .label .label-config .label-purple }
default: 26379
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

The port of this [redis sentinel] node.

#### route_by_latency
<div markdown="1">
type: boolean
{: .label .label-config .label-purple }
default: false
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Prioritizes low latency [redis sentinel] nodes when set to true.

#### route_randomly
<div markdown="1">
type: boolean
{: .label .label-config .label-purple }
default: false
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Randomly chooses [redis sentinel] nodes when set to true.

[redis]: https://redis.io
[redis sentinel]: https://redis.io/topics/sentinel
