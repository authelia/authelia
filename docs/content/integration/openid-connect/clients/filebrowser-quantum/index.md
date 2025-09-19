---
title: "FileBrowser Quantum"
description: "Integrating FileBrowser Quantum with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-07-14T23:37:32+10:00
draft: false
images: []
weight: 620
toc: true
support:
  level: community
  versions: true
  integration: true
seo:
  title: "FileBrowser Quantum | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring FileBrowser Quantum with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.10](https://github.com/authelia/authelia/releases/tag/v4.39.10)
- [FileBrowser Quantum]
  - [v0.7.18-beta](https://github.com/gtsteffaniak/filebrowser/releases/tag/v0.7.18-beta)

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://filebrowser.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `filebrowser`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [FileBrowser Quantum] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'filebrowser'
        client_name: 'FileBrowser Quantum'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://filebrowser.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/oidc/callback'
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

To configure [FileBrowser Quantum] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

```yaml
auth:
  methods:
    password:
      # Set to false if you only want to allow OIDC.
      enabled: true
    oidc:
      enabled: true
      clientId: 'filebrowser'
      clientSecret: 'insecure_secret'
      issuerUrl: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      scopes: 'email openid profile groups'
      userIdentifier: 'preferred_username'
      disableVerifyTLS: false
      logoutRedirectUrl: ''
      createUser: true
      adminGroup: 'admin'
      groupsClaim: 'groups'
```

## See Also

- [FileBrowser Quantum Configuration (Auth Examples) Documentation](https://github.com/gtsteffaniak/filebrowser/wiki/Configuration-And-Examples#auth-config-examples)

[Authelia]: https://www.authelia.com
[FileBrowser Quantum]: https://github.com/gtsteffaniak/filebrowser
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
