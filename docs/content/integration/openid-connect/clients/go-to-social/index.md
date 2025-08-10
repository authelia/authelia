---
title: "GoToSocial"
description: "Integrating GoToSocial with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-07-14T23:37:32+10:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/go-to-social/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "GoToSocial | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring GoToSocial with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.19](https://github.com/authelia/authelia/releases/tag/v4.38.19)
- [GoToSocial]
  - [v0.19.1](https://codeberg.org/superseriousbusiness/gotosocial/releases/tag/v0.19.1)

{{% oidc-common bugs="claims-hydration" %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://gotosocial.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `gotosocial`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [GoToSocial] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'gotosocial'
        client_name: 'GoToSocial'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://gotosocial.{{< sitevar name="domain" nojs="example.com" >}}/auth/callback'
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

#### Configuration Escape Hatch

{{% oidc-escape-hatch-claims-hydration client_id="gotosocial" claims="preferred_username" %}}

### Application

To configure [GoToSocial] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

```shell
oidc-enabled: true
oidc-idp-name: 'Authelia'
oidc-issuer: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
oidc-client-id: 'gotosocial'
oidc-client-secret: 'insecure_secret'
oidc-scopes:
  - 'openid'
  - 'email'
  - 'profile'
  - 'groups'
oidc-allowed-groups: []
oidc-admin-groups:
  - 'admin'
```

## See Also

- [GoToSocial OpenID Connect (OIDC) Documentation](https://docs.gotosocial.org/en/latest/configuration/oidc/)

[Authelia]: https://www.authelia.com
[GoToSocial]: https://gotosocial.org/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
