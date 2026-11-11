---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Huly"
description: "A guide on integrating Huly with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Huly | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Huly with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Huly]
  - [v0.7.426](https://github.com/hcengineering/platform/releases/tag/v0.7.426)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://huly.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Application Account Service URL:** `https://huly.{{< sitevar name="domain" nojs="example.com" >}}/_accounts`
  - This option determines the redirect URI in the format of
    `https://huly.{{< sitevar name="domain" nojs="example.com" >}}/_accounts/auth/openid/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `huly`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Huly] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'huly'
        client_name: 'Huly'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://huly.{{< sitevar name="domain" nojs="example.com" >}}/_accounts/auth/openid/callback'
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

To configure [Huly] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [Huly] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment variables
on the `account` service. The `ACCOUNTS_URL` environment variable of the `account` and `front` services must be the
externally accessible URL of the account service.

##### Standard

```shell {title=".env"}
ACCOUNTS_URL=https://huly.{{< sitevar name="domain" nojs="example.com" >}}/_accounts
OPENID_DISPLAY_NAME=Authelia
OPENID_ISSUER=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
OPENID_CLIENT_ID=huly
OPENID_CLIENT_SECRET=insecure_secret
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  account:
    environment:
      ACCOUNTS_URL: 'https://huly.{{< sitevar name="domain" nojs="example.com" >}}/_accounts'
      OPENID_DISPLAY_NAME: 'Authelia'
      OPENID_ISSUER: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      OPENID_CLIENT_ID: 'huly'
      OPENID_CLIENT_SECRET: 'insecure_secret'
```

## See Also

- [Huly Self Host OpenID Connect Documentation](https://github.com/hcengineering/huly-selfhost#configure-openid-connect-oidc)

[Authelia]: https://www.authelia.com
[Huly]: https://huly.io/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
