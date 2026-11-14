---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Pomerium"
description: "A guide on integrating Pomerium with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Pomerium | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Pomerium with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Pomerium]
  - [v0.34.1](https://github.com/pomerium/pomerium/releases/tag/v0.34.1)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://authenticate.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This is the [Pomerium] authenticate service URL and it determines the redirect URI in the format of
    `https://authenticate.{{< sitevar name="domain" nojs="example.com" >}}/oauth2/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `pomerium`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Pomerium] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'pomerium'
        client_name: 'Pomerium'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://authenticate.{{< sitevar name="domain" nojs="example.com" >}}/oauth2/callback'
        scopes:
          - 'openid'
          - 'offline_access'
          - 'email'
          - 'profile'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
          - 'refresh_token'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [Pomerium] there are two methods, using the [Configuration File](#configuration-file), or using
[Environment Variables](#environment-variables).

#### Configuration File

To configure [Pomerium] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```yaml {title="config.yaml"}
authenticate_service_url: 'https://authenticate.{{< sitevar name="domain" nojs="example.com" >}}'
idp_provider: 'oidc'
idp_provider_url: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
idp_client_id: 'pomerium'
idp_client_secret: 'insecure_secret'
```

#### Environment Variables

To configure [Pomerium] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
AUTHENTICATE_SERVICE_URL=https://authenticate.{{< sitevar name="domain" nojs="example.com" >}}
IDP_PROVIDER=oidc
IDP_PROVIDER_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
IDP_CLIENT_ID=pomerium
IDP_CLIENT_SECRET=insecure_secret
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  pomerium:
    environment:
      AUTHENTICATE_SERVICE_URL: 'https://authenticate.{{< sitevar name="domain" nojs="example.com" >}}'
      IDP_PROVIDER: 'oidc'
      IDP_PROVIDER_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      IDP_CLIENT_ID: 'pomerium'
      IDP_CLIENT_SECRET: 'insecure_secret'
```

## See Also

- [Pomerium OIDC Identity Provider Documentation](https://www.pomerium.com/docs/integrations/user-identity/oidc)
- [Pomerium Identity Provider Settings Documentation](https://www.pomerium.com/docs/reference/identity-provider-settings)

[Authelia]: https://www.authelia.com
[Pomerium]: https://www.pomerium.com/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
