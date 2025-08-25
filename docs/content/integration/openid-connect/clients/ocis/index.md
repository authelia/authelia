---
title: "ownCloud Infinite Scale"
description: "Integrating ownCloud Infinite Scale with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/ocis/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "ownCloud Infinite Scale | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring ownCloud Infinite Scale with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
- [ownCloud Infinite Scale]
  - [v4.0.5](https://github.com/owncloud/ocis/releases/tag/v4.0.5)

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
    lifespans:
      custom:
        ocis:
          access_token: '2 days'
          refresh_token: '3 days'
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
        lifespan: 'ocis'
        public: true
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        scopes:
          - 'openid'
          - 'offline_access'
          - 'groups'
          - 'profile'
          - 'email'
        redirect_uris:
          - 'https://owncloud.{{< sitevar name="domain" nojs="example.com" >}}/'
          - 'https://owncloud.{{< sitevar name="domain" nojs="example.com" >}}/oidc-callback.html'
          - 'https://owncloud.{{< sitevar name="domain" nojs="example.com" >}}/oidc-silent-redirect.html'
          - 'https://owncloud.{{< sitevar name="domain" nojs="example.com" >}}/apps/openidconnect/redirect'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
          - 'refresh_token'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'none'
      - client_id: 'xdXOt13JKxym1B1QcEncf2XDkLAexMBFwiT9j6EfhhHFJhs2KM9jbjTmf8JBXE69'
        client_name: 'ownCloud Infinite Scale (Desktop Client)'
        client_secret: 'UBntmLjC2yYCeHwsyj73Uwo9TAaecAetRwMw0xYcvNL9yRdLSUi0hUAHfvCHFeFh'
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        scopes:
          - 'openid'
          - 'offline_access'
          - 'groups'
          - 'profile'
          - 'email'
        redirect_uris:
          - 'http://127.0.0.1'
          - 'http://localhost'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
          - 'refresh_token'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
      - client_id: 'e4rAsNUSIUs0lF4nbv9FmCeUkTlV9GdgTLDH1b5uie7syb90SzEVrbN7HIpmWJeD'
        client_name: 'ownCloud Infinite Scale (Android)'
        client_secret: 'dInFYGV33xKzhbRmpqQltYNdfLdJIfJ9L5ISoKhNoT9qZftpdWSP71VrpGR9pmoD'
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'oc://android.owncloud.com'
        scopes:
          - 'openid'
          - 'offline_access'
          - 'groups'
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
      - client_id: 'mxd5OQDk6es5LzOzRvidJNfXLUZS2oN3oUFeXPP8LpPrhx3UroJFduGEYIBOxkY1'
        client_name: 'ownCloud Infinite Scale (iOS)'
        client_secret: 'KFeFWWEZO9TkisIQzR3fo7hfiMXlOpaqP8CFuTbSHzV1TUuGECglPxpiVKJfOXIx'
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'oc://ios.owncloud.com'
          - 'oc.ios://ios.owncloud.com'
        scopes:
          - 'openid'
          - 'offline_access'
          - 'groups'
          - 'profile'
          - 'email'
        grant_types:
          - 'authorization_code'
          - 'refresh_token'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [ownCloud Infinite Scale] there is one method, using the [Environment Variables](#environment-variables).

#### Environment Variables

To configure [ownCloud Infinite Scale] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
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

- [ownCloud Infinite Scale]
- [ownCloud Infinite Scale IDP Service Configuration Documentation](https://doc.owncloud.com/ocis/next/deployment/services/s-list/idp.html)

[Authelia]: https://www.authelia.com
[ownCloud Infinite Scale]: https://owncloud.com/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
