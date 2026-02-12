---
title: "Wanderer"
description: "Integrating Wanderer with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2026-02-12T00:00:00+00:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/wanderer/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Wanderer | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Wanderer with OpenID Connect 1.0 for secure SSO."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.15](https://github.com/authelia/authelia/releases/tag/v4.39.15)
- [Wanderer]
  - [v0.18.4](https://github.com/open-wanderer/wanderer/releases/tag/v0.18.4)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://wanderer.{{< sitevar name="domain" nojs="example.com" >}}/`
  - Wanderer uses the `ORIGIN` environment variable as the public URL. The redirect URL is `${ORIGIN}/login/redirect`.
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `wanderer`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Wanderer] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'wanderer'
        client_name: 'Wanderer'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'deny_guests'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://wanderer.{{< sitevar name="domain" nojs="example.com" >}}/login/redirect'
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

Wanderer uses PocketBase for authentication configuration. To configure [Wanderer] to utilize Authelia as an
[OpenID Connect 1.0] Provider:

1. Sign in to the PocketBase admin UI using your superuser.
1. Navigate to the `users` collection.
1. Click the gear icon to open the collection settings.
1. Navigate to `Options`.
1. In the `OAuth2` tab, add a new provider with type `OpenID Connect (oidc)`.
1. Configure the provider with these options:

    - Client ID: `wanderer`
    - Client secret: `insecure_secret`
    - Display name: `Authelia`
    - Auth URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization`
    - Token URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token`
    - Fetch user info from: `User info URL`
    - User info URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo`
    - Support PKCE: optional

1. Save your changes.

Ensure the redirect URL configured at the provider is exactly `${ORIGIN}/login/redirect` and that the same value is
present in the Authelia `redirect_uris` list.

## See Also

- [Wanderer OAuth2 Documentation](https://wanderer.to/run/backend-configuration/oauth2/)

[Authelia]: https://www.authelia.com
[Wanderer]: https://wanderer.to/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
