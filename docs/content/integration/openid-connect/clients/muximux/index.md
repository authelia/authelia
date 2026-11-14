---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Muximux"
description: "A guide on integrating Muximux with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Muximux | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Muximux with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Muximux]
  - [v3.7.2](https://github.com/mescon/Muximux/releases/tag/v3.7.2)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://muximux.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://muximux.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/oidc/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `muximux`
- **Client Secret:** `insecure_secret`
- **Administrator Group:** `muximux-admins`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Muximux] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'muximux'
        client_name: 'Muximux'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://muximux.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/oidc/callback'
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
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [Muximux] there are two methods, using the [Configuration File](#configuration-file), or using
[Environment Variables](#environment-variables).

#### Configuration File

To configure [Muximux] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```yaml {title="config.yaml"}
auth:
  method: 'oidc'
  oidc:
    enabled: true
    issuer_url: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
    client_id: 'muximux'
    client_secret: 'insecure_secret'
    redirect_url: 'https://muximux.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/oidc/callback'
    scopes:
      - 'openid'
      - 'profile'
      - 'email'
      - 'groups'
    username_claim: 'preferred_username'
    email_claim: 'email'
    groups_claim: 'groups'
    display_name_claim: 'name'
    admin_groups:
      - 'muximux-admins'
```

#### Environment Variables

To configure [Muximux] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
MUXIMUX_AUTH_METHOD=oidc
MUXIMUX_OIDC_ISSUER_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
MUXIMUX_OIDC_CLIENT_ID=muximux
MUXIMUX_OIDC_CLIENT_SECRET=insecure_secret
MUXIMUX_OIDC_REDIRECT_URL=https://muximux.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/oidc/callback
MUXIMUX_OIDC_SCOPES=openid profile email groups
MUXIMUX_OIDC_USERNAME_CLAIM=preferred_username
MUXIMUX_OIDC_EMAIL_CLAIM=email
MUXIMUX_OIDC_GROUPS_CLAIM=groups
MUXIMUX_OIDC_DISPLAY_NAME_CLAIM=name
MUXIMUX_OIDC_ADMIN_GROUPS=muximux-admins
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  muximux:
    environment:
      MUXIMUX_AUTH_METHOD: 'oidc'
      MUXIMUX_OIDC_ISSUER_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      MUXIMUX_OIDC_CLIENT_ID: 'muximux'
      MUXIMUX_OIDC_CLIENT_SECRET: 'insecure_secret'
      MUXIMUX_OIDC_REDIRECT_URL: 'https://muximux.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/oidc/callback'
      MUXIMUX_OIDC_SCOPES: 'openid profile email groups'
      MUXIMUX_OIDC_USERNAME_CLAIM: 'preferred_username'
      MUXIMUX_OIDC_EMAIL_CLAIM: 'email'
      MUXIMUX_OIDC_GROUPS_CLAIM: 'groups'
      MUXIMUX_OIDC_DISPLAY_NAME_CLAIM: 'name'
      MUXIMUX_OIDC_ADMIN_GROUPS: 'muximux-admins'
```

## See Also

- [Muximux Authentication Documentation](https://github.com/mescon/Muximux/blob/main/docs/wiki/authentication.md)
- [Muximux Authelia Documentation](https://github.com/mescon/Muximux/blob/main/docs/wiki/forward-auth-authelia.md)

[Authelia]: https://www.authelia.com
[Muximux]: https://github.com/mescon/Muximux
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
