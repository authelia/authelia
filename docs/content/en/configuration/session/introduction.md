---
title: "Session"
description: "Session Configuration"
lead: "Configuring the Session / Cookie settings."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  configuration:
    parent: "session"
weight: 105100
toc: true
aliases:
  - /c/session
  - /docs/configuration/session/
---

__Authelia__ relies on session cookies to authenticate users. When the user visits a website of the protected domain
`example.com` for the first time, Authelia detects that there is no cookie for that user. Consequently, Authelia
redirects the user to the login portal through which the user should authenticate to get a cookie which is valid for
`*.example.com`, meaning all websites of the domain. At the next request, Authelia receives the cookie associated to the
authenticated user and can then order the reverse proxy to let the request pass through to the application.

## Configuration

```yaml
session:
  name: authelia_session
  domain: example.com
  same_site: lax
  secret: unsecure_session_secret
  expiration: 1h
  inactivity: 5m
  remember_me_duration:  1M
```

## Providers

There are currently two providers for session storage (three if you count Redis Sentinel as a separate provider):

* Memory (default, stateful, no additional configuration)
* [Redis](redis.md) (stateless).
* [Redis Sentinel](redis.md#high_availability) (stateless, highly available).

### Kubernetes or High Availability

It's important to note when picking a provider, the stateful providers are not recommended in High Availability
scenarios like Kubernetes. Each provider has a note beside it indicating it is *stateful* or *stateless* the stateless
providers are recommended.

## Options

### name

{{< confkey type="string" default="authelia_session" required="no" >}}

The name of the session cookie. By default this is set to authelia_session. It's mostly useful to change this if you are
doing development or running multiple instances of Authelia.

### domain

{{< confkey type="string" required="yes" >}}

The domain the cookie is assigned to protect. This must be the same as the domain Authelia is served on or the root
of the domain. For example if listening on auth.example.com the cookie should be auth.example.com or example.com.

### same_site

{{< confkey type="string" default="lax" required="no" >}}

Sets the cookies SameSite value. Prior to offering the configuration choice this defaulted to None. The new default is
Lax. This option is defined in lower-case. So for example if you want to set it to `Strict`, the value in configuration
needs to be `strict`.

You can read about the SameSite cookie in detail on the
[MDN](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie/SameSite). In short setting SameSite to Lax
is generally the most desirable option for Authelia. None is not recommended unless you absolutely know what you're
doing and trust all the protected apps. Strict is not going to work in many use cases and we have not tested it in this
state but it's available as an option anyway.

### secret

{{< confkey type="string" required="yes" >}}

*__Important Note:__ This can also be defined using a [secret](../methods/secrets.md) which is __strongly recommended__
especially for containerized deployments.*

The secret key used to encrypt session data in Redis.

It's __strongly recommended__ this is a
[Random Alphanumeric String](../miscellaneous/guides.md#generating-a-random-alphanumeric-string) with 64 or more
characters.

### expiration

{{< confkey type="duration" default="1h" required="no" >}}

*__Note:__ This setting uses the [duration notation format](../prologue/common.md#duration-notation-format). Please see
the [common options](../prologue/common.md#duration-notation-format) documentation for information on this format.*

The period of time before the cookie expires and the session is destroyed. This is overriden by
[remember_me_duration](#remember_me_duration) when the remember me box is checked.

### inactivity

{{< confkey type="duration" default="5m" required="no" >}}

*__Note:__ This setting uses the [duration notation format](../prologue/common.md#duration-notation-format). Please see
the [common options](../prologue/common.md#duration-notation-format) documentation for information on this format.*

The period of time the user can be inactive for until the session is destroyed. Useful if you want long session timers
but don't want unused devices to be vulnerable.

### remember_me_duration

{{< confkey type="duration" default="1M" required="no" >}}

*__Note:__ This setting uses the [duration notation format](../prologue/common.md#duration-notation-format). Please see
the [common options](../prologue/common.md#duration-notation-format) documentation for information on this format.*

The period of time before the cookie expires and the session is destroyed when the remember me box is checked. Setting
this to `-1` disables this feature entirely.

## Security

Configuration of this section has an impact on security. You should read notes in
[security measures](../../overview/security/measures.md#session-security) for more information.

