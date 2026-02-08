---
title: "Papra"
description: "Integrating Papra with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 620
toc: true
aliases: []
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Papra | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Papra with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.15](https://github.com/authelia/authelia/releases/tag/v4.39.15)
- [Papra]
  - [v26.1.0](https://github.com/papra-hq/papra/pkgs/container/papra/656320542?tag=26.1.0-rootless)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://papra.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `papra`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Papra] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'papra'
        client_name: 'Papra'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://papra.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/oauth2/callback/authelia'
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

To configure [Papra] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [Papra] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment variables:

##### Standard

```shell {title=".env"}
AUTH_PROVIDERS_CUSTOMS=[{"providerId": "authelia","providerName": "Authelia","providerIconUrl": "https://www.authelia.com/images/branding/logo-cropped.png","clientId": "papra","clientSecret": "insecure_secret","type": "oidc","discoveryUrl": "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration","scopes": ["openid", "profile", "email"]}]
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  papra:
    environment:
      AUTH_PROVIDERS_CUSTOMS: '[{"providerId": "authelia","providerName": "Authelia","providerIconUrl": "https://www.authelia.com/images/branding/logo-cropped.png","clientId": "papra","clientSecret": "insecure_secret","pkce": true,"type": "oidc","discoveryUrl": "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration","scopes": ["openid", "profile", "email"]}]'
```

## See Also

- [Papra OAuth2 Documentation](https://docs.papra.app/guides/setup-custom-oauth2-providers/)

[Papra]: https://papra.app/
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
