---
title: "Matomo"
description: "Integrating Matomo with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-10-05T22:31:30+10:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/matomo/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Matomo | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Matomo with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.14](https://github.com/authelia/authelia/releases/tag/v4.38.14)
- [Matomo]
  - [v5.1.2](https://github.com/matomo-org/matomo/releases/tag/5.1.2)
- [Login OIDC Plugin]:
  - [v5.0.0](https://github.com/dominik-th/matomo-plugin-LoginOIDC/releases/tag/5.0.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://matomo.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `matomo`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

The following example uses the [Login OIDC Plugin] which is assumed to be installed when following this
section of the guide.

To install the [Login OIDC Plugin] for [Matomo] via the Web GUI:

1. Visit the [Matomo] `Administration` page.
2. Click `Plugins`.
3. Click `Manage Plugins`.
4. Click `installing plugins from the Marketplace`.
5. Install `Login OIDC` by `dominik-th`.

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Matomo] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'matomo'
        client_name: 'Matomo'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://matomo.{{< sitevar name="domain" nojs="example.com" >}}/index.php?module=LoginOIDC&action=callback&provider=oidc'
        scopes:
          - 'openid'
          - 'groups'
          - 'email'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [Matomo] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Matomo] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Click `System`.
2. Click `General settings`.
3. Click `Login OIDC`.
4. Configure the following options:
   - Authorize URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization`
   - Token URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token`
   - Userinfo URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo`
   - Userinfo ID: `sub`
   - Client ID: `matomo`
   - Client Secret: `insecure_secret`
   - OAuth Scope: `openid email`

## See Also

- [Matomo Login OIDC FAQ](https://plugins.matomo.org/LoginOIDC/#faq)

[Matomo]: https://matomo.org/
[Authelia]: https://www.authelia.com
[Login OIDC Plugin]: https://plugins.matomo.org/LoginOIDC/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
