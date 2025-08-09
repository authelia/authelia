---
title: "Uptime Kuma"
description: "Integrating Uptime Kuma status monitors with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/uptime-kuma/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Uptime Kuma | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Uptime Kuma with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
- [Uptime Kuma]
  - [v1.23.11](https://github.com/louislam/uptime-kuma/releases/tag/1.23.11)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://uptime-kuma.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `uptime-kuma`
- __Client Secret:__ `insecure_secret`
- __Secured Resource URL:__ `https://application.{{< sitevar name="domain" nojs="example.com" >}}/`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

### Important Notes

This implementation has several facets which must be configured as a security precaution. It's advised people read the
[OAuth 2.0 Bearer Token Usage](../oauth-2.0-bearer-token-usage.md) integration guide in addition to this guide to
properly understand this process.

For example this guide has a requirement to adapt a fairly new and special section of Authelia. It's important to take
the time to understand it before you attempt to do it. Some notes about this are below.

1. The `implementation` value of the server authz endpoints section must be the appropriate implementation for your
   proxy.
2. The `endpoint_name` in the server authz endpoints section is the actual name of the endpoint which must be configured
   in your proxy for the Forwarded / Redirected Authorization Flow:
   1. You can customize this name but by configuring just one all other default endpoints for authorization are removed
      such as `/api/verify`, `/api/authz/forward-auth`, etc.
   2. The name represents the endpoint path, for example setting `endpoint_name` will configure an endpoint at
      `/api/authz/endpoint_name`.
3. The use of the `HeaderAuthorization` strategy and how it's configured here accepts bearer tokens in the Authorization
   header as one of the possible ways to authenticate, but still allows cookie-based authorization.

See more information about the server authz endpoints section in the
[Configuration Guide](../../../configuration/miscellaneous/server-endpoints-authz.md) and
[Reference Guide](../../../reference/guides/proxy-authorization.md).

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Uptime Kuma] which will
operate with the above example:

```yaml {title="configuration.yml"}
server:
  endpoints:
    authz:
      endpoint_name:
        implementation: ''
        authn_strategies:
          - name: 'HeaderAuthorization'
            schemes:
              - 'Basic'
              - 'Bearer'
          - name: 'CookieSession'
access_control:
  rules:
    - domain:
      - 'application.{{< sitevar name="domain" nojs="example.com" >}}'
      subject: 'oauth2:client:uptime-kuma'
      policy: 'one_factor'
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'uptime-kuma'
        client_name: 'Uptime Kuma'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        require_pkce: false
        pkce_challenge_method: ''
        requested_audience_mode: 'implicit'
        scopes:
          - 'authelia.bearer.authz'
        audience:
          - 'https://application.{{< sitevar name="domain" nojs="example.com" >}}/'
        grant_types:
          - 'client_credentials'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

Notes:

- The configuration has a [requested_audience_mode] value of `implicit` which is used to automatically grant all audiences the client is permitted to request, the default is `explicit` which does not do this and the client must also request the audience using the `audience` form parameter. As [Uptime Kuma] does not currently support this configuration is required.
- The `audience` (or multiple) is the endpoints of the secured resource you want to monitor using [Uptime Kuma].
- The `oauth2:client:uptime-kuma` is a special subject which refers to the `uptime-kuma` client id and allows Access
  Tokens granted via the Client Credentials Flow to be used provided they were granted to this client.

### Application

To configure [Uptime Kuma] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Uptime Kuma] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Create a new status monitor or configure an existing one
2. Choose monitor type e.g. HTTP(s) Keyword and set a keyword you want to find
3. Set the URL to be monitored (this corresponds to the `audience` parameter in Authelia)
4. Configure the following options:
   - Method: `OAuth2: Client Credentials`
   - Authentication Method: `Authorization Header`
   - OAuth Token URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token`
   - Client ID: `uptime-kuma`
   - Client Secret: `insecure_secret`
   - OAuth Scope: `authelia.bearer.authz`

See the following screenshot for an authentication example of the above:

{{< figure src="uptime-kuma-authentication.png" alt="Uptime Kuma Authentication example" width="300" >}}


[Authelia]: https://www.authelia.com
[Uptime Kuma]: https://github.com/louislam/uptime-kuma
[OpenID Connect 1.0]: ../introduction.md
[requested_audience_mode]: ../../configuration/identity-providers/openid-connect/clients/#requested_audience_mode
[Server Authz Endpoints]: ../../configuration/miscellaneous/server-endpoints-authz/
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
