---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "LobeHub"
description: "A guide on integrating LobeHub with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
summary: ""
date: 2026-11-12T00:00:00+00:00
draft: false
images: []
weight: 620
toc: true
support:
  level: community
  versions: true
  integration: true
seo:
  title: "LobeHub | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring LobeHub with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [LobeHub] (previously named LobeChat)
  - [v2.3.2](https://github.com/lobehub/lobehub/releases/tag/v2.3.2)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://lobehub.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://lobehub.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/callback/authelia`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `lobehub`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [LobeHub] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'lobehub'
        client_name: 'LobeHub'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://lobehub.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/callback/authelia'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [LobeHub] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [LobeHub] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
APP_URL=https://lobehub.{{< sitevar name="domain" nojs="example.com" >}}
AUTH_SECRET=<a random secret generated with openssl rand -base64 32>
AUTH_SSO_PROVIDERS=authelia
AUTH_AUTHELIA_ID=lobehub
AUTH_AUTHELIA_SECRET=insecure_secret
AUTH_AUTHELIA_ISSUER=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  lobehub:
    environment:
      APP_URL: 'https://lobehub.{{< sitevar name="domain" nojs="example.com" >}}'
      AUTH_SECRET: '<a random secret generated with openssl rand -base64 32>'
      AUTH_SSO_PROVIDERS: 'authelia'
      AUTH_AUTHELIA_ID: 'lobehub'
      AUTH_AUTHELIA_SECRET: 'insecure_secret'
      AUTH_AUTHELIA_ISSUER: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
```

## See Also

- [LobeHub Authelia Authentication Documentation](https://lobehub.com/docs/self-hosting/auth/providers/authelia)
- [LobeHub Authentication Environment Variables Documentation](https://lobehub.com/docs/self-hosting/environment-variables/auth)

[Authelia]: https://www.authelia.com
[LobeHub]: https://lobehub.com/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
