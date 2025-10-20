---
title: "Ryot (Roll Your Own Tracker)"
description: "Integrating Ryot with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-07-26T19:41:59+10:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/ryot/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Ryot (Roll Your Own Tracker) | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Ryot (Roll Your Own Tracker) with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.13](https://github.com/authelia/authelia/releases/tag/v4.39.13)
- [Ryot]
  - [v8.9.0](https://github.com/IgnisDa/ryot/releases/tag/v8.9.0)

{{% oidc-common bugs="claims-hydration" %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://ryot.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://ryot.{{< sitevar name="domain" nojs="example.com" >}}/auth/oidc/callback`.
    This means if you change this value, you need to update the redirect URI.
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `ryot`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Ryot] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'ryot'
        client_name: 'Ryot'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        require_pkce: false
        pkce_challenge_method: ''
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://ryot.{{< sitevar name="domain" nojs="example.com" >}}/api/auth'
        scopes:
          - 'openid'
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

{{% oidc-escape-hatch-claims-hydration client_id="ryot" claims="email" %}}

### Application

To configure [Ryot] there are two methods, using the [Configuration File](#configuration-file) or using the
[Environment Variables](#environment-variables).

#### Configuration File

To configure [Ryot] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```toml {title="config.yaml"}
frontend:
  url: 'https://ryot.{{< sitevar name="domain" nojs="example.com" >}}'
  oidc_button_label: 'Use Authelia'
server:
  odic:
    client_id: 'ryot'
    client_secret: 'insecure_secret'
    issuer_url: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
```

#### Environment Variables

To configure [Ryot] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
FRONTEND_URL=https://ryot.{{< sitevar name="domain" nojs="example.com" >}}
SERVER_OIDC_CLIENT_ID=ryot
SERVER_OIDC_CLIENT_SECRET=insecure_secret
SERVER_OIDC_ISSUER_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
FRONTEND_OIDC_BUTTON_LABEL=Use Authelia
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  ryot:
    environment:
      FRONTEND_URL: 'https://ryot.{{< sitevar name="domain" nojs="example.com" >}}'
      SERVER_OIDC_CLIENT_ID: 'ryot'
      SERVER_OIDC_CLIENT_SECRET: 'insecure_secret'
      SERVER_OIDC_ISSUER_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      FRONTEND_OIDC_BUTTON_LABEL: 'Use Authelia'
```

## See Also

- [Ryot Authentication Guide Documentation](https://docs.ryot.io/guides/authentication.html)

[Authelia]: https://www.authelia.com
[Ryot]: https://ryot.io/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
