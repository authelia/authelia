---
title: "Xen Orchestra"
description: "Integrating Xen Orchestra with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-04-26T18:35:57+10:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/xen-orchestra/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Xen Orchestra | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Xen Orchestra with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.6](https://github.com/authelia/authelia/releases/tag/v4.39.6)
- [Xen Orchestra]
  - [v5.105](https://xen-orchestra.com/blog/xen-orchestra-5-105/)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://xen-orchestra.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `xen-orchestra`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Xen Orchestra] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'xen-orchestra'
        client_name: 'Xen Orchestra'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://xen-orchestra.{{< sitevar name="domain" nojs="example.com" >}}/signin/oidc/callback'
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

To configure [Xen Orchestra] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Xen Orchestra] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [Xen Orchestra].
2. Navigate to Settings.
3. Navigate to Plugins.
4. Navigate to the `auth-oidc` plugin and click `+`.
5. Configure the following options:
   - Auto-discovery Endpoint: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration`
   - Client identifier (key): `xen-orchestra`
   - Client secret: `insecure_secret`
   - Fill information (optional): Enabled
   - Username field: `preferred_username`
   - Scopes: `openid profile email`
6. Press `Save configuration`.
7. Toggle the switch next to the `auth-oidc` plugin name.

## See Also

- [Xen Orchestra OpenID Connect Guide](https://docs.xen-orchestra.com/users#openid-connect)

[Xen Orchestra]: https://xen-orchestra.com/
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
