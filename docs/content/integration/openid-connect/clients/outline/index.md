---
title: "Outline"
description: "Integrating Outline with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/outline/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Outline | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Outline with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
- [Outline]
  - [v0.65.2](https://github.com/outline/outline/releases/tag/v0.65.2)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://outline.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `outline`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
At the time of this writing [Outline](https://www.getoutline.com/) requires the `offline_access` scope by default. Failure to
include this scope will result in an error as [Outline](https://www.getoutline.com/) will attempt to use a refresh token that is never issued.
{{< /callout >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Outline] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'outline'
        client_name: 'Outline'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://outline.{{< sitevar name="domain" nojs="example.com" >}}/auth/oidc.callback'
        scopes:
          - 'openid'
          - 'offline_access'
          - 'profile'
          - 'email'
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

To configure [Outline] there is one method, using the [Environment Variables](#environment-variables).

#### Environment Variables

To configure [Outline] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment variables:

##### Standard

```shell {title=".env"}
URL=https://outline.{{< sitevar name="domain" nojs="example.com" >}}
OIDC_CLIENT_ID=outline
OIDC_CLIENT_SECRET=insecure_secret
OIDC_AUTH_URI=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization
OIDC_TOKEN_URI=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token
OIDC_USERINFO_URI=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo
OIDC_USERNAME_CLAIM=preferred_username
OIDC_DISPLAY_NAME=Authelia
OIDC_SCOPES=openid offline_access profile email
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  outline:
    environment:
      URL: 'https://outline.{{< sitevar name="domain" nojs="example.com" >}}'
      OIDC_CLIENT_ID: 'outline'
      OIDC_CLIENT_SECRET: 'insecure_secret'
      OIDC_AUTH_URI: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization'
      OIDC_TOKEN_URI: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token'
      OIDC_USERINFO_URI: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo'
      OIDC_USERNAME_CLAIM: 'preferred_username'
      OIDC_DISPLAY_NAME: 'Authelia'
      OIDC_SCOPES: 'openid offline_access profile email'
```

## See Also

- [Outline OpenID Connect Documentation](https://app.getoutline.com/share/770a97da-13e5-401e-9f8a-37949c19f97e/doc/oidc-8CPBm6uC0I)

[Authelia]: https://www.authelia.com
[Outline]: https://www.getoutline.com/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
