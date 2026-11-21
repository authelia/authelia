---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Wakapi"
description: "A guide on integrating Wakapi with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Wakapi | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Wakapi with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Wakapi]
  - [v2.18.2](https://github.com/muety/wakapi/releases/tag/2.18.2)

{{% oidc-common bugs="claims-hydration" %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://wakapi.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://wakapi.{{< sitevar name="domain" nojs="example.com" >}}/oidc/authelia/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `wakapi`
- **Client Secret:** `insecure_secret`
- **Provider Name:** `authelia`
  - This option determines the redirect URI in the format of
    `https://wakapi.{{< sitevar name="domain" nojs="example.com" >}}/oidc/authelia/callback`.
    This means if you change this value, you need to update the redirect URI.

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Wakapi] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'wakapi'
        client_name: 'Wakapi'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://wakapi.{{< sitevar name="domain" nojs="example.com" >}}/oidc/authelia/callback'
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

#### Configuration Escape Hatch

{{% oidc-escape-hatch-claims-hydration client_id="wakapi" claims="email,preferred_username" %}}

### Application

To configure [Wakapi] there are two methods, using the [Configuration File](#configuration-file), or using
[Environment Variables](#environment-variables).

#### Configuration File

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
Generally the configuration file is named `config.yml`.
{{< /callout >}}

To configure [Wakapi] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```yaml {title="config.yml"}
server:
  public_url: 'https://wakapi.{{< sitevar name="domain" nojs="example.com" >}}'
security:
  oidc_allow_signup: true
  oidc:
    - name: 'authelia'
      display_name: 'Authelia'
      client_id: 'wakapi'
      client_secret: 'insecure_secret'
      endpoint: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
```

#### Environment Variables

To configure [Wakapi] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
WAKAPI_PUBLIC_URL=https://wakapi.{{< sitevar name="domain" nojs="example.com" >}}
WAKAPI_OIDC_ALLOW_SIGNUP=true
WAKAPI_OIDC_PROVIDERS_0_NAME=authelia
WAKAPI_OIDC_PROVIDERS_0_DISPLAY_NAME=Authelia
WAKAPI_OIDC_PROVIDERS_0_CLIENT_ID=wakapi
WAKAPI_OIDC_PROVIDERS_0_CLIENT_SECRET=insecure_secret
WAKAPI_OIDC_PROVIDERS_0_ENDPOINT=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  wakapi:
    environment:
      WAKAPI_PUBLIC_URL: 'https://wakapi.{{< sitevar name="domain" nojs="example.com" >}}'
      WAKAPI_OIDC_ALLOW_SIGNUP: 'true'
      WAKAPI_OIDC_PROVIDERS_0_NAME: 'authelia'
      WAKAPI_OIDC_PROVIDERS_0_DISPLAY_NAME: 'Authelia'
      WAKAPI_OIDC_PROVIDERS_0_CLIENT_ID: 'wakapi'
      WAKAPI_OIDC_PROVIDERS_0_CLIENT_SECRET: 'insecure_secret'
      WAKAPI_OIDC_PROVIDERS_0_ENDPOINT: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
```

## See Also

- [Wakapi OpenID Connect Login (SSO) Documentation](<https://github.com/muety/wakapi/wiki/OpenID-Connect-login-(SSO)>)
- [Wakapi Configuration Options Documentation](https://github.com/muety/wakapi#-configuration-options)

[Authelia]: https://www.authelia.com
[Wakapi]: https://wakapi.dev
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
