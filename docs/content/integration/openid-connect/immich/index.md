---
title: "immich"
description: "Integrating immich with the Authelia OpenID Connect 1.0 Provider."
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
* [immich]
  * [v1.101.0](https://github.com/immich-app/immich/releases/tag/v1.101.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://immich.example.com/`
* __Authelia Root URL:__ `https://auth.example.com/`
* __Client ID:__ `immich`
* __Client Secret:__ `insecure_secret`

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [immich] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'immich'
        client_name: 'immich'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://immich.example.com/auth/login'
          - 'https://immich.example.com/user-settings'
          - 'app.immich:/'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
        userinfo_signed_response_alg: 'none'
```

### Application

To configure [immich] to utilize Authelia as an [OpenID Connect 1.0] Provider:

1. Login to [immich] and visit the OAuth Settings.
2. On the screen that appears, enter the following information:
    - Issuer URL: `https://auth.example.com/.well-known/openid-configuration`.
    - Client ID: `immich`.
    - Client Secret: `insecure_secret`.
    - Scope: `openid profile email`.
    - Button Text: `Login with Authelia`.
    - Auto Register: Enable if desired.
3. Press `Save` at the bottom

## See Also

- [immich OAuth Authentication Guide](https://immich.app/docs/administration/oauth)

[immich]: https://immich.app/
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
