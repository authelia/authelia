---
title: "Kasm Workspaces"
description: "Integrating Kasm Workspaces with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/kasm-workspaces/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Kasm Workspaces | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Kasm Workspaces with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
- [Kasm Workspaces]
  - [v1.13.0](https://kasmweb.com/docs/latest/release_notes/1.13.0.html)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://kasm.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `kasm`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Kasm Workspaces] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'kasm'
        client_name: 'Kasm Workspaces'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://kasm.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc_callback'
        scopes:
          - 'openid'
          - 'profile'
          - 'groups'
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

To configure [Kasm Workspaces] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Kasm Workspaces] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Visit Authentication
2. Visit OpenID
3. Configure the following options:
   - Automatic User Provision: Enable if you want users to automatically be created in [Kasm Workspaces].
   - Auto Login: Enable if you want automatic user login.
   - Default: Enable if you want Authelia to be the default sign-in method.
   - Client ID: `kasm`
   - Client Secret: `insecure_secret`
   - Authorization URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization`
   - Token URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token`
   - User Info URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo`
   - Scope (One Per Line): `openid profile groups email`
   - User Identifier: `preferred_username`

{{< figure src="kasm.png" alt="Kasam Workspaces" width="736" style="padding-right: 10px" >}}

## See Also

- [Kasm Workspaces OpenID Connect Authentication Documentation](https://kasmweb.com/docs/latest/guide/oidc.html)

[Authelia]: https://www.authelia.com
[Kasm Workspaces]: https://kasmweb.com/
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
