---
title: "Seafile"
description: "Integrating Seafile with the Authelia OpenID Connect Provider."
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
  * [v4.36.9](https://github.com/authelia/authelia/releases/tag/v4.36.9)
* [Seafile] Server
  * [9.0.9](https://manual.seafile.com/changelog/server-changelog/#909-2022-09-22)

## Before You Begin

You are required to utilize a unique client id and a unique and random client secret for all [OpenID Connect] relying
parties. You should not use the client secret in this example, you should randomly generate one yourself. You may also
choose to utilize a different client id, it's completely up to you.

This example makes the following assumptions:

* __Application Root URL:__ `https://seafile.example.com`
* __Authelia Root URL:__ `https://auth.example.com`
* __Client ID:__ `seafile`
* __Client Secret:__ `seafile_client_secret`

*__Important Note:__ [Seafile] only uses email to identify unique user accounts[^1]. When using the
[File First Factor](../../../configuration/first-factor/file.md) there may be multiple Authelia users with the same
email, resulting in them being assigned the same [Seafile] account. This issue could be mitigated by tuning
the `OAUTH_ATTRIBUTE_MAP` or modifying `seahub/seahub/oauth/views.py` in your [Seafile] installation[^2].*

## Configuration

### Application

To configure [Seafile] to utilize Authelia as an [OpenID Connect] Provider:

1. Install the `requests_oauthlib` pip package which is required but not installed with [Seafile] by default[^3].

2. Edit your [Seafile] `seahub_settings.py` configuration file and add configure the following:

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
]
OAUTH_ATTRIBUTE_MAP = {
    "email": (True, "email"),
    "name": (False, "name"),
    "id": (False, "not used"),
}
```

Remember to restart [Seafile] so your configuration changes take effect.

### Authelia

The following YAML configuration is an example __Authelia__
[client configuration](../../../configuration/identity-providers/open-id-connect.md#clients) for use with [Seafile]
which will operate with the above example:

```yaml
- id: seafile
  description: Seafile
  secret: seafile_client_secret
  public: false
  authorization_policy: two_factor
  redirect_uris:
    - https://seafile.example.com/oauth/callback/
  scopes:
    - openid
    - profile
    - email
  userinfo_signing_algorithm: none
```

## See Also

* [Seafile OAuth Authentication Documentation](https://manual.seafile.com/deploy/oauth/)

[^1]:https://forum.seafile.com/t/oauth-question-error-with-sso/5481/2
[^2]:https://forum.seafile.com/t/oauth-question-error-with-sso/5481/3
[^3]:https://manual.seafile.com/deploy/oauth/#oauth

[Authelia]: https://www.authelia.com
[Seafile]: https://www.seafile.com/
[OpenID Connect]: ../../openid-connect/introduction.md
