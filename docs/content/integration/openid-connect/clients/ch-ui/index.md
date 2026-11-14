---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "CH-UI"
description: "A guide on integrating CH-UI with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "CH-UI | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring CH-UI with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [CH-UI]
  - [v2.33.0](https://github.com/caioricciuti/ch-ui/releases/tag/v2.33.0)

{{% oidc-common bugs="claims-hydration" %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://ch-ui.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://ch-ui.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/oidc/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `ch-ui`
- **Client Secret:** `insecure_secret`
- **Admin Group:** `ch-ui-admins`
- **Analyst Group:** `ch-ui-analysts`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

### Special Notes

1. The [CH-UI] [OpenID Connect 1.0] integration is a Pro feature and requires a Pro license.
2. [CH-UI] runs queries for users who log in via [OpenID Connect 1.0] using a ClickHouse service account which must be
   configured on the relevant connection. See the [CH-UI SSO Documentation](https://ch-ui.com/docs/sso) for more
   information.

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [CH-UI] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'ch-ui'
        client_name: 'CH-UI'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://ch-ui.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/oidc/callback'
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

#### Configuration Escape Hatch

{{% oidc-escape-hatch-claims-hydration client_id="ch-ui" claims="email,email_verified,name,groups" %}}

### Application

To configure [CH-UI] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [CH-UI] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
OIDC_ISSUER_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
OIDC_CLIENT_ID=ch-ui
OIDC_CLIENT_SECRET=insecure_secret
OIDC_REDIRECT_URL=https://ch-ui.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/oidc/callback
OIDC_GROUPS_CLAIM=groups
OIDC_ADMIN_GROUPS=ch-ui-admins
OIDC_ANALYST_GROUPS=ch-ui-analysts
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  ch-ui:
    environment:
      OIDC_ISSUER_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      OIDC_CLIENT_ID: 'ch-ui'
      OIDC_CLIENT_SECRET: 'insecure_secret'
      OIDC_REDIRECT_URL: 'https://ch-ui.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/oidc/callback'
      OIDC_GROUPS_CLAIM: 'groups'
      OIDC_ADMIN_GROUPS: 'ch-ui-admins'
      OIDC_ANALYST_GROUPS: 'ch-ui-analysts'
```

## See Also

- [CH-UI Single Sign-On (SSO) Documentation](https://ch-ui.com/docs/sso)
- [CH-UI Configuration Documentation](https://ch-ui.com/docs/configuration)

[Authelia]: https://www.authelia.com
[CH-UI]: https://ch-ui.com/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
