---
title: "Kasm Workspaces"
description: "Integrating Kasm Workspaces with the Authelia OpenID Connect Provider."
lead: ""
date: 2023-04-25T23:07:05+2:00
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
* [Kasm Workspaces]
  * [1.13.0](https://kasmweb.com/docs/latest/release_notes/1.13.0.html)

## Before You Begin

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Authorization URL:__ `https://auth.example.com/api/oidc/authorization`
* __Token URL:__ `https://auth.example.com/api/oidc/token`
* __User Info URL:__ `https://auth.example.com/api/oidc/userinfo`
* __Scope (One per line):__ `openid profile groups email`
* __Username Attribute:__ `preferred_username`
* __Application Root URL:__ `https://kasm.example.com`
* __Client ID:__ `kasm`
* __Client Secret:__ `insecure_secret`

## Configuration

### Application

To configure [Kasm Workspaces] to utilize Authelia as an [OpenID Connect 1.0] Provider use the following configuration:

![Kasm](kasm.png)

### Authelia

The following YAML configuration is an example __Authelia__
[client configuration](../../../configuration/identity-providers/open-id-connect.md#clients) for use with
[Kasm Workspaces] which will operate with the above example:

```yaml
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - id: kasm
        description: Kasm Workspaces
        secret: 'insecure-secret'
        public: false
        authorization_policy: two_factor
        redirect_uris:
          - https://kasm.example.com/api/oidc_callback
        scopes:
          - openid
          - profile
          - groups
          - email
        consent_mode: implicit
        userinfo_signing_algorithm: none
```

## See Also

* [Kasm Workspaces OpenID Connect Authentication Documentation](https://kasmweb.com/docs/latest/guide/oidc.html)

[Authelia]: https://www.authelia.com
[Kasm Workspaces]: https://kasmweb.com/
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
