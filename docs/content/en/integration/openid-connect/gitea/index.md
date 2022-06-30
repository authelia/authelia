---
title: "Gitea"
description: "Integrating Gitea with the Authelia OpenID Connect Provider."
lead: ""
date: 2022-06-30T12:35:54+10:00
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
  * [v4.36.1](https://github.com/authelia/authelia/releases/tag/v4.36.1)
* [Gitea]
  * TBA

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

1. Steps to come.

{{< figure src="gitea.png" alt="Gitea" width="736" style="padding-right: 10px" >}}

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
    - https://gitea.example.com
  userinfo_signing_algorithm: none
```

[Authelia]: https://www.authelia.com
[Gitea]: https://gitea.io/
[OpenID Connect]: ../../openid-connect/introduction.md
