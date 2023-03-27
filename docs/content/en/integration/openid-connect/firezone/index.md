---
title: "Firezone"
description: "Integrating Firezone with the Authelia OpenID Connect Provider."
lead: ""
date: 2023-03-25T13:07:02+10:00
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
* [Firezone]
  * [0.7.25](https://github.com/firezone/firezone/releases/tag/0.7.25)

## Before You Begin

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://firezone.example.com`
* __Authelia Root URL:__ `https://auth.example.com`
* __Client ID:__ `firezone`
* __Client Secret:__ `insecure_secret`
* __Authentication Name (Gitea):__ `authelia`:
    * This option determines the redirect URI in the format of
      `https://firezone.example.com/user/oauth2/<Authentication Name>/callback`.
      This means if you change this value you need to update the redirect URI.

## Configuration

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
   7. Redirect URI (optional): `https://firezone.example.com/auth/oidc/authelia/callback/`
   8. Auto-create users (checkbox): `true`

{{< figure src="firezone.png" alt="Firezone" width="500" >}}

Take a look at the [See Also](#see-also) section for the cheatsheets corresponding to the sections above for their
descriptions.

### Authelia

The following YAML configuration is an example __Authelia__
[client configuration](../../../configuration/identity-providers/open-id-connect.md#clients) for use with [Firezone] which
will operate with the above example:

```yaml
- id: firezone
  description: Firezone
  secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
  public: false
  authorization_policy: two_factor
  redirect_uris:
    - https://firezone.example.com/auth/oidc/authelia/callback
  scopes:
    - openid
    - email
    - profile
  userinfo_signing_algorithm: none
```

## See Also

- [Firezone OIDC documentation](https://www.firezone.dev/docs/authenticate/oidc/)

[Authelia]: https://www.authelia.com
[Firezone]: https://www.firezone.dev
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
