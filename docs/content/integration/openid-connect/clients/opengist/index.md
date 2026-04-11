---
title: "Opengist"
description: "Integrating Opengist with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2026-03-07T23:00:00+11:00
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
  title: "Opengist | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Opengist with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.19](https://github.com/authelia/authelia/releases/tag/v4.39.19)
- [Opengist]
  - [v1.12.1](https://github.com/thomiceli/opengist/releases/tag/v1.12.1)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://opengist.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `opengist`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Opengist

The following YAML configuration is an example Opengist [client configuration] for use with [Opengist] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'opengist'
        client_name: 'Opengist'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://opengist.{{< sitevar name="domain" nojs="example.com" >}}/oauth/openid-connect/callback'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
          - 'groups' # Supports https://opengist.io/docs/configuration/oauth-providers.html#oidc-admin-group
        grant_types:
          - 'authorization_code'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [Opengist] there are two methods, using the [Environment Variables](#environment-variables), or using the [Configuration File](#configuration-file).

See https://opengist.io/docs/configuration/oauth-providers.html#openid-connect.

#### Environment Variables

To configure [Opengist] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
OG_OIDC_PROVIDER_NAME=authelia
OG_OIDC_CLIENT_KEY=opengist
OG_OIDC_SECRET=insecure_secret
OG_OIDC_DISCOVERY_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration
OG_OIDC_GROUP_CLAIM_NAME=groups
OG_OIDC_ADMIN_GROUP=admin-group-name
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  opengist:
    environment:
      OG_OIDC_PROVIDER_NAME: 'authelia'
      OG_OIDC_CLIENT_KEY: 'opengist'
      OG_OIDC_SECRET: 'insecure_secret'
      OG_OIDC_DISCOVERY_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration'
      OG_OIDC_GROUP_CLAIM_NAME: 'groups'
      OG_OIDC_ADMIN_GROUP: 'admin-group-name'
```

#### Configuration file

```yaml {title="configuration.yml"}
oidc.provider-name: authelia
oidc.client-key: opengist
oidc.secret: insecure_secret
oidc.discovery-url: https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration

oidc.group-claim-name: groups # Name of the claim containing the groups
oidc.admin-group: admin-group-name # Name of the group that should receive admin rights

# Required for correct setting of ?redirect_uri= in OIDC callback URL
external-url: https://opengist.{{< sitevar name="domain" nojs="example.com" >}}
```

## See Also

- [Opengist 'Use OAuth providers' documentation](https://opengist.io/docs/configuration/oauth-providers.html#openid-connect)

[Authelia]: https://www.authelia.com
[Opengist]: https://opengist.io
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
