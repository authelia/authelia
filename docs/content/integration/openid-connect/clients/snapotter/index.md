---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "SnapOtter"
description: "A guide on integrating SnapOtter with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "SnapOtter | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring SnapOtter with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [SnapOtter]
  - [v2.3.0](https://github.com/snapotter-hq/SnapOtter/releases/tag/v2.3.0)

{{% oidc-common bugs="claims-hydration" %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://snapotter.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://snapotter.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/oidc/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `snapotter`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [SnapOtter] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'snapotter'
        client_name: 'SnapOtter'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://snapotter.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/oidc/callback'
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

#### Configuration Escape Hatch

{{% oidc-escape-hatch-claims-hydration client_id="snapotter" claims="email,email_verified,preferred_username,name" %}}

### Application

To configure [SnapOtter] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [SnapOtter] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
EXTERNAL_URL=https://snapotter.{{< sitevar name="domain" nojs="example.com" >}}
OIDC_ENABLED=true
OIDC_ISSUER_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
OIDC_CLIENT_ID=snapotter
OIDC_CLIENT_SECRET=insecure_secret
OIDC_SCOPES=openid profile email
OIDC_PROVIDER_NAME=Authelia
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  snapotter:
    environment:
      EXTERNAL_URL: 'https://snapotter.{{< sitevar name="domain" nojs="example.com" >}}'
      OIDC_ENABLED: 'true'
      OIDC_ISSUER_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      OIDC_CLIENT_ID: 'snapotter'
      OIDC_CLIENT_SECRET: 'insecure_secret'
      OIDC_SCOPES: 'openid profile email'
      OIDC_PROVIDER_NAME: 'Authelia'
```

## See Also

- [SnapOtter OIDC / Single Sign-On Documentation](https://docs.snapotter.com/guide/oidc)

[Authelia]: https://www.authelia.com
[SnapOtter]: https://snapotter.com/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
