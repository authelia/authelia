---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Manyfold"
description: "A guide on integrating Manyfold with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Manyfold | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Manyfold with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Manyfold]
  - [v0.154.1](https://github.com/manyfold3d/manyfold/releases/tag/v0.154.1)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://manyfold.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://manyfold.{{< sitevar name="domain" nojs="example.com" >}}/users/auth/openid_connect/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `manyfold`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Manyfold] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'manyfold'
        client_name: 'Manyfold'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://manyfold.{{< sitevar name="domain" nojs="example.com" >}}/users/auth/openid_connect/callback'
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

To configure [Manyfold] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
[Manyfold] derives the redirect URI from the `PUBLIC_HOSTNAME` and `PUBLIC_PORT` options, and only uses the `https`
scheme when `HTTPS_ONLY` is enabled. Single sign-on also requires the multiuser mode to be enabled.
{{< /callout >}}

To configure [Manyfold] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
MULTIUSER=enabled
HTTPS_ONLY=enabled
PUBLIC_HOSTNAME=manyfold.{{< sitevar name="domain" nojs="example.com" >}}
OIDC_ISSUER=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
OIDC_CLIENT_ID=manyfold
OIDC_CLIENT_SECRET=insecure_secret
OIDC_NAME=Authelia
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  manyfold:
    environment:
      MULTIUSER: 'enabled'
      HTTPS_ONLY: 'enabled'
      PUBLIC_HOSTNAME: 'manyfold.{{< sitevar name="domain" nojs="example.com" >}}'
      OIDC_ISSUER: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      OIDC_CLIENT_ID: 'manyfold'
      OIDC_CLIENT_SECRET: 'insecure_secret'
      OIDC_NAME: 'Authelia'
```

Optionally `FORCE_OIDC=enabled` can be set to disable logging in with local accounts.

## See Also

- [Manyfold Configuration Documentation](https://manyfold.app/sysadmin/configuration.html)

[Authelia]: https://www.authelia.com
[Manyfold]: https://manyfold.app
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
