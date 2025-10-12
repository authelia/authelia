---
title: "Bitwarden"
description: "Integrating Bitwarden with the Authelia OpenID Connect 1.0 Provider."
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
  title: "Bitwarden | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Bitwarden with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.12](https://github.com/authelia/authelia/releases/tag/v4.39.12)
- [Bitwarden]
  - [v2025.7.3](https://github.com/bitwarden/server/releases/tag/v2025.7.3)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://bitwarden.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `bitwarden`
- __Client Secret:__ `insecure_secret`

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
This setup assumes you're using the self-hosted version of Bitwarden. If you're using the SaaS version the `redirect_uris` are either
usually `https://sso.bitwarden.com/oidc-signin` or if you're in the EU `https://sso.bitwarden.eu/oidc-signin`.
{{< /callout >}}

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Bitwarden] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'bitwarden'
        client_name: 'Bitwarden'
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://bitwarden.{{< sitevar name="domain" nojs="example.com" >}}/oidc-signin'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
        response_modes:
          - 'form_post'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [Bitwarden] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Bitwarden] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to your Bitwarden administrator account.
2. Select `Admin Console`.
3. Select `Settings`.
4. Select `Single sign-on`.
5. Select `Allow SSO authentication`.
6. If you're using [Bitwarden] SaaS configure the SSO Identifier per their instructions.
7. Select `OpenID Connect` for the type.
8. Enter the following values:
  - Authority: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
  - Client ID: `bitwarden`
  - Client Secret: `insecure_secret`
  - Metadata Address: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration`
  - OIDC redirect behavior: `Form POST`
  - Get claims from user info endpoint: Enabled
9. Click Save.

## See Also

- [Bitwarden Configure Unlock Bitwarden with SSO using OpenID Connect Documentation](https://support.bitwarden.com/sso-configure-generic/)

[Authelia]: https://www.authelia.com
[Bitwarden]: https://bitwarden.com/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
