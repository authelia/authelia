---
title: "Nextcloud"
description: "Integrating Nextcloud with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/nextcloud/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Nextcloud | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Nextcloud with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.17](https://github.com/authelia/authelia/releases/tag/v4.38.17)
- Application:
  - [Nextcloud OpenID Connect Login app]:
    - Nextcloud [v31.0.5](https://github.com/nextcloud/server/releases/tag/v31.0.5)
    - App [v3.2.2](https://github.com/pulsejet/nextcloud-oidc-login/releases/tag/v3.2.2) ([see also](https://apps.nextcloud.com/apps/oidc_login/releases?platform=31#31))
  - [Nextcloud OpenID Connect user backend app]:
    - Nextcloud [v31.0.4](https://github.com/nextcloud/server/releases/tag/v31.0.4)
    - App [v7.2.0](https://apps.nextcloud.com/apps/user_oidc/releases?platform=31#31)

{{% oidc-common %}}

## Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://nextcloud.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `nextcloud`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
It has been reported that some of the [Nextcloud](https://nextcloud.com/) plugins do not properly encode the client secret.
as such it's important to only use alphanumeric characters as well as the other
[RFC3986 Unreserved Characters](https://datatracker.ietf.org/doc/html/rfc3986#section-2.3). We recommend using the
generating client secrets guidance above.
{{< /callout >}}

## Available Options

The following two tested options exist for Nextcloud:

1. [OpenID Connect Login App](#openid-connect-login-app)
2. [OpenID Connect user backend App](#openid-connect-user-backend-app)

## OpenID Connect Login App

The following example uses the [Nextcloud OpenID Connect Login app] which is assumed to be installed, as well as have [pretty urls](https://docs.nextcloud.com/server/latest/admin_manual/installation/source_installation.html#pretty-urls) enabled when following this section of the guide.

### Configuration

#### Authelia

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
The `is_nextcloud_admin` user attribute renders the value `true` if the user is in the `nextcloud-admins` group within
Authelia, otherwise it renders `false`. You can adjust this to your preference to assign the admin role to the
appropriate user groups.
{{< /callout >}}

The following YAML configuration is an example __Authelia__
[client configuration] for use with [Nextcloud]
which will operate with the application example:

```yaml {title="configuration.yml"}
definitions:
  user_attributes:
    is_nextcloud_admin:
      ## Expression to evaluate admin privilege for Nextcloud.
      expression: '"nextcloud-admins" in groups'

identity_providers:
  oidc:
    claims_policies:
      nextcloud_userinfo:
        custom_claims:
          is_nextcloud_admin: {}

    scopes:
      nextcloud_userinfo:
        claims:
          - 'is_nextcloud_admin'

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
        claims_policy: 'nextcloud_userinfo'
        redirect_uris:
          - 'https://nextcloud.{{< sitevar name="domain" nojs="example.com" >}}/apps/oidc_login/oidc'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
          - 'groups'
          - 'nextcloud_userinfo'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

#### Application

To configure [Nextcloud] and the [Nextcloud OpenID Connect Login app] there is one method, using the
[Configuration File](#configuration-file).

#### Configuration File

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
Generally the configuration file is named `config.php`.
{{< /callout >}}

To configure [Nextcloud] and the [Nextcloud OpenID Connect Login app] to utilize Authelia as an [OpenID Connect 1.0]
Provider use the following configuration:

```php {title="config.php"}
$CONFIG = array (
    'allow_user_to_change_display_name' => false,
    'lost_password_link' => 'disabled',
    'oidc_login_provider_url' => 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}',
    'oidc_login_client_id' => 'nextcloud',
    'oidc_login_client_secret' => 'insecure_secret',
    'oidc_login_auto_redirect' => false,
    'oidc_login_end_session_redirect' => false,
    'oidc_login_button_text' => 'Log in with Authelia',
    'oidc_login_hide_password_form' => false,
    'oidc_login_use_id_token' => false,
    'oidc_login_attributes' => array (
        'id' => 'preferred_username',
        'name' => 'name',
        'mail' => 'email',
        'groups' => 'groups',
        'is_admin' => 'is_nextcloud_admin',
    ),
    'oidc_login_default_group' => 'oidc',
    'oidc_login_use_external_storage' => false,
    'oidc_login_scope' => 'openid profile email groups nextcloud_userinfo',
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

The following example uses the [Nextcloud OpenID Connect user backend app] which is assumed to be installed, as well as have [pretty urls](https://docs.nextcloud.com/server/latest/admin_manual/installation/source_installation.html#pretty-urls) enabled when following this section of the guide.

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
          - 'https://nextcloud.{{< sitevar name="domain" nojs="example.com" >}}/apps/user_oidc/code'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
          - 'groups'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

#### Application

To configure [Nextcloud] and the [Nextcloud OpenID Connect Login app] there is one method, using the
[Configuration File](#configuration-file-1).

##### Configuration File

To configure [Nextcloud] and the [Nextcloud OpenID Connect user backend app] to utilize Authelia as an
[OpenID Connect 1.0] Provider, use the following instructions:

1. Edit the `OpenID Connect` configuration:
   - Identifier: `Authelia`
   - Client ID: `nextcloud`
   - Client secret: `insecure_secret`
   - Discovery endpoint: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration`
   - Scope: openid email profile

## See Also

- [Nextcloud OpenID Connect Login app]
  - [Documentation](https://github.com/pulsejet/nextcloud-oidc-login)
- [Nextcloud OpenID Connect user backend app]
  - [Documentation](https://github.com/nextcloud/user_oidc)

[Authelia]: https://www.authelia.com
[Nextcloud]: https://nextcloud.com/
[Nextcloud OpenID Connect Login app]: https://apps.nextcloud.com/apps/oidc_login
[Nextcloud OpenID Connect user backend app]: https://apps.nextcloud.com/apps/user_oidc
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
