---
title: "LinkAce"
description: "Integrate LinkAce with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2026-01-21T20:56:59+02:00
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
  title: "LinkAce | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configure LinkAce with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.18](https://github.com/authelia/authelia/releases/tag/v4.38.18)
- [LinkAce]
  - [v2.4.2](https://github.com/Kovah/LinkAce/releases/tag/v2.4.2)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://linkace.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `linkace`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [LinkAce] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'linkace'
        client_name: 'linkace'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        redirect_uris:
          - 'https://linkace.{{< sitevar name="domain" nojs="example.com" >}}/auth/sso/oidc/callback'
        scopes:
          - 'openid'
          - 'groups'
          - 'email'
          - 'profile'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
          - 'refresh_token'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [LinkAce] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [LinkAce] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment variables in the '.env' file:

##### Standard

```shell {title=".env"}
SSO_ENABLED=true
SSO_OIDC_ENABLED=true
SSO_OIDC_BASE_URL=https://auth.{{< sitevar name="domain" nojs="example.com" >}}
SSO_OIDC_CLIENT_ID=linkace
SSO_OIDC_CLIENT_SECRET=insecure_secret
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  linkace:
    environment:
      SSO_ENABLED: 'true'
      SSO_OIDC_ENABLED: 'true'
      SSO_OIDC_BASE_URL: 'https://auth.{{< sitevar name="domain" nojs="example.com" >}}'
      SSO_OIDC_CLIENT_ID: 'linkace'
      SSO_OIDC_CLIENT_SECRET: 'insecure_secret'

      
```

## See Also

- [LinkAce OIDC documentation](https://www.linkace.org/docs/v2/configuration/sso-oauth-oidc/)

[Authelia]: https://www.authelia.com
[LinkAce]: https://www.linkace.org/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
