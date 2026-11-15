---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "n8n"
description: "A guide on integrating n8n with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "n8n | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring n8n with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [n8n]
  - [v2.47.5](https://github.com/n8n-io/n8n/releases/tag/n8n%402.47.5)

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
The [n8n] OpenID Connect 1.0 single sign-on feature requires an n8n Enterprise plan.
{{< /callout >}}

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://n8n.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://n8n.{{< sitevar name="domain" nojs="example.com" >}}/rest/sso/oidc/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `n8n`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [n8n] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'n8n'
        client_name: 'n8n'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://n8n.{{< sitevar name="domain" nojs="example.com" >}}/rest/sso/oidc/callback'
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

To configure [n8n] there are two methods, using
[Environment Variables](#environment-variables), or using the [Web GUI](#web-gui).

#### Environment Variables

To configure [n8n] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
N8N_EDITOR_BASE_URL=https://n8n.{{< sitevar name="domain" nojs="example.com" >}}
N8N_SSO_MANAGED_BY_ENV=true
N8N_SSO_OIDC_LOGIN_ENABLED=true
N8N_SSO_OIDC_DISCOVERY_ENDPOINT=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration
N8N_SSO_OIDC_CLIENT_ID=n8n
N8N_SSO_OIDC_CLIENT_SECRET=insecure_secret
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  n8n:
    environment:
      N8N_EDITOR_BASE_URL: 'https://n8n.{{< sitevar name="domain" nojs="example.com" >}}'
      N8N_SSO_MANAGED_BY_ENV: 'true'
      N8N_SSO_OIDC_LOGIN_ENABLED: 'true'
      N8N_SSO_OIDC_DISCOVERY_ENDPOINT: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration'
      N8N_SSO_OIDC_CLIENT_ID: 'n8n'
      N8N_SSO_OIDC_CLIENT_SECRET: 'insecure_secret'
```

#### Web GUI

To configure [n8n] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [n8n] as an instance owner or administrator.
2. Navigate to `Settings`.
3. Navigate to `Authentication`.
4. Navigate to `Single sign-on (SSO)`.
5. Select `OIDC` from the `Select Authentication Protocol` dropdown.
6. Ensure the displayed redirect URL is `https://n8n.{{< sitevar name="domain" nojs="example.com" >}}/rest/sso/oidc/callback`.
7. Configure the following options:
   - Discovery Endpoint: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration`
   - Client ID: `n8n`
   - Client Secret: `insecure_secret`
8. Click `Save settings`.
9. Set `OIDC` to `Activated`.

If [n8n] is running behind a reverse proxy and displays an internal redirect URL, set the `N8N_EDITOR_BASE_URL`
environment variable to `https://n8n.{{< sitevar name="domain" nojs="example.com" >}}`.

## See Also

- [n8n Set up OIDC Documentation](https://docs.n8n.io/user-management/oidc/setup/)
- [n8n SSO Environment Variables Documentation](https://docs.n8n.io/deploy/host-n8n/configure-n8n/basic-configuration/use-environment-variables/sso/)

[Authelia]: https://www.authelia.com
[n8n]: https://n8n.io/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
