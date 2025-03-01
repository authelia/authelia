---
title: "Linkwarden"
description: "Integrating Linkwarden with the Authelia OpenID Connect 1.0 Provider."
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
  * [v4.38.18](https://github.com/authelia/authelia/releases/tag/v4.38.18)
* [Linkwarden]
  * [2.9.2](https://github.com/linkwarden/linkwarden/releases/tag/v2.9.2)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://linkwarden.{{< sitevar name="domain" nojs="example.com" >}}/`
* __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
* __Client ID:__ `linkwarden`
* __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Linkwarden] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'linkwarden'
        client_name: 'Linkwarden'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://linkwarden.{{< sitevar name="domain" nojs="example.com" >}}/api/v1/auth/callback/authelia'
        scopes:
          - 'openid'
          - 'groups'
          - 'email'
          - 'profile'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

1. To configure [Linkwarden] to utilize Authelia as an [OpenID Connect 1.0](https://www.authelia.com/integration/openid-connect/introduction/) Provider, specify the below environment
   variables.

```yaml
services:
  linkwarden:
    image: 'ghcr.io/linkwarden/linkwarden:latest'
    environment:
      NEXT_PUBLIC_AUTHELIA_ENABLED: 'true'
      AUTHELIA_WELLKNOWN_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration'
      AUTHELIA_CLIENT_ID: 'linkwarden'
      AUTHELIA_CLIENT_SECRET: 'insecure_secret'
```

## See Also

- [Linkwarden OIDC documentation](https://docs.linkwarden.app/self-hosting/sso-oauth#authelia)

[Authelia]: https://www.authelia.com
[Linkwarden]: https://linkwarden.app/
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
