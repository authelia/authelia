---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "GitLab"
description: "Integrating GitLab as an OpenID Connect 1.0 external identity provider users can sign in to Authelia with."
summary: ""
date: 2026-09-13T18:00:00+10:00
draft: false
images: []
weight: 654
toc: true
seo:
  title: "GitLab | External Identity | Integration"
  description: "Step-by-step guide to letting users sign in to Authelia with their GitLab account using OpenID Connect 1.0."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

[GitLab] is an [OpenID Connect 1.0] Provider, so it is configured as an `openid_connect` provider. The same steps apply
to any other [OpenID Connect 1.0] Provider, with the values of that provider in place of GitLab's.

## Assumptions

This example makes the following assumptions:

- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **GitLab URL:** `https://gitlab.com`
- **Provider ID:** `gitlab`

For a self-managed GitLab instance, replace `https://gitlab.com` with the external URL of the instance, e.g.
`https://gitlab.{{< sitevar name="domain" nojs="example.com" >}}`.

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### GitLab

To register Authelia with [GitLab]:

1. Sign in to GitLab, select your avatar, select _Edit profile_, then select _Applications_. To register the
   application for a group instead, open the group and select _Settings_ then _Applications_. On a self-managed
   instance an administrator can register an instance-wide application from the _Admin area_ under _Applications_.
2. Select _Add new application_ and configure the following options:
   - **Name:** a name users will recognize, such as `Authelia`.
   - **Redirect URI:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/firstfactor/external-identity/gitlab/callback`
   - **Confidential:** enabled.
   - **Scopes:** `openid`, `profile`, and `email`.
3. Select _Save application_.
4. Copy the _Application ID_ and the _Secret_. GitLab only shows the secret once.

### Authelia

The following YAML configuration is an example **Authelia** [external identity configuration] for use with [GitLab],
where the `client_id` is the _Application ID_ and the `client_secret` is the _Secret_ copied from GitLab:

```yaml {title="configuration.yml"}
authentication_backend:
  external_identity:
    providers:
      - id: 'gitlab'
        type: 'openid_connect'
        name: 'GitLab'
        issuer: 'https://gitlab.com'
        client_id: 'b2f4c6d8e0a1b3c5d7e9f1a3b5c7d9e1f3a5b7c9d1e3f5a7b9c1d3e5f7a9b1c3'
        client_secret: 'insecure_secret'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
        token_endpoint_auth_method: 'client_secret_basic'
```

The [issuer](../../configuration/first-factor/external-identity.md#issuer) must be exactly the URL of the GitLab
instance without a trailing slash, as that is the issuer GitLab identifies itself with. Authelia discovers every
endpoint it needs from GitLab's discovery document, so no endpoints need to be configured.

## Linking Accounts

Each user links their GitLab account from _Settings_ → _Linked Accounts_ once. See
[Linking Accounts](introduction.md#linking-accounts).

Accounts are identified by the following claims. The Issuer and Subject are the claims of the ID Token, and the
Username, Display Name, and Email are the claims of the UserInfo response, taken from the ID Token only when the
UserInfo response does not include them. Only the issuer and the subject identify the account.
The other claims are only displayed, so a user can change their GitLab username or email address without affecting
the link.

| Value        | Claim                                                                 |
| :----------- | :-------------------------------------------------------------------- |
| Issuer       | `iss`, which must be the configured issuer, i.e. `https://gitlab.com` |
| Subject      | `sub`, which is the GitLab user ID                                    |
| Username     | `preferred_username`, which is the GitLab username                    |
| Display Name | `name`                                                                |
| Email        | `email`, only when the `email` scope is granted                       |

## Authentication Level

A sign in with GitLab satisfies the `one_factor` policy. Unlike the `discord`, `github`, and `plex` provider types, an
`openid_connect` provider adopts no Authentication Method References by default: it only adopts the values it asserts
in the `amr` claim when [trust](../../configuration/first-factor/external-identity.md#trust) is enabled, and GitLab does
not include an `amr` claim in its ID Tokens at the time of writing.

To have a sign in with GitLab treated as a password sign in with Authelia, configure the
[default](../../configuration/first-factor/external-identity.md#default) values:

```yaml {title="configuration.yml"}
authentication_backend:
  external_identity:
    providers:
      - id: 'gitlab'
        authentication_methods_reference:
          default:
            - 'pwd'
            - 'kba'
```

See [Authentication Level](../../configuration/first-factor/external-identity.md#authentication-level) for how these
values interact with access control.

## See Also

- [External Identity configuration reference][external identity configuration]
- [GitLab OpenID Connect Identity Provider Documentation](https://docs.gitlab.com/integration/openid_connect_provider/)
- [GitLab OAuth 2.0 Identity Provider Documentation](https://docs.gitlab.com/integration/oauth_provider/)

[GitLab]: https://about.gitlab.com
[OpenID Connect 1.0]: https://openid.net/specs/openid-connect-core-1_0.html
[external identity configuration]: ../../configuration/first-factor/external-identity.md
