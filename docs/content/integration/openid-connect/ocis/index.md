---
title: "ownCloud Infinite Scale"
description: "Integrating ownCloud Infinite Scale with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-03-05T21:58:32+11:00
draft: false
images: []
weight: 620
toc: true
community: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

* [Authelia]
  * [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
* [ownCloud Infinite Scale]
  * 4.0.5

## Before You Begin

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://owncloud.example.com`
* __Authelia Root URL:__ `https://auth.example.com`
* __Client ID:__
  * Web Application: `ownCloud`
  * Other Clients: the values
* __Client Secret:__ `insecure_secret`

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__
[client configuration](../../../configuration/identity-providers/openid-connect/clients.md) for use with [Nextcloud]
which will operate with the above example:

```yaml
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
      - client_id: 'ownCloud'
        client_name: 'ownCloud Infinite Scale'
        public: true
        redirect_uris:
          - 'https://owncloud.home.yourdomain.com/'
          - 'https://owncloud.home.yourdomain.com/oidc-callback.html'
          - 'https://owncloud.home.yourdomain.com/oidc-silent-redirect.html'
      - client_id: 'xdXOt13JKxym1B1QcEncf2XDkLAexMBFwiT9j6EfhhHFJhs2KM9jbjTmf8JBXE69'
        client_name: 'ownCloud desktop client'
        client_secret: 'UBntmLjC2yYCeHwsyj73Uwo9TAaecAetRwMw0xYcvNL9yRdLSUi0hUAHfvCHFeFh'
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

To configure [Nextcloud] to utilize Authelia as an [OpenID Connect 1.0] Provider:

1. Install the [Nextcloud OpenID Connect Login app]
2. Add the following to the [Nextcloud] `config.php` configuration:

```php
WEB_OIDC_CLIENT_ID=ownCloud

```

## See Also

* [Nextcloud OpenID Connect Login app]
* [Nextcloud OpenID Connect Login Documentation](https://github.com/pulsejet/nextcloud-oidc-login)

[Authelia]: https://www.authelia.com
[Nextcloud]: https://nextcloud.com/
[Nextcloud OpenID Connect Login app]: https://apps.nextcloud.com/apps/oidc_login
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
