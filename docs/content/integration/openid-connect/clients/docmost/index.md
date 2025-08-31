---
title: "Docmost"
description: "Integrating Docmost with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-04-26T18:35:57+10:00
draft: false
images: []
weight: 620
toc: true
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Docmost | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Docmost with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.7](https://github.com/authelia/authelia/releases/tag/v4.39.7)
- [Docmost]
  - [v0.22.2](https://github.com/docmost/docmost/releases/tag/v0.22.2)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://docmost.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `docmost`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Docmost] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'docmost'
        client_name: 'Docmost'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://docmost.{{< sitevar name="domain" nojs="example.com" >}}'
        scopes:
          - 'openid'
          - 'profile'
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

To configure [Docmost] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Docmost] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [Docmost].
2. Navigate to Settings.
3. Navigate to Security & SSO.
4. Select `Create SSO`.
5. Select `OpenID (OIDC)` from the dropdown menu.
6. Copy the `Callback URL` and replace the configured `redirect_uri` value in the Authelia configuration.
7. Configure the following options:
  - Display name: `Authelia`
  - Issuer URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
  - Client ID: `docmost`
  - Client Secret: `insecure_secret`
  - Allow signup: Disabled
  - Enabled: Enabled
8. Press `Save` at the bottom.

## See Also

- [Docmost OIDC Authentication Documentation](https://docmost.com/docs/user-guide/authentication/oidc)

[Docmost]: https://docmost.com/
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
