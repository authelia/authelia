---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Hoop.dev"
description: "A guide on integrating Hoop.dev with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Hoop.dev | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Hoop.dev with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Hoop.dev]
  - [v1.305.0](https://github.com/hoophq/hoop/releases/tag/1.305.0)

{{% oidc-common bugs="claims-hydration" %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://hoop.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://hoop.{{< sitevar name="domain" nojs="example.com" >}}/api/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `hoop-dev`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Hoop.dev] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'hoop-dev'
        client_name: 'Hoop.dev'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://hoop.{{< sitevar name="domain" nojs="example.com" >}}/api/callback'
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

{{% oidc-escape-hatch-claims-hydration client_id="hoop-dev" claims="email,email_verified,name,groups" %}}

### Application

To configure [Hoop.dev] there are two methods, using the [Web GUI](#web-gui) or using
[Environment Variables](#environment-variables).

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
[Hoop.dev] builds the redirect URI from the `API_URL` environment variable, which must be set to
`https://hoop.{{< sitevar name="domain" nojs="example.com" >}}` regardless of the method used. When the identity
provider is configured using the [Web GUI](#web-gui) the [Environment Variables](#environment-variables) for the
identity provider are ignored.
{{< /callout >}}

#### Web GUI

To configure [Hoop.dev] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [Hoop.dev] as an administrator.
2. Navigate to `Integrations`, `Authentication`.
3. Select `Identity Provider`.
4. Select `OIDC` as the protocol.
5. Select `Other` as the identity provider.
6. Configure the following options:
   - Client ID: `hoop-dev`
   - Client Authentication: `Client secret`
   - Client secret: `insecure_secret`
   - Issuer URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - Custom Scopes: `groups`
   - Groups Claim: `groups`
7. Press `Save`.

#### Environment Variables

To configure [Hoop.dev] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
API_URL=https://hoop.{{< sitevar name="domain" nojs="example.com" >}}
IDP_ISSUER=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
IDP_CLIENT_ID=hoop-dev
IDP_CLIENT_SECRET=insecure_secret
IDP_CUSTOM_SCOPES=groups
IDP_GROUPS_CLAIM=groups
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  gateway:
    environment:
      API_URL: 'https://hoop.{{< sitevar name="domain" nojs="example.com" >}}'
      IDP_ISSUER: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      IDP_CLIENT_ID: 'hoop-dev'
      IDP_CLIENT_SECRET: 'insecure_secret'
      IDP_CUSTOM_SCOPES: 'groups'
      IDP_GROUPS_CLAIM: 'groups'
```

## See Also

- [Hoop.dev Identity Providers Documentation](https://hoop.dev/docs/setup/configuration/idp/get-started)
- [Hoop.dev Environment Variables Documentation](https://hoop.dev/docs/setup/configuration/env-vars)

[Authelia]: https://www.authelia.com
[Hoop.dev]: https://hoop.dev/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
