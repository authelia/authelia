---
title: "Seafile"
description: "Integrating Seafile with Authelia via OpenID Connect."
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
  * [v4.35.5](https://github.com/authelia/authelia/releases/tag/v4.35.5)
* [Seafile] Server
  * 9.0.4

## Before You Begin

You are required to utilize a unique client id and a unique and random client secret for all [OpenID Connect] relying
parties. You should not use the client secret in this example, you should randomly generate one yourself. You may also
choose to utilize a different client id, it's completely up to you.

This example makes the following assumptions:

* __Application Root URL:__ `https://seafile.example.com`
* __Authelia Root URL:__ `https://auth.example.com`
* __Client ID:__ `seafile`
* __Client Secret:__ `seafile_client_secret`

## Configuration

### Application

To configure [Seafile] to utilize Authelia as an [OpenID Connect] Provider:

1. Edit your [Seafile] `seahub_settings.py` configuration file and add configure the following:

```python
ENABLE_OAUTH = True
OAUTH_ENABLE_INSECURE_TRANSPORT = False
OAUTH_CLIENT_ID = "seafile"
OAUTH_CLIENT_SECRET = "seafile_client_secret"
OAUTH_REDIRECT_URL = 'https://seafile.example.com/oauth/callback/'
OAUTH_PROVIDER_DOMAIN = 'auth.example.com'
OAUTH_AUTHORIZATION_URL = 'https://auth.example.com/api/oidc/authorization'
OAUTH_TOKEN_URL = 'https://auth.example.com/api/oidc/token'
OAUTH_USER_INFO_URL = 'https://auth.example.com/api/oidc/userinfo'
OAUTH_SCOPE = [
  "openid",
  "profile",
  "email",
  "groups",
]
OAUTH_ATTRIBUTE_MAP = {
    "id": (True, "preferred_username"),
    "name": (False, "name"),
    "email": (False, "email"),
}
```

### Authelia

The following YAML configuration is an example __Authelia__
[client configuration](../../../configuration/identity-providers/open-id-connect.md#clients) for use with [Seafile]
which will operate with the above example:

```yaml
- id: seafile
  secret: seafile_client_secret
  public: false
  authorization_policy: two_factor
  scopes:
    - openid
    - profile
    - email
  redirect_uris:
    - https://seafile.example.com/oauth/callback/
  userinfo_signing_algorithm: none
```

## See Also

* [Seafile OAuth Authentication Documentation](https://manual.seafile.com/deploy/oauth/)

[Authelia]: https://www.authelia.com
[Seafile]: https://www.seafile.com/
[OpenID Connect]: ../../openid-connect/introduction.md
