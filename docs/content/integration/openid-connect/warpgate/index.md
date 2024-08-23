---
title: "Warpgate"
description: "Integrating Warpgate with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2023-12-10T10:52:22+11:00
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

- [Authelia]
  - [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
- [Warpgate]
  - [0.9.1](https://github.com/warp-tech/warpgate/releases/tag/v0.9.1)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://warpgate.{{< sitevar name="domain" nojs="example.com" >}}/`
* __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
* __Client ID:__ `warpgate`
* __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Warpgate]
which will operate with the application example:

```yaml {title="configuration.yml"}
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
          - 'https://warpgate.{{< sitevar name="domain" nojs="example.com" >}}/@warpgate/api/sso/return'
        scopes:
          - 'openid'
          - 'email'
        userinfo_signed_response_alg: 'none'
```

## Application

```toml
external_host: warpgate.{{< sitevar name="domain" nojs="example.com" >}}
sso_providers:
- name: authelia
  label: Authelia
  provider:
    type: custom
    client_id: warpgate
    client_secret: insecure_secret
    issuer_url: https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
    scopes: ["openid", "email"]
```

## See Also

- [Warpgate OpenID Connect Documentation](https://github.com/warp-tech/warpgate/wiki/SSO-Authentication)

[Authelia]: https://www.authelia.com
[Warpgate]: https://github.com/warp-tech/warpgate
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
