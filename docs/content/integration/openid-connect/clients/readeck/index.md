---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Readeck"
description: "A guide on integrating Readeck with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Readeck | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Readeck with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Readeck]
  - [v0.23.5](https://codeberg.org/readeck/readeck/releases/tag/0.23.5)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://readeck.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://readeck.{{< sitevar name="domain" nojs="example.com" >}}/login/oidc`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `readeck`
- **Client Secret:** `insecure_secret`
- **Admin Group:** `readeck-admins`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Readeck] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'readeck'
        client_name: 'Readeck'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://readeck.{{< sitevar name="domain" nojs="example.com" >}}/login/oidc'
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

To configure [Readeck] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [Readeck] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
READECK_AUTH_OIDC_PROVIDERS_1_NAME=Authelia
READECK_AUTH_OIDC_PROVIDERS_1_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
READECK_AUTH_OIDC_PROVIDERS_1_CLIENT_ID=readeck
READECK_AUTH_OIDC_PROVIDERS_1_CLIENT_SECRET=insecure_secret
READECK_AUTH_OIDC_PROVIDERS_1_PROVISIONING=true
READECK_AUTH_OIDC_PROVIDERS_1_GROUPS="readeck-admins=admin"
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  readeck:
    environment:
      READECK_AUTH_OIDC_PROVIDERS_1_NAME: 'Authelia'
      READECK_AUTH_OIDC_PROVIDERS_1_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      READECK_AUTH_OIDC_PROVIDERS_1_CLIENT_ID: 'readeck'
      READECK_AUTH_OIDC_PROVIDERS_1_CLIENT_SECRET: 'insecure_secret'
      READECK_AUTH_OIDC_PROVIDERS_1_PROVISIONING: 'true'
      READECK_AUTH_OIDC_PROVIDERS_1_GROUPS: 'readeck-admins=admin'
```

## See Also

- [Readeck External Authentication Documentation](https://readeck.org/en/docs/external-auth)

[Authelia]: https://www.authelia.com
[Readeck]: https://readeck.org/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
