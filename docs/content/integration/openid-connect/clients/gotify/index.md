---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Gotify"
description: "A guide on integrating Gotify with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Gotify | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Gotify with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Gotify]
  - [v3.1.1](https://github.com/gotify/server/releases/tag/v3.1.1)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://gotify.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://gotify.{{< sitevar name="domain" nojs="example.com" >}}/auth/oidc/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `gotify`
- **Client Secret:** `insecure_secret`
- **Administrator Group:** `gotify-admins`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Gotify] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'gotify'
        client_name: 'Gotify'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://gotify.{{< sitevar name="domain" nojs="example.com" >}}/auth/oidc/callback'
          - 'gotify://oidc/callback'  # Only required for the Gotify Android application.
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
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

To configure [Gotify] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [Gotify] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
GOTIFY_OIDC_ENABLED=true
GOTIFY_OIDC_ISSUER=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
GOTIFY_OIDC_CLIENTID=gotify
GOTIFY_OIDC_CLIENTSECRET=insecure_secret
GOTIFY_OIDC_REDIRECTURL=https://gotify.{{< sitevar name="domain" nojs="example.com" >}}/auth/oidc/callback
GOTIFY_OIDC_IDP_NAME=Authelia
GOTIFY_OIDC_SCOPES=openid,profile,email,groups
GOTIFY_OIDC_USERNAMECLAIM=preferred_username
GOTIFY_OIDC_AUTOREGISTER=true
GOTIFY_OIDC_GROUPS_CLAIM=groups
GOTIFY_OIDC_GROUPS_ADMIN=gotify-admins
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  gotify:
    environment:
      GOTIFY_OIDC_ENABLED: 'true'
      GOTIFY_OIDC_ISSUER: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      GOTIFY_OIDC_CLIENTID: 'gotify'
      GOTIFY_OIDC_CLIENTSECRET: 'insecure_secret'
      GOTIFY_OIDC_REDIRECTURL: 'https://gotify.{{< sitevar name="domain" nojs="example.com" >}}/auth/oidc/callback'
      GOTIFY_OIDC_IDP_NAME: 'Authelia'
      GOTIFY_OIDC_SCOPES: 'openid,profile,email,groups'
      GOTIFY_OIDC_USERNAMECLAIM: 'preferred_username'
      GOTIFY_OIDC_AUTOREGISTER: 'true'
      GOTIFY_OIDC_GROUPS_CLAIM: 'groups'
      GOTIFY_OIDC_GROUPS_ADMIN: 'gotify-admins'
```

If a Gotify user with the same username already exists, the login is rejected unless `GOTIFY_OIDC_LINK_BY_USERNAME`
is set to `true`. Only enable this option if the usernames in Authelia belong to the same people as the matching Gotify
usernames.

## See Also

- [Gotify OpenID Connect Documentation](https://gotify.net/docs/oidc)
- [Gotify Configuration Documentation](https://gotify.net/docs/config)

[Authelia]: https://www.authelia.com
[Gotify]: https://gotify.net/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
