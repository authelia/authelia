---
title: "Arcane"
description: "Integrating Arcane with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-01-25T12:36:00+11:00
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
  title: "Arcane | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Arcane with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.15](https://github.com/authelia/authelia/releases/tag/v4.39.15)
- [Arcane]
  - [v1.1.0](https://github.com/ofkm/arcane/releases/tag/v1.1.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://arcane.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
        `https://arcane.{{< sitevar name="domain" nojs="example.com" >}}/login`.
        This means if you change this value, you need to update the redirect URI.
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `arcane`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Arcane] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'arcane'
        client_name: 'Arcane'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://arcane.{{< sitevar name="domain" nojs="example.com" >}}/auth/oidc/callback'
        scopes:
          - 'openid'
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

### Application

To configure [Arcane] there are two methods, using
[Environment Variables](#environment-variables), or using the [Web GUI](#web-gui).

#### Environment Variables

To configure [Arcane] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
APP_URL=https://arcane.{{< sitevar name="domain" nojs="example.com" >}}
OIDC_ENABLED=true
OIDC_ISSUER_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
OIDC_CLIENT_ID=arcane
OIDC_CLIENT_SECRET=insecure_secret
OIDC_SCOPES=openid email profile
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  arcane:
    environment:
      APP_URL: 'https://arcane.{{< sitevar name="domain" nojs="example.com" >}}'
      OIDC_ENABLED: 'true'
      OIDC_ISSUER_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      OIDC_CLIENT_ID: 'arcane'
      OIDC_CLIENT_SECRET: 'insecure_secret'
      OIDC_SCOPES: 'openid email profile'
```

#### Web GUI

{{< callout context="tip" title="Help Wanted" icon="outline/rocket" >}}
We would love screenshots of this configuration!
{{< /callout >}}

To configure [Arcane] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Navigate to Settings.
2. Navigate to Authentication.
3. Configure the following options:
   - Issuer URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - Client ID: `arcane`
   - Client Secret: `insecure_secret`
   - Redirect URI: 'https://arcane.{{< sitevar name="domain" nojs="example.com" >}}/auth/oidc/callback'
4. Click Save.

## See Also

- [Arcane OIDC Single Sign-On Documentation](https://arcane.ofkm.dev/docs/users/sso)

[Authelia]: https://www.authelia.com
[Arcane]: https://arcane.ofkm.dev/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
