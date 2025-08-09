---
title: "LibreChat"
description: "Integrating LibreChat with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-09-18T18:02:11+10:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/librechat/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "LibreChat | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring LibreChat with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.10](https://github.com/authelia/authelia/releases/tag/v4.38.10)
- [LibreChat]
  - [v0.7.5](https://www.librechat.ai/changelog/v0.7.5)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://librechat.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Application Session Secret:__ `insecure_session_secret`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `librechat`
- __Client Secret:__ `insecure_secret`

_**Note:** The application session secret should be randomly generated in a similar fashion to the client secret, but should
not be the same value as the session secret. Users should refer to LibreChat support for more information._

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [LibreChat] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'librechat'
        client_name: 'LibreChat'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://librechat.{{< sitevar name="domain" nojs="example.com" >}}/oauth/openid/callback'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
          - 'refresh_token'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [LibreChat] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [LibreChat] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment variables:

##### Standard

```shell {title=".env"}
ALLOW_SOCIAL_LOGIN=true
OPENID_BUTTON_LABEL=Log in with Authelia
OPENID_ISSUER=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration
OPENID_CLIENT_ID=librechat
OPENID_CLIENT_SECRET=insecure_secret
OPENID_SESSION_SECRET=insecure_session_secret
OPENID_CALLBACK_URL=/oauth/openid/callback
OPENID_SCOPE=openid profile email
OPENID_IMAGE_URL=https://www.authelia.com/images/branding/logo-cropped.png
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  librechat:
    environment:
      ALLOW_SOCIAL_LOGIN: 'true'
      OPENID_BUTTON_LABEL: 'Log in with Authelia'
      OPENID_ISSUER: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration'
      OPENID_CLIENT_ID: 'librechat'
      OPENID_CLIENT_SECRET: 'insecure_secret'
      OPENID_SESSION_SECRET: 'insecure_session_secret'
      OPENID_CALLBACK_URL: '/oauth/openid/callback'
      OPENID_SCOPE: 'openid profile email'
      OPENID_IMAGE_URL: 'https://www.authelia.com/images/branding/logo-cropped.png'
```

[LibreChat]: https://www.librechat.ai/
