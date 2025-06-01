---
title: "ROM Manager (RomM)"
description: "Integrating ROM Manager (RomM) with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
weight: 620
toc: true
support:
  level: community
  versions: true
  integration: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.17](https://github.com/authelia/authelia/releases/tag/v4.38.17)
- [ROM Manager]
  - [v3.9.0](https://github.com/rommapp/romm/releases/tag/3.9.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://romm.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `romm`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

{{% oidc-conformance-claims claims="email,email_verified,alt_emails,preferred_username,name" %}}

The following YAML configuration is an example __Authelia__ [client configuration] for use with [ROM Manager] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'romm'
        client_name: 'ROM Manager'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://romm.{{< sitevar name="domain" nojs="example.com" >}}/api/oauth/openid'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [ROM Manager] there is one method, using the [Environment Variables](#environment-variables).

#### Environment Variables

To configure [ROM Manager] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

##### Standard

```shell {title=".env"}
GF_SERVER_ROOT_URL=https://romm.{{< sitevar name="domain" nojs="example.com" >}}
GF_AUTH_GENERIC_OAUTH_ENABLED=true
GF_AUTH_GENERIC_OAUTH_NAME=Authelia
GF_AUTH_GENERIC_OAUTH_ICON=signin
GF_AUTH_GENERIC_OAUTH_CLIENT_ID=romm
GF_AUTH_GENERIC_OAUTH_CLIENT_SECRET=insecure_secret
GF_AUTH_GENERIC_OAUTH_SCOPES=openid profile email groups
GF_AUTH_GENERIC_OAUTH_EMPTY_SCOPES=false
GF_AUTH_GENERIC_OAUTH_AUTH_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization
GF_AUTH_GENERIC_OAUTH_TOKEN_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token
GF_AUTH_GENERIC_OAUTH_API_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo
GF_AUTH_GENERIC_OAUTH_LOGIN_ATTRIBUTE_PATH=preferred_username
GF_AUTH_GENERIC_OAUTH_GROUPS_ATTRIBUTE_PATH=groups
GF_AUTH_GENERIC_OAUTH_NAME_ATTRIBUTE_PATH=name
GF_AUTH_GENERIC_OAUTH_USE_PKCE=true
GF_AUTH_GENERIC_OAUTH_ROLE_ATTRIBUTE_PATH=
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  romm:
    environment:
      GF_SERVER_ROOT_URL: 'https://romm.{{< sitevar name="domain" nojs="example.com" >}}'
      GF_AUTH_GENERIC_OAUTH_ENABLED: 'true'
      GF_AUTH_GENERIC_OAUTH_NAME: 'Authelia'
      GF_AUTH_GENERIC_OAUTH_ICON: 'signin'
      GF_AUTH_GENERIC_OAUTH_CLIENT_ID: 'romm'
      GF_AUTH_GENERIC_OAUTH_CLIENT_SECRET: 'insecure_secret'
      GF_AUTH_GENERIC_OAUTH_SCOPES: 'openid profile email groups'
      GF_AUTH_GENERIC_OAUTH_EMPTY_SCOPES: 'false'
      GF_AUTH_GENERIC_OAUTH_AUTH_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization'
      GF_AUTH_GENERIC_OAUTH_TOKEN_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token'
      GF_AUTH_GENERIC_OAUTH_API_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo'
      GF_AUTH_GENERIC_OAUTH_LOGIN_ATTRIBUTE_PATH: 'preferred_username'
      GF_AUTH_GENERIC_OAUTH_GROUPS_ATTRIBUTE_PATH: 'groups'
      GF_AUTH_GENERIC_OAUTH_NAME_ATTRIBUTE_PATH: 'name'
      GF_AUTH_GENERIC_OAUTH_USE_PKCE: 'true'
      GF_AUTH_GENERIC_OAUTH_ROLE_ATTRIBUTE_PATH: ''
```

## See Also

- [ROM Manager OIDC Setup With Authelia Documentation](https://docs.romm.app/latest/OIDC-Guides/OIDC-Setup-With-Authelia/)

[Authelia]: https://www.authelia.com
[ROM Manager]: https://romm.app/
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
