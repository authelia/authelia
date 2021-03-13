---
layout: default
title: Session
parent: Configuration
nav_order: 8
has_children: true
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
  domain: example.com
  secret: unsecure_session_secret
  expiration: 1h
  inactivity: 5m
  remember_me_duration:  1M
```

## Providers

There are currently two providers for session storage (three if you count Redis Sentinel as a separate provider):
* Memory (default, stateful, no additional configuration)
* [Redis](./redis.md) (stateless).
* [Redis Sentinel](./redis.md#high_availability) (stateless, highly available).

### Kubernetes or High Availability

It's important to note when picking a provider, the stateful providers are not recommended in High Availability
scenarios like Kubernetes. Each provider has a note beside it indicating it is *stateful* or *stateless* the stateless
providers are recommended. 

## Options

### name

The name of the session cookie. By default this is set to authelia_session. It's mostly useful to change this if you are
doing development or running multiple instances of Authelia.

### domain

The domain the cookie is assigned to protect. This must be the same as the domain Authelia is served on or the root
of the domain. For example if listening on auth.example.com the cookie should be auth.example.com or example.com.

### secret

The secret key used to encrypt session data in Redis. It's recommended this is set using a [secret](../secrets.md).

### expiration

The time in [duration notation format](../index.md#duration-notation-format) before the cookie expires and the session 
is destroyed. This is overriden by remember_me_duration when the remember me box is checked.

### inactivity

The time in [duration notation format](../index.md#duration-notation-format) the user can be inactive for until the 
session is destroyed. Useful if you want long session timers but don't want unused devices to be vulnerable.

### remember_me_duration

The time in [duration notation format](../index.md#duration-notation-format) the cookie expires and the session is
destroyed when the remember me box is checked.

## Security

Configuration of this section has an impact on security. You should read notes in
[security measures](../../security/measures.md#session-security) for more information.

## Loading a password from a secret instead of inside the configuration

Password can also be defined using a [secret](../secrets.md).
