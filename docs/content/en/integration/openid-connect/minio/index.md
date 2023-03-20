---
title: "Minio"
description: "Integrating Minio with the Authelia OpenID Connect Provider."
lead: ""
date: 2022-06-15T17:51:47+10:00
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
* [Minio]
  * [2023-03-13T19:46:17Z](https://github.com/minio/minio/releases/tag/RELEASE.2023-03-13T19-46-17Z)

## Before You Begin

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://minio.example.com`
* __Authelia Root URL:__ `https://auth.example.com`
* __Client ID:__ `minio`
* __Client Secret:__ `insecure_secret`

## Configuration

### Application

To configure [Minio] to utilize Authelia as an [OpenID Connect 1.0] Provider:

1. Login to Minio
2. On the left hand menu, go to `Identity`, then `OpenID`.
3. On the top right, click `Create Configuration`
4. On the screen that appears, enter the following information  - 

| Field | Info to enter |
| ----- | ----- | 
| Name | `authelia` |
| Config URL | `https://auth.example.com/.well-known/openid-configuration` |
| Client ID | `minio` |
| Client Secret | `insecure_secret` |
| Claim Name | <blank> |
| Display Name | `Authelia` |
| Claim Prefix | `authelia` |
| Scopes | `openid,profile,email` |
| Redirect URI | `https://minio.example.com/oauth_callback` |
| Role Policy | `readonly` |
| Claim User Info | Disabled |
| Redirect URI Dynamic | Disabled |

5. Press `Save` at the bottom
6. At the top, accept the offer of a server restart
7. When the login screen appears again, click the `Other Authentication Methods` open, then select `Authelia` from the list.
8. Login!

  
### Authelia

The following YAML configuration is an example __Authelia__
[client configuration](../../../configuration/identity-providers/open-id-connect.md#clients) for use with [Minio]
which will operate with the above example:

```yaml
- id: minio
  description: Minio
  secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
  public: false
  authorization_policy: two_factor
  redirect_uris:
    - https://minio.example.com/apps/oidc_login/oidc
  scopes:
    - openid
    - profile
    - email
    - groups
  userinfo_signing_algorithm: none
```

## See Also

- [Minio](https://minio.com/) 
- [Minio OpenID Identiy Management](https://min.io/docs/minio/linux/reference/minio-server/minio-server.html#minio-server-envvar-external-identity-management-openid)
- [Authelia]: https://www.authelia.com
- [OpenID Connect 1.0]: ../../openid-connect/introduction.md 
