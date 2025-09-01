---
title: "Passbolt"
description: "Integrating Passbolt with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-07-19T02:33:13+00:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/passbolt/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Passbolt | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Passbolt with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.7](https://github.com/authelia/authelia/releases/tag/v4.39.7)
- [Passbolt]
  - [v5.3.2](https://www.passbolt.com/changelog/api-bext/somebody-to-love-browser-extension-api)

{{% oidc-common bugs="claims-hydration" %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://passbolt.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `passbolt`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Passbolt] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'passbolt'
        client_name: 'Passbolt'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - '<copy from the passbolt setup>'
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
        token_endpoint_auth_method: 'client_secret_post'
```

#### Configuration Escape Hatch

{{% oidc-escape-hatch-claims-hydration client_id="passbolt" claims="email,email_verified,alt_emails,preferred_username,name" %}}

### Application

To configure [Passbolt] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Passbolt] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following
instructions:

1. Visit the [Passbolt] admin panel
2. Visit `Org Settings`
3. Visit `Auth`
4. Visit `SSO`
5. Configure the following values:
   1. Login URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
   2. OpenID configuration URI: `/.well-known/openid-configuration`
   3. Scope: `openid email profile`
   4. Client Login: `passbolt`
   5. Client Secret: `insecure_secret`
6. Copy the callback URL into the `redirect_uris` of the Authelia configuration

## See Also

- [Passbolt OpenID Blog Post](https://www.passbolt.com/blog/openid-for-sso)

[Authelia]: https://www.authelia.com
[Passbolt]: https://www.passbolt.com/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
