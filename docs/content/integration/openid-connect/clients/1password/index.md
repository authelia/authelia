---
title: "1Password"
description: "Integrating 1Password with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-04-26T18:35:57+10:00
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
  title: "1Password | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring 1Password with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.11](https://github.com/authelia/authelia/releases/tag/v4.39.11)
- [1Password]

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `1password`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [1Password] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: '1password'
        client_name: '1Password'
        public: true
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - ''  # See step 5 below.
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
        token_endpoint_auth_method: 'none'
```

### Application

To configure [1Password] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [1Password] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. [Sign in to your 1Password](https://start.1password.com/policies/sso/configure-idp).
2. Select `Other` from the list of identity providers and select `Next`.
3. Select `Other` from the identity provider list.
4. Configure the following options:
  - Client ID: `1password`
  - Well-known URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration`
5. Copy the listed `Redirect URIs` and configure the Authelia `redirect_uris` option.

## See Also

- [1Password Configure Unlock 1Password with SSO using OpenID Connect Documentation](https://support.1password.com/sso-configure-generic/)

[Authelia]: https://www.authelia.com
[1Password]: https://1password.com/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
