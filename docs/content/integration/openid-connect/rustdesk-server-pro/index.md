---
title: "RustDesk Server Pro"
description: "Integrating RustDesk Server Pro with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-01-25T10:04:53+11:00
draft: false
images: []
weight: 620
toc: true
support:
  level: community
  versions: true
  integration: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.4](https://github.com/authelia/authelia/releases/tag/v4.39.4)
- [RustDesk Server Pro]
  - [v1.3.9](https://github.com/rustdesk/rustdesk/releases/tag/1.3.9)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://rustdesk.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `rustdesk`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [RustDesk Server Pro] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'rustdesk'
        client_name: 'RustDesk Server Pro'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://rustdesk.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/callback'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [RustDesk Server Pro] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [RustDesk Server Pro] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [RustDesk Server Pro].
2. Navigate to Settings.
3. Navigate to OIDC.
4. Click `+ New Auth Provider`.
5. Configure the following options:
   - Name: `Authelia`
   - Client ID: `rustdesk`
   - Client Secret: `insecure_secret`
   - Issuer: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - Authorization Endpoint: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization`
   - Token Endpoint: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token`
   - Userinfo Endpoint: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo`
   - JWKS Endpoint: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/jwks.json`
6. Press `Submit` at the bottom.

## See Also

- [RustDesk Server Pro OIDC documentation](https://rustdesk.com/docs/en/self-host/rustdesk-server-pro/oidc/)

[Authelia]: https://www.authelia.com
[RustDesk Server Pro]: https://rustdesk.com
[OAuth login Extension]: https://www.rustdesk.com/extensions/oauth/
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
