---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Yuvomi"
description: "A guide on integrating Yuvomi with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Yuvomi | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Yuvomi with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Yuvomi]
  - [v2.89.1](https://github.com/ulsklyc/yuvomi/releases/tag/v2.89.1)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://yuvomi.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://yuvomi.{{< sitevar name="domain" nojs="example.com" >}}/api/v1/auth/oidc/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `yuvomi`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Yuvomi] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'yuvomi'
        client_name: 'Yuvomi'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://yuvomi.{{< sitevar name="domain" nojs="example.com" >}}/api/v1/auth/oidc/callback'
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
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [Yuvomi] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [Yuvomi] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
OIDC_ISSUER=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
OIDC_CLIENT_ID=yuvomi
OIDC_CLIENT_SECRET=insecure_secret
OIDC_REDIRECT_URI=https://yuvomi.{{< sitevar name="domain" nojs="example.com" >}}/api/v1/auth/oidc/callback
OIDC_TOKEN_AUTH_METHOD=client_secret_basic
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  yuvomi:
    environment:
      OIDC_ISSUER: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      OIDC_CLIENT_ID: 'yuvomi'
      OIDC_CLIENT_SECRET: 'insecure_secret'
      OIDC_REDIRECT_URI: 'https://yuvomi.{{< sitevar name="domain" nojs="example.com" >}}/api/v1/auth/oidc/callback'
      OIDC_TOKEN_AUTH_METHOD: 'client_secret_basic'
```

## See Also

- [Yuvomi SSO / OpenID Connect Documentation](https://github.com/ulsklyc/yuvomi/blob/main/docs/installation.md#sso--openid-connect-optional)

[Authelia]: https://www.authelia.com
[Yuvomi]: https://github.com/ulsklyc/yuvomi
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
