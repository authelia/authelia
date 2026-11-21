---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Dreeve"
description: "A guide on integrating Dreeve with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
summary: ""
date: 2026-11-02T10:39:00+00:00
draft: false
images: []
weight: 620
toc: true
aliases: []
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Dreeve | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Dreeve with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Dreeve]
  - [v5.8.0](https://github.com/dreeveapp/dreeve/releases/tag/v5.8.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://dreeve.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://dreeve.{{< sitevar name="domain" nojs="example.com" >}}/admin/login/oidc/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `dreeve`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Dreeve] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    claims_policies:
      dreeve:
        id_token: ['email', 'email_verified', 'preferred_username', 'name']
    clients:
      - client_id: 'dreeve'
        client_name: 'Dreeve'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        claims_policy: 'dreeve'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://dreeve.{{< sitevar name="domain" nojs="example.com" >}}/admin/login/oidc/callback'
        scopes: ['openid', 'email', 'profile']
        response_types: ['code']
        grant_types: ['authorization_code']
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [Dreeve] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [Dreeve] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
OIDC_ISSUER_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
OIDC_CLIENT_ID=dreeve
OIDC_CLIENT_SECRET=insecure_secret
# Comma-separated. Each entry is a subject ("sub" claim) or a verified email address.
OIDC_ALLOWED_USERS=athlete@{{< sitevar name="domain" nojs="example.com" >}}
```

## See Also

- [Dreeve Single sign-on (OIDC) Documentation](https://docs.dreeve.app/integrations/single-sign-on/)

[Authelia]: https://www.authelia.com
[Dreeve]: https://dreeve.app/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
