---
title: "Nextcloud"
description: "Integrating Nextcloud with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2022-06-15T17:51:47+10:00
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
* [Nextcloud]
  * 22.1.0 with the application oidc_login
  * 28.0.4 with the application user_oidc

{{% oidc-common %}}

## Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://nextcloud.example.com/`
* __Authelia Root URL:__ `https://auth.example.com/`
* __Client ID:__ `nextcloud`
* __Client Secret:__ `insecure_secret`

*__Important Note:__ it has been reported that some of the [Nextcloud] plugins do not properly encode the client secret.
as such it's important to only use alphanumeric characters as well as the other
[RFC3986 Unreserved Characters](https://datatracker.ietf.org/doc/html/rfc3986#section-2.3). We recommend using the
generating client secrets guidance above.*

## Available Options

The following two tested options exist for Nextcloud:

1. [OpenID Connect Login App](#openid-connect-login-app)
2. [OpenID Connect user backend App](#openid-connect-user-backend-app)

## OpenID Connect Login App

The following example uses the [OpenID Connect Login App](https://apps.nextcloud.com/apps/oidc_login) app.

### Configuration

#### Authelia

The following YAML configuration is an example __Authelia__
[client configuration] for use with [Nextcloud]
which will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'nextcloud'
        client_name: 'NextCloud'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://nextcloud.example.com/apps/oidc_login/oidc'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
          - 'groups'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

#### Application

To configure [Nextcloud] to utilize Authelia as an [OpenID Connect 1.0] Provider:

1. Install the [Nextcloud OpenID Connect Login app]
2. Add the following to the [Nextcloud] `config.php` configuration:

```php
$CONFIG = array (
    'allow_user_to_change_display_name' => false,
    'lost_password_link' => 'disabled',
    'oidc_login_provider_url' => 'https://auth.example.com',
    'oidc_login_client_id' => 'nextcloud',
    'oidc_login_client_secret' => 'insecure_secret',
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
    'oidc_login_tls_verify' => true,
    'oidc_create_groups' => false,
    'oidc_login_webdav_enabled' => false,
    'oidc_login_password_authentication' => false,
    'oidc_login_public_key_caching_time' => 86400,
    'oidc_login_min_time_between_jwks_requests' => 10,
    'oidc_login_well_known_caching_time' => 86400,
    'oidc_login_update_avatar' => false,
    'oidc_login_code_challenge_method' => 'S256'
);
```

## OpenID Connect user backend App

The following example uses the [OpenID Connect user backend](https://apps.nextcloud.com/apps/user_oidc) app.

### Configuration

#### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Nextcloud] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'nextcloud'
        client_name: 'NextCloud'
        client_secret: 'insecure_secret'
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://nextcloud.example.com/apps/user_oidc/code'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
          - 'groups'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

#### Application

To configure [Nextcloud] to utilize Authelia as an [OpenID Connect 1.0] Provider:

1. Install the [Nextcloud OpenID Connect user backend app]
2. Edit the 'OpenID Connect' configuration:

* Identifier : Authelia
* Client ID : nextcloud
* Client secret : insecure_secret
* Discovery endpoint : https://auth.example.com/.well-known/openid-configuration
* Scope : openid email profile

3. Add the following to the [Nextcloud] `config.php` configuration:
``` php
'user_oidc' => [
    'use_pkce' => true,
],
```

## See Also

* [Nextcloud OpenID Connect user backend app]
* [Nextcloud OpenID Connect Login app]
* [Nextcloud OpenID Connect Login Documentation](https://github.com/pulsejet/nextcloud-oidc-login)

[Authelia]: https://www.authelia.com
[Nextcloud]: https://nextcloud.com/
[Nextcloud OpenID Connect Login app]: https://apps.nextcloud.com/apps/oidc_login
[Nextcloud OpenID Connect user backend app]: https://apps.nextcloud.com/apps/user_oidc
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
