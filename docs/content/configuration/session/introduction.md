---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Session"
description: "Configuring the Authelia session cookies including the secret, cookie name, domain settings, expiration, inactivity timeout, and remember me duration."
summary: "Configuring the Session / Cookie settings."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 106100
toc: true
aliases:
  - /c/session
  - /docs/configuration/session/
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

**Authelia** relies on session cookies to authorize user access to various protected websites. This section configures
the session cookie behavior and the domains which Authelia can service authorization requests for.

## Variables

Some of the values within this page can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
session:
  secret: 'insecure_session_secret'
  storage: 'internal'
  name: 'authelia_session'
  same_site: 'lax'
  inactivity: '5m'
  expiration: '1h'
  remember_me: '1M'
  cookies:
    - domain: '{{< sitevar name="domain" nojs="example.com" >}}'
      authelia_url: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      default_redirection_url: 'https://www.{{< sitevar name="domain" nojs="example.com" >}}'
      name: 'authelia_session'
      same_site: 'lax'
      inactivity: '5m'
      expiration: '1h'
      remember_me: '1d'
```

## Providers

Session data is persisted to one of two backends, selected with the [storage](#storage) option:

- `internal` (default): the configured [storage](../storage/introduction.md) provider. This is stateless when using the
  [PostgreSQL](../storage/postgres.md) or [MySQL](../storage/mysql.md) providers, and stateful when using the
  [SQLite](../storage/sqlite.md) provider.
- `cache`: the configured [cache](../cache/introduction.md) provider, which is one of [Redis](../cache/redis.md),
  [Redis Sentinel](../cache/redis-sentinel.md), or [Redis Cluster](../cache/redis-cluster.md) (stateless).

### Kubernetes or High Availability

It's important to note when picking a backend, the stateful backends are not recommended in High Availability
scenarios like Kubernetes. Each backend has a note beside it indicating it is _stateful_ or _stateless_ the stateless
backends are recommended. See [statelessness](../../overview/authorization/statelessness.md) for more information.

## Options

This section describes the individual configuration options.

### secret

{{< confkey type="string" required="situational" secret="yes" >}}

The secret used to derive the key which encrypts session data before it's persisted to the [storage](#storage) backend.

This option is required when [storage](#storage) is `cache`. When [storage](#storage) is `internal` and this option is
not configured, the storage [encryption_key](../storage/introduction.md#encryption_key) is used to derive the session
encryption key instead and a warning is logged. It's strongly recommended this option is explicitly configured
regardless, as changing the storage [encryption_key](../storage/introduction.md#encryption_key) would otherwise
invalidate every session.

It's **strongly recommended** this is a
[Random Alphanumeric String](../../reference/guides/generating-secure-values.md#generating-a-random-alphanumeric-string) with 64 or more
characters.

### storage

{{< confkey type="string" default="internal" required="no" >}}

The backend session data is persisted to. Must be one of `internal` or `cache`. See [Providers](#providers) for more
information about each backend.

### name

{{< confkey type="string" default="authelia_session" required="no" >}}

The default `name` value for all [cookies](#cookies) configurations.

### same_site

{{< confkey type="string" default="lax" required="no" >}}

The default `same_site` value for all `cookies` configurations. See the cookie [same_site](#same_site-1) option for
what this value applies to.

### inactivity

{{< confkey type="string,integer" syntax="duration" default="5 minutes" required="no" >}}

The default `inactivity` value for all [cookies](#cookies) configurations.

### expiration

{{< confkey type="string,integer" syntax="duration" default="1 hour" required="no" >}}

The default `expiration` value for all [cookies](#cookies) configurations.

### remember_me

{{< confkey type="string,integer" syntax="duration" default="1 month" required="no" >}}

The default `remember_me` value for all [cookies](#cookies) configurations.

### cookies

The list of specific cookie domains that Authelia is configured to handle. Domains not properly configured will
automatically be denied by Authelia. The list allows administrators to define multiple session cookie domain
configurations with individual settings.

#### domain

{{< confkey type="string" required="yes" >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
Browsers have rules regarding which cookie domains a website can write. In particular the
[Public Suffix List](https://publicsuffix.org/list/) usually contains domains which are not permitted.
{{< /callout >}}

The domain the session cookie is assigned to protect.

This value:

1. Must be standard use domain i.e. the hostname portion of a URL (exclude the port) which a globally trusted
   Certificate Authority is willing to issue X.509 Certificates for (the use of TLS is a strict requirement).
2. Must either be a valid suffix of or exactly match the hostname portion of the [authelia_url](#authelia_url).
3. Must include at least one domain segment i.e. period.
4. Must not exactly match a public suffix domain i.e. a domain on the [Public Suffix List] like `duckdns.org` and should
   instead be something like `example.duckdns.org` as browsers will refuse to set cookies at all for these domains.
   Consequently, if you have `example.duckdns.org` and `example-auth.duckdns.org` you cannot share cookies between these
   domains.
5. Must not be a special use domain like `localhost` as browsers do not allow sites to set the cookies Authelia
   requires to operate properly for these domains.

For example if Authelia is accessible via the URL
`https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}` the
domain should be either
`{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}` or
`{{< sitevar name="domain" nojs="example.com" >}}`.

[Public Suffix List]: https://publicsuffix.org/list/

#### authelia_url

{{< confkey type="string" required="yes" >}}

This is a required URL which is the root URL of your Authelia installation for this cookie domain which can
be used to generate the appropriate redirection URL when authentication is required. This URL must:

1. Be able to read and write cookies for the configured [domain](#domain-1).
2. Use the `https://` scheme.
3. Include the path if relevant (i.e. `https://{{< sitevar name="domain" nojs="example.com" >}}/authelia` rather than `https://{{< sitevar name="domain" nojs="example.com" >}}` if you're using
   the [server address option](../miscellaneous/server.md#address) of `authelia` to specify a subpath and if the
   Authelia portal is inaccessible from `https://{{< sitevar name="domain" nojs="example.com" >}}`).

The appropriate query parameter or header for your relevant proxy can override this behavior.

#### default_redirection_url

{{< confkey type="string" required="no" >}}

This is a completely optional URL which is used as the redirection location when visiting Authelia directly. This option
deprecates the global [default_redirection_url](../miscellaneous/introduction.md#default_redirection_url) option. This URL
must:

1. Be able to read and write cookies for the configured [domain](#domain-1).
2. Use the `https://` scheme.
3. Not be the same as the [authelia_url](#authelia_url)

If this option is absent you must use the appropriate query parameter or header for your relevant proxy.

#### name

{{< confkey type="string" required="no" >}}

_**Default Value:** This option takes its default value from the [name](#name) setting above._

The name of the session cookie. By default this is set to the `name` value in the main session configuration section.

#### same_site

{{< confkey type="string" required="no" >}}

_**Default Value:** This option takes its default value from the [same_site](#same_site) setting above._

Sets the SameSite value of the authorization cookie. This option is defined in lower-case. So for example if you want
to set it to `Strict`, the value in configuration needs to be `strict`.

Two cookies are issued for every session once the user has signed in:

- The session cookie, which uses the configured [name](#name-1). It is always issued with a SameSite value of Strict
  regardless of this option, and it is the only cookie the portal and the API accept. It is accepted on every endpoint
  except the [authorization endpoints] (`/api/authz/*` and the legacy `/api/verify`).
- The authorization cookie, which uses the configured name with the suffix `_authz`. It is issued with the SameSite
  value configured here, and it is only accepted by the [authorization endpoints] used by your proxy
  (`/api/authz/*` and the legacy `/api/verify`) and by no other endpoint. It can't be
  used to sign in, to change credentials, or to perform any other action against the API. The only exception is that
  signing in or signing out without a session cookie ends the session the authorization cookie belongs to.

The authorization cookie is linked to the session. It is replaced whenever the session cookie is replaced and stops
working as soon as the session is destroyed. When the session cookie is valid but the authorization cookie is missing
or no longer matches the session, Authelia issues the authorization cookie again the next time the user visits the
portal.

Requests to the authorization endpoints record activity for the [inactivity](#inactivity-1) check, but they never
extend the [expiration](#expiration-1) of the session and don't reissue either cookie. Both cookies are issued with a
lifetime of at least 30 days, or the [remember_me](#remember_me-1) or [expiration](#expiration-1) value when either is
longer. The cookies only identify the session: whether the session is still valid, including whether it has expired,
is decided by Authelia alone, so a cookie which outlives its session is simply rejected.

You can read about the SameSite cookie in detail on the
[MDN](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie/SameSite). In short setting SameSite to Lax
is generally the most desirable option for Authelia. None is not recommended unless you absolutely know what you're
doing and trust all the protected apps. Strict prevents a user who follows a link from another site to a protected
application from being recognized until they navigate within the application.

Because the session cookie is always Strict, it is not sent when a user arrives at Authelia from another site, such
as an OpenID Connect 1.0 relying party hosted on a different site. Authelia handles this for authorization requests
by continuing the request from a page it serves itself, after which the session cookie is sent. This only applies
when the user's browser performs an ordinary page navigation to Authelia; a relying party which performs silent
renewal in a hidden frame from another site, or which makes the request by script from another site, still receives
the `login_required` error. See the
[OpenID Connect 1.0 frequently asked questions](../../integration/openid-connect/frequently-asked-questions.md#why-does-an-authorization-request-with-promptnone-return-login_required-when-the-user-is-signed-in)
for the details and the limits of this.

[authorization endpoints]: ../miscellaneous/server-endpoints-authz.md

#### inactivity

{{< confkey type="string,integer" syntax="duration" required="no" >}}

_**Default Value:** This option takes its default value from the [inactivity](#inactivity) setting above._

The period of time the user can be inactive for until the session is destroyed. Useful if you want long session timers
but don't want unused devices to be vulnerable.

#### expiration

{{< confkey type="string,integer" syntax="duration" required="no" >}}

_**Default Value:** This option takes its default value from the [expiration](#expiration) setting above._

The period of time before the session is destroyed. It is extended whenever the portal saves the session, such as
when the user signs in, but never by requests to the authorization endpoints, so a user who only uses protected
applications is asked to sign in again once it elapses. Use [inactivity](#inactivity-1) to end sessions which are no
longer being used. The cookies themselves are issued with a longer lifetime, see [same_site](#same_site-1) for more
information. This is overridden by [remember_me](#remember_me) when the remember me box is checked.

#### remember_me

{{< confkey type="string,integer" syntax="duration" required="no" >}}

_**Default Value:** This option takes its default value from the [remember_me](#remember_me) setting above._

The period of time before the session is destroyed when the remember me box is checked. Setting
this to `-1` disables this feature entirely for this session cookie domain.

It should be noted that the Passkey Login Flow currently requires remembering sessions if enabled. This is because the
WebAuthn dialog is disruptive to the UI control, and several future features including Conditional Mediation and
essentially OpenID Connect 1.0 Relying Party logins feel clunky having to click the checkbox first.

As such the ability to control the remembered state in these scenarios will come as a feature which changes how
remember me operates today, where it will be a question asked of the user just after they complete any initial login.

## Security

Configuration of this section has an impact on security. You should read notes in
[security measures](../../overview/security/measures.md#session-security) for more information.
