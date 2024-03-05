---
title: "PocketBase"
description: "Integrating PocketBase with the Authelia OpenID Connect 1.0 Provider."
lead: ""
date: 2024-03-03T10:46:12+02:00
draft: false
images: []
menu:
  integration:
    parent: "openid-connect"
weight: 620
toc: true
community: true
---

## Tested Versions

* [Authelia]
  * [v4.37.5](https://github.com/authelia/authelia/releases/tag/v4.37.5)
* [PocketBase](https://pocketbase.io/docs/authentication/#oauth2-integration)
  * 4.2.3

## Before You Begin

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://you-app-using-pocket-base.com/`
* __Authelia Root URL:__ `https://auth.example.com/`
* __Client ID:__ `upsnap`
* __Client Secret:__ `insecure_secret`

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__
[client configuration](../../../configuration/identity-providers/openid-connect/clients.md) for use with [Grafana]
which will operate with the above example:

```yaml
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
    - id: 'upsnap'
      description: 'UpSnap'
      secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
      public: false
      authorization_policy: 'two_factor'
      redirect_uris:
        - 'https://you-app-using-pocket-base.com/api/oauth2-redirect'
      scopes:
        - 'email'
        - 'groups'
        - 'openid'
        - 'profile'
      userinfo_signed_response_alg: 'none'
```

### Application

To configure [PocketBase] to utilize Authelia as an [OpenID Connect 1.0], please follow this options:

1. Connect to PocketBase admin view
2. On the left menu, go to "Settings" (3rd icon)
3. In "Authentication" section, go to "Auth providers"
4. Select the gear on "OpenID Connect (oidc)"
5. Configure:
   1. ClientID: `upsnap`
   2. Client secret: `insecure_secret`
   3. Display name: `Authelia`(or whatever you want)
   4. Auth URL: https://auth.example.com/api/oidc/authorization
   5. Token URL: https://auth.example.com/api/oidc/token
   6. User API URL: https://auth.example.com/api/oidc/userinfo
   7. You can leave "Support PKCE" checked
6. Save changes

## See Also

* [PocketBase OAuth Documentation](https://pocketbase.io/docs/authentication/#oauth2-integration)

[Authelia]: https://www.authelia.com
[PocketBase]: https://pocketbase.io
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
