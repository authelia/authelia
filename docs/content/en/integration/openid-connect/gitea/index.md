---
title: "Gitea"
description: "Integrating Gitea with the Authelia OpenID Connect Provider."
lead: ""
date: 2022-07-01T13:07:02+10:00
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
  * [v4.36.3](https://github.com/authelia/authelia/releases/tag/v4.36.3)
* [Gitea]
  * [1.17.0](https://github.com/go-gitea/gitea/releases/tag/v1.17.0)

## Before You Begin

You are required to utilize a unique client id and a unique and random client secret for all [OpenID Connect] relying
parties. You should not use the client secret in this example, you should randomly generate one yourself. You may also
choose to utilize a different client id, it's completely up to you.

This example makes the following assumptions:

* __Application Root URL:__ `https://gitea.example.com`
* __Authelia Root URL:__ `https://auth.example.com`
* __Client ID:__ `gitea`
* __Client Secret:__ `gitea_client_secret`

## Configuration

### Application

To configure [Gitea] to utilize Authelia as an [OpenID Connect] Provider:

1. Expand User Options
2. Visit Site Administration
3. Visit Authentication Sources
4. Visit Add Authentication Source
5. Configure:
   1. Authentication Name: `authelia`
   2. OAuth2 Provider: `OpenID Connect`
   3. Client ID (Key): `gitea`
   4. Client Secret: `gitea_client_secret`
   5. OpenID Connect Auto Discovery URL: `https://auth.example.com/.well-known/openid-configuration`

{{< figure src="gitea.png" alt="Gitea" width="300" >}}

#### Gitea App.ini
The following variables should be setup to enable 'auto' user creation -  
```
[openid]
ENABLE_OPENID_SIGNIN = false
ENABLE_OPENID_SIGNUP = true
WHITELISTED_URIS     = auth.domain.com

[service]
DISABLE_REGISTRATION                          = false
ALLOW_ONLY_EXTERNAL_REGISTRATION              = true
SHOW_REGISTRATION_BUTTON                      = false
```
* 'ENABLE_OPENID_SIGNIN' - shows an additional OpenID section on the login page. Setting this to false removes it - this allows all auth to be instigated by Authelia.
* 'ENABLE_OPENID_SIGNUP' - this allows the OIDC user to sign up with Gitea, with the user passed from your Authelia login session.
* 'WHITELISTED_URIS' - this restricts the login/OIDC source to be Authelia
* 'DISABLE_REGISTRATION' - this disables general user (non-OIDC) registration
* 'ALLOW_ONLY_EXTERNAL_REGISTRATION' - this allows registration via OIDC only
* 'SHOW_REGISTRATION_BUTTON' - This hides the sign up button so manual registrations via the home page are not possible

### Authelia

The following YAML configuration is an example __Authelia__
[client configuration](../../../configuration/identity-providers/open-id-connect.md#clients) for use with [Gitea] which
will operate with the above example:

```yaml
- id: gitea
  secret: gitea_client_secret
  public: false
  authorization_policy: two_factor
  scopes:
    - openid
    - email
    - profile
  redirect_uris:
    - https://gitea.example.com/user/oauth2/authelia/callback
  userinfo_signing_algorithm: none
```

[Authelia]: https://www.authelia.com
[Gitea]: https://gitea.io/
[OpenID Connect]: ../../openid-connect/introduction.md
