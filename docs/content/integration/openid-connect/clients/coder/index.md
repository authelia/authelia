---
title: "Coder"
description: "Integrating Coder with the Authelia OpenID Connect 1.0 Provider."
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
  title: "Coder | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Coder with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.7](https://github.com/authelia/authelia/releases/tag/v4.39.7)
- [Coder]
  - [v2.24.2](https://github.com/coder/coder/releases/tag/v2.24.2)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://coder.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `coder`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Coder] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'coder'
        client_name: 'Coder'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://coder.{{< sitevar name="domain" nojs="example.com" >}}/api/v2/users/oidc/callback'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
          - 'offline_access'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
          - 'refresh_token'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [Coder] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [Coder] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment variables:

##### Standard

```shell {title=".env"}
CODER_OIDC_ISSUER_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
CODER_OIDC_EMAIL_DOMAIN={{< sitevar name="domain" nojs="example.com" >}}
CODER_OIDC_CLIENT_ID=coder
CODER_OIDC_CLIENT_SECRET=insecure_secret
CODER_OIDC_SCOPES=openid,profile,email,offline_access
CODER_OIDC_EMAIL_FIELD=email
CODER_OIDC_IGNORE_EMAIL_VERIFIED=false
CODER_OIDC_USERNAME_FIELD=preferred_username
CODER_OIDC_SIGN_IN_TEXT=Sign in with Authelia
CODER_OIDC_ICON_URL=https://www.authelia.com/images/branding/logo-cropped.png
CODER_DISABLE_PASSWORD_AUTH=false
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  coder:
    environment:
      CODER_OIDC_ISSUER_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      CODER_OIDC_EMAIL_DOMAIN: '{{< sitevar name="domain" nojs="example.com" >}}'
      CODER_OIDC_CLIENT_ID: 'coder'
      CODER_OIDC_CLIENT_SECRET: 'insecure_secret'
      CODER_OIDC_SCOPES: 'openid,profile,email,offline_access'
      CODER_OIDC_EMAIL_FIELD: 'email'
      CODER_OIDC_IGNORE_EMAIL_VERIFIED: 'false'
      CODER_OIDC_USERNAME_FIELD: 'preferred_username'
      CODER_OIDC_SIGN_IN_TEXT: 'Sign in with Authelia'
      CODER_OIDC_ICON_URL: 'https://www.authelia.com/images/branding/logo-cropped.png'
      CODER_DISABLE_PASSWORD_AUTH: 'false'
```

## See Also

- [Coder OIDC Authentication Documentation](https://coder.com/docs/admin/users/oidc-auth)

[Coder]: https://coder.com/
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
