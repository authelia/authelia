---
title: "Nextcloud"
description: "Integrating Nextcloud with the Authelia OpenID Connect Provider."
lead: ""
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  integration:
    parent: "openid-connect"
weight: 620
toc: true
community: true
---

## Tested Versions

* [Authelia]
  * [v4.35.5](https://github.com/authelia/authelia/releases/tag/v4.35.5)
* [Nextcloud]
  * 22.1.0

## Before You Begin

### Common Notes

1. You are *__required__* to utilize a unique client id for every client.
2. The client id on this page is merely an example and you can theoretically use any alphanumeric string.
3. You *__should not__* use the client secret in this example, We *__strongly recommend__* reading the
   [Generating Client Secrets] guide instead.

[Generating Client Secrets]: ../specific-information.md#generating-client-secrets

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://nextcloud.example.com`
* __Authelia Root URL:__ `https://auth.example.com`
* __Client ID:__ `nextcloud`
* __Client Secret:__ `nextcloud_client_secret`

## Configuration

### Application

To configure [Nextcloud] to utilize Authelia as an [OpenID Connect] Provider:

1. Install the [Nextcloud OpenID Connect Login app]
2. Add the following to the [Nextcloud] `config.php` configuration:

```php
$CONFIG = array (
    'allow_user_to_change_display_name' => false,
    'lost_password_link' => 'disabled',
    'oidc_login_provider_url' => 'https://auth.example.com',
    'oidc_login_client_id' => 'nextcloud',
    'oidc_login_client_secret' => 'nextcloud_client_secret',
    'oidc_login_auto_redirect' => false,
    'oidc_login_end_session_redirect' => false,
    'oidc_login_button_text' => 'Log in with Authelia',
    'oidc_login_hide_password_form' => false,
    'oidc_login_use_id_token' => true,
    'oidc_login_attributes' => array (
        'id' => 'preferred_username',
        'name' => 'name',
        'mail' => 'email',
        'groups' => 'groups',
    ),
    'oidc_login_default_group' => 'oidc',
    'oidc_login_use_external_storage' => false,
    'oidc_login_scope' => 'openid profile email groups',
    'oidc_login_proxy_ldap' => false,
    'oidc_login_disable_registration' => true,
    'oidc_login_redir_fallback' => false,
    'oidc_login_alt_login_page' => 'assets/login.php',
    'oidc_login_tls_verify' => true,
    'oidc_create_groups' => false,
    'oidc_login_webdav_enabled' => false,
    'oidc_login_password_authentication' => false,
    'oidc_login_public_key_caching_time' => 86400,
    'oidc_login_min_time_between_jwks_requests' => 10,
    'oidc_login_well_known_caching_time' => 86400,
    'oidc_login_update_avatar' => false,
);
```

### Authelia

The following YAML configuration is an example __Authelia__
[client configuration](../../../configuration/identity-providers/open-id-connect.md#clients) for use with [Nextcloud]
which will operate with the above example:

```yaml
- id: nextcloud
  description: NextCloud
  secret: '$plaintext$nextcloud_client_secret'
  public: false
  authorization_policy: two_factor
  redirect_uris:
    - https://nextcloud.example.com/apps/oidc_login/oidc
  scopes:
    - openid
    - profile
    - email
    - groups
  userinfo_signing_algorithm: none
```

## See Also

* [Nextcloud OpenID Connect Login app]
* [Nextcloud OpenID Connect Login Documentation](https://github.com/pulsejet/nextcloud-oidc-login)

[Authelia]: https://www.authelia.com
[Nextcloud]: https://nextcloud.com/
[Nextcloud OpenID Connect Login app]: https://apps.nextcloud.com/apps/oidc_login
[OpenID Connect]: ../../openid-connect/introduction.md
