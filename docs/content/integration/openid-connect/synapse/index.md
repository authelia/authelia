---
title: "Synapse"
description: "Integrating Synapse with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
weight: 620
toc: true
community: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

* [Authelia]
  * [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
* [Synapse]
  * [v1.60.0](https://github.com/matrix-org/synapse/releases/tag/v1.60.0)

## Before You Begin

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://matrix.example.com/`
* __Authelia Root URL:__ `https://auth.example.com/`
* __Client ID:__ `synapse`
* __Client Secret:__ `insecure_secret`

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__
[client configuration](../../../configuration/identity-providers/openid-connect/clients.md) for use with [Synapse]
which will operate with the above example:

```yaml
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'synapse'
        client_name: 'Synapse'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://synapse.example.com/_synapse/client/oidc/callback'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
        userinfo_signed_response_alg: 'none'
```

### Application

To configure [Synapse] to utilize Authelia as an [OpenID Connect 1.0] Provider:

1. Edit your [Synapse] `homeserver.yaml` configuration file and add configure the following:

```yaml
oidc_providers:
  - idp_id: authelia
    idp_name: "Authelia"
    idp_icon: "mxc://authelia.com/cKlrTPsGvlpKxAYeHWJsdVHI"
    discover: true
    issuer: "https://auth.example.com"
    client_id: "synapse"
    client_secret: "insecure_secret"
    scopes: ["openid", "profile", "email"]
    allow_existing_users: true
    user_mapping_provider:
      config:
        subject_claim: "sub"
        localpart_template: "{{ user.preferred_username }}"
        display_name_template: "{{ user.name }}"
        email_template: "{{ user.email }}"
```

## See Also

* [Synapse OpenID Connect Authentication Documentation](https://matrix-org.github.io/synapse/latest/openid.html)

[Authelia]: https://www.authelia.com
[Synapse]: https://github.com/matrix-org/synapse
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
