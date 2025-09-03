---
title: "Karakeep"
description: "Integrating Karakeep with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-04-19T01:05:09+00:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/karakeep/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Karakeep | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Karakeep with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.8](https://github.com/authelia/authelia/releases/tag/v4.39.8)
- [Karakeep] (previously named Hoarder)
  - [v0.26.0](https://github.com/karakeep-app/karakeep/releases/tag/v0.26.0)

{{% oidc-common bugs="claims-hydration" %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://karakeep.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `karakeep`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [karakeep] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'karakeep'
        client_name: 'Karakeep'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng' # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://karakeep.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/callback/custom'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

#### Configuration Escape Hatch

{{% oidc-escape-hatch-claims-hydration client_id="karakeep" claims="email" %}}

### Application

To configure [karakeep] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [karakeep] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment variables:

##### Standard

```shell {title=".env"}
OAUTH_WELLKNOWN_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration
OAUTH_CLIENT_ID=karakeep
OAUTH_CLIENT_SECRET=insecure_secret
OAUTH_PROVIDER_NAME=Authelia
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  karakeep:
    environment:
      OAUTH_WELLKNOWN_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration'
      OAUTH_CLIENT_ID: 'karakeep'
      OAUTH_CLIENT_SECRET: 'insecure_secret'
      OAUTH_PROVIDER_NAME: 'Authelia'
```

## See Also

- [Karakeep OAuth OIDC config](https://docs.karakeep.app/configuration#authentication--signup)

[karakeep]: https://karakeep.app/
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
