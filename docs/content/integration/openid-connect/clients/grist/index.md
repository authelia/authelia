---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Grist"
description: "A guide on integrating Grist with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Grist | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Grist with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Grist]
  - [v1.7.22](https://github.com/gristlabs/grist-core/releases/tag/v1.7.22)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://grist.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://grist.{{< sitevar name="domain" nojs="example.com" >}}/oauth2/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `grist`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Grist] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'grist'
        client_name: 'Grist'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://grist.{{< sitevar name="domain" nojs="example.com" >}}/oauth2/callback'
        scopes: ['openid', 'email', 'profile']
        response_types: ['code']
        grant_types: ['authorization_code']
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [Grist] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [Grist] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
GRIST_OIDC_SP_HOST=https://grist.{{< sitevar name="domain" nojs="example.com" >}}
GRIST_OIDC_IDP_ISSUER=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
GRIST_OIDC_IDP_CLIENT_ID=grist
GRIST_OIDC_IDP_CLIENT_SECRET=insecure_secret
GRIST_OIDC_IDP_SCOPES=openid email profile
GRIST_OIDC_IDP_SKIP_END_SESSION_ENDPOINT=true
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  grist:
    environment:
      GRIST_OIDC_SP_HOST: 'https://grist.{{< sitevar name="domain" nojs="example.com" >}}'
      GRIST_OIDC_IDP_ISSUER: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      GRIST_OIDC_IDP_CLIENT_ID: 'grist'
      GRIST_OIDC_IDP_CLIENT_SECRET: 'insecure_secret'
      GRIST_OIDC_IDP_SCOPES: 'openid email profile'
      GRIST_OIDC_IDP_SKIP_END_SESSION_ENDPOINT: 'true'
```

Notes:

- The `GRIST_OIDC_IDP_SKIP_END_SESSION_ENDPOINT` option is required as Authelia does not currently advertise an
  `end_session_endpoint` and [Grist] will otherwise fail to initialize the [OpenID Connect 1.0] integration.
- The `GRIST_DEFAULT_EMAIL` environment variable should be set to the email address of the Authelia user who should be
  the [Grist] administrator.

## See Also

- [Grist OpenID Connect Documentation](https://support.getgrist.com/install/oidc/)

[Authelia]: https://www.authelia.com
[Grist]: https://www.getgrist.com/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
