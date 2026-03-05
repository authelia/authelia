---
title: "HomeBox"
description: "Integrate HomeBox with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2026-02-01T22:46:39+00:00
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
  title: "HomeBox | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring HomeBox with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.15](https://github.com/authelia/authelia/releases/tag/v4.39.15)
- [HomeBox]
  - [v0.23.1](https://github.com/sysadminsmedia/homebox/releases/tag/v0.23.1)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://homebox.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `homebox`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [HomeBox] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'homebox'
        client_name: 'homebox'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://homebox.{{< sitevar name="domain" nojs="example.com" >}}/api/v1/users/login/oidc/callback'
        scopes:
          - 'openid'
          - 'groups'
          - 'email'
          - 'profile'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [HomeBox] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [HomeBox] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment variables in the '.env' file:

##### Standard

```shell {title=".env"}
HBOX_OIDC_ENABLED=true
HBOX_OIDC_ISSUER_URL=https://auth.{{< sitevar name="domain" nojs="example.com" >}}
HBOX_OIDC_CLIENT_ID=homebox
HBOX_OIDC_CLIENT_SECRET=insecure_secret
HBOX_OIDC_SCOPE=openid profile email groups
HBOX_OPTIONS_TRUST_PROXY=true # this is only needed if you are running HomeBox behind a reverse proxy
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  homebox:
    environment:
      HBOX_OIDC_ENABLED: 'true'
      HBOX_OIDC_ISSUER_URL: 'https://auth.{{< sitevar name="domain" nojs="example.com" >}}'
      HBOX_OIDC_CLIENT_ID: 'homebox'
      HBOX_OIDC_CLIENT_SECRET: 'insecure_secret'
      HBOX_OIDC_SCOPE: 'openid profile email groups'
      HBOX_OPTIONS_TRUST_PROXY: 'true' # this is only needed if you are running HomeBox behind a reverse proxy
```

## See Also

- [HomeBox OIDC documentation](https://homebox.software/en/configure/oidc)

[Authelia]: https://www.authelia.com
[HomeBox]: https://homebox.software/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
