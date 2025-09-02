---
title: "Beszel"
description: "Integrating Beszel with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-04-26T11:03:16+00:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/beszel/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Beszel | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Beszel with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.8](https://github.com/authelia/authelia/releases/tag/v4.39.8)
- [Beszel]
  - [v0.10.2](https://github.com/henrygd/beszel/releases/tag/v0.10.2)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://beszel.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `beszel`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Beszel] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'beszel'
        client_name: 'Beszel'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://beszel.{{< sitevar name="domain" nojs="example.com" >}}/api/oauth2-redirect'
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

To configure [Beszel] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Beszel] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [Beszel].
2. Navigate to the settings dashboard by going to `https://beszel.{{< sitevar name="domain" nojs="example.com" >}}/_/#/settings`.
3. Disable the `Hide collection create and edit controls`.
4. Edit the `users` collection by clicking the cog and selecting `Options`.
5. Expand OAuth2.
6. Toggle Enable to the on position.
7. Click `Add Provider`.
8. Select `OpenID Connect`.
9. Configure the following options:
   - Client ID: `beszel`
   - Client secret: `insecure_secret`
   - Display name: `Authelia`
   - Auth URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization`
   - Token URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token`
   - Fetch user info from: `User info URL`
   - User info URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo`
10. Press `Save` at the bottom.

## See Also

- [Beszel OIDC documentation](https://beszel.dev/guide/oauth)

[Authelia]: https://www.authelia.com
[Beszel]: https://beszel.dev/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
