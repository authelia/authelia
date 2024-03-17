---
title: "Server Authz Endpoints"
description: "Configuring the Server Authz Endpoint Settings."
summary: "Authelia supports several authorization endpoints on the internal web server. This section describes how to configure and tune them."
date: 2023-01-25T20:36:40+11:00
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
          - name: 'CookieSession'
      ext-authz:
        implementation: 'ExtAuthz'
        authn_strategies:
          - name: 'HeaderAuthorization'
            schemes:
              - 'Basic'
          - name: 'CookieSession'
      auth-request:
        implementation: 'AuthRequest'
        authn_strategies:
          - name: 'HeaderAuthRequestAuthorization'
            schemes:
              - 'Basic'
          - name: 'CookieSession'
      legacy:
        implementation: 'Legacy'
        authn_strategies:
          - name: 'HeaderLegacy'
          - name: 'CookieSession'
```

## Name

{{< confkey type="string" required="yes" >}}

The first level under the `authz` directive is the name of the endpoint. In the example these names are `forward-auth`,
`ext-authz`, `auth-request`, and `legacy`.

The name correlates with the path of the endpoint. All endpoints start with `/api/authz/`, and end with the name. In the
example the `forward-auth` endpoint has a full path of `/api/authz/forward-auth`.

Valid characters for the name are alphanumeric as well as `-` and `_`. They MUST start AND end with an
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
