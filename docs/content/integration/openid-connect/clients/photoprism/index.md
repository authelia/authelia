---
title: "PhotoPrism"
description: "Integrating PhotoPrism with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-10-09T07:24:14+11:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/photoprism/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "PhotoPrism | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring PhotoPrism with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.16](https://github.com/authelia/authelia/releases/tag/v4.38.16)
- [PhotoPrism]
  - [v240915](https://github.com/photoprism/photoprism/releases/tag/240915-e1280b2fb)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://photoprism.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `photoprism`
- __Client Secret:__ `insecure_secret`

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
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://photoprism.{{< sitevar name="domain" nojs="example.com" >}}/api/v1/oidc/redirect'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
          - 'address'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [PhotoPrism] there is one method, using the [Environment Variables](#environment-variables).

#### Environment Variables

To configure [PhotoPrism] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment variables:

##### Standard

```shell {title=".env"}
PHOTOPRISM_OIDC_URI=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
PHOTOPRISM_OIDC_CLIENT=photoprism
PHOTOPRISM_OIDC_SECRET=insecure_secret
PHOTOPRISM_OIDC_PROVIDER=authelia
PHOTOPRISM_OIDC_REGISTER=true
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  photoprism:
    environment:
      PHOTOPRISM_OIDC_URI: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      PHOTOPRISM_OIDC_CLIENT: 'photoprism'
      PHOTOPRISM_OIDC_SECRET: 'insecure_secret'
      PHOTOPRISM_OIDC_PROVIDER: 'authelia'
      PHOTOPRISM_OIDC_REGISTER: 'true'
```

## See Also

- [PhotoPrism Single Sign-On via OpenID Connect](https://docs.photoprism.app/getting-started/advanced/openid-connect/)

[PhotoPrism]: https://photoprism.app/
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
