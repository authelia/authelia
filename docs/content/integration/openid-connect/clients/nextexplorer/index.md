---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "nextExplorer"
description: "A guide on integrating nextExplorer with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "nextExplorer | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring nextExplorer with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [nextExplorer]
  - [v4.0.1](https://github.com/nxzai/NextExplorer/releases/tag/v4.0.1)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://nextexplorer.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://nextexplorer.{{< sitevar name="domain" nojs="example.com" >}}/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `nextexplorer`
- **Client Secret:** `insecure_secret`
- **Administrators Group:** `admins`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [nextExplorer] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'nextexplorer'
        client_name: 'nextExplorer'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://nextexplorer.{{< sitevar name="domain" nojs="example.com" >}}/callback'
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

To configure [nextExplorer] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [nextExplorer] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
PUBLIC_URL=https://nextexplorer.{{< sitevar name="domain" nojs="example.com" >}}
AUTH_MODE=both
OIDC_ENABLED=true
OIDC_ISSUER=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
OIDC_CLIENT_ID=nextexplorer
OIDC_CLIENT_SECRET=insecure_secret
OIDC_SCOPES=openid profile email groups
OIDC_ADMIN_GROUPS=admins
OIDC_REQUIRE_EMAIL_VERIFIED=true
OIDC_AUTO_CREATE_USERS=true
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  nextexplorer:
    environment:
      PUBLIC_URL: 'https://nextexplorer.{{< sitevar name="domain" nojs="example.com" >}}'
      AUTH_MODE: 'both'
      OIDC_ENABLED: 'true'
      OIDC_ISSUER: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      OIDC_CLIENT_ID: 'nextexplorer'
      OIDC_CLIENT_SECRET: 'insecure_secret'
      OIDC_SCOPES: 'openid profile email groups'
      OIDC_ADMIN_GROUPS: 'admins'
      OIDC_REQUIRE_EMAIL_VERIFIED: 'true'
      OIDC_AUTO_CREATE_USERS: 'true'
```

## See Also

- [nextExplorer OIDC Documentation](https://explorer.nxz.ai/integrations/oidc.html)

[Authelia]: https://www.authelia.com
[nextExplorer]: https://explorer.nxz.ai/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
