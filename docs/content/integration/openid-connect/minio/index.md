---
title: "MinIO"
description: "Integrating MinIO with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2023-03-21T11:21:23+11:00
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
* [MinIO]
  * [2024-01-05T22-17-24Z](https://github.com/minio/minio/releases/tag/RELEASE.2024-01-05T22-17-24Z)

## Before You Begin

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://minio.example.com/`
* __Authelia Root URL:__ `https://auth.example.com/`
* __Client ID:__ `minio`
* __Client Secret:__ `insecure_secret`

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__
[client configuration](../../../configuration/identity-providers/openid-connect/clients.md) for use with [MinIO]
which will operate with the above example:

```yaml
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'minio'
        client_name: 'MinIO'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://minio.example.com/oauth_callback'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
          - 'groups'
        userinfo_signed_response_alg: 'none'
```

### Application

To configure [MinIO] to utilize Authelia as an [OpenID Connect 1.0] Provider:

1. Login to [MinIO]
2. On the left hand menu, go to `Identity`, then `OpenID`
3. On the top right, click `Create Configuration`
4. On the screen that appears, enter the following information:
    - Name: `authelia`
    - Config URL: `https://auth.example.com/.well-known/openid-configuration`
    - Client ID: `minio`
    - Client Secret: `insecure_secret`
    - Claim Name: `groups`
    - Display Name: `Authelia`
    - Claim Prefix: Leave Empty
    - Scopes: `openid,profile,email,groups`
    - Redirect URI: `https://minio.example.com/oauth_callback`
    - Role Policy: Leave Empty
    - Claim User Info: Disabled
    - Redirect URI Dynamic: Disabled
5. Press `Save` at the bottom
6. Accept the offer of a server restart at the top
    - Refresh the page and sign out if not done so automatically
7. Add a [default policy](https://min.io/docs/minio/linux/administration/identity-access-management/policy-based-access-control.html#built-in-policies) to your user groups in Authelia
8. When the login screen appears again, click the `Other Authentication Methods` open, then select `Authelia` from the list.
9. Login

## See Also

- [MinIO OpenID Identity Management](https://min.io/docs/minio/linux/reference/minio-server/minio-server.html#minio-server-envvar-external-identity-management-openid)

[MinIO]: https://minio.com/
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
