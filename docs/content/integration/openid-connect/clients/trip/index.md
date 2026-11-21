---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "TRIP"
description: "A guide on integrating TRIP with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "TRIP | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring TRIP with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [TRIP]
  - [v1.50.1](https://github.com/itskovacs/trip/releases/tag/1.50.1)

{{% oidc-common bugs="client-credentials-encoding,claims-hydration" %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://trip.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://trip.{{< sitevar name="domain" nojs="example.com" >}}/auth`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `trip`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [TRIP] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'trip'
        client_name: 'TRIP'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://trip.{{< sitevar name="domain" nojs="example.com" >}}/auth'
        scopes:
          - 'openid'
          - 'profile'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

#### Configuration Escape Hatch

{{% oidc-escape-hatch-claims-hydration client_id="trip" claims="preferred_username" %}}

### Application

To configure [TRIP] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [TRIP] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title="storage/config.env"}
OIDC_DISCOVERY_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration
OIDC_CLIENT_ID=trip
OIDC_CLIENT_SECRET=insecure_secret
OIDC_REDIRECT_URI=https://trip.{{< sitevar name="domain" nojs="example.com" >}}/auth
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  app:
    environment:
      OIDC_DISCOVERY_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration'
      OIDC_CLIENT_ID: 'trip'
      OIDC_CLIENT_SECRET: 'insecure_secret'
      OIDC_REDIRECT_URI: 'https://trip.{{< sitevar name="domain" nojs="example.com" >}}/auth'
```

## See Also

- [TRIP OIDC Authelia Documentation](https://github.com/itskovacs/trip/blob/main/docs/docs/misc/authelia.md)
- [TRIP Configuration Documentation](https://github.com/itskovacs/trip/blob/main/docs/docs/getting-started/configuration.md)

[Authelia]: https://www.authelia.com
[TRIP]: https://github.com/itskovacs/trip
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
