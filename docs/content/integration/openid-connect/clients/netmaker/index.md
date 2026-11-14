---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Netmaker"
description: "A guide on integrating Netmaker with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Netmaker | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Netmaker with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Netmaker]
  - [v1.7.0](https://github.com/gravitl/netmaker/releases/tag/v1.7.0)

{{% oidc-common bugs="claims-hydration" %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://dashboard.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Application API URL:** `https://api.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://api.{{< sitevar name="domain" nojs="example.com" >}}/api/oauth/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `netmaker`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Netmaker] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'netmaker'
        client_name: 'Netmaker'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://api.{{< sitevar name="domain" nojs="example.com" >}}/api/oauth/callback'
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
        token_endpoint_auth_method: 'client_secret_basic'
```

#### Configuration Escape Hatch

{{% oidc-escape-hatch-claims-hydration client_id="netmaker" claims="email,name" %}}

### Application

To configure [Netmaker] there are two methods, using the [Web GUI](#web-gui), or using
[Environment Variables](#environment-variables).

{{< callout context="note" title="Netmaker Pro" icon="outline/info-circle" >}}
The OAuth 2.0 and OpenID Connect 1.0 integrations are only available in Netmaker Pro.
{{< /callout >}}

#### Web GUI

To configure [Netmaker] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to the [Netmaker] dashboard as an administrator.
2. Navigate to Settings.
3. Navigate to Security & Authentication.
4. Select the OIDC identity provider integration.
5. Configure the following options:
   - Client ID: `netmaker`
   - Client Secret: `insecure_secret`
   - OIDC Issuer: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
6. Click Save.

#### Environment Variables

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
Starting with [Netmaker] v0.90.0 the OAuth 2.0 integrations are managed using the [Web GUI](#web-gui). The environment
variables are only used for the initial server settings.
{{< /callout >}}

To configure [Netmaker] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
FRONTEND_URL=https://dashboard.{{< sitevar name="domain" nojs="example.com" >}}
AUTH_PROVIDER=oidc
OIDC_ISSUER=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
CLIENT_ID=netmaker
CLIENT_SECRET=insecure_secret
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  netmaker:
    environment:
      FRONTEND_URL: 'https://dashboard.{{< sitevar name="domain" nojs="example.com" >}}'
      AUTH_PROVIDER: 'oidc'
      OIDC_ISSUER: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      CLIENT_ID: 'netmaker'
      CLIENT_SECRET: 'insecure_secret'
```

## See Also

- [Netmaker Identity Provider Integration Guide](https://docs.netmaker.io/docs/how-to-guides/identity-provider-integration-guide)
- [Netmaker Integrating OAuth Documentation](https://learn.netmaker.io/getting-started/server-installation/integrating-oauth)

[Authelia]: https://www.authelia.com
[Netmaker]: https://www.netmaker.io/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
