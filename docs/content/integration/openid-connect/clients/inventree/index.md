---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "InvenTree"
description: "A guide on integrating InvenTree with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "InvenTree | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring InvenTree with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [InvenTree]
  - [1.7.1](https://github.com/inventree/InvenTree/releases/tag/1.7.1)

{{% oidc-common bugs="client-credentials-encoding" %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://inventree.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://inventree.{{< sitevar name="domain" nojs="example.com" >}}/accounts/oidc/authelia/login/callback/`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `inventree`
- **Client Secret:** `insecure_secret`
- **Provider ID:** `authelia`
  - This option determines the redirect URI in the format of
    `https://inventree.{{< sitevar name="domain" nojs="example.com" >}}/accounts/oidc/authelia/login/callback/`.
    This means if you change this value, you need to update the redirect URI.

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [InvenTree] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'inventree'
        client_name: 'InvenTree'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://inventree.{{< sitevar name="domain" nojs="example.com" >}}/accounts/oidc/authelia/login/callback/'
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

To configure [InvenTree] there are two methods, using the [Configuration File](#configuration-file), or using
[Environment Variables](#environment-variables). Regardless of the method single sign-on must also be enabled using the
[Web GUI](#web-gui).

#### Configuration File

To configure [InvenTree] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```yaml {title="config.yaml"}
social_backends:
  - 'allauth.socialaccount.providers.openid_connect'
social_providers:
  openid_connect:
    SCOPE:
      - 'openid'
      - 'profile'
      - 'email'
      - 'groups'
    OAUTH_PKCE_ENABLED: true
    APPS:
      - provider_id: 'authelia'
        name: 'Authelia'
        client_id: 'inventree'
        secret: 'insecure_secret'
        settings:
          server_url: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
          token_auth_method: 'client_secret_basic'
```

#### Environment Variables

To configure [InvenTree] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables. The `INVENTREE_SOCIAL_PROVIDERS` value is the minified JSON format of the `social_providers` value from the
[Configuration File](#configuration-file) example:

##### Standard

```shell {title=".env"}
INVENTREE_SOCIAL_BACKENDS=allauth.socialaccount.providers.openid_connect
INVENTREE_SOCIAL_PROVIDERS={"openid_connect":{"SCOPE":["openid","profile","email","groups"],"OAUTH_PKCE_ENABLED":true,"APPS":[{"provider_id":"authelia","name":"Authelia","client_id":"inventree","secret":"insecure_secret","settings":{"server_url":"https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}","token_auth_method":"client_secret_basic"}}]}}
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  inventree-server:
    environment:
      INVENTREE_SOCIAL_BACKENDS: 'allauth.socialaccount.providers.openid_connect'
      INVENTREE_SOCIAL_PROVIDERS: '{"openid_connect":{"SCOPE":["openid","profile","email","groups"],"OAUTH_PKCE_ENABLED":true,"APPS":[{"provider_id":"authelia","name":"Authelia","client_id":"inventree","secret":"insecure_secret","settings":{"server_url":"https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}","token_auth_method":"client_secret_basic"}}]}}'
```

#### Web GUI

To enable single sign-on in [InvenTree], use the following instructions:

1. Login to [InvenTree] as an administrator.
2. Navigate to `System Settings` > `Authentication`.
3. Enable the following options:
   - `Enable SSO`
   - `Enable SSO registration`
   - `Auto-fill SSO users`
4. Optionally enable `Enable SSO group sync` and configure the following options:
   - SSO group key: `groups`
   - SSO group map: a JSON mapping of Authelia groups to [InvenTree] groups, for example
     `{"inventree-admins": "admins"}`

## See Also

- [InvenTree Single Sign On Documentation](https://docs.inventree.org/en/latest/settings/SSO/)
- [django-allauth OpenID Connect Provider Documentation](https://docs.allauth.org/en/latest/socialaccount/providers/openid_connect.html)

[Authelia]: https://www.authelia.com
[InvenTree]: https://inventree.org/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
