---
title: "PocketBase"
description: "Integrating PocketBase with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
weight: 720
toc: true
support:
  level: community
  versions: true
  integration: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
- [PocketBase]
  - [v0.27.1](https://github.com/pocketbase/pocketbase/releases/tag/v0.27.1)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://pocketbase.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `pocketbase`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [PocketBase] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'pocketbase'
        client_name: 'PocketBase'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://pocketbase.{{< sitevar name="domain" nojs="example.com" >}}/api/oauth2-redirect'
        scopes:
          - 'email'
          - 'groups'
          - 'openid'
          - 'profile'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [PocketBase] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [PocketBase] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Connect to PocketBase admin view.
2. On the left menu, go to `Settings`.
3. In `Authentication` section, go to `Auth providers`.
4. Select the gear on `OpenID Connect (oidc)`
5. Configure the following options:
   - ClientID: `pocketbase`
   - Client secret: `insecure_secret`
   - Display name: `Authelia` (or whatever you want)
   - Auth URL: https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization
   - Token URL: https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token
   - User API URL: https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo
   - You can leave `Support PKCE` checked.
6. Save changes.

## See Also

- [PocketBase OAuth Documentation](https://pocketbase.io/docs/authentication/#oauth2-integration)

[Authelia]: https://www.authelia.com
[PocketBase]: https://pocketbase.io
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
