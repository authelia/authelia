---
title: "HashiCorp Vault"
description: "Integrating HashiCorp Vault with the Authelia OpenID Connect Provider."
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
* [HashiCorp Vault]
  * 1.8.1

## Before You Begin

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://vault.example.com`
* __Authelia Root URL:__ `https://auth.example.com`
* __Client ID:__ `vault`
* __Client Secret:__ `insecure_secret`

## Configuration

### Application

To configure [HashiCorp Vault] to utilize Authelia as an [OpenID Connect 1.0] Provider please see the links in the
[see also](#see-also) section.

### Authelia

The following YAML configuration is an example __Authelia__
[client configuration](../../../configuration/identity-providers/open-id-connect.md#clients) for use with [HashiCorp Vault]
which will operate with the above example:

```yaml
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
    - id: vault
      description: HashiCorp Vault
      secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
      public: false
      authorization_policy: two_factor
      redirect_uris:
        - https://vault.example.com/oidc/callback
        - https://vault.example.com/ui/vault/auth/oidc/oidc/callback
      scopes:
        - openid
        - profile
        - groups
        - email
      userinfo_signing_algorithm: none
```

## See Also

* [HashiCorp Vault JWT/OIDC Auth Documentation](https://www.vaultproject.io/docs/auth/jwt)
* [HashiCorp Vault OpenID Connect Providers Documentation](https://www.vaultproject.io/docs/auth/jwt/oidc_providers)

[Authelia]: https://www.authelia.com
[HashiCorp Vault]: https://www.vaultproject.io/
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
