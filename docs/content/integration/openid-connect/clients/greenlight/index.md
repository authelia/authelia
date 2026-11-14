---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Greenlight"
description: "A guide on integrating Greenlight with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Greenlight | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Greenlight with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Greenlight]
  - [v3.9.1](https://github.com/bigbluebutton/greenlight/releases/tag/release-3.9.1)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://greenlight.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://greenlight.{{< sitevar name="domain" nojs="example.com" >}}/auth/openid_connect/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `greenlight`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Greenlight] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'greenlight'
        client_name: 'Greenlight'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://greenlight.{{< sitevar name="domain" nojs="example.com" >}}/auth/openid_connect/callback'
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

To configure [Greenlight] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [Greenlight] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
OPENID_CONNECT_CLIENT_ID=greenlight
OPENID_CONNECT_CLIENT_SECRET=insecure_secret
OPENID_CONNECT_ISSUER=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
OPENID_CONNECT_REDIRECT=https://greenlight.{{< sitevar name="domain" nojs="example.com" >}}/
OPENID_CONNECT_UID_FIELD=sub
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  greenlight:
    environment:
      OPENID_CONNECT_CLIENT_ID: 'greenlight'
      OPENID_CONNECT_CLIENT_SECRET: 'insecure_secret'
      OPENID_CONNECT_ISSUER: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      OPENID_CONNECT_REDIRECT: 'https://greenlight.{{< sitevar name="domain" nojs="example.com" >}}/'
      OPENID_CONNECT_UID_FIELD: 'sub'
```

## See Also

- [Greenlight External Authentication Documentation](https://docs.bigbluebutton.org/greenlight/v3/external-authentication)

[Authelia]: https://www.authelia.com
[Greenlight]: https://docs.bigbluebutton.org/greenlight/v3/install/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
