---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "OpenCTI"
description: "A guide on integrating OpenCTI with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "OpenCTI | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring OpenCTI with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [OpenCTI]
  - [7.261109.0](https://github.com/OpenCTI-Platform/opencti/releases/tag/7.261109.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://opencti.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://opencti.{{< sitevar name="domain" nojs="example.com" >}}/auth/oic/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `opencti`
- **Client Secret:** `insecure_secret`
- **Groups:** The `opencti-admins` and `opencti-users` groups are mapped to the `Administrators` and `Default`
  [OpenCTI] groups respectively.

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [OpenCTI] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'opencti'
        client_name: 'OpenCTI'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://opencti.{{< sitevar name="domain" nojs="example.com" >}}/auth/oic/callback'
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

To configure [OpenCTI] there is one method, using [Environment Variables](#environment-variables).

{{< callout context="note" title="Enterprise Edition" icon="outline/info-circle" >}}
Starting with [OpenCTI] 7.260224.0 authentication strategies other than local authentication are an Enterprise
Edition feature. Authentication providers can also be managed from the [OpenCTI] interface, and providers defined by
environment variables are migrated to it.
{{< /callout >}}

#### Environment Variables

To configure [OpenCTI] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
APP__BASE_URL=https://opencti.{{< sitevar name="domain" nojs="example.com" >}}/
PROVIDERS__OPENID__STRATEGY=OpenIDConnectStrategy
PROVIDERS__OPENID__CONFIG__LABEL=Authelia
PROVIDERS__OPENID__CONFIG__ISSUER=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
PROVIDERS__OPENID__CONFIG__CLIENT_ID=opencti
PROVIDERS__OPENID__CONFIG__CLIENT_SECRET=insecure_secret
PROVIDERS__OPENID__CONFIG__REDIRECT_URIS=["https://opencti.{{< sitevar name="domain" nojs="example.com" >}}/auth/oic/callback"]
PROVIDERS__OPENID__CONFIG__LOGOUT_REMOTE=false
PROVIDERS__OPENID__CONFIG__GROUPS_MANAGEMENT__GROUPS_SCOPE=groups
PROVIDERS__OPENID__CONFIG__GROUPS_MANAGEMENT__GROUPS_PATH=["groups"]
PROVIDERS__OPENID__CONFIG__GROUPS_MANAGEMENT__GROUPS_MAPPING=["opencti-admins:Administrators","opencti-users:Default"]
PROVIDERS__OPENID__CONFIG__GROUPS_MANAGEMENT__READ_USERINFO=true
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  opencti:
    environment:
      APP__BASE_URL: 'https://opencti.{{< sitevar name="domain" nojs="example.com" >}}/'
      PROVIDERS__OPENID__STRATEGY: 'OpenIDConnectStrategy'
      PROVIDERS__OPENID__CONFIG__LABEL: 'Authelia'
      PROVIDERS__OPENID__CONFIG__ISSUER: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      PROVIDERS__OPENID__CONFIG__CLIENT_ID: 'opencti'
      PROVIDERS__OPENID__CONFIG__CLIENT_SECRET: 'insecure_secret'
      PROVIDERS__OPENID__CONFIG__REDIRECT_URIS: '["https://opencti.{{< sitevar name="domain" nojs="example.com" >}}/auth/oic/callback"]'
      PROVIDERS__OPENID__CONFIG__LOGOUT_REMOTE: 'false'
      PROVIDERS__OPENID__CONFIG__GROUPS_MANAGEMENT__GROUPS_SCOPE: 'groups'
      PROVIDERS__OPENID__CONFIG__GROUPS_MANAGEMENT__GROUPS_PATH: '["groups"]'
      PROVIDERS__OPENID__CONFIG__GROUPS_MANAGEMENT__GROUPS_MAPPING: '["opencti-admins:Administrators","opencti-users:Default"]'
      PROVIDERS__OPENID__CONFIG__GROUPS_MANAGEMENT__READ_USERINFO: 'true'
```

## See Also

- [OpenCTI Authentication Documentation](https://docs.opencti.io/latest/deployment/authentication/)

[Authelia]: https://www.authelia.com
[OpenCTI]: https://filigran.io/solutions/open-cti/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
