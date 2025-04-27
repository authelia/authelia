---
title: "Chronograf"
description: "Integrating Chronograf with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-04-26T11:03:16+00:00
draft: false
images: []
weight: 620
toc: true
support:
  level: community
  versions: true
  integration: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.1](https://github.com/authelia/authelia/releases/tag/v4.39.1)
- [Chronograf]
  - [v1.10.7](https://docs.influxdata.com/chronograf/v1/about_the_project/release-notes/#v1107)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://chronograf.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `chronograf`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Chronograf] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'chronograf'
        client_name: 'Chronograf'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://chronograf.{{< sitevar name="domain" nojs="example.com" >}}/oauth/authelia/callback'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [Chronograf] there is one method, using the [Environment Variables](#environment-variables).

#### Environment Variables

To configure [Chronograf] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment variables:

##### Standard

```shell {title=".env"}
PUBLIC_URL=https://chronograf.{{< sitevar name="domain" nojs="example.com" >}}
TOKEN_SECRET=insecure_random_secret
GENERIC_CLIENT_ID=chronograf
GENERIC_CLIENT_SECRET=insecure_secret
GENERIC_AUTH_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization
GENERIC_TOKEN_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token
JWKS_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/jwks.json
GENERIC_API_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo
GENERIC_API_KEY=email
GENERIC_SCOPES=openid,email,profile
GENERIC_NAME=authelia
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  chronograf:
    environment:
      PUBLIC_URL: 'https://chronograf.{{< sitevar name="domain" nojs="example.com" >}}'
      TOKEN_SECRET: 'insecure_random_secret'
      GENERIC_CLIENT_ID: 'chronograf'
      GENERIC_CLIENT_SECRET: 'insecure_secret'
      GENERIC_AUTH_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization'
      GENERIC_TOKEN_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token'
      JWKS_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/jwks.json'
      GENERIC_API_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo'
      GENERIC_API_KEY: 'email'
      GENERIC_SCOPES: 'openid,email,profile'
      GENERIC_NAME: 'authelia'
```

## See Also

- [Chronograf Security OAuth 2.0 Documentation](https://docs.influxdata.com/chronograf/v1/administration/managing-security/#configure-chronograf-to-use-any-oauth-20-provider)

[Authelia]: https://www.authelia.com
[Chronograf]: https://www.influxdata.com/time-series-platform/chronograf/
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
