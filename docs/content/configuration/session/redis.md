---
title: "Redis"
description: "Redis Session Configuration"
summary: "Configuring the Redis Session Storage."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 106200
toc: true
aliases:
  - /docs/configuration/session/redis.html
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

This is a session provider. By default Authelia uses an in-memory provider. Not configuring redis leaves Authelia
[stateful](../../overview/authorization/statelessness.md). It's important in highly available scenarios to configure
this option and we highly recommend it in production environments. It requires you setup [redis] as well.

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
session:
  redis:
    host: '127.0.0.1'
    port: 6379
    timeout: '5s'
    max_retries: 0
    username: 'authelia'
    password: 'authelia'
    database_index: 0
    maximum_active_connections: 8
    minimum_idle_connections: 0
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
    high_availability:
      sentinel_name: 'mysentinel'
      # If `sentinel_username` is supplied, Authelia will connect using ACL-based
      # authentication. Otherwise, it will use traditional `requirepass` auth.
      sentinel_username: 'sentinel_user'
      sentinel_password: 'sentinel_specific_pass'
      nodes:
        - host: 'sentinel-node1'
          port: 26379
        - host: 'sentinel-node2'
          port: 26379
      route_by_latency: false
      route_randomly: false
```

## Options

This section describes the individual configuration options.

### host

{{< confkey type="string" required="yes" >}}

The [redis] host or unix socket path. If utilising an IPv6 literal address it must be enclosed by square brackets and
quoted:

```yaml {title="configuration.yml"}
host: '[fd00:1111:2222:3333::1]'
```

### timeout

{{< confkey type="string,integer" syntax="duration" default="5 seconds" required="no" >}}

The Redis connection timeout.

### max_retries

{{< confkey type="integer" default="0" required="no" >}}

The maximum number of retries on a failed command. Setting this option to 0 disables retries entirely.

### port

{{< confkey type="integer" default="6379" required="no" >}}

The port [redis] is listening on.

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

### database_index

{{< confkey type="integer" default="0" required="no" >}}

The index number of the [redis] database, the same value as specified with the redis SELECT command.

### maximum_active_connections

{{< confkey type="integer" default="8" required="no" >}}

The maximum connections open to [redis] at the same time.

### minimum_idle_connections

{{< confkey type="integer" default="0" required="no" >}}

The minimum number of [redis] connections to keep open as long as they don't exceed the maximum active connections. This
is useful if there are long delays in establishing connections.

### tls

{{< confkey type="structure" structure="tls" required="no" >}}

If defined enables connecting over a TLS socket and additionally controls the TLS connection
verification parameters for the [redis] server.

By default Authelia uses the system certificate trust for TLS certificate verification of TLS connections and the
[certificates_directory](../miscellaneous/introduction.md#certificates_directory) global option can be used to augment
this.

### high_availability

When defining this session it enables [redis sentinel] connections. It's possible in
the future we may add [redis cluster](https://redis.io/topics/cluster-tutorial).

#### sentinel_name

{{< confkey type="string" required="yes" >}}

The [redis sentinel] master name. This is defined in your [redis sentinel] configuration, it is not a hostname. This
must be defined currently for a high availability configuration.

#### sentinel_username

{{< confkey type="string" required="no" >}}

The username for the [redis sentinel] connection. If this is provided, it will be used along with the sentinel_password
for ACL-based authentication to the Redis Sentinel. If only a password is provided, the [redis sentinel] connection will
be authenticated with traditional [requirepass] authentication.

#### sentinel_password

{{< confkey type="string" required="no (yes if sentinel_username is supplied)" secret="yes" >}}

The password for the [redis sentinel] connection. If specified with sentinel_username, configures Authelia to
authenticate to the Redis Sentinel with ACL-based authentication. Otherwise, this is used for [requirepass]
authentication.

It's __strongly recommended__ this is a
[Random Alphanumeric String](../../reference/guides/generating-secure-values.md#generating-a-random-alphanumeric-string) with 64 or more
characters and the user password is changed to this value.

#### nodes

A list of [redis sentinel] nodes to load balance over. This list is added to the host in the [redis] section above. It
is required you either define the [redis] host or one [redis sentinel] node. The [redis] host must be a [redis sentinel]
host, not a regular one. The individual [redis] hosts are determined using [redis sentinel] commands.

Each node has a host and port configuration. Example:

```yaml {title="configuration.yml"}
- host: redis-sentinel-0
  port: 26379
```

##### host

{{< confkey type="boolean" default="false" required="no" >}}

The host of this [redis sentinel] node.

##### port

{{< confkey type="integer" default="26379" required="no" >}}

The port of this [redis sentinel] node.

#### route_by_latency

{{< confkey type="boolean" default="false" required="no" >}}

Prioritizes low latency [redis sentinel] nodes when set to true.

#### route_randomly

{{< confkey type="boolean" default="false" required="no" >}}

Randomly chooses [redis sentinel] nodes when set to true.

[redis]: https://redis.io
[redis sentinel]: https://redis.io/topics/sentinel
[requirepass]: https://redis.io/topics/config
