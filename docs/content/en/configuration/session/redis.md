---
title: "Redis"
description: "Redis Session Configuration"
lead: "Configuring the Redis Session Storage."
date: 2021-04-11T21:25:03+10:00
draft: false
images: []
menu:
  configuration:
    parent: "session"
weight: 105200
toc: true
aliases:
  - /docs/configuration/session/redis.html
---

This is a session provider. By default Authelia uses an in-memory provider. Not configuring redis leaves Authelia
[stateful](../../overview/authorization/statelessness.md). It's important in highly available scenarios to configure
this option and we highly recommend it in production environments. It requires you setup [redis] as well.

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
      maximum_version: TLS1.3
      client_auth_keypair: |
        -----BEGIN CERTIFICATE-----
        MIIC5jCCAc6gAwIBAgIRANQn+N/s2XpbyLjHhrhbLMYwDQYJKoZIhvcNAQELBQAw
        EzERMA8GA1UEChMIQXV0aGVsaWEwHhcNMjIwOTIzMDA1NTM5WhcNMjMwOTIzMDA1
        NTM5WjATMREwDwYDVQQKEwhBdXRoZWxpYTCCASIwDQYJKoZIhvcNAQEBBQADggEP
        ADCCAQoCggEBAJrdtDhT0LXvenka1pxPKN3ay82gGMuUHQPEwer+vhOjQmTuiNzz
        lYOQ5HRKkRippfCNwxf1QKpvllLwWxWh55MeFwgoIdy1ro2Q/sCIHsRdfIIgFxXu
        CBcyAPnfxUrujCIKhbvE72GC1MJWGCwtCJjPWSlqGUZPMFEy8n4WTtzgBDzx7tPU
        roH6tMyERO+LzLex6udNaY3L43TwIIdXfiK7X6tFtIcgySGNMoEoA3TzXvpr8N+5
        oPtFMx7nf8sBjV85AZ134ZAzsSVv/pPF1JyAasBT9n/b4zhr8t/km8BKxrwU4OZ6
        rx7ftZGtqx1k7czy3u0gCtFUcShItfL/vSECAwEAAaM1MDMwDgYDVR0PAQH/BAQD
        AgWgMBMGA1UdJQQMMAoGCCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwDQYJKoZIhvcN
        AQELBQADggEBAFZ5MPqio6v/4YysGXD+ljgXYTRdqb11FA1iMbEYTT80RGdD5Q/D
        f3MKgYfcq7tQTTYD05u4DEvxIv0VtFK9uyG3W13n/Bt+2Wv4bKuTIJEwpdnbFq8G
        nq2dmRtZL4K+oesdWOUXWcXouCN/M+b12Ik+9NlBbXkIBbz9/ni3i2FgaeN+cfGE
        ik4MjWBclSTMWQCB4jPhkunybzgdpTW+zhFBoZFHdbM3LlMTXJ5LXvWPGCcHy3c+
        XXgc6RG3GfuKWBOUfKJ/ejt6lKSI3vGkKgHjCAoHVsgHFz5CuGK3YISeX54sXA2D
        WXAcqD7v1ddNQKmE2eWZU4+2boBdXKMPtUQ=
        -----END CERTIFICATE-----
        -----BEGIN RSA PRIVATE KEY-----
        MIIEowIBAAKCAQEAmt20OFPQte96eRrWnE8o3drLzaAYy5QdA8TB6v6+E6NCZO6I
        3POVg5DkdEqRGKml8I3DF/VAqm+WUvBbFaHnkx4XCCgh3LWujZD+wIgexF18giAX
        Fe4IFzIA+d/FSu6MIgqFu8TvYYLUwlYYLC0ImM9ZKWoZRk8wUTLyfhZO3OAEPPHu
        09Sugfq0zIRE74vMt7Hq501pjcvjdPAgh1d+Irtfq0W0hyDJIY0ygSgDdPNe+mvw
        37mg+0UzHud/ywGNXzkBnXfhkDOxJW/+k8XUnIBqwFP2f9vjOGvy3+SbwErGvBTg
        5nqvHt+1ka2rHWTtzPLe7SAK0VRxKEi18v+9IQIDAQABAoIBAHS8dBIVk/jgmOhb
        A7T1sq9xMzk/2hDzB+AEW8yA09THts+QQxiSgHyZJqxGXRNDJjOrGImhtGoFDUJd
        rbsjvQTXpLLgVY4iYX6S8oU81jxc3/LSr7Q3JmAdsECqnfR61qT+W4qLy4osbaZD
        8ZqzI4zUl7gxIvYt0RUUG1hSBoZVJo4RjY2JG+YQbD1XYA7RYV9E/OKCH7yX/7T4
        iaIC0vN79yVHGPW6mvYE2INmCOCwabvSRKKqngVS/usm3cgo7LqL4OSqFVpHqf9F
        b+iP5fmzhexwQ/iZeXrv0Y9mGBjX0XpCNZRxiyDLuCXWvZeX/13Ae4/xEOipDw28
        mWBzyvUCgYEAx4PW0gEpVDMX2u3GmlK4WhCV3ArRDnYL0mU6xobQP6MuNYoyHh4L
        9yAt4YCmkTXx2Mc6kN6BJDG+6d0PwKnaMXqQGXjCo5+kdqzIYLCmHgG2Xv6SSqAU
        0KfaOZPgMaOmghGw2QqrBwWo+LEc92Fj4bmpNCjl1rZ+ZdpdGI7FoVsCgYEAxrXc
        goQAffsPIrNFOPKk5sOpN5cBHpaX62SZmTCEZjzvquBzE5b4jb5mbbw/5uFyeH39
        s74LRS24UrVjhJ5VaFKUNmUrAoUihb8eUl5lb3eUYIaKNz81YhDq369G0Ku4KRtc
        mwlnCFqY0G0ijK400lExlZR4ogIA9V1QHZHtSDMCgYEAjQ+p0tD/Z4i4VRHIWVQj
        A4q2ad078f2EXj00USkAE/5LrY8H4ENeMluOFOHg4spBNAOoZMTsiaqiULb7bDyr
        CFCfkWLQOt+kaEPBaJt817peNsvGovyLuvryT8M9v9r03wGjB9GDGnPmA+81i7JP
        7EhYWYiQ+D4PH/RD3hkTogECgYBPAwMyVmCHt2tWRegxc7IEHCrN8to8Gm8/5xl4
        IyWSLYqy7Fp1oKMmYV4DJkZWfLByns5hSSDcGgjfwkZW9kpJmARc+K84ak3G1q6s
        2+IDh43VL8oHm7eTTdzGosBKuu0YU0voTb3NQZDf13VUcPSJ6EUKECZDbP6KkdcI
        Wvz5pwKBgCffF+1xaMPOY8cVomvbnPfEAGes3EZGDq9TE36jDbQkOomBYZJ7+kPp
        mzhg2cUnDBiZ3eRGlfmBmgxIHdSweZQ5yxAyDBLfWxw9yqIX8bLyP7rBzxSy4efu
        6C3Nu7c0HX6rhcgrDFzZKc2TMuV8ihsMPfp5WSOqOIHXhUTfxPdh
        -----END RSA PRIVATE KEY-----
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

{{< confkey type="string" required="yes" >}}

The [redis] host or unix socket path. If utilising an IPv6 literal address it must be enclosed by square brackets and
quoted:

```yaml
host: "[fd00:1111:2222:3333::1]"
```

### port

{{< confkey type="integer" default="6379" required="no" >}}

The port [redis] is listening on.

### username

{{< confkey type="string" required="no" >}}

The username for [redis authentication](https://redis.io/commands/auth). Only supported in [redis] 6.0+, and [redis]
currently offers backwards compatibility with password-only auth. You probably do not need to set this unless you went
through the process of setting up [redis ACLs](https://redis.io/topics/acl).

### password

{{< confkey type="string" required="no" >}}

*__Important Note:__ This can also be defined using a [secret](../methods/secrets.md) which is __strongly recommended__
especially for containerized deployments.*

The password for [redis authentication](https://redis.io/commands/auth).

It's __strongly recommended__ this is a
[Random Alphanumeric String](../miscellaneous/guides.md#generating-a-random-alphanumeric-string) with 64 or more
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

If defined enables [redis] over TLS, and additionally controls the TLS connection validation process. You can see how to
configure the tls section [here](../prologue/common.md#tls-configuration).

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

{{< confkey type="string" required="no (yes if sentinel_username is supplied)" >}}

*__Important Note:__ This can also be defined using a [secret](../methods/secrets.md) which is __strongly recommended__
especially for containerized deployments.*

The password for the [redis sentinel] connection. If specified with sentinel_username, configures Authelia to
authenticate to the Redis Sentinel with ACL-based authentication. Otherwise, this is used for [requirepass]
authentication.

It's __strongly recommended__ this is a
[Random Alphanumeric String](../miscellaneous/guides.md#generating-a-random-alphanumeric-string) with 64 or more
characters and the user password is changed to this value.

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
