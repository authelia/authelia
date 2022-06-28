---
title: "Harbor"
description: "Integrating Harbor with Authelia via OpenID Connect."
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
* [Harbor]
  * 2.5.0

## Before You Begin

You are required to utilize a unique client id and a unique and random client secret for all [OpenID Connect] relying
parties. You should not use the client secret in this example, you should randomly generate one yourself. You may also
choose to utilize a different client id, it's completely up to you.

This example makes the following assumptions:

* __Application Root URL:__ `https://harbor.example.com`
* __Authelia Root URL:__ `https://auth.example.com`
* __Client ID:__ `harbor`
* __Client Secret:__ `harbor_client_secret`

## Configuration

### Application

To configure [Harbor] to utilize Authelia as an [OpenID Connect] Provider:

1. Visit Administration
2. Visit Configuration
3. Visit Authentication
4. Select `OIDC` from the `Auth Mode` drop down
5. Enter the following information:
   1. OIDC Provider Name: `Authelia`
   2. OIDC Provider Endpoint: `https://auth.example.com`
   3. OIDC Client ID: `harbor`
   4. OIDC Client Secret: `harbor_client_secret`
   5. Group Claim Name: `groups`
   6. OIDC Scope: `openid,profile,email,groups`
   7. For OIDC Admin Group you can specify a group name that matches your authentication backend.
   8. Ensure `Verify Certificate` is checked.
   9. Ensure `Automatic onboarding` is checked if you want users to be created by default.
   10. Username Claim: `preferred_username`
6. Click `Test OIDC Server`
7. Click `Save`

### Authelia

The following YAML configuration is an example __Authelia__
[client configuration](../../../configuration/identity-providers/open-id-connect.md#clients) for use with [Harbor]
which will operate with the above example:

```yaml
- id: harbor
  secret: harbor_client_secret
  public: false
  authorization_policy: two_factor
  scopes:
    - openid
    - profile
    - groups
    - email
  redirect_uris:
    - https://vault.example.com/oidc/callback
    - https://vault.example.com/ui/vault/auth/oidc/oidc/callback
  userinfo_signing_algorithm: none
```

## See Also

* [Harbor OpenID Connect Provider Documentation](https://goharbor.io/docs/2.5.0/administration/configure-authentication/oidc-auth/)

[Authelia]: https://www.authelia.com
[Harbor]: https://goharbor.io/
[OpenID Connect]: ../../openid-connect/introduction.md
