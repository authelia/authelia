---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "PLANKA"
description: "A guide on integrating PLANKA with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "PLANKA | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring PLANKA with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [PLANKA]
  - [v2.1.1](https://github.com/plankanban/planka/releases/tag/2.1.1)

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
Starting with [v2.2.0](https://github.com/plankanban/planka/releases/tag/v2.2.0) OpenID Connect 1.0 single sign-on
was removed from the PLANKA Community edition and is only available in [PLANKA Pro](https://planka.app/pro). This
guide was tested with the last Community edition release which included OpenID Connect 1.0 support.
{{< /callout >}}

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://planka.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://planka.{{< sitevar name="domain" nojs="example.com" >}}/oidc-callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `planka`
- **Client Secret:** `insecure_secret`
- **Administrator Group:** `planka-admins`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [PLANKA] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'planka'
        client_name: 'PLANKA'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://planka.{{< sitevar name="domain" nojs="example.com" >}}/oidc-callback'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
          - 'groups'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [PLANKA] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [PLANKA] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
BASE_URL=https://planka.{{< sitevar name="domain" nojs="example.com" >}}
OIDC_ISSUER=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
OIDC_CLIENT_ID=planka
OIDC_CLIENT_SECRET=insecure_secret
OIDC_SCOPES=openid email profile groups
OIDC_USE_DEFAULT_RESPONSE_MODE=true
OIDC_ROLES_ATTRIBUTE=groups
OIDC_ADMIN_ROLES=planka-admins
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  planka:
    environment:
      BASE_URL: 'https://planka.{{< sitevar name="domain" nojs="example.com" >}}'
      OIDC_ISSUER: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      OIDC_CLIENT_ID: 'planka'
      OIDC_CLIENT_SECRET: 'insecure_secret'
      OIDC_SCOPES: 'openid email profile groups'
      OIDC_USE_DEFAULT_RESPONSE_MODE: 'true'
      OIDC_ROLES_ATTRIBUTE: 'groups'
      OIDC_ADMIN_ROLES: 'planka-admins'
```

To disable local authentication and only allow single sign-on, additionally set `OIDC_ENFORCED` to `true`.

## See Also

- [PLANKA OIDC Example Configuration](https://github.com/plankanban/planka/blob/2.1.1/server/.env.sample)
- [PLANKA Pro](https://planka.app/pro)

[Authelia]: https://www.authelia.com
[PLANKA]: https://planka.app/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
