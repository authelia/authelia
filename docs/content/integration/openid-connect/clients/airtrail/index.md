---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "AirTrail"
description: "A guide on integrating AirTrail with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "AirTrail | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring AirTrail with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [AirTrail]
  - [v3.14.0](https://github.com/johanohly/AirTrail/releases/tag/v3.14.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://airtrail.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://airtrail.{{< sitevar name="domain" nojs="example.com" >}}/login`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `airtrail`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [AirTrail] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'airtrail'
        client_name: 'AirTrail'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://airtrail.{{< sitevar name="domain" nojs="example.com" >}}/login'
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
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [AirTrail] there are two methods, using
[Environment Variables](#environment-variables), or using the [Web GUI](#web-gui).

#### Environment Variables

To configure [AirTrail] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
ORIGIN=https://airtrail.{{< sitevar name="domain" nojs="example.com" >}}
OAUTH_ENABLED=true
OAUTH_ISSUER_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
OAUTH_CLIENT_ID=airtrail
OAUTH_CLIENT_SECRET=insecure_secret
OAUTH_TOKEN_ENDPOINT_AUTH_METHOD=client_secret_post
OAUTH_SCOPE=openid profile email
OAUTH_BUTTON_TEXT=Log in with Authelia
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  airtrail:
    environment:
      ORIGIN: 'https://airtrail.{{< sitevar name="domain" nojs="example.com" >}}'
      OAUTH_ENABLED: 'true'
      OAUTH_ISSUER_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      OAUTH_CLIENT_ID: 'airtrail'
      OAUTH_CLIENT_SECRET: 'insecure_secret'
      OAUTH_TOKEN_ENDPOINT_AUTH_METHOD: 'client_secret_post'
      OAUTH_SCOPE: 'openid profile email'
      OAUTH_BUTTON_TEXT: 'Log in with Authelia'
```

#### Web GUI

{{< callout context="tip" title="Help Wanted" icon="outline/rocket" >}}
We would love screenshots of this configuration!
{{< /callout >}}

To configure [AirTrail] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [AirTrail] as an administrator.
2. Navigate to Settings.
3. Navigate to OAuth.
4. Configure the following options:
   - Enable OAuth: Enabled
   - Issuer URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - Client ID: `airtrail`
   - Client Secret: `insecure_secret`
   - Scope: `openid profile email`
   - Token Endpoint Auth Method: `client_secret_post`
   - Button Text: `Log in with Authelia`
5. Click Save.

## See Also

- [AirTrail OAuth Authentication Documentation](https://airtrail.johan.ohly.dk/docs/features/oauth)

[Authelia]: https://www.authelia.com
[AirTrail]: https://airtrail.johan.ohly.dk/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
