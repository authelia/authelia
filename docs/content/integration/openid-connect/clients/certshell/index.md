---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "CertShell"
description: "A guide on integrating CertShell with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "CertShell | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring CertShell with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [CertShell]
  - [v1.16.3](https://github.com/maca134/certshell/releases/tag/v1.16.3)

{{% oidc-common bugs="claims-hydration" %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://certshell.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://certshell.{{< sitevar name="domain" nojs="example.com" >}}/auth/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `certshell`
- **Client Secret:** `insecure_secret`
- **Admin Group:** `certshell-admins`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [CertShell] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'certshell'
        client_name: 'CertShell'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://certshell.{{< sitevar name="domain" nojs="example.com" >}}/auth/callback'
        scopes: ['openid', 'email', 'profile', 'groups']
        response_types: ['code']
        grant_types: ['authorization_code']
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

Notes:

- Members of the `certshell-admins` group are able to access the **Admin** section of [CertShell]. Access to individual
  hosts and logins is granted to groups from within the **Admin** section.

#### Configuration Escape Hatch

{{% oidc-escape-hatch-claims-hydration client_id="certshell" claims="email,groups" %}}

### Application

To configure [CertShell] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [CertShell] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
APP_URL=https://certshell.{{< sitevar name="domain" nojs="example.com" >}}
OIDC_ISSUER=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
OIDC_CLIENT_ID=certshell
OIDC_CLIENT_SECRET=insecure_secret
OIDC_ADMIN_GROUP=certshell-admins
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  certshell:
    environment:
      APP_URL: 'https://certshell.{{< sitevar name="domain" nojs="example.com" >}}'
      OIDC_ISSUER: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      OIDC_CLIENT_ID: 'certshell'
      OIDC_CLIENT_SECRET: 'insecure_secret'
      OIDC_ADMIN_GROUP: 'certshell-admins'
```

## See Also

- [CertShell Configuration Documentation](https://github.com/maca134/certshell#configuration)

[Authelia]: https://www.authelia.com
[CertShell]: https://github.com/maca134/certshell
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
