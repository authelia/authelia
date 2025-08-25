---
title: "Mastodon"
description: "Integrating Mastodon with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-04-13T13:46:05+10:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/mastodon/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Mastodon | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Mastodon with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
- [Mastodon]
  - [v4.2.8](https://github.com/mastodon/mastodon/releases/tag/v4.2.8)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://mastodon.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `mastodon`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Mastodon] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'mastodon'
        client_name: 'Mastodon'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://mastodon.{{< sitevar name="domain" nojs="example.com" >}}/auth/auth/openid_connect/callback'
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

To configure [Mastodon] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [Mastodon] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment variables:

##### Standard

```shell {title=".env"}
OIDC_ENABLED=true
OIDC_DISPLAY_NAME=Authelia
OIDC_DISCOVERY=true
OIDC_ISSUER=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
OIDC_SCOPE=openid,profile,email
OIDC_UID_FIELD=preferred_username
OIDC_CLIENT_ID=mastodon
OIDC_CLIENT_SECRET=insecure_secret
OIDC_REDIRECT_URI=https://mastodon.{{< sitevar name="domain" nojs="example.com" >}}/auth/auth/openid_connect/callback
OIDC_SECURITY_ASSUME_EMAIL_IS_VERIFIED=true
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  mastodon:
    environment:
      OIDC_ENABLED: 'true'
      OIDC_DISPLAY_NAME: 'Authelia'
      OIDC_DISCOVERY: 'true'
      OIDC_ISSUER: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      OIDC_SCOPE: 'openid,profile,email'
      OIDC_UID_FIELD: 'preferred_username'
      OIDC_CLIENT_ID: 'mastodon'
      OIDC_CLIENT_SECRET: 'insecure_secret'
      OIDC_REDIRECT_URI: 'https://mastodon.{{< sitevar name="domain" nojs="example.com" >}}/auth/auth/openid_connect/callback'
      OIDC_SECURITY_ASSUME_EMAIL_IS_VERIFIED: 'true'
```

## See Also

- [OmniAuth OpenID Connect 1.0 Docs](https://github.com/omniauth/omniauth_openid_connect)

[Mastodon]: https://joinmastodon.org/
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
