---
layout: default
title: Session
parent: Configuration
nav_order: 8
---

# Session

**Authelia** relies on session cookies to authenticate users. When the user visits
a website of the protected domain `example.com` for the first time, Authelia detects
that there is no cookie for that user. Consequently, Authelia redirects the user
to the login portal through which the user should authenticate to get a cookie which
is valid for `*.example.com`, meaning all websites of the domain.
At the next request, Authelia receives the cookie associated to the authenticated user
and can then order the reverse proxy to let the request pass through to the application.

## Configuration

```yaml
session:
  name: authelia_session
  secret: unsecure_session_secret
  expiration: 1h
  inactivity: 5m
  remember_me_duration:  1M
  domain: example.com
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

### name

The name of the session cookie. By default this is set to authelia_session. It's mostly useful to change this if you are
doing development or running multiple instances of Authelia.

### secret

The secret key used to encrypt session data in Redis. It's recommended this is set using a [secret](./secrets.md).

### expiration

The time in [duration notation format](index.md#duration-notation-format) before the cookie expires and the session is 
destroyed. This is overriden by remember_me_duration when the remember me box is checked.

### inactivity

The time in [duration notation format](index.md#duration-notation-format) the user can be inactive for until the session
is destroyed. Useful if you want long session timers but don't want unused devices to be vulnerable.

### remember_me_duration

The time in [duration notation format](index.md#duration-notation-format) the cookie expires and the session is 
destroyed when the remember me box is checked.

### domain

The domain the cookie is assigned to protect. This must be the same as the domain Authelia is served on or the root
of the domain. For example if listening on auth.example.com the cookie should be auth.example.com or example.com.

### redis

This is a session provider. By default Authelia uses an in-memory provider. Not configuring redis leaves Authelia 
[stateful](../features/statelessness.md). It's important in highly available scenarios to configure this option and
we highly recommend it in production environments. It requires you setup redis as well.

#### host

The redis host or unix socket path. If utilising an IPv6 literal address it must be enclosed by square brackets and quoted:
```yaml
host: "[fd00:1111:2222:3333::1]"
```

#### port

The port redis is listening on.

#### password

The password for redis authentication.

#### database_index

The index number of the redis database, the same value as specified with the redis SELECT command.

## Security

Configuration of this section has an impact on security. You should read notes in
[security measures](../security/measures.md#session-security) for more information.

## Loading a password from a secret instead of inside the configuration

Password can also be defined using a [secret](../secrets.md).

## Redis Sentinel

When using Redis Sentinel, the host specified in the main redis section is added (it will be the first node) to the 
nodes in the high availability section. This however is optional.