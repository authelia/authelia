---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Directus"
description: "A guide on integrating Directus with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Directus | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Directus with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Directus]
  - [v12.6.1](https://github.com/directus/directus/releases/tag/v12.6.1)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://directus.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://directus.{{< sitevar name="domain" nojs="example.com" >}}/auth/login/authelia/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `directus`
- **Client Secret:** `insecure_secret`
- **Provider Name:** `authelia`
  - This option determines the redirect URI and the environment variable names. This means if you change this value,
    you need to update the redirect URI and the environment variable names.

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Directus] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'directus'
        client_name: 'Directus'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://directus.{{< sitevar name="domain" nojs="example.com" >}}/auth/login/authelia/callback'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [Directus] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [Directus] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
PUBLIC_URL=https://directus.{{< sitevar name="domain" nojs="example.com" >}}
AUTH_PROVIDERS=authelia
AUTH_AUTHELIA_DRIVER=openid
AUTH_AUTHELIA_CLIENT_ID=directus
AUTH_AUTHELIA_CLIENT_SECRET=insecure_secret
AUTH_AUTHELIA_ISSUER_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration
AUTH_AUTHELIA_SCOPE=openid profile email
AUTH_AUTHELIA_ALLOW_PUBLIC_REGISTRATION=true
AUTH_AUTHELIA_LABEL=Authelia
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  directus:
    environment:
      PUBLIC_URL: 'https://directus.{{< sitevar name="domain" nojs="example.com" >}}'
      AUTH_PROVIDERS: 'authelia'
      AUTH_AUTHELIA_DRIVER: 'openid'
      AUTH_AUTHELIA_CLIENT_ID: 'directus'
      AUTH_AUTHELIA_CLIENT_SECRET: 'insecure_secret'
      AUTH_AUTHELIA_ISSUER_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration'
      AUTH_AUTHELIA_SCOPE: 'openid profile email'
      AUTH_AUTHELIA_ALLOW_PUBLIC_REGISTRATION: 'true'
      AUTH_AUTHELIA_LABEL: 'Authelia'
```

If [Directus] already has other SSO providers configured, add `authelia` to the comma separated `AUTH_PROVIDERS` value
instead of replacing it. Users registered automatically are assigned the role configured by the
`AUTH_AUTHELIA_DEFAULT_ROLE_ID` environment variable which should be set to the ID of an existing [Directus] role.

## See Also

- [Directus Auth and SSO Configuration Documentation](https://directus.com/docs/configuration/auth-sso)
- [Directus Single Sign-On Documentation](https://directus.com/docs/guides/auth/sso)

[Authelia]: https://www.authelia.com
[Directus]: https://directus.io/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
