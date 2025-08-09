---
title: "HumHub"
description: "Integrating HumHub with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-03-30T07:14:05+11:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/humhub/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "HumHub | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring HumHub with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.6](https://github.com/authelia/authelia/releases/tag/v4.38.6)
- [HumHub]
  - [v1.15.4](https://github.com/humhub/humhub/releases/tag/v1.15.4)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://humhub.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `humhub`
- __Client Secret:__ `insecure_secret`
- The [HumHub] instance you're using was built using the [OIDC Connector](https://github.com/Worteks/humhub-auth-oidc).

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [HumHub] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'humhub'
        client_name: 'HumHub'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://humhub.{{< sitevar name="domain" nojs="example.com" >}}/user/auth/external?authclient=oidc'
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

To configure [HumHub] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
Generally the configuration file is named `config/common.php`.
{{< /callout >}}

To configure [HumHub] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```php {title="config/common.php"}
return [
    'components' => [
        'urlManager' => [
            'enablePrettyUrl' => true,
        ],
        'authClientCollection' => [
          'clients' => [
            'oidc' => [
              'class' => 'worteks\humhub\authclient\OIDC',
              'domain' => 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}',
              'clientId' => 'humhub',
              'clientSecret' => 'insecure_secret',
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

 - [HumHub OpenID Connect Repository](https://github.com/Worteks/humhub-auth-oidc?tab=readme-ov-file)

[Authelia]: https://www.authelia.com
[HumHub]: https://www.humhub.com
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
