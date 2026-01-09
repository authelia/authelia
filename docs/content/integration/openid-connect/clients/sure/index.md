---
title: "Sure"
description: "Integrating Sure with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-01-25T12:36:00+11:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/sure/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Sure | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Sure with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.15](https://github.com/authelia/authelia/releases/tag/v4.39.15)
- [Sure]
  - [v0.6.6](https://github.com/we-promise/sure/releases/tag/v0.6.6)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://sure.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
        `https://sure.{{< sitevar name="domain" nojs="example.com" >}}/auth/openid_connect/callback`.
        This means if you change this value, you need to update the redirect URI.
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `sure`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Sure] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'sure'
        client_name: 'Sure'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://sure.{{< sitevar name="domain" nojs="example.com" >}}/auth/openid_connect/callback'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
          - 'groups'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [Sure] there are two methods, using [Environment Variables](#environment-variables), or using the [Web GUI](#web-gui).

#### Environment Variables

To configure [Sure] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
OIDC_ISSUER=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
OIDC_CLIENT_ID=sure
OIDC_CLIENT_SECRET=insecure_secret
OIDC_REDIRECT_URI=https://sure.{{< sitevar name="domain" nojs="example.com" >}}/auth/openid_connect/callback
OIDC_BUTTON_LABEL=Sign in with Authelia
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  sure:
    environment:
      OIDC_ISSUER: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      OIDC_CLIENT_ID: 'sure'
      OIDC_CLIENT_SECRET: 'insecure_secret'
      OIDC_REDIRECT_URI: 'https://sure.{{< sitevar name="domain" nojs="example.com" >}}/auth/openid_connect/callback'
      OIDC_BUTTON_LABEL: 'Sign in with Authelia'
```

#### Web GUI

To configure [Sure] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Log in as the admin user.
2. Navigate to Settings.
3. In the Advanced section click on SSO Providers.
4. Configure the following options:
   - Strategy: `OpenID Connect`
   - Name: `authelia`
   - Button Label: `Sign in with Authelia`
   - Issuer URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - Client ID: `sure`
   - Client Secret: `insecure_secret`
5. Click Test Connection.
6. Click Update Provider.

## See Also

- [Sure OpenID Provider Documentation](https://github.com/we-promise/sure/blob/main/docs/hosting/oidc.md)

[Authelia]: https://www.authelia.com
[Sure]: https://sure.am/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
