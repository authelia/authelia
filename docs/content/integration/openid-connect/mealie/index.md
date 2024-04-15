---
title: "Mealie"
description: "Integrating Mealie with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-04-13T21:01:17+10:00
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
* [Mealie]
  * [v1.4.0](https://github.com/mealie-recipes/mealie/releases/tag/v1.4.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://mealie.example.com/`
* __Authelia Root URL:__ `https://auth.example.com/`
* __Client ID:__ `mealie`

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Mealie] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'mealie'
        client_name: 'Mealie'
        public: true
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://mealie.example.com/login'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
          - 'groups'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'none'
```

### Application

_**Important Note:** This configuration assumes [Mealie] administrators are part of the `mealie-admins` group, and
[Mealie] users are part of the `mealie-users` group. Depending on your specific group configuration, you will have to
adapt the `OIDC_ADMIN_GROUP` and `OIDC_USER_GROUP` nodes respectively. Alternatively you may elect to create a new
authorization policy in [provider authorization policies] then utilize that policy as the
[client authorization policy]._

To configure [Mealie] to utilize Authelia as an [OpenID Connect 1.0] Provider use the following environment variables:

```env
OIDC_AUTH_ENABLED=true
OIDC_SIGNUP_ENABLED=true
OIDC_CONFIGURATION_URL=https://auth.example.com/.well-known/openid-configuration
OIDC_CLIENT_ID=mealie
OIDC_AUTO_REDIRECT=false
OIDC_ADMIN_GROUP=mealie-admins
OIDC_USER_GROUP=mealie-users
```

## See Also

- [Mealie OpenID Connect Documentation](https://docs.mealie.io/documentation/getting-started/authentication/oidc/)

[Mealie]: https://mealie.io/
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
