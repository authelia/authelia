---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "MISP"
description: "A guide on integrating MISP with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "MISP | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring MISP with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [MISP]
  - [v2.5.51](https://github.com/MISP/MISP/releases/tag/v2.5.51)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://misp.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://misp.{{< sitevar name="domain" nojs="example.com" >}}/users/login`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `misp`
- **Client Secret:** `insecure_secret`
- **Groups:**
  - `misp-admin` is mapped to the `Admin` role (role ID `1`).
  - `misp-user` is mapped to the `User` role (role ID `3`).
- **Default Organization:** `ORGNAME`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [MISP] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'misp'
        client_name: 'MISP'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://misp.{{< sitevar name="domain" nojs="example.com" >}}/users/login'
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
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [MISP] there are two methods, using the [Configuration File](#configuration-file), or using
[Environment Variables](#environment-variables) with the official Docker image.

#### Configuration File

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
Generally the configuration file is named `app/Config/config.php`. The `OidcAuth` plugin must also be loaded by adding
`CakePlugin::load('OidcAuth');` to the end of `app/Config/bootstrap.php`.
{{< /callout >}}

To configure [MISP] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```php {title="app/Config/config.php"}
$config = array(
    // ...
    'Security' => array(
        // ...
        'auth' => array('OidcAuth.Oidc'),
    ),
    // ...
    'OidcAuth' => [
        'provider_url' => 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}',
        'client_id' => 'misp',
        'client_secret' => 'insecure_secret',
        'authentication_method' => 'client_secret_basic',
        'code_challenge_method' => 'S256',
        'disable_request_object' => true,
        'scopes' => ['profile', 'email', 'groups'],
        'roles_property' => 'groups',
        'role_mapper' => [
            'misp-user' => 3,
            'misp-admin' => 1,
        ],
        'default_org' => 'ORGNAME',
        'mixedAuth' => true,
        'login_button_text' => 'Login with Authelia',
    ],
    // ...
);
```

#### Environment Variables

To configure [MISP] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
OIDC_ENABLE=true
OIDC_PROVIDER_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
OIDC_CLIENT_ID=misp
OIDC_CLIENT_SECRET=insecure_secret
OIDC_AUTH_METHOD=client_secret_basic
OIDC_CODE_CHALLENGE_METHOD=S256
OIDC_DISABLE_REQUEST_OBJECT=true
OIDC_SCOPES='["profile", "email", "groups"]'
OIDC_ROLES_PROPERTY=groups
OIDC_ROLES_MAPPING='{"misp-user": 3, "misp-admin": 1}'
OIDC_DEFAULT_ORG=ORGNAME
OIDC_MIXEDAUTH=true
OIDC_LOGIN_TEXT="Login with Authelia"
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  misp-core:
    environment:
      OIDC_ENABLE: 'true'
      OIDC_PROVIDER_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      OIDC_CLIENT_ID: 'misp'
      OIDC_CLIENT_SECRET: 'insecure_secret'
      OIDC_AUTH_METHOD: 'client_secret_basic'
      OIDC_CODE_CHALLENGE_METHOD: 'S256'
      OIDC_DISABLE_REQUEST_OBJECT: 'true'
      OIDC_SCOPES: '["profile", "email", "groups"]'
      OIDC_ROLES_PROPERTY: 'groups'
      OIDC_ROLES_MAPPING: '{"misp-user": 3, "misp-admin": 1}'
      OIDC_DEFAULT_ORG: 'ORGNAME'
      OIDC_MIXEDAUTH: 'true'
      OIDC_LOGIN_TEXT: 'Login with Authelia'
```

## See Also

- [MISP OpenID Connect Authentication Plugin Documentation](https://github.com/MISP/MISP/blob/2.5/app/Plugin/OidcAuth/README.md)
- [MISP Docker Environment Template](https://github.com/MISP/misp-docker/blob/master/template.env)

[Authelia]: https://www.authelia.com
[MISP]: https://www.misp-project.org
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
