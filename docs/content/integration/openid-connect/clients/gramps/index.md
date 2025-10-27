---
title: "Gramps Web"
description: "Integrating Gramps Web with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-10-25T18:05:17+01:00
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
  title: "Gramps Web | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Gramps Web with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.13](https://github.com/authelia/authelia/releases/tag/v4.39.13)
- [Gramps Web]
  - [v25.10.1](https://github.com/gramps-project/gramps-web/releases/tag/v25.10.1)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://gramps.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `gramps`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Gramps Web] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'gramps'
        client_name: 'gramps'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'http://gramps.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/callback/?provider=custom'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
          - 'groups' # optional: include if you want to link Authelia groups to Gramps roles
```

### Application

To configure [Gramps Web] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [Gramps Web] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
GRAMPSWEB_BASE_URL=https://gramps.{{< sitevar name="domain" nojs="example.com" >}}
GRAMPSWEB_OIDC_ENABLED=True
GRAMPSWEB_OIDC_ISSUER=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/
GRAMPSWEB_OIDC_NAME=Authelia
GRAMPSWEB_OIDC_CLIENT_ID=gramps
GRAMPSWEB_OIDC_CLIENT_SECRET=insecure_secret
GRAMPSWEB_OIDC_SCOPES="openid email profile"
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  grampsweb:
    environment:
      GRAMPSWEB_BASE_URL: https://gramps.{{< sitevar name="domain" nojs="example.com" >}}
      GRAMPSWEB_OIDC_ENABLED: True
      GRAMPSWEB_OIDC_ISSUER: https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/
      GRAMPSWEB_OIDC_NAME: Authelia
      GRAMPSWEB_OIDC_CLIENT_ID: gramps
      GRAMPSWEB_OIDC_CLIENT_SECRET: insecure_secret
      GRAMPSWEB_OIDC_SCOPES: openid email profile
```

## See Also

- [Gramps Web OIDC configuration instructions](https://www.grampsweb.org/install_setup/oidc/).

[Gramps Web]: https://www.grampsweb.org/
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
