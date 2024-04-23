---
title: "WeKan"
description: "Integrating WeKan with the Authelia OpenID Connect 1.0 Provider."
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
* [WeKan]
  * [v7.42](https://github.com/wekan/wekan/releases/tag/v7.42)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://wekan.example.com/`
* __Authelia Root URL:__ `https://auth.example.com/`
* __Client ID:__ `wekan`
* __Client Secret:__ `insecure_secret`

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [WeKan] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'wekan'
        client_name: 'WeKan'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://wekan.example.com/_oauth/oidc'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [WeKan] to utilize Authelia as an [OpenID Connect 1.0] Provider:

1. Add the following YAML to your configuration:

```env
OAUTH2_ENABLED=true
OAUTH2_LOGIN_STYLE=redirect
OAUTH2_CLIENT_ID=wekan
OAUTH2_SECRET=insecure_secret
OAUTH2_SERVER_URL=https://auth.example.com
OAUTH2_AUTH_ENDPOINT=/api/oidc/authorization
OAUTH2_TOKEN_ENDPOINT=/api/oidc/token
OAUTH2_USERINFO_ENDPOINT=/api/oidc/userinfo
OAUTH2_ID_MAP=sub
OAUTH2_USERNAME_MAP=email
OAUTH2_FULLNAME_MAP=name
OAUTH2_EMAIL_MAP=email
```

## See Also

- [WeKan OAuth2 Documentation](https://github.com/wekan/wekan/wiki/OAuth2)

[WeKan]: https://github.com/wekan/wekan
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
