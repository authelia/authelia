---
title: "WeKan"
description: "Integrating WeKan with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-04-13T13:46:05+10:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/wekan/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "WeKan | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring WeKan with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
- [WeKan]
  - [v7.42](https://github.com/wekan/wekan/releases/tag/v7.42)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://wekan.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `wekan`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [WeKan] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'wekan'
        client_name: 'WeKan'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://wekan.{{< sitevar name="domain" nojs="example.com" >}}/_oauth/oidc'
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

To configure [WeKan] there is one method, using the [Environment Variables](#environment-variables).

#### Environment Variables

To configure [WeKan] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment variables:

##### Standard

```shell {title=".env"}
OAUTH2_ENABLED=true
OAUTH2_LOGIN_STYLE=redirect
OAUTH2_CLIENT_ID=wekan
OAUTH2_SECRET=insecure_secret
OAUTH2_SERVER_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
OAUTH2_AUTH_ENDPOINT=/api/oidc/authorization
OAUTH2_TOKEN_ENDPOINT=/api/oidc/token
OAUTH2_USERINFO_ENDPOINT=/api/oidc/userinfo
OAUTH2_ID_MAP=sub
OAUTH2_USERNAME_MAP=email
OAUTH2_FULLNAME_MAP=name
OAUTH2_EMAIL_MAP=email
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  wekan:
    environment:
      OAUTH2_ENABLED: 'true'
      OAUTH2_LOGIN_STYLE: 'redirect'
      OAUTH2_CLIENT_ID: 'wekan'
      OAUTH2_SECRET: 'insecure_secret'
      OAUTH2_SERVER_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      OAUTH2_AUTH_ENDPOINT: '/api/oidc/authorization'
      OAUTH2_TOKEN_ENDPOINT: '/api/oidc/token'
      OAUTH2_USERINFO_ENDPOINT: '/api/oidc/userinfo'
      OAUTH2_ID_MAP: 'sub'
      OAUTH2_USERNAME_MAP: 'email'
      OAUTH2_FULLNAME_MAP: 'name'
      OAUTH2_EMAIL_MAP: 'email'
```

## See Also

- [WeKan OAuth2 Documentation](https://github.com/wekan/wekan/wiki/OAuth2)

[WeKan]: https://github.com/wekan/wekan
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
