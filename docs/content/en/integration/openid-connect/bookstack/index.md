---
title: "BookStack"
description: "Integrating BookStack with the Authelia OpenID Connect Provider."
lead: ""
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  integration:
    parent: "openid-connect"
weight: 620
toc: true
community: true
---

## Tested Versions

* [Authelia]
  * [v4.35.5](https://github.com/authelia/authelia/releases/tag/v4.35.5)
* [BookStack]
  * 20.10

## Before You Begin

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://bookstack.example.com`
* __Authelia Root URL:__ `https://auth.example.com`
* __Client ID:__ `bookstack`
* __Client Secret:__ `insecure_secret`

*__Important Note:__ [BookStack] does not properly URL encode the secret per [RFC6749 Appendix B] at the time this
article was last modified (noted at the bottom). This means you'll either have to use only alphanumeric characters for
the secret or URL encode the secret yourself.*

[RFC6749 Appendix B]: https://datatracker.ietf.org/doc/html/rfc6749#appendix-B

## Configuration

### Application

To configure [BookStack] to utilize Authelia as an [OpenID Connect 1.0] Provider:

1. Edit your .env file
2. Set the following values:
   1. AUTH_METHOD: `oidc`
   2. OIDC_NAME: `Authelia`
   3. OIDC_DISPLAY_NAME_CLAIMS: `name`
   4. OIDC_CLIENT_ID: `bookstack`
   5. OIDC_CLIENT_SECRET: `insecure_secret`
   6. OIDC_ISSUER: `https://auth.example.com`
   7. OIDC_ISSUER_DISCOVER: `true`

### Authelia

The following YAML configuration is an example __Authelia__
[client configuration](../../../configuration/identity-providers/openid-connect/clients.md) for use with [BookStack]
which will operate with the above example:

```yaml
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
    - id: 'bookstack'
      description: 'BookStack'
      secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
      public: false
      authorization_policy: 'two_factor'
      redirect_uris:
        - 'https://bookstack.example.com/oidc/callback'
      scopes:
        - 'openid'
        - 'profile'
        - 'email'
      userinfo_signing_alg: 'none'
```

## See Also

* [BookStack OpenID Connect Documentation](https://www.bookstackapp.com/docs/admin/oidc-auth/)

[Authelia]: https://www.authelia.com
[BookStack]: https://www.bookstackapp.com/
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
