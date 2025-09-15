---
title: "Dashy"
description: "Integrating Dashy with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-06-13T14:12:09+00:00
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
  title: "Dashy | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Dashy with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.9](https://github.com/authelia/authelia/releases/tag/v4.39.9)
- [Dashy]
  - [v3.1.1](https://github.com/Lissy93/dashy/releases/tag/3.1.1)

{{% oidc-common bugs="claims-hydration" %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://dashy.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `dashy`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Dashy] which will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'dashy'
        client_name: 'Dashy'
        public: true
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://dashy.{{< sitevar name="domain" nojs="example.com" >}}'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
          - 'groups'
          - 'roles'
        grant_types:
          - 'authorization_code'
        response_types:
          - 'code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'none'
```

#### Configuration Escape Hatch

{{% oidc-escape-hatch-claims-hydration client_id="dashy" claims="email,email_verified,alt_emails,preferred_username,name" %}}

### Application

To configure [Dashy] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

To configure [Dashy] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```yaml
appConfig:
  auth:
    enableOidc: true
    oidc:
      clientId: 'dashy'
      endpoint: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
```

## See Also

- [Dashy OIDC Authentication Documentation](https://dashy.to/docs/authentication#oidc)

[Authelia]: https://www.authelia.com
[Dashy]: https://dashy.to/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
