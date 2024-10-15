---
title: "PhotoPrism"
description: "Integrating PhotoPrism with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2022-06-15T17:51:47+10:00
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
  * [v4.38.16](https://github.com/authelia/authelia/releases/tag/v4.38.16)
* [PhotoPrism]
  * [v240915](https://github.com/photoprism/photoprism/releases/tag/240915-e1280b2fb)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://photoprism.{{< sitevar name="domain" nojs="example.com" >}}/`
* __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
* __Client ID:__ `photoprism`
* __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [PhotoPrism] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'photoprism'
        client_name: 'photoprism'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://photoprism.{{< sitevar name="domain" nojs="example.com" >}}/api/v1/oidc/redirect
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
          - 'address'
        userinfo_signed_response_alg: 'none'
```

### Application

To configure [PhotoPrism] to utilize Authelia as an [OpenID Connect 1.0] Provider, specify the below environment variables:

```yaml
environment:
  PHOTOPRISM_OIDC_URI: https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
  PHOTOPRISM_OIDC_CLIENT: photoprism
  PHOTOPRISM_OIDC_SECRET: insecure_secret
  PHOTOPRISM_OIDC_PROVIDER: authelia
  PHOTOPRISM_OIDC_REGISTER: true
```

## See Also

- [PhotoPrism Single Sign-On via OpenID Connect](https://docs.photoprism.app/getting-started/advanced/openid-connect/)

[PhotoPrism]: https://photoprism.app/
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
