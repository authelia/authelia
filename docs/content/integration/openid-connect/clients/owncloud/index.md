---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "ownCloud"
description: "A guide on integrating ownCloud with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
summary: ""
date: 2026-11-12T00:00:00+00:00
draft: false
images: []
weight: 620
toc: true
support:
  level: community
  versions: true
  integration: true
seo:
  title: "ownCloud | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring ownCloud with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [ownCloud] Server
  - [v11.1.0](https://github.com/owncloud/core/releases/tag/v11.1.0)
- [ownCloud OpenID Connect App]
  - [v2.4.2](https://github.com/owncloud/openidconnect/releases/tag/v2.4.2)

{{% oidc-common %}}

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
This guide is for the classic [ownCloud] Server. For ownCloud Infinite Scale see the
[ownCloud Infinite Scale](../ocis/index.md) guide.
{{< /callout >}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://owncloud.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://owncloud.{{< sitevar name="domain" nojs="example.com" >}}/index.php/apps/openidconnect/redirect`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:**
  - Web Application: `owncloud`
  - Other Clients: the values of the other clients are static for compatibility with the native apps
- **Client Secret:**
  - Web Application: `insecure_secret`
  - Other Clients: the values of the other clients are static for compatibility with the native apps
- **Access Token Audience:** `https://owncloud.{{< sitevar name="domain" nojs="example.com" >}}`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [ownCloud] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'owncloud'
        client_name: 'ownCloud'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        audience:
          - 'https://owncloud.{{< sitevar name="domain" nojs="example.com" >}}'
        requested_audience_mode: 'implicit'
        redirect_uris:
          - 'https://owncloud.{{< sitevar name="domain" nojs="example.com" >}}/index.php/apps/openidconnect/redirect'
        scopes:
          - 'openid'
          - 'offline_access'
          - 'profile'
          - 'email'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
          - 'refresh_token'
        access_token_signed_response_alg: 'RS256'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
      - client_id: 'xdXOt13JKxym1B1QcEncf2XDkLAexMBFwiT9j6EfhhHFJhs2KM9jbjTmf8JBXE69'
        client_name: 'ownCloud (Desktop Client)'
        client_secret: 'UBntmLjC2yYCeHwsyj73Uwo9TAaecAetRwMw0xYcvNL9yRdLSUi0hUAHfvCHFeFh'
        public: false
        application_type: 'native'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        audience:
          - 'https://owncloud.{{< sitevar name="domain" nojs="example.com" >}}'
        requested_audience_mode: 'implicit'
        redirect_uris:
          - 'http://127.0.0.1'
          - 'http://localhost'
        scopes:
          - 'openid'
          - 'offline_access'
          - 'profile'
          - 'email'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
          - 'refresh_token'
        access_token_signed_response_alg: 'RS256'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
      - client_id: 'e4rAsNUSIUs0lF4nbv9FmCeUkTlV9GdgTLDH1b5uie7syb90SzEVrbN7HIpmWJeD'
        client_name: 'ownCloud (Android)'
        client_secret: 'dInFYGV33xKzhbRmpqQltYNdfLdJIfJ9L5ISoKhNoT9qZftpdWSP71VrpGR9pmoD'
        public: false
        application_type: 'native'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        audience:
          - 'https://owncloud.{{< sitevar name="domain" nojs="example.com" >}}'
        requested_audience_mode: 'implicit'
        redirect_uris:
          - 'oc://android.owncloud.com'
        scopes:
          - 'openid'
          - 'offline_access'
          - 'profile'
          - 'email'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
          - 'refresh_token'
        access_token_signed_response_alg: 'RS256'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
      - client_id: 'mxd5OQDk6es5LzOzRvidJNfXLUZS2oN3oUFeXPP8LpPrhx3UroJFduGEYIBOxkY1'
        client_name: 'ownCloud (iOS)'
        client_secret: 'KFeFWWEZO9TkisIQzR3fo7hfiMXlOpaqP8CFuTbSHzV1TUuGECglPxpiVKJfOXIx'
        public: false
        application_type: 'native'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        audience:
          - 'https://owncloud.{{< sitevar name="domain" nojs="example.com" >}}'
        requested_audience_mode: 'implicit'
        redirect_uris:
          - 'oc://ios.owncloud.com'
        scopes:
          - 'openid'
          - 'offline_access'
          - 'profile'
          - 'email'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
          - 'refresh_token'
        access_token_signed_response_alg: 'RS256'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

Notes:

- [ownCloud] verifies the access token presented by the web application and the native apps. The access tokens are
  issued as JSON Web Tokens via the `access_token_signed_response_alg` option so [ownCloud] can verify them using the
  Authelia JSON Web Key Set.
- The access tokens for every client contain the `https://owncloud.{{< sitevar name="domain" nojs="example.com" >}}`
  audience via the `audience` and `requested_audience_mode` options. This allows [ownCloud] to accept the access tokens
  issued to the native apps which have a different client ID to the web application.

### Application

To configure [ownCloud] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
The [ownCloud OpenID Connect App] requires a distributed memory cache such as Redis to be configured. In addition, the
desktop and mobile clients require the `/.well-known/openid-configuration` path of [ownCloud] to be rewritten (not
redirected) to `/index.php/apps/openidconnect/config`. When using the official Docker image this can be done by setting
the `OWNCLOUD_ENABLE_OIDC_REWRITE_URL` environment variable to `true`.
{{< /callout >}}

To configure [ownCloud] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [ownCloud] as an administrator.
2. Navigate to the `Market` and install the `OpenID Connect` app, or alternatively run `occ app:enable openidconnect`.
3. Create the following configuration file in the same directory as the `config.php` file of [ownCloud]:

```php {title="config/oidc.config.php"}
<?php
$CONFIG = [
  'openid-connect' => [
    'provider-url' => 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}',
    'client-id' => 'owncloud',
    'client-secret' => 'insecure_secret',
    'audience' => 'https://owncloud.{{< sitevar name="domain" nojs="example.com" >}}',
    'scopes' => ['openid', 'offline_access', 'profile', 'email'],
    'loginButtonName' => 'Authelia',
    'mode' => 'userid',
    'search-attribute' => 'preferred_username',
    'auto-provision' => [
      'enabled' => true,
      'email-claim' => 'email',
      'display-name-claim' => 'name',
      'update' => [
        'enabled' => true,
      ],
    ],
  ],
];
```

## See Also

- [ownCloud Server OpenID Connect Documentation](https://doc.owncloud.com/server/latest/admin_manual/configuration/user/oidc/oidc.html)
- [ownCloud OpenID Connect App README](https://github.com/owncloud/openidconnect/blob/master/README.md)

[Authelia]: https://www.authelia.com
[ownCloud]: https://owncloud.com/
[ownCloud OpenID Connect App]: https://github.com/owncloud/openidconnect
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
