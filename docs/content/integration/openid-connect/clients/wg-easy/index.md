---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "wg-easy"
description: "A guide on integrating wg-easy with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "wg-easy | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring wg-easy with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [wg-easy]
  - [v15.5.0](https://github.com/wg-easy/wg-easy/releases/tag/v15.5.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://wg-easy.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URIs in the format of
    `https://wg-easy.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/oidc/callback` and
    `https://wg-easy.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/oidc/link`.
    This means if you change this value, you need to update the redirect URIs.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `wg-easy`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [wg-easy] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'wg-easy'
        client_name: 'wg-easy'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://wg-easy.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/oidc/callback'
          - 'https://wg-easy.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/oidc/link'
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

To configure [wg-easy] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [wg-easy] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
Users registered automatically via `OAUTH_AUTO_REGISTER` are created with administrator permissions as [wg-easy] does
not yet implement a permissions system. Ensure the Authelia [authorization_policy] only permits users who should be able
to manage the WireGuard server, or set `OAUTH_AUTO_REGISTER` to `false` and have existing users link their accounts.
{{< /callout >}}

##### Standard

```shell {title=".env"}
OAUTH_PROVIDERS=oidc
OAUTH_OIDC_NAME=Authelia
OAUTH_OIDC_SERVER=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
OAUTH_OIDC_CLIENT_ID=wg-easy
OAUTH_OIDC_CLIENT_SECRET=insecure_secret
OAUTH_AUTO_REGISTER=true
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  wg-easy:
    environment:
      OAUTH_PROVIDERS: 'oidc'
      OAUTH_OIDC_NAME: 'Authelia'
      OAUTH_OIDC_SERVER: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      OAUTH_OIDC_CLIENT_ID: 'wg-easy'
      OAUTH_OIDC_CLIENT_SECRET: 'insecure_secret'
      OAUTH_AUTO_REGISTER: 'true'
```

## See Also

- [wg-easy External Authentication Documentation](https://wg-easy.github.io/wg-easy/latest/advanced/config/external-authentication/)

[Authelia]: https://www.authelia.com
[wg-easy]: https://wg-easy.github.io/wg-easy/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
[authorization_policy]: ../../../../configuration/identity-providers/openid-connect/clients.md#authorization_policy
