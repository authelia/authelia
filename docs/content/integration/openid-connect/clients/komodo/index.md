---
title: "Komodo"
description: "Integrating Komodo with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-05-10T10:01:23+00:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/komodo/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Komodo | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Komodo with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.7](https://github.com/authelia/authelia/releases/tag/v4.39.7)
- [Komodo]
  - [v1.17.5](https://github.com/moghtech/komodo/releases/tag/v1.17.5)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://komodo.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://komodo.{{< sitevar name="domain" nojs="example.com" >}}/auth/oidc/callback`.
    This means if you change this value, you need to update the redirect URI.
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `komodo`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Komodo] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'komodo'
        client_name: 'Komodo'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        require_pkce: true
        pkce_challenge_method: 'S256'
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://komodo.{{< sitevar name="domain" nojs="example.com" >}}/auth/oidc/callback'
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

### Application

To configure [Komodo] there are two methods, using the [Configuration File](#configuration-file) or using the
[Environment Variables](#environment-variables).

#### Configuration File

To configure [Komodo] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```toml {title="config.toml"}
host = "https://komodo.{{< sitevar name="domain" nojs="example.com" >}}"
oidc_enabled = true
oidc_provider = "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}"
oidc_client_id = "komodo"
oidc_client_secret = "insecure_secret"
```

#### Environment Variables

To configure [Komodo] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
KOMODO_HOST=https://komodo.{{< sitevar name="domain" nojs="example.com" >}}
KOMODO_OIDC_ENABLED=true
KOMODO_OIDC_PROVIDER=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
KOMODO_OIDC_CLIENT_ID=komodo
KOMODO_OIDC_CLIENT_SECRET=insecure_secret
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  komodo:
    environment:
      KOMODO_HOST: 'https://komodo.{{< sitevar name="domain" nojs="example.com" >}}'
      KOMODO_OIDC_ENABLED: 'true'
      KOMODO_OIDC_PROVIDER: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      KOMODO_OIDC_CLIENT_ID: 'komodo'
      KOMODO_OIDC_CLIENT_SECRET: 'insecure_secret'
```

## See Also

- [Komodo Advanced Configuration OIDC/OAuth2.0 Documentation](https://komo.do/docs/setup/advanced#oidc--oauth2)

[Authelia]: https://www.authelia.com
[Komodo]: https://komo.do/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
