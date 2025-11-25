---
title: "Donetick"
description: "Integrating Donetick with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-08-19T08:32:16+00:00
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
  title: "Donetick | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Donetick with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.15](https://github.com/authelia/authelia/releases/tag/v4.39.15)
- [Donetick]
  - [v0.1.53](https://github.com/donetick/donetick/releases/tag/v0.1.53)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://donetick.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `donetick`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Donetick] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'donetick'
        client_name: 'Donetick'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://donetick.{{< sitevar name="domain" nojs="example.com" >}}/auth/oauth2'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [Donetick] there are two methods, using [Environment Variables](#environment-variables), or using the
[Configuration File](#configuration-file).

#### Environment Variables

To configure [Donetick] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment variables:

##### Standard

```shell {title=".env"}
DT_OAUTH2_NAME=Authelia
DT_OAUTH2_CLIENT_ID=donetick
DT_OAUTH2_CLIENT_SECRET=insecure_secret
DT_OAUTH2_SCOPE=openid profile email
DT_OAUTH2_AUTH_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization
DT_OAUTH2_TOKEN_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token
DT_OAUTH2_INFO_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo
DT_OAUTH2_REDIRECT_URL=https://donetick.{{< sitevar name="domain" nojs="example.com" >}}/auth/oauth2
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  donetick:
    environment:
      DT_OAUTH2_NAME: 'Authelia'
      DT_OAUTH2_CLIENT_ID: 'donetick'
      DT_OAUTH2_CLIENT_SECRET: 'insecure_secret'
      DT_OAUTH2_SCOPES: 'openid profile email'
      DT_OAUTH2_AUTH_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization'
      DT_OAUTH2_TOKEN_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token'
      DT_OAUTH2_INFO_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo'
      DT_OAUTH2_REDIRECT_URL: 'https://donetick.{{< sitevar name="domain" nojs="example.com" >}}/auth/oauth2'
```

##### Configuration File

To configure [Donetick] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```yaml {title="selfhosted.yml"}
oauth2:
  name: 'Authelia'
  client_id: 'donetick'
  client_secret: 'insecure_secret'
  scopes:
    - 'openid'
    - 'profile'
    - 'email'
  auth_url: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization'
  token_url: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token'
  user_info_url: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo'
  redirect_url: 'https://donetick.{{< sitevar name="domain" nojs="example.com" >}}/auth/oauth2'

```

### Android App

Please note that the Donetick app in version [0.1.34](https://github.com/donetick/donetick/releases/tag/v0.1.34) does not work with OpenID on self-hosted Donetick instances. See [issue #268](https://github.com/donetick/donetick/issues/268) for details.


## See Also

- [Donetick Configuration Documentation](https://docs.donetick.com/getting-started/configration)

[Donetick]: https://docs.donetick.com/
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
