---
title: "Server Authz Endpoints"
description: "Configuring the Server Authz Endpoint Settings."
summary: "Authelia supports several authorization endpoints on the internal web server. This section describes how to configure and tune them."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
menu:
configuration:
parent: "miscellaneous"
weight: 199210
toc: true
aliases:
  - /c/authz
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Configuration

{{< config-alert-example >}}

```yaml {title=configuration.yml}
server:
  endpoints:
    authz:
      forward-auth:
        implementation: 'ForwardAuth'
        authn_strategies:
          - name: 'HeaderAuthorization'
            schemes:
              - 'Basic'
            scheme_basic_cache_lifespan: 0
          - name: 'CookieSession'
      ext-authz:
        implementation: 'ExtAuthz'
        authn_strategies:
          - name: 'HeaderAuthorization'
            schemes:
              - 'Basic'
            scheme_basic_cache_lifespan: 0
          - name: 'CookieSession'
      auth-request:
        implementation: 'AuthRequest'
        authn_strategies:
          - name: 'HeaderAuthorization'
            schemes:
              - 'Basic'
            scheme_basic_cache_lifespan: 0
          - name: 'CookieSession'
      legacy:
        implementation: 'Legacy'
```

## name

{{< confkey type="string" required="yes" >}}

The first level under the `authz` directive is the name of the endpoint. In the example these names are `forward-auth`,
`ext-authz`, `auth-request`, and `legacy`.

The name correlates with the path of the endpoint. All endpoints start with `/api/authz/`, and end with the name. In the
example the `forward-auth` endpoint has a full path of `/api/authz/forward-auth`.

Valid characters for the name are alphanumeric as well as `-`, `_`, and `/`. They MUST start AND end with an
alphanumeric character.

### implementation

{{< confkey type="string" required="yes" >}}

The underlying implementation for the endpoint. Valid case-sensitive values are `ForwardAuth`, `ExtAuthz`,
`AuthRequest`, and `Legacy`. Read more about the implementations in the
[reference guide](../../reference/guides/proxy-authorization.md#implementations).

### authn_strategies

{{< confkey type="list" required="no" >}}

A list of authentication strategies and their configuration options. These strategies are in order, and the first one
which succeeds is used. Failures other than lacking the sufficient information in the request to perform the strategy
immediately short-circuit the authentication, otherwise the next strategy in the list is attempted.

#### name

{{< confkey type="string" required="yes" >}}

The name of the strategy. Valid case-sensitive values are `CookieSession`, `HeaderAuthorization`,
`HeaderProxyAuthorization`, `HeaderAuthRequestProxyAuthorization`, and `HeaderLegacy`. Read more about the strategies in
the [reference guide](../../reference/guides/proxy-authorization.md#authn-strategies).

#### schemes

{{< confkey type="list(string)" default="Basic" required="no" >}}

The list of schemes allowed on this endpoint. Options are `Basic`, and `Bearer`. This option is only applicable to the
`HeaderAuthorization`, `HeaderProxyAuthorization`, and `HeaderAuthRequestProxyAuthorization` strategies and unavailable
with the `legacy` endpoint which only uses `Basic`.

#### scheme_basic_cache_lifespan

{{< confkey type="string,integer" syntax="duration" default="0 seconds" required="no" >}}

The lifespan to cache username and password combinations when using the `Basic` scheme. This option enables the use
of the caching which is completely disabled by default. This option must only be used when the `Basic` scheme is
configured, and like all new options may not be used with the `Legacy` implementation.
