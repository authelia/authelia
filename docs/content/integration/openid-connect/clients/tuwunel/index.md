---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Tuwunel"
description: "A guide on integrating Tuwunel with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Tuwunel | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Tuwunel with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Tuwunel]
  - [v1.9.4](https://github.com/matrix-construct/tuwunel/releases/tag/v1.9.4)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://tuwunel.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://tuwunel.{{< sitevar name="domain" nojs="example.com" >}}/_matrix/client/unstable/login/sso/callback/tuwunel`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `tuwunel`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Tuwunel] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'tuwunel'
        client_name: 'Tuwunel'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://tuwunel.{{< sitevar name="domain" nojs="example.com" >}}/_matrix/client/unstable/login/sso/callback/tuwunel'
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

To configure [Tuwunel] there are two methods, using the [Configuration File](#configuration-file), or using
[Environment Variables](#environment-variables).

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
[Tuwunel] uses the `client_id` as the unique identifier of the provider and associates stored identities with it. The
`client_id` and `issuer_url` values must never change after users have logged in.
{{< /callout >}}

#### Configuration File

To configure [Tuwunel] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```toml {title="tuwunel.toml"}
[[global.identity_provider]]
brand = "Authelia"
name = "Authelia"
client_id = "tuwunel"
client_secret = "insecure_secret"
issuer_url = "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}"
callback_url = "https://tuwunel.{{< sitevar name="domain" nojs="example.com" >}}/_matrix/client/unstable/login/sso/callback/tuwunel"
```

#### Environment Variables

To configure [Tuwunel] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
TUWUNEL_IDENTITY_PROVIDER__0__BRAND=Authelia
TUWUNEL_IDENTITY_PROVIDER__0__NAME=Authelia
TUWUNEL_IDENTITY_PROVIDER__0__CLIENT_ID=tuwunel
TUWUNEL_IDENTITY_PROVIDER__0__CLIENT_SECRET=insecure_secret
TUWUNEL_IDENTITY_PROVIDER__0__ISSUER_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
TUWUNEL_IDENTITY_PROVIDER__0__CALLBACK_URL=https://tuwunel.{{< sitevar name="domain" nojs="example.com" >}}/_matrix/client/unstable/login/sso/callback/tuwunel
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  tuwunel:
    environment:
      TUWUNEL_IDENTITY_PROVIDER__0__BRAND: 'Authelia'
      TUWUNEL_IDENTITY_PROVIDER__0__NAME: 'Authelia'
      TUWUNEL_IDENTITY_PROVIDER__0__CLIENT_ID: 'tuwunel'
      TUWUNEL_IDENTITY_PROVIDER__0__CLIENT_SECRET: 'insecure_secret'
      TUWUNEL_IDENTITY_PROVIDER__0__ISSUER_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      TUWUNEL_IDENTITY_PROVIDER__0__CALLBACK_URL: 'https://tuwunel.{{< sitevar name="domain" nojs="example.com" >}}/_matrix/client/unstable/login/sso/callback/tuwunel'
```

## See Also

- [Tuwunel Identity Providers Documentation](https://github.com/matrix-construct/tuwunel/blob/main/docs/authentication/providers.md)
- [Tuwunel Authelia Provider Documentation](https://github.com/matrix-construct/tuwunel/blob/main/docs/authentication/providers/authelia.md)

[Authelia]: https://www.authelia.com
[Tuwunel]: https://github.com/matrix-construct/tuwunel
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
