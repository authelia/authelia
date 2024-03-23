---
title: "Memos"
description: "Integrating Memos with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2023-11-12T21:18:09+11:00
draft: false
images: []
weight: 620
toc: true
community: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

* [Authelia]
  * [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
* [Memos](https://github.com/usememos/memos)
  * [0.16.1](https://github.com/usememos/memos/tree/v0.16.1)

## Before You Begin

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://memos.example.com/`
* __Authelia Root URL:__ `https://auth.example.com/`
* __Client ID:__ `memos`
* __Client Secret:__ `insecure_secret`


## Configuration

### Application

To configure [Memos](https://github.com/usememos/memos) to utilize Authelia as an [OpenID Connect 1.0] Provider:

1. Go to the settings menu, choose `SSO`, `create` and `OAuth2`
2. Choose template `custom`
3. Set the following values:
   1. Name: `Authelia`
   2. Identifier Filter:
   3. Client ID: `memos`
   4. Client secret: `insecure_secret`
   5. Authorization endpoint: 	`https://auth.example.com/api/oidc/authorization`
   6. Token endpoint: 	`https://auth.example.com/api/oidc/token`
   7. User endpoint: 	`https://auth.example.com/api/oidc/userinfo`
   8. Scopes: 	`openid profile email`
   9. Identifier: 	`preferred_username`
   10. Display Name: 	`given_name`
   11. Email: 	`email`


### Authelia

The following YAML configuration is an example __Authelia__
[client configuration](../../../configuration/identity-providers/openid-connect/clients.md) for use with [Memos]
which will operate with the above example:

```yaml
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'memos'
        client_name: 'Memos'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://memos.example.com/auth/callback'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
        grant_types:
          - 'authorization_code'
        token_endpoint_auth_method: 'client_secret_post'
        userinfo_signed_response_alg: 'none'
```

[Authelia]: https://www.authelia.com
[Memos]: https://github.com/usememos/memos
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
