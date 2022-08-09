---
title: "Argo CD"
description: "Integrating Argo CD with the Authelia OpenID Connect Provider."
lead: ""
date: 2022-07-13T03:42:47+10:00
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
  * [v4.36.2](https://github.com/authelia/authelia/releases/tag/v4.36.2)
* [Argo CD]
  * v2.4.5

## Before You Begin

You are required to utilize a unique client id and a unique and random client secret for all [OpenID Connect] relying
parties. You should not use the client secret in this example, you should randomly generate one yourself. You may also
choose to utilize a different client id, it's completely up to you.

This example makes the following assumptions:

* __Application Root URL:__ `https://argocd.example.com`
* __Authelia Root URL:__ `https://auth.example.com`
* __Client ID:__ `argocd`
* __Client Secret:__ `argocd_client_secret`
* __CLI Client ID:__ `argocd-cli`

## Configuration

### Application

To configure [Argo CD] to utilize Authelia as an [OpenID Connect] Provider use the following configuration:

```yaml
name: Authelia
issuer: https://auth.example.com
clientID: argocd
clientSecret: argocd_client_secret
cliClientID: argocd-cli
requestedScopes:
  - openid
  - profile
  - email
  - groups
```

### Authelia

The following YAML configuration is an example __Authelia__
[client configuration](../../../configuration/identity-providers/open-id-connect.md#clients) for use with [Argo CD]
which will operate with the above example:

```yaml
- id: argocd
  description: Argo CD
  redirect_uris:
    - https://argocd.example.com/auth/callback
  scopes:
    - openid
    - groups
    - email
    - profile
  secret: argocd_client_secret
  userinfo_signing_algorithm: none
- id: argocd-cli
  description: Argo CD (CLI)
  public: true
  redirect_uris:
    - http://localhost:8085/auth/callback
  scopes:
    - openid
    - groups
    - email
    - profile
    - offline_access
  userinfo_signing_algorithm: none
```

## See Also

* [Argo CD OpenID Connect Documentation](https://argo-cd.readthedocs.io/en/stable/operator-manual/user-management/#existing-oidc-provider)

[Authelia]: https://www.authelia.com
[Argo CD]: https://argo-cd.readthedocs.io/en/stable/
[OpenID Connect]: ../../openid-connect/introduction.md




