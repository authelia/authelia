---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "tududi"
description: "A guide on integrating tududi with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "tududi | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring tududi with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [tududi]
  - [v2.1.0](https://github.com/chrisvel/tududi/releases/tag/v2.1.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://tududi.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://tududi.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/callback/authelia`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `tududi`
- **Client Secret:** `insecure_secret`
- **Provider Slug:** `authelia`
  - This option determines the last segment of the redirect URI.

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [tududi] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'tududi'
        client_name: 'tududi'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://tududi.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/callback/authelia'
        scopes: ['openid', 'email', 'profile']
        response_types: ['code']
        grant_types: ['authorization_code']
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [tududi] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [tududi] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
BASE_URL=https://tududi.{{< sitevar name="domain" nojs="example.com" >}}
TUDUDI_TRUST_PROXY=true
OIDC_ENABLED=true
OIDC_PROVIDER_NAME=Authelia
OIDC_PROVIDER_SLUG=authelia
OIDC_ISSUER_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
OIDC_CLIENT_ID=tududi
OIDC_CLIENT_SECRET=insecure_secret
OIDC_SCOPE=openid profile email
OIDC_AUTO_PROVISION=true
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  tududi:
    environment:
      BASE_URL: 'https://tududi.{{< sitevar name="domain" nojs="example.com" >}}'
      TUDUDI_TRUST_PROXY: 'true'
      OIDC_ENABLED: 'true'
      OIDC_PROVIDER_NAME: 'Authelia'
      OIDC_PROVIDER_SLUG: 'authelia'
      OIDC_ISSUER_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      OIDC_CLIENT_ID: 'tududi'
      OIDC_CLIENT_SECRET: 'insecure_secret'
      OIDC_SCOPE: 'openid profile email'
      OIDC_AUTO_PROVISION: 'true'
```

Notes:

- The `TUDUDI_TRUST_PROXY` option is required when [tududi] is deployed behind a reverse proxy.

## See Also

- [tududi OIDC/SSO Authentication Documentation](https://github.com/chrisvel/tududi/blob/main/docs/10-oidc-sso.md)

[Authelia]: https://www.authelia.com
[tududi]: https://tududi.com/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
