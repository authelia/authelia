---
title: "engomo"
description: "Integrating engomo with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-04-26T18:35:57+10:00
draft: false
images: []
weight: 620
toc: true
aliases: []
support:
  level: community
  versions: true
  integration: true
seo:
  title: "engomo | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring engomo with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.13](https://github.com/authelia/authelia/releases/tag/v4.39.13)
- [engomo]

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://engomo.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `engomo`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [engomo] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'engomo'
        client_name: 'engomo'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://engomo.{{< sitevar name="domain" nojs="example.com" >}}/auth'
          - 'com.engomo.engomo://callback/'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
        response_modes:
          - 'form_post'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [engomo] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [engomo] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to your [engomo] composer as an administrator.
2. Select `Server`.
3. Select `Authentication`.
4. Select `+` to add a new method.
5. Set the `Name` to `Authelia`.
6. Select the `OpenID Connect` value for `Type`
7. Click `Create`.
8. Set the following values:
  - Issuer: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
  - Client ID: `engomo`
  - Client Secret: `insecure_secret`
9. Click Save.

[Authelia]: https://www.authelia.com
[engomo]: https://engomo.com/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
