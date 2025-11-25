---
title: "Wallos"
description: "Integrating Wallos with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-08-19T20:35:59+02:00
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
  title: "Wallos | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Wallos with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.15](https://github.com/authelia/authelia/releases/tag/v4.39.15)
- [Wallos]
  - [v4.1.1](https://github.com/ellite/Wallos/releases/tag/v4.1.1)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://wallos.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `wallos`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Wallos] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'wallos'
        client_name: 'Wallos'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://wallos.{{< sitevar name="domain" nojs="example.com" >}}/index.php'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [Wallos] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Wallos] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [Wallos] using the admin account.
2. Navigate to the Admin panel and scroll down to OIDC settings.
3. Click `Enable OIDC/OAuth`.
4. Configure the following options:
    - Provider Name: `Authelia`.
    - Client ID: `wallos`.
    - Client Secret: `insecure_secret`.
    - Auth URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization`.
    - Token URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token`.
    - User Info URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo`.
    - Redirect URL: `https://wallos.{{< sitevar name="domain" nojs="example.com" >}}/index.php`.
    - (Optional) Logout URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/logout`.
    - (Default) User Identifier Field: `sub`.
    - (Default) Scopes: `openid email profile`.
5. Press `Save` at the bottom.

## See Also

- [Wallos OIDC Documentation](https://github.com/ellite/Wallos?tab=readme-ov-file#oidc)

[Authelia]: https://www.authelia.com
[Wallos]: https://www.wallosapp.com
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
