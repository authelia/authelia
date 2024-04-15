---
title: "Homarr"
description: "Integrating Homarr with the Authelia OpenID Connect 1.0 Provider."
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
* [Homarr]
  * 0.15.2

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://homarr.example.com/`
* __Authelia Root URL:__ `https://auth.example.com/`
* __Client ID:__ `homarr`
* __Client Secret:__ `insecure_secret`

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Homarr] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'homarr'
        client_name: 'Homarr'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://homarr.example.com/api/auth/callback/oidc'
        scopes:
          - 'openid'
          - 'profile'
          - 'groups'
          - 'email'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

_**Important Note:** The following example assumes you want users with the `homarr-admins` group to be administrators in
[Homarr], and users with the `homarr-owners` group to be owners in [Homarr]. You may be required to adjust this._

To configure [Homarr] to utilize Authelia as an [OpenID Connect 1.0] Provider:

1. Include the [Homarr] environment variables for [OpenID Connect 1.0] configuration:

```env
AUTH_PROVIDER=oidc
AUTH_OIDC_URI=https://auth.example.com
AUTH_OIDC_CLIENT_SECRET=insecure_secret
AUTH_OIDC_CLIENT_ID=homarr
AUTH_OIDC_CLIENT_NAME=Authelia
AUTH_OIDC_ADMIN_GROUP=homarr-admins
AUTH_OIDC_OWNER_GROUP=homarr-owners
```

## See Also

* [Homarr SSO Documentation](https://homarr.dev/docs/advanced/sso)

[Authelia]: https://www.authelia.com
[Homarr]: https://homarr.dev
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
