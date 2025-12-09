---
title: "Wapy.dev"
description: "Integrating Wapy.dev with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-12-09T19:19:00+01:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/wapy-dev/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Wapy.dev | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Wapy.Deb with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.0](https://github.com/authelia/authelia/releases/tag/v4.39.0)
- [Wapy.dev]
  - [v2.1.2](https://github.com/meceware/wapy.dev/releases/tag/v2.1.2)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://payments.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `{< sitevar name="client_id" nojs="wapydev" >}}`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Wapy.dev] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'wapydev'
        client_name: 'wapydev'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        sector_identifier_uri: ''
        public: false
        redirect_uris:
          - 'https://payments.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/oauth2/callback/{{< sitevar name="client_id" nojs="wapydev" >}}'
        audience: []
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
        grant_types:
          - 'authorization_code'
        response_types:
          - 'code'
        response_modes:
          - 'form_post'
          - 'query'
          - 'fragment'
        authorization_policy: 'two_factor'
        token_endpoint_auth_method: 'client_secret_post'

```

### Application

To configure [Wapy.dev] there is one method, using the [Environment options](#environment-options).

#### Environment options

To configure [Wapy.dev] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Edit your environment variables for Wapy.dev.
2. Configure the following options:
    - GENERIC_AUTH_PROVIDER=`authelia`
    - GENERIC_AUTH_CLIENT_ID=`{{< sitevar name="client_id" nojs="wapydev" >}}`
    - GENERIC_AUTH_CLIENT_SECRET=`insecure_secret`
    - GENERIC_AUTH_ISSUER=`https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration`
    - GENERIC_AUTH_SCOPE=`openid email profile`
4. Save the variables and restart container if needed.

## See Also

- [Single Sign-On Alternatives Wapy.dev](https://github.com/meceware/wapy.dev/wiki/Single-Sign%E2%80%90On-(SSO)-Alternatives)

[Wapy.dev]: https://www.wapy.dev
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
