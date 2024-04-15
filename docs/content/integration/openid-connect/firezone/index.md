---
title: "Firezone"
description: "Integrating Firezone with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2023-03-28T20:29:13+11:00
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
* [Firezone]
  * [0.7.25](https://github.com/firezone/firezone/releases/tag/0.7.25)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://firezone.example.com/`
* __Authelia Root URL:__ `https://auth.example.com/`
* __Client ID:__ `firezone`
* __Client Secret:__ `insecure_secret`
* __Config ID (Firezone):__ `authelia`:
    * This option determines the redirect URI in the format of
      `https://firezone.example.com/auth/oidc/<Config ID>/callback`.
      This means if you change this value you need to update the redirect URI.

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Firezone] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'firezone'
        client_name: 'Firezone'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://firezone.example.com/auth/oidc/authelia/callback'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
        userinfo_signed_response_alg: 'none'
```

### Application

To configure [Firezone] to utilize Authelia as an [OpenID Connect 1.0] Provider:

1. Visit your [Firezone] site
2. Sign in as an admin
3. Visit:
    1. Settings
    2. Security
4. In the `Single Sign-On` section, click on the `Add OpenID Connect Provider` button
5. Configure:
   1. Config ID: `authelia`
   2. Label: `Authelia`
   3. Scope: `openid email profile`
   4. Client ID: `firezone`
   5. Client secret: `insecure_secret`
   6. Discovery Document URI: `https://auth.example.com/.well-known/openid-configuration`
   7. Redirect URI (optional): `https://firezone.example.com/auth/oidc/authelia/callback`
   8. Auto-create users (checkbox): `true`

{{< figure src="firezone.png" alt="Firezone" width="500" >}}

Take a look at the [See Also](#see-also) section for the cheatsheets corresponding to the sections above for their
descriptions.

## See Also

- [Firezone OIDC documentation](https://www.firezone.dev/docs/authenticate/oidc/)

[Authelia]: https://www.authelia.com
[Firezone]: https://www.firezone.dev
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
