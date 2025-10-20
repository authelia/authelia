---
title: "BookStack"
description: "Integrating BookStack with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/bookstack/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "BookStack | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring BookStack with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.13](https://github.com/authelia/authelia/releases/tag/v4.39.13)
- [BookStack]
  - [v25.07](https://github.com/BookStackApp/BookStack/releases/tag/v25.07)

{{% oidc-common bugs="client-credentials-encoding" %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://bookstack.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `bookstack`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [BookStack] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'bookstack'
        client_name: 'BookStack'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://bookstack.{{< sitevar name="domain" nojs="example.com" >}}/oidc/callback'
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

To configure [BookStack] there is one method, using the [Environment Variables](#environment-variables).

#### Environment Variables

To configure [BookStack] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
AUTH_METHOD=oidc
AUTH_AUTO_INITIATE=false
OIDC_NAME=Authelia
OIDC_ISSUER=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
OIDC_ISSUER_DISCOVER=true
OIDC_CLIENT_ID=bookstack
OIDC_CLIENT_SECRET=insecure_secret
OIDC_ADDITIONAL_SCOPES=groups
OIDC_DISPLAY_NAME_CLAIMS=name
OIDC_GROUPS_CLAIM=groups
OIDC_USER_TO_GROUPS=true
OIDC_REMOVE_FROM_GROUPS=true
OIDC_END_SESSION_ENDPOINT=false
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  bookstack:
    environment:
      AUTH_METHOD: 'oidc'
      AUTH_AUTO_INITIATE: 'false'
      OIDC_NAME: 'Authelia'
      OIDC_ISSUER: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      OIDC_ISSUER_DISCOVER: 'true'
      OIDC_CLIENT_ID: 'bookstack'
      OIDC_CLIENT_SECRET: 'insecure_secret'
      OIDC_ADDITIONAL_SCOPES: 'groups'
      OIDC_DISPLAY_NAME_CLAIMS: 'name'
      OIDC_GROUPS_CLAIM: 'groups'
      OIDC_USER_TO_GROUPS: 'true'
      OIDC_REMOVE_FROM_GROUPS: 'true'
      OIDC_END_SESSION_ENDPOINT: 'false'
```

## See Also

- [BookStack OpenID Connect Documentation](https://www.bookstackapp.com/docs/admin/oidc-auth/)

[Authelia]: https://www.authelia.com
[BookStack]: https://www.bookstackapp.com/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
