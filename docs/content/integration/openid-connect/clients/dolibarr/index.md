---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Dolibarr"
description: "A guide on integrating Dolibarr with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Dolibarr | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Dolibarr with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Dolibarr]
  - [v24.0.3](https://github.com/Dolibarr/dolibarr/releases/tag/24.0.3)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://dolibarr.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://dolibarr.{{< sitevar name="domain" nojs="example.com" >}}/core/modules/openid_connect/callback.php`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `dolibarr`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

### Special Notes

1. By default [Dolibarr] matches the `email` claim against the login or email of an existing [Dolibarr] user, so the
   email address of the Authelia user must match the email address of the [Dolibarr] user unless automatic user
   creation is enabled.

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Dolibarr] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'dolibarr'
        client_name: 'Dolibarr'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://dolibarr.{{< sitevar name="domain" nojs="example.com" >}}/core/modules/openid_connect/callback.php'
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

To configure [Dolibarr] both the [Web GUI](#web-gui) and the [Configuration File](#configuration-file) steps are required.

#### Web GUI

To configure [Dolibarr] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [Dolibarr] as an administrator.
2. Navigate to `Home` > `Setup` > `Security`.
3. Select `OpenID authentication parameters`.
4. Toggle `Enable the OpenID Connect authentication mode` to the on position.
5. Configure the following options:
   - Client ID: `dolibarr`
   - Client secret: `insecure_secret`
   - Scopes: `openid profile email`
   - Authorize URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization`
   - Token URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token`
   - User info URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo`
   - Login claim: `email`
6. Click `Save`.

#### Configuration File

To enable the [OpenID Connect 1.0] authentication mode in [Dolibarr], replace the value of
`$dolibarr_main_authentication` in the [Dolibarr] configuration file with the following value:

```php {title="conf/conf.php"}
$dolibarr_main_authentication='openid_connect,dolibarr';
```

## See Also

- [Dolibarr Authentication, SSO and SSL Documentation](https://wiki.dolibarr.org/index.php?title=Authentication,_SSO_and_SSL#Mode_openid_connect)

[Authelia]: https://www.authelia.com
[Dolibarr]: https://www.dolibarr.org/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
