---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "ExcaliDash"
description: "A guide on integrating ExcaliDash with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "ExcaliDash | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring ExcaliDash with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [ExcaliDash]
  - [v0.6.7](https://github.com/ZimengXiong/ExcaliDash/releases/tag/v0.6.7)

{{% oidc-common bugs="claims-hydration" %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://excalidash.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://excalidash.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/oidc/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `excalidash`
- **Client Secret:** `insecure_secret`
- **Administrator Group:** `excalidash-admins`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [ExcaliDash] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'excalidash'
        client_name: 'ExcaliDash'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://excalidash.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/oidc/callback'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
          - 'groups'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

#### Configuration Escape Hatch

{{% oidc-escape-hatch-claims-hydration client_id="excalidash" claims="email,email_verified,name,preferred_username,groups" %}}

### Application

To configure [ExcaliDash] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [ExcaliDash] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables on the backend:

##### Standard

```shell {title=".env"}
AUTH_MODE=oidc_enforced
FRONTEND_URL=https://excalidash.{{< sitevar name="domain" nojs="example.com" >}}
OIDC_PROVIDER_NAME=Authelia
OIDC_ISSUER_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
OIDC_CLIENT_ID=excalidash
OIDC_CLIENT_SECRET=insecure_secret
OIDC_REDIRECT_URI=https://excalidash.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/oidc/callback
OIDC_SCOPES=openid profile email groups
OIDC_TOKEN_ENDPOINT_AUTH_METHOD=client_secret_basic
OIDC_ADMIN_GROUPS=excalidash-admins
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  backend:
    environment:
      AUTH_MODE: 'oidc_enforced'
      FRONTEND_URL: 'https://excalidash.{{< sitevar name="domain" nojs="example.com" >}}'
      OIDC_PROVIDER_NAME: 'Authelia'
      OIDC_ISSUER_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      OIDC_CLIENT_ID: 'excalidash'
      OIDC_CLIENT_SECRET: 'insecure_secret'
      OIDC_REDIRECT_URI: 'https://excalidash.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/oidc/callback'
      OIDC_SCOPES: 'openid profile email groups'
      OIDC_TOKEN_ENDPOINT_AUTH_METHOD: 'client_secret_basic'
      OIDC_ADMIN_GROUPS: 'excalidash-admins'
```

To keep password login enabled alongside Authelia, set `AUTH_MODE` to `hybrid` instead of `oidc_enforced`.

## See Also

- [ExcaliDash Authentication and OIDC Documentation](https://excalidash.xyz/guide/authentication)

[Authelia]: https://www.authelia.com
[ExcaliDash]: https://excalidash.xyz/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
