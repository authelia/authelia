---
title: "Leantime"
description: "Integrating Leantime with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-06-29T00:59:33+00:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/leantime/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Leantime | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Leantime with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.15](https://github.com/authelia/authelia/releases/tag/v4.39.15)
- [Leantime]
  - [v3.5.8](https://github.com/Leantime/leantime/releases/tag/v3.5.8)

{{% oidc-common bugs="claims-hydration" %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://leantime.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `leantime`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Leantime] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'leantime'
        client_name: 'Leantime'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://leantime.{{< sitevar name="domain" nojs="example.com" >}}/oidc/callback'
        scopes:
          - 'openid'
          - 'groups'
          - 'email'
          - 'profile'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

#### Configuration Escape Hatch

{{% oidc-escape-hatch-claims-hydration client_id="leantime" claims="email" %}}

### Application

To configure [leantime] there is one method, using the [Environment Variables](#environment-variables).

#### Environment Variables

To configure [leantime] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
LEAN_OIDC_ENABLE=true
LEAN_OIDC_CLIENT_ID=leantime
LEAN_OIDC_CLIENT_SECRET=insecure_secret
LEAN_OIDC_PROVIDER_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
LEAN_OIDC_CREATE_USER=true
LEAN_OIDC_DEFAULT_ROLE=20
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  leantime:
    environment:
      LEAN_OIDC_ENABLE: 'true'
      LEAN_OIDC_CLIENT_ID: 'leantime'
      LEAN_OIDC_CLIENT_SECRET: 'insecure_secret'
      LEAN_OIDC_PROVIDER_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      LEAN_OIDC_CREATE_USER: 'true'
      LEAN_OIDC_DEFAULT_ROLE: '20'
```

## See Also

- [Leantime OpenID Connect Documentation](https://docs.leantime.io/installation/configuration?id=openid-conenct-oidc-configuration)

[Authelia]: https://www.authelia.com
[Leantime]: https://leantime.io/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
