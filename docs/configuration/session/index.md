---
layout: default
title: Session
parent: Configuration
nav_order: 13
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
  same_site: lax
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
<div markdown="1">
type: string
{: .label .label-config .label-purple }
default: authelia_session
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

The name of the session cookie. By default this is set to authelia_session. It's mostly useful to change this if you are
doing development or running multiple instances of Authelia.

### domain
<div markdown="1">
type: string
{: .label .label-config .label-purple }
required: yes
{: .label .label-config .label-red }
</div>

The domain the cookie is assigned to protect. This must be the same as the domain Authelia is served on or the root
of the domain. For example if listening on auth.example.com the cookie should be auth.example.com or example.com.

### same_site
<div markdown="1">
type: string
{: .label .label-config .label-purple }
default: lax
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Sets the cookies SameSite value. Prior to offering the configuration choice this defaulted to None. The new default is
Lax. This option is defined in lower-case. So for example if you want to set it to `Strict`, the value in configuration
needs to be `strict`.

You can read about the SameSite cookie in detail on the 
[MDN](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie/SameSite). In short setting SameSite to Lax
is generally the most desirable option for Authelia. None is not recommended unless you absolutely know what you're
doing and trust all the protected apps. Strict is not going to work in many use cases and we have not tested it in this
state but it's available as an option anyway.

### secret
<div markdown="1">
type: string
{: .label .label-config .label-purple }
required: yes
{: .label .label-config .label-red }
</div>

The secret key used to encrypt session data in Redis. It's recommended this is set using a [secret](../secrets.md).

### expiration
<div markdown="1">
type: string (duration)
{: .label .label-config .label-purple }
default: 1h
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

The time in [duration notation format](../index.md#duration-notation-format) before the cookie expires and the session
is destroyed. This is overriden by remember_me_duration when the remember me box is checked.

### inactivity
<div markdown="1">
type: string (duration)
{: .label .label-config .label-purple }
default: 5m
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

The time in [duration notation format](../index.md#duration-notation-format) the user can be inactive for until the
session is destroyed. Useful if you want long session timers but don't want unused devices to be vulnerable.

### remember_me_duration
<div markdown="1">
type: string (duration)
{: .label .label-config .label-purple }
default: 1M
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

The time in [duration notation format](../index.md#duration-notation-format) the cookie expires and the session is
destroyed when the remember me box is checked.

## Security

Configuration of this section has an impact on security. You should read notes in
[security measures](../../security/measures.md#session-security) for more information.

## Loading a password from a secret instead of inside the configuration

Password can also be defined using a [secret](../secrets.md).
