---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Penpot"
description: "A guide on integrating Penpot with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Penpot | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Penpot with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Penpot]
  - [2.19.1](https://github.com/penpot/penpot/releases/tag/2.19.1)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://penpot.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://penpot.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/oidc/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `penpot`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Penpot] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'penpot'
        client_name: 'Penpot'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://penpot.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/oidc/callback'
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

To configure [Penpot] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [Penpot] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
The `PENPOT_FLAGS` variable must be set on both the frontend and backend services and should include any flags you
already use. The `enable-oidc-registration` flag allows users without an existing [Penpot] profile to register using
Authelia. The `PENPOT_OIDC_USER_INFO_SOURCE` variable must be `userinfo` so that [Penpot] retrieves the `email` and
`name` claims from the UserInfo Endpoint.
{{< /callout >}}

##### Standard

```shell {title=".env"}
PENPOT_FLAGS=enable-login-with-oidc enable-oidc-registration
PENPOT_OIDC_NAME=Authelia
PENPOT_OIDC_BASE_URI=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
PENPOT_OIDC_CLIENT_ID=penpot
PENPOT_OIDC_CLIENT_SECRET=insecure_secret
PENPOT_OIDC_SCOPES=openid email profile
PENPOT_OIDC_USER_INFO_SOURCE=userinfo
```

##### Docker Compose

```yaml {title="compose.yml"}
x-flags: &penpot-flags
  PENPOT_FLAGS: 'enable-login-with-oidc enable-oidc-registration'

services:
  penpot-frontend:
    environment:
      <<: *penpot-flags
      PENPOT_OIDC_NAME: 'Authelia'
  penpot-backend:
    environment:
      <<: *penpot-flags
      PENPOT_OIDC_BASE_URI: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      PENPOT_OIDC_CLIENT_ID: 'penpot'
      PENPOT_OIDC_CLIENT_SECRET: 'insecure_secret'
      PENPOT_OIDC_SCOPES: 'openid email profile'
      PENPOT_OIDC_USER_INFO_SOURCE: 'userinfo'
```

## See Also

- [Penpot Authentication Providers Documentation](https://help.penpot.app/technical-guide/configuration/#authentication-providers)

[Authelia]: https://www.authelia.com
[Penpot]: https://penpot.app/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
