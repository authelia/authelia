---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "GitHub"
description: "Integrating GitHub as an external identity provider users can sign in to Authelia with."
summary: ""
date: 2026-09-13T18:00:00+10:00
draft: false
images: []
weight: 653
toc: true
seo:
  title: "GitHub | External Identity | Integration"
  description: "Step-by-step guide to letting users sign in to Authelia with their GitHub account."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Assumptions

This example makes the following assumptions:

- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Provider ID:** `github`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### GitHub

To register Authelia with [GitHub] as an OAuth App:

1. Open [GitHub](https://github.com/settings/developers), and select _OAuth Apps_ then _New OAuth App_. To register
   the app for an organization instead, open the organization's _Settings_ and select _Developer settings_ then
   _OAuth Apps_.
2. Configure the following options, and select _Register application_:
   - **Application name:** a name users will recognize, such as `Authelia`.
   - **Homepage URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
   - **Authorization callback URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/firstfactor/external-identity/github/callback`
3. Copy the _Client ID_.
4. Select _Generate a new client secret_ and copy the client secret. GitHub only shows it once.

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
An OAuth App has a single callback URL. If the portal is reached over more than one name, register a GitHub App
instead, which allows several callback URLs, and grant it the _Email addresses_ account permission with read-only access
so the email address of the linked account can be shown.
{{< /callout >}}

### Authelia

The following YAML configuration is an example **Authelia** [external identity configuration] for use with [GitHub],
where the `client_id` and `client_secret` are the values copied from GitHub:

```yaml {title="configuration.yml"}
authentication_backend:
  external_identity:
    providers:
      - id: 'github'
        type: 'github'
        name: 'GitHub'
        client_id: 'Ov23liAbCdEfGhIjKlMn'
        client_secret: 'insecure_secret'
        scopes:
          - 'read:user'
          - 'user:email'
```

The `user:email` scope is optional. Without it the linked account is shown without an email address. With it, the
primary email address of the GitHub account is shown when GitHub has verified it.

## Linking Accounts

Each user links their GitHub account from _Settings_ → _Linked Accounts_ once. See
[Linking Accounts](introduction.md#linking-accounts).

Accounts are identified by the following attributes of the user returned by GitHub's authenticated user endpoint.
Only the issuer and the subject identify the account. The other attributes are only displayed, so a user can change
their GitHub username or email address without affecting the link.

| Value        | Attribute                                                                                 |
| :----------- | :---------------------------------------------------------------------------------------- |
| Issuer       | always `https://github.com`                                                               |
| Subject      | `id`, recorded as its decimal representation                                              |
| Username     | `login`                                                                                   |
| Display Name | `name`, or `login` when it is absent                                                      |
| Email        | `email` of the entry from the user emails endpoint which is both `primary` and `verified` |

## Authentication Level

A sign in with GitHub is treated as a password sign in with Authelia: the session adopts the `pwd` and `kba`
Authentication Method References, exactly as signing in with a username and password does, because GitHub asserts
nothing about how the user authenticated. The session satisfies the `one_factor` policy, and a resource with the
`two_factor` policy still asks the user for a second factor with Authelia.

To decide the Authentication Method References yourself, configure the
[default](../../configuration/first-factor/external-identity.md#default) values for the provider. They are adopted in
place of the password values. For example, the following treats a sign in with GitHub as a multi-factor sign in, which
satisfies the `two_factor` policy on GitHub's authentication alone:

```yaml {title="configuration.yml"}
authentication_backend:
  external_identity:
    providers:
      - id: 'github'
        authentication_methods_reference:
          default:
            - 'pwd'
            - 'otp'
```

See [Authentication Level](../../configuration/first-factor/external-identity.md#authentication-level) for how these
values interact with access control.

## See Also

- [External Identity configuration reference][external identity configuration]
- [GitHub Creating an OAuth App Documentation](https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/creating-an-oauth-app)
- [GitHub Authorizing OAuth Apps Documentation](https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/authorizing-oauth-apps)

[GitHub]: https://github.com
[external identity configuration]: ../../configuration/first-factor/external-identity.md#github
