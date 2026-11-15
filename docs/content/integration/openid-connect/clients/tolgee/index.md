---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Tolgee"
description: "A guide on integrating Tolgee with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Tolgee | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Tolgee with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Tolgee]
  - [v3.242.7](https://github.com/tolgee/tolgee-platform/releases/tag/v3.242.7)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://tolgee.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://tolgee.{{< sitevar name="domain" nojs="example.com" >}}/login/auth_callback/oauth2`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `tolgee`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Tolgee] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'tolgee'
        client_name: 'Tolgee'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://tolgee.{{< sitevar name="domain" nojs="example.com" >}}/login/auth_callback/oauth2'
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

To configure [Tolgee] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [Tolgee] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
TOLGEE_AUTHENTICATION_ENABLED=true
TOLGEE_AUTHENTICATION_OAUTH2_CLIENT_ID=tolgee
TOLGEE_AUTHENTICATION_OAUTH2_CLIENT_SECRET=insecure_secret
TOLGEE_AUTHENTICATION_OAUTH2_SCOPES=openid,email,profile
TOLGEE_AUTHENTICATION_OAUTH2_AUTHORIZATION_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization
TOLGEE_AUTHENTICATION_OAUTH2_TOKEN_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token
TOLGEE_AUTHENTICATION_OAUTH2_USER_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  tolgee:
    environment:
      TOLGEE_AUTHENTICATION_ENABLED: 'true'
      TOLGEE_AUTHENTICATION_OAUTH2_CLIENT_ID: 'tolgee'
      TOLGEE_AUTHENTICATION_OAUTH2_CLIENT_SECRET: 'insecure_secret'
      TOLGEE_AUTHENTICATION_OAUTH2_SCOPES: 'openid,email,profile'
      TOLGEE_AUTHENTICATION_OAUTH2_AUTHORIZATION_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization'
      TOLGEE_AUTHENTICATION_OAUTH2_TOKEN_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token'
      TOLGEE_AUTHENTICATION_OAUTH2_USER_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo'
```

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
New users signing in via [OpenID Connect 1.0] are only able to register when they have been invited or when the
`TOLGEE_AUTHENTICATION_REGISTRATIONS_ALLOWED` option is set to `true`.
{{< /callout >}}

## See Also

- [Tolgee Configuration Documentation](https://docs.tolgee.io/platform/self_hosting/configuration)

[Authelia]: https://www.authelia.com
[Tolgee]: https://tolgee.io/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
