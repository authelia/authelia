---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Discord"
description: "Integrating Discord as an external identity provider users can sign in to Authelia with."
summary: ""
date: 2026-09-13T18:00:00+10:00
draft: false
images: []
weight: 652
toc: true
seo:
  title: "Discord | External Identity | Integration"
  description: "Step-by-step guide to letting users sign in to Authelia with their Discord account."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Assumptions

This example makes the following assumptions:

- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Provider ID:** `discord`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Discord

To register Authelia with [Discord]:

1. Open the [Discord Developer Portal](https://discord.com/developers/applications) and select _New Application_.
2. Enter a name users will recognize, such as `Authelia`, accept the terms, and select _Create_.
3. Select _OAuth2_ from the menu.
4. Copy the _Client ID_.
5. Select _Reset Secret_ and copy the _Client Secret_. Discord only shows it once.
6. Under _Redirects_ select _Add Redirect_, enter the following Redirect URI, and select _Save Changes_:

   ```text
   https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/firstfactor/external-identity/discord/callback
   ```

### Authelia

The following YAML configuration is an example **Authelia** [external identity configuration] for use with [Discord],
where the `client_id` and `client_secret` are the values copied from the Discord Developer Portal:

```yaml {title="configuration.yml"}
authentication_backend:
  external_identity:
    providers:
      - id: 'discord'
        type: 'discord'
        name: 'Discord'
        client_id: '123456789012345678'
        client_secret: 'insecure_secret'
        scopes:
          - 'identify'
          - 'email'
```

The `email` scope is optional. Without it, or when Discord has not verified the address, the linked account is shown
without an email address.

## Linking Accounts

Each user links their Discord account from _Settings_ → _Linked Accounts_ once. See
[Linking Accounts](introduction.md#linking-accounts).

Accounts are identified by the following attributes of the user returned by Discord's current user endpoint.
Only the issuer and the subject identify the account. The other attributes are only displayed, so a user can change
their Discord username or email address without affecting the link.

| Value        | Attribute                                      |
| :----------- | :--------------------------------------------- |
| Issuer       | always `https://discord.com`                   |
| Subject      | `id`                                           |
| Username     | `username`                                     |
| Display Name | `global_name`, or `username` when it is absent |
| Email        | `email`, only when `verified` is `true`        |

## Authentication Level

A sign in with Discord is treated as a password sign in with Authelia: the session adopts the `pwd` and `kba`
Authentication Method References, exactly as signing in with a username and password does, because Discord asserts
nothing about how the user authenticated. The session satisfies the `one_factor` policy, and a resource with the
`two_factor` policy still asks the user for a second factor with Authelia.

To decide the Authentication Method References yourself, configure the
[default](../../configuration/first-factor/external-identity.md#default) values for the provider. They are adopted in
place of the password values. For example, the following treats a sign in with Discord as a multi-factor sign in, which
satisfies the `two_factor` policy on Discord's authentication alone:

```yaml {title="configuration.yml"}
authentication_backend:
  external_identity:
    providers:
      - id: 'discord'
        authentication_methods_reference:
          default:
            - 'pwd'
            - 'otp'
```

See [Authentication Level](../../configuration/first-factor/external-identity.md#authentication-level) for how these
values interact with access control.

## See Also

- [External Identity configuration reference][external identity configuration]
- [Discord OAuth2 Documentation](https://discord.com/developers/docs/topics/oauth2)

[Discord]: https://discord.com
[external identity configuration]: ../../configuration/first-factor/external-identity.md#discord
