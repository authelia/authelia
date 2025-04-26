---
title: "ownCloud Infinite Scale"
description: "Integrating ownCloud Infinite Scale with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-03-05T21:58:32+11:00
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
  - [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
- [ownCloud Infinite Scale]
  - 4.0.5

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://owncloud.{{< sitevar name="domain" nojs="example.com" >}}`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
- __Client ID:__
  - Web Application: `ocis`
  - Other Clients: the values of the other clients are static for compatibility with the native app
- __Client Secret:__
  - Web Application: `insecure_secret`
  - Other Clients: the values of the other clients are static for compatibility with the native app

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with
[ownCloud Infinite Scale] which will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    # Extend the access and refresh token lifespan from the default 30m to work around ownCloud client re-authentication prompts every few hours.
    # It should be possible to remove this once Authelia supports dynamic client registration (DCR).
    # Note: ownCloud's built-in IDP uses a value of 30d.
    access_token_lifespan: '2d'
    refresh_token_lifespan: '3d'

    cors:
      endpoints:
        - 'authorization'
        - 'token'
        - 'revocation'
        - 'introspection'
        - 'userinfo'
    clients:
      - client_id: 'ocis'
        client_name: 'ownCloud Infinite Scale'
        public: true
        redirect_uris:
          - 'https://owncloud.home.yourdomain.com/'
          - 'https://owncloud.home.yourdomain.com/oidc-callback.html'
          - 'https://owncloud.home.yourdomain.com/oidc-silent-redirect.html'
      - client_id: 'xdXOt13JKxym1B1QcEncf2XDkLAexMBFwiT9j6EfhhHFJhs2KM9jbjTmf8JBXE69'
        client_name: 'ownCloud desktop client'
        client_secret: 'UBntmLjC2yYCeHwsyj73Uwo9TAaecAetRwMw0xYcvNL9yRdLSUi0hUAHfvCHFeFh'
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        scopes:
          - 'openid'
          - 'groups'
          - 'profile'
          - 'email'
          - 'offline_access'
        redirect_uris:
          - 'http://127.0.0.1'
          - 'http://localhost'
      - client_id: 'e4rAsNUSIUs0lF4nbv9FmCeUkTlV9GdgTLDH1b5uie7syb90SzEVrbN7HIpmWJeD'
        client_name: 'ownCloud Android app'
        client_secret: 'dInFYGV33xKzhbRmpqQltYNdfLdJIfJ9L5ISoKhNoT9qZftpdWSP71VrpGR9pmoD'
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        scopes:
          - 'openid'
          - 'groups'
          - 'profile'
          - 'email'
          - 'offline_access'
        redirect_uris:
          - 'oc://android.owncloud.com'
      - client_id: 'mxd5OQDk6es5LzOzRvidJNfXLUZS2oN3oUFeXPP8LpPrhx3UroJFduGEYIBOxkY1'
        client_name: 'ownCloud iOS app'
        client_secret: 'KFeFWWEZO9TkisIQzR3fo7hfiMXlOpaqP8CFuTbSHzV1TUuGECglPxpiVKJfOXIx'
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        scopes:
          - 'openid'
          - 'groups'
          - 'profile'
          - 'email'
          - 'offline_access'
        redirect_uris:
          - 'oc://ios.owncloud.com'
          - 'oc.ios://ios.owncloud.com'
```

### Application

To configure [ownCloud Infinite Scale] there is one method, using the [Environment Variables](#environment-variables).

#### Environment Variables

To configure [ownCloud Infinite Scale] to utilize Authelia as an [OpenID Connect 1.0] Provider use the following environment
variables:

##### Standard

```shell {title=".env"}
WEB_OIDC_CLIENT_ID=ocis
PROXY_OIDC_ISSUER=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
PROXY_OIDC_REWRITE_WELLKNOWN=true
PROXY_OIDC_ACCESS_TOKEN_VERIFY_METHOD=none
PROXY_OIDC_SKIP_USER_INFO=false
PROXY_AUTOPROVISION_ACCOUNTS=false
PROXY_AUTOPROVISION_CLAIM_USERNAME=preferred_username
PROXY_AUTOPROVISION_CLAIM_EMAIL=email
PROXY_AUTOPROVISION_CLAIM_DISPLAYNAME=name
PROXY_AUTOPROVISION_CLAIM_GROUPS=groups
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  oics:
    environment:
      WEB_OIDC_CLIENT_ID: 'ocis'
      PROXY_OIDC_ISSUER: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      PROXY_OIDC_REWRITE_WELLKNOWN: 'true'
      PROXY_OIDC_ACCESS_TOKEN_VERIFY_METHOD: 'none'
      PROXY_OIDC_SKIP_USER_INFO: 'false'
      PROXY_AUTOPROVISION_ACCOUNTS: 'false'
      PROXY_AUTOPROVISION_CLAIM_USERNAME: 'preferred_username'
      PROXY_AUTOPROVISION_CLAIM_EMAIL: 'email'
      PROXY_AUTOPROVISION_CLAIM_DISPLAYNAME: 'name'
      PROXY_AUTOPROVISION_CLAIM_GROUPS: 'groups'
```

## See Also

- [Nextcloud OpenID Connect Login app]
- [Nextcloud OpenID Connect Login Documentation](https://github.com/pulsejet/nextcloud-oidc-login)

[Authelia]: https://www.authelia.com
[ownCloud Infinite Scale]: https://nextcloud.com/
[Nextcloud OpenID Connect Login app]: https://apps.nextcloud.com/apps/oidc_login
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
