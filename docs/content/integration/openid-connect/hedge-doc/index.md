---
title: "HedgeDoc"
description: "Integrating HedgeDoc with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-04-13T13:46:05+10:00
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
  * [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
* [HedgeDoc]
  * [v1.9.9](https://github.com/hedgedoc/hedgedoc/releases/tag/1.9.9)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://hedgedoc.example.com/`
* __Authelia Root URL:__ `https://auth.example.com/`
* __Client ID:__ `hedgedoc`
* __Client Secret:__ `insecure_secret`

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [HedgeDoc] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'hedgedoc'
        client_name: 'HedgeDoc'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://hedgedoc.example.com/auth/oauth2/callback'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
          - 'groups'
        userinfo_signed_response_alg: 'none'
```

### Application

_**Important Note:** This configuration assumes [HedgeDoc] users are part of the `hedgedoc-users` group. Depending on
your specific group configuration, you will have to adapt the `CMD_OAUTH2_ACCESS_ROLE` variable. Alternatively you may
elect to create a new authorization policy in [provider authorization policies] then utilize that policy as the
[client authorization policy]._

[client authorization policy]: ../../../configuration/identity-providers/openid-connect/clients.md#authorization_policy
[provider authorization policies]: ../../../configuration/identity-providers/openid-connect/provider.md#authorization_policies

To configure [HedgeDoc] to utilize Authelia as an [OpenID Connect 1.0] Provider:

1. Add the following configuration variables:

```env
CMD_OAUTH2_PROVIDERNAME=Authelia
CMD_OAUTH2_AUTHORIZATION_URL=https://auth.example.com/api/oidc/authorization
CMD_OAUTH2_TOKEN_URL=https://auth.example.com/api/oidc/token
CMD_OAUTH2_USER_PROFILE_URL=https://auth.example.com/api/oidc/userinfo
CMD_OAUTH2_CLIENT_ID=hedgedoc
CMD_OAUTH2_CLIENT_SECRET=insecure_secret
CMD_OAUTH2_SCOPE=openid email profile groups
CMD_OAUTH2_USER_PROFILE_USERNAME_ATTR=preferred_username
CMD_OAUTH2_USER_PROFILE_DISPLAY_NAME_ATTR=name
CMD_OAUTH2_USER_PROFILE_EMAIL_ATTR=email
CMD_OAUTH2_ROLES_CLAIM=groups
CMD_OAUTH2_ACCESS_ROLE=hedgedoc
```

## See Also

- [HedgeDoc OAuth2 Login Documentation](https://docs.hedgedoc.org/configuration/#oauth2-login)

[HedgeDoc]: https://hedgedoc.org/
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
