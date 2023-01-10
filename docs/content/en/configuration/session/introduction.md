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
  secret: insecure_session_secret

  name: authelia_session
  same_site: lax
  inactivity: 5m
  expiration: 1h
  remember_me: 1M

  cookies:
    - name: authelia_session
      domain: example.com
      same_site: lax
      inactivity: 5m
      expiration: 1h
      remember_me: 1d
```

## Providers

There are currently two providers for session storage (three if you count Redis Sentinel as a separate provider):

* Memory (default, stateful, no additional configuration)
* [Redis](redis.md) (stateless).
* [Redis Sentinel](redis.md#highavailability) (stateless, highly available).

### Kubernetes or High Availability

It's important to note when picking a provider, the stateful providers are not recommended in High Availability
scenarios like Kubernetes. Each provider has a note beside it indicating it is *stateful* or *stateless* the stateless
providers are recommended.

## Options

### secret

{{< confkey type="string" required="yes" >}}

*__Important Note:__ This can also be defined using a [secret](../methods/secrets.md) which is __strongly recommended__
especially for containerized deployments.*

The secret key used to encrypt session data in Redis.

It's __strongly recommended__ this is a
[Random Alphanumeric String](../../reference/guides/generating-secure-values.md#generating-a-random-alphanumeric-string) with 64 or more
characters.

### domain

{{< confkey type="string" required="no" >}}

_**Deprecation Notice:** This option is deprecated. See the [cookies](#cookies) section instead._

The domain the cookie is assigned to protect. This must be the same as the domain Authelia is served on or the root
of the domain. For example if listening on auth.example.com the cookie should be auth.example.com or example.com.

This value automatically maps to a single cookies configuration using the default values. It cannot be assigned at the
same time as a `cookies` configuration.

### name

{{< confkey type="string" default="authelia_session" required="no" >}}

The default `name` value for all [cookies](#cookies) configurations.

### same_site

{{< confkey type="string" default="lax" required="no" >}}

The default `same_site` value for all `cookies` configurations.

### inactivity

{{< confkey type="duration" default="5m" required="no" >}}

*__Note:__ This setting uses the [duration notation format](../prologue/common.md#duration-notation-format). Please see
the [common options](../prologue/common.md#duration-notation-format) documentation for information on this format.*

The default `inactivity` value for all [cookies](#cookies) configurations.

### expiration

{{< confkey type="duration" default="1h" required="no" >}}

*__Note:__ This setting uses the [duration notation format](../prologue/common.md#duration-notation-format). Please see
the [common options](../prologue/common.md#duration-notation-format) documentation for information on this format.*

The default `expiration` value for all [cookies](#cookies) configurations.

### remember_me

{{< confkey type="duration" default="1M" required="no" >}}

*__Note:__ This setting uses the [duration notation format](../prologue/common.md#duration-notation-format). Please see
the [common options](../prologue/common.md#duration-notation-format) documentation for information on this format.*

The default `remember_me` value for all [cookies](#cookies) configurations.

### cookies

The list of specific cookie domains that Authelia is configured to handle. Domains not properly configured will
automatically be denied by Authelia. The list allows administrators to define multiple session cookie domain
configurations with individual settings.

#### name

{{< confkey type="string" required="no" >}}

*__Default Value:__ This option takes its default value from the [name](#name) setting above.*

The name of the session cookie. By default this is set to the `name` value in the main session configuration section.

#### domain

{{< confkey type="string" required="yes" >}}

The domain the cookie is assigned to protect. This must be the same as the domain Authelia is served on or the root
of the domain, and consequently if the [portal_url](#portal_url) is configured must be able to read and write cookies
for the domain. For example if listening on `auth.example.com` the cookie should be either `auth.example.com` or
`example.com`.

Please note most good DynamicDNS solutions fall into a specially protected group of domains and browsers do not allow
you to write cookies for the root domain. i.e. if you have been assigned `john.duckdns.org` you can't use `duckdns.org`
for the domain value as browsers will not allow `john.duckdns.org` to read or write cookies for `duckdns.org`.

Consequently, if you have `john.duckdns.org` and `mary.duckdns.org` you cannot share cookies between these domains.

#### portal_url

{{< confkey type="string" required="no" >}}

*__Note:__ The AuthRequest implementation does not support redirection control on the authorization server. This means
that the `portal_url` option is ineffectual for both NGINX and HAProxy, or any other proxy which uses the AuthRequest
implementation.*

This is a completely optional URL which is the root URL of your Authelia installation for this cookie domain which can
be used to generate the appropriate redirection for proxies which support this.

If this option is absent you must use the appropriate query parameter or header for your relevant proxy.

#### same_site

{{< confkey type="string" required="no" >}}

*__Default Value:__ This option takes its default value from the [same_site](#samesite) setting above.*

Sets the cookies SameSite value. Prior to offering the configuration choice this defaulted to None. The new default is
Lax. This option is defined in lower-case. So for example if you want to set it to `Strict`, the value in configuration
needs to be `strict`.

You can read about the SameSite cookie in detail on the
[MDN](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie/SameSite). In short setting SameSite to Lax
is generally the most desirable option for Authelia. None is not recommended unless you absolutely know what you're
doing and trust all the protected apps. Strict is not going to work in many use cases and we have not tested it in this
state but it's available as an option anyway.

#### inactivity

{{< confkey type="duration" required="no" >}}

*__Default Value:__ This option takes its default value from the [inactivity](#inactivity) setting above.*

*__Note:__ This setting uses the [duration notation format](../prologue/common.md#duration-notation-format). Please see
the [common options](../prologue/common.md#duration-notation-format) documentation for information on this format.*

The period of time the user can be inactive for until the session is destroyed. Useful if you want long session timers
but don't want unused devices to be vulnerable.

#### expiration

{{< confkey type="duration" required="no" >}}

*__Default Value:__ This option takes its default value from the [expiration](#expiration) setting above.*

*__Note:__ This setting uses the [duration notation format](../prologue/common.md#duration-notation-format). Please see
the [common options](../prologue/common.md#duration-notation-format) documentation for information on this format.*

The period of time before the cookie expires and the session is destroyed. This is overriden by
[remember_me](#rememberme) when the remember me box is checked.

#### remember_me

{{< confkey type="duration" required="no" >}}

*__Default Value:__ This option takes its default value from the [remember_me](#rememberme) setting above.*

*__Note:__ This setting uses the [duration notation format](../prologue/common.md#duration-notation-format). Please see
the [common options](../prologue/common.md#duration-notation-format) documentation for information on this format.*

The period of time before the cookie expires and the session is destroyed when the remember me box is checked. Setting
this to `-1` disables this feature entirely for this session cookie domain.

## Security

Configuration of this section has an impact on security. You should read notes in
[security measures](../../overview/security/measures.md#session-security) for more information.

