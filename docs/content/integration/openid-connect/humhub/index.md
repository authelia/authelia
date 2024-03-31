---
title: "Humhub"
description: "Integrating Humhub with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-03-29T11:23:00+01:00
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
  * [v4.38.6](https://github.com/authelia/authelia/releases/tag/v4.38.6)
* [Humhub]
  * [1.15.4](https://github.com/humhub/humhub/releases/tag/v1.15.4)

## Before You Begin

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://humhub.example.com/`
* __Authelia Root URL:__ `https://auth.example.com/`
* __Client ID:__ `client_id_humhub`

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__
[client configuration](../../../configuration/identity-providers/openid-connect/clients.md) for use with [Humhub]
which will operate with the above example:

```yaml
identity_providers:
  oidc:
    clients:
      - client_id: 'client_id_humhub'
        client_name: 'humhub.example.com'
        client_secret: '<yoursecret>'
        redirect_uris:
          - 'https://humhub.example.com/user/auth/external?authclient=oidc'
        authorization_policy: 'one_factor'
        token_endpoint_auth_method: 'client_secret_post'
        consent_mode: 'pre-configured'
```

### Application

To configure [Humhub] to utilize Authelia as an [OpenID Connect 1.0] Provider:

1. Build your own Humhub with the [OIDC Connector](https://github.com/Worteks/humhub-auth-oidc)
2. Configure a new oidc provider in config/common.php:
```php
return [
    'components' => [
        'urlManager' => [
            'enablePrettyUrl' => true,
        ],
        'authClientCollection' => [
          'clients' => [
            'oidc' => [
              'class' => 'worteks\humhub\authclient\OIDC',
              'domain' => 'https://auth.example.com',
              'clientId' => 'client_id_humhub',
              'clientSecret' => '<yoursecret>',
              'defaultTitle' => 'login with SSO',
              'cssIcon' => 'fa fa-sign-in',
              'authUrl' => '/api/oidc/authorization',
              'tokenUrl' => '/api/oidc/token',
              'apiBaseUrl' => '/api/oidc',
              'userInfoUrl' => 'userinfo',
              'scope' => 'openid profile email'
            ],
         ],
       ],
    ],
];

```

## See Also
 * [Odoo Authentication OpenID Connect]

[Authelia]: https://www.authelia.com
[Humhub]: https://www.humhub.com
[OpenID Connect 1.0]: ../../openid-connect/introduction.md

