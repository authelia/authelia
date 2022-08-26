---
title: "Outline"
description: "Integrating Outline with the Authelia OpenID Connect Provider."
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
  * [v4.36.4](https://github.com/authelia/authelia/releases/tag/v4.36.4)
* [Outline]
  * 0.65.2

## Before You Begin

You are required to utilize a unique client id and a unique and random client secret for all [OpenID Connect] relying
parties. You should not use the client secret in this example, you should randomly generate one yourself. You may also
choose to utilize a different client id, it's completely up to you.

This example makes the following assumptions:

* __Application Root URL:__ `https://outline.example.com`
* __Authelia Root URL:__ `https://auth.example.com`
* __Client ID:__ `outline`
* __Client Secret:__ `outline_client_secret`

## Configuration

### Application

To configure [Outline] to utilize Authelia as an [OpenID Connect] Provider:

1. Configure the following environment options:
```text
URL=https://outline.example.com
FORCE_HTTPS=true

OIDC_CLIENT_ID=outline
OIDC_CLIENT_SECRET=outline_client_secret
OIDC_AUTH_URI=https://auth.example.com/api/oidc/authorization
OIDC_TOKEN_URI=https://auth.example.com/api/oidc/token
OIDC_USERINFO_URI=https://auth.example.com/api/oidc/userinfo
OIDC_USERNAME_CLAIM=preferred_username
OIDC_DISPLAY_NAME=Authelia
OIDC_SCOPES="openid profile email"
```

### Authelia

The following YAML configuration is an example __Authelia__
[client configuration](../../../configuration/identity-providers/open-id-connect.md#clients) for use with [Outline]
which will operate with the above example:

```yaml
- id: outline
  description: Outline
  secret: outline_client_secret
  public: false
  authorization_policy: two_factor
  redirect_uris:
    - https://outline.example.com/auth/oidc.callback
  scopes:
    - openid
    - profile
    - email
  userinfo_signing_algorithm: none
```

## See Also

* [Outline OpenID Connect Documentation](https://app.getoutline.com/share/770a97da-13e5-401e-9f8a-37949c19f97e/doc/oidc-8CPBm6uC0I)

[Authelia]: https://www.authelia.com
[Outline]: https://www.getoutline.com/
[OpenID Connect]: ../../openid-connect/introduction.md
