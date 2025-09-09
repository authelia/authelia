---
title: "OpenID Connect Playground"
description: "Integrating OpenID Connect Playground with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-05-07T09:48:38+10:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/openid-connect-playground/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "OpenID Connect Playground | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring OpenID Connect Playground with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.9](https://github.com/authelia/authelia/releases/tag/v4.39.9)
- [OpenID Connect Playground]
  - Not Applicable

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `openid-connect-playground`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [OpenID Connect Playground] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'openid-connect-playground'
        client_name: 'OpenID Connect Playground'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://openidconnect.net/callback'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
          - 'phone'
          - 'address'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [OpenID Connect Playground] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [OpenID Connect Playground] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Visit [OpenID Connect Playground].
2. Visit `Configuration`.
3. Configure the following options:
   - Server Template: `Custom`
   - Discovery Document URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration`
   - OIDC Client ID: `openid-connect-playground`
   - OIDC Client Secret: `insecure_secret`
   - Scope: `openid profile email phone address`
4. Click `Use Discovery Document`.
5. Verify the following options:
   - Authorization Token Endpoint: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization`
   - Token Endpoint: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token`
   - Token Keys Endpoint: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/jwks.json`
6. Click Save.

[Authelia]: https://www.authelia.com
[OpenID Connect Playground]: https://openidconnect.net/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
