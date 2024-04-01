---
title: "Warpgate"
description: "Integrating Warpgate with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2023-12-10T10:52:22+11:00
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

- [Authelia]
  - [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
- [Warpgate]
  - [0.9.1](https://github.com/warp-tech/warpgate/releases/tag/v0.9.1)

## Before You Begin

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://warpgate.example.com/`
* __Authelia Root URL:__ `https://auth.example.com/`
* __Client ID:__ `warpgate`
* __Client Secret:__ `insecure_secret`

### Authelia

Authelia configuration.yml

```yaml
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'warpgate'
        client_name: 'Warpgate'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://warpgate.example.com/@warpgate/api/sso/return'
        scopes:
          - 'openid'
          - 'email'
        userinfo_signed_response_alg: 'none'
```

## Application

```toml
external_host: warpgate.example.com
sso_providers:
- name: authelia
  label: Authelia
  provider:
    type: custom
    client_id: warpgate
    client_secret: insecure_secret
    issuer_url: https://auth.example.com
    scopes: ["openid", "email"]
```

## See Also

- [Warpgate OpenID Connect Documentation](https://github.com/warp-tech/warpgate/wiki/SSO-Authentication)

[Authelia]: https://www.authelia.com
[Warpgate]: https://github.com/warp-tech/warpgate
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
