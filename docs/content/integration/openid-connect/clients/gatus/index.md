---
title: "Gatus"
description: "Integrating Gatus with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-04-26T11:03:16+00:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/gatus/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Gatus | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Gatus with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.8](https://github.com/authelia/authelia/releases/tag/v4.39.8)
- [Gatus]
  - [v5.17.0](https://github.com/TwiN/gatus/releases/tag/v5.17.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://gatus.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `gatus`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Gatus] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'gatus'
        client_name: 'Gatus'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://gatus.{{< sitevar name="domain" nojs="example.com" >}}/authorization-code/callback'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [Gatus] there are three methods, using the [Configuration File](#configuration-file), using
[Environment Variables](#environment-variables), or using the [Web GUI](#web-gui).

#### Configuration File

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
Generally the configuration file is named `config.yaml`.
{{< /callout >}}

To configure [Gatus] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```yaml {title="config.yaml"}
security:
  oidc:
    issuer-url: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
    client-id: 'gatus'
    client-secret: 'insecure_secret'
    redirect-url: 'https://gatus.{{< sitevar name="domain" nojs="example.com" >}}/authorization-code/callback'
    scopes: ['openid', 'profile', 'email']
```

## See Also

- [Gatus Custom Single Sign-On (SSO) Documentation](https://gatus.io/docs/private-status-page)

[Authelia]: https://www.authelia.com
[Gatus]: https://gatus.io/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
