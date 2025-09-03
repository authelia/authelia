---
title: "Glitchtip"
description: "Integrating Glitchtip with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-04-26T11:03:16+00:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/glitchtip/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Glitchtip | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Glitchtip with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.8](https://github.com/authelia/authelia/releases/tag/v4.39.8)
- [Glitchtip]
  - [v4.2](https://glitchtip.com/blog/2024-11-01-glitchtip-4-2-release)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://glitchtip.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `glitchtip`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

The following instructions assume you've setup a Django Admin / Super User. See the
[Django Admin](https://glitchtip.com/documentation/install#django-admin) guide for more information.

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Glitchtip] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'glitchtip'
        client_name: 'Glitchtip'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://glitchtip.{{< sitevar name="domain" nojs="example.com" >}}/accounts/authelia/login/callback/'
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
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [Glitchtip] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Glitchtip] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Visit `https://glitchtip.{{< sitevar name="domain" nojs="example.com" >}}/admin/socialaccount/socialapp/`.
2. Click `Add Social Application`.
3. Configure the following options:
   - Provider: `OpenID Connect`
   - Provider ID: `authelia`
   - Provider Name: `Authelia`
   - Client ID: `glitchtip`
   - Secret Key: `insecure_secret`
   - Settings: `{"server_url":"https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration"}`
6. Press `Save` at the bottom.

## See Also

- [Glitchtip Configuring OpenID Connect based SSO Documentation](https://glitchtip.com/documentation/install#configuring-openid-connect-based-sso)

[Authelia]: https://www.authelia.com
[Glitchtip]: https://glitchtip.com/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
