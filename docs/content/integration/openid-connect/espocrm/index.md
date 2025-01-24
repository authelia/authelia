---
title: "EspoCRM"
description: "Integrating EspoCRM with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-01-21T22:32:51+11:00
draft: false
images: []
weight: 620
toc: true
support:
  level: community
  versions: true
  integration: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

* [Authelia]
  * [v4.38.8](https://github.com/authelia/authelia/releases/tag/v4.38.8)
* [EspoCRM]
  * [2.0.1 3b9ed2f](https://github.com/m4rc3l-h3/espocrm)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://espocrm.{{< sitevar name="domain" nojs="example.com" >}}/`
  * This option determines the redirect URI in the format of
        `https://espocrm.{{< sitevar name="domain" nojs="example.com" >}}/login`.
        This means if you change this value, you need to update the redirect URI.
* __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
* __Client ID:__ `espocrm`
* __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [EspoCRM] which will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'espocrm'
        client_name: 'EspoCRM'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://espocrm.{{< sitevar name="domain" nojs="example.com" >}}/oauth-callback.php'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [EspoCRM] to utilize Authelia as an [OpenID Connect 1.0] Provider you must use the GUI to configure it.

1. Visit [EspoCRM].
2. Login as an Administration user.
3. Visit Authentication.
4. Select OIDC as the method.
5. Configure the following options:
   - Client ID: `espocrm`.
   - Client Secret: `insecure_secret`
   - Authorization Redirect URI: `https://espocrm.{{< sitevar name="domain" nojs="example.com" >}}/oauth-callback.php`
   - Fallback Login: it's recommended this option is enabled to allow you to login with internal users.
   - Allow OIDC Login for admin users: it's recommended this option is enabled, it allows admins to login via
     [OpenID Connect 1.0].
   - Authorization Endpoint: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization`
   - Token Endpoint: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token`
   - JSON Web Key Set Endpoint: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/jwks.json`

## See Also

- [EspoCRM]
  - [OpenID Connect (OIDC) Authentication Documentation](https://docs.espocrm.com/administration/oidc/)

[Authelia]: https://www.authelia.com
[EspoCRM]: https://www.espocrm.com/
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
