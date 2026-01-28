---
title: "Session"
description: "Session Configuration"
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

__Authelia__ relies on session cookies to authorize user access to various protected websites. This section configures
the session cookie behavior and the domains which Authelia can service authorization requests for.

## Variables

Some of the values within this page can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
session:
  secret: 'insecure_session_secret'
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

There are currently three providers for session storage (four if you count Redis Sentinel as a separate provider):

* Memory (default, stateful, no additional configuration)
* [File](file.md) (stateful, no external dependencies).
* [Redis](redis.md) (stateless).
* [Redis Sentinel](redis.md#high_availability) (stateless, highly available).

### Kubernetes or High Availability

It's important to note when picking a provider, the stateful providers are not recommended in High Availability
scenarios like Kubernetes. Each provider has a note beside it indicating it is *stateful* or *stateless* the stateless
providers are recommended.

## Options

This section describes the individual configuration options.

### secret

{{< confkey type="string" required="yes" secret="yes" >}}

The secret key used to encrypt session data in the [Redis](redis.md) or [File](file.md) session providers.

It's __strongly recommended__ this is a
[Random Alphanumeric String](../../reference/guides/generating-secure-values.md#generating-a-random-alphanumeric-string) with 64 or more
characters.

### name

{{< confkey type="string" default="authelia_session" required="no" >}}

The default `name` value for all [cookies](#cookies) configurations.

### same_site

{{< confkey type="string" default="lax" required="no" >}}

The default `same_site` value for all `cookies` configurations.

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

The domain the session cookie is assigned to protect. This must be the same as the domain Authelia is served on or the
root of the domain, and consequently if the [authelia_url](#authelia_url) is configured must be able to read and write
cookies for this domain.

For example if Authelia is accessible via the URL
`https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}` the domain should be either
`{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}` or `{{< sitevar name="domain" nojs="example.com" >}}`.

The value must not match a domain on the [Public Suffix List] as browsers do not allow
websites to write cookies for these domains. This includes most Dynamic DNS services such as `duckdns.org`. You should
use your domain instead of `duckdns.org` for this value, for example `example.duckdns.org`.

Consequently, if you have `example.duckdns.org` and `example-auth.duckdns.org` you cannot share cookies between these
domains.

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

*__Default Value:__ This option takes its default value from the [name](#name) setting above.*

The name of the session cookie. By default this is set to the `name` value in the main session configuration section.

#### same_site

{{< confkey type="string" required="no" >}}

*__Default Value:__ This option takes its default value from the [same_site](#same_site) setting above.*

Sets the cookies SameSite value. Prior to offering the configuration choice this defaulted to None. The new default is
Lax. This option is defined in lower-case. So for example if you want to set it to `Strict`, the value in configuration
needs to be `strict`.

You can read about the SameSite cookie in detail on the
[MDN](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie/SameSite). In short setting SameSite to Lax
is generally the most desirable option for Authelia. None is not recommended unless you absolutely know what you're
doing and trust all the protected apps. Strict is not going to work in many use cases and we have not tested it in this
state but it's available as an option anyway.

#### inactivity

{{< confkey type="string,integer" syntax="duration" required="no" >}}

*__Default Value:__ This option takes its default value from the [inactivity](#inactivity) setting above.*

The period of time the user can be inactive for until the session is destroyed. Useful if you want long session timers
but don't want unused devices to be vulnerable.

#### expiration

{{< confkey type="string,integer" syntax="duration" required="no" >}}

*__Default Value:__ This option takes its default value from the [expiration](#expiration) setting above.*

The period of time before the cookie expires and the session is destroyed. This is overridden by
[remember_me](#remember_me) when the remember me box is checked.

#### remember_me

{{< confkey type="string,integer" syntax="duration" required="no" >}}

*__Default Value:__ This option takes its default value from the [remember_me](#remember_me) setting above.*

The period of time before the cookie expires and the session is destroyed when the remember me box is checked. Setting
this to `-1` disables this feature entirely for this session cookie domain.

## Security

Configuration of this section has an impact on security. You should read notes in
[security measures](../../overview/security/measures.md#session-security) for more information.

