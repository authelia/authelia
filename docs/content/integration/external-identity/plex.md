---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Plex"
description: "Integrating Plex as an external identity provider users can sign in to Authelia with."
summary: ""
date: 2026-09-13T18:00:00+10:00
draft: false
images: []
weight: 655
toc: true
seo:
  title: "Plex | External Identity | Integration"
  description: "Step-by-step guide to letting users sign in to Authelia with their Plex account."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Assumptions

This example makes the following assumptions:

- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Provider ID:** `plex`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Plex

There is nothing to register with [Plex]. Plex signs users in with the same PIN flow Plex apps use, so there is no
client secret and no Redirect URI to register.

Instead, Authelia identifies itself to Plex with a client identifier, which Plex treats as a device. Generate a random
value once and keep it: Plex treats every different value as a different device.

```shell
authelia crypto rand --length 32 --charset alphanumeric
```

### Authelia

The following YAML configuration is an example **Authelia** [external identity configuration] for use with [Plex],
where the `client_id` is the value generated above:

```yaml {title="configuration.yml"}
authentication_backend:
  external_identity:
    providers:
      - id: 'plex'
        type: 'plex'
        name: 'Plex'
        client_id: 'Wk3n9Qd7xLp2Rv8Hs5Tj1Ym6Bc4Fg0Za'
```

## Linking Accounts

Each user links their Plex account from _Settings_ → _Linked Accounts_ once. See
[Linking Accounts](introduction.md#linking-accounts).

Accounts are identified by the following attributes of the account returned by Plex's user endpoint. Only the issuer
and the subject identify the account. The other attributes are only displayed, so a user can change their Plex
username or email address without affecting the link.

| Value        | Attribute                                |
| :----------- | :--------------------------------------- |
| Issuer       | always `https://plex.tv`                 |
| Subject      | `uuid`                                   |
| Username     | `username`                               |
| Display Name | `title`, or `username` when it is absent |
| Email        | `email`                                  |

When a user signs in, Plex asks them to approve a device named
`Authelia (https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}})`,
shown in the language they chose in Authelia. The device then appears in the authorized devices of their Plex account,
where they can remove it at any time.

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
Any Plex account can be linked, whether or not it has access to a Plex Media Server. A Plex account only ever signs in
to the Authelia account it was linked to by the owner of that Authelia account.
{{< /callout >}}

## Authentication Level

A sign in with Plex is treated as a password sign in with Authelia: the session adopts the `pwd` and `kba`
Authentication Method References, exactly as signing in with a username and password does, because Plex asserts
nothing about how the user authenticated. The session satisfies the `one_factor` policy, and a resource with the
`two_factor` policy still asks the user for a second factor with Authelia.

To decide the Authentication Method References yourself, configure the
[default](../../configuration/first-factor/external-identity.md#default) values for the provider. They are adopted in
place of the password values. For example, the following treats a sign in with Plex as a multi-factor sign in, which
satisfies the `two_factor` policy on Plex's authentication alone:

```yaml {title="configuration.yml"}
authentication_backend:
  external_identity:
    providers:
      - id: 'plex'
        authentication_methods_reference:
          default:
            - 'pwd'
            - 'otp'
```

See [Authentication Level](../../configuration/first-factor/external-identity.md#authentication-level) for how these
values interact with access control.

## See Also

- [External Identity configuration reference][external identity configuration]
- [Plex Authenticating with Plex Documentation](https://forums.plex.tv/t/authenticating-with-plex/609370)

[Plex]: https://www.plex.tv
[external identity configuration]: ../../configuration/first-factor/external-identity.md#plex
