---
title: "Synapse"
description: "Integrating Synapse with Authelia via OpenID Connect."
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
  * [v4.35.6](https://github.com/authelia/authelia/releases/tag/v4.35.6)
* [Synapse]
  * [v1.60.0](https://github.com/matrix-org/synapse/releases/tag/v1.60.0)

## Before You Begin

You are required to utilize a unique client id and a unique and random client secret for all [OpenID Connect] relying
parties. You should not use the client secret in this example, you should randomly generate one yourself. You may also
choose to utilize a different client id, it's completely up to you.

This example makes the following assumptions:

* __Application Root URL:__ `https://matrix.example.com/`
* __Authelia Root URL:__ `https://auth.example.com`
* __Client ID:__ `synapse`
* __Client Secret:__ `synapse_client_secret`

## Configuration

### Application

To configure [Synapse] to utilize Authelia as an [OpenID Connect] Provider:

1. Edit your [Synapse] `homeserver.yaml` configuration file and add configure the following:

```yaml
oidc_providers:
  - idp_id: synapse
    idp_name: "Authelia"
    issuer: "https://auth.example.com"
    client_id: "synapse"
    client_secret: "synapse_client_secret"
    allow_existing_users: true
    scopes: ["openid", "profile"]
    user_mapping_provider:
      config:
        localpart_template: "{{ openid.preferred_username }}"
```

### Authelia

The following YAML configuration is an example __Authelia__
[client configuration](../../../configuration/identity-providers/open-id-connect.md#clients) for use with [Synapse]
which will operate with the above example:

```yaml
- id: synapse
  secret: synapse_client_secret
  public: false
  authorization_policy: two_factor
  scopes:
    - openid
    - profile
  redirect_uris:
    - https://synapse.example.com/_synapse/client/oidc/callback
  userinfo_signing_algorithm: none
```

## See Also

* [Synapse OpenID Connect Authentication Documentation](https://matrix-org.github.io/synapse/latest/openid.html)

[Authelia]: https://www.authelia.com
[Synapse]: https://github.com/matrix-org/synapse
[OpenID Connect]: ../../openid-connect/introduction.md
