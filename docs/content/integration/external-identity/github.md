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
   - **Authorization callback URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/identity/github/callback`
3. Copy the _Client ID_.
4. Select _Generate a new client secret_ and copy the client secret. GitHub only shows it once.

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
An OAuth App supports [up to 10 callback URLs][GitHub Creating an OAuth App]. If the portal is reached over more than
one name, select _Add callback URL_ and register the
[Redirect URI](../../configuration/first-factor/external-identity.md#redirect-uri) for every name users reach it with.
GitHub requires an exact match unless wildcard matching is enabled for the callback URL, which Authelia does not need.

If a GitHub App is registered instead, grant it the _Email addresses_ account permission with read-only access so the
email address of the linked account can be shown.
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
primary email address of the GitHub account is shown when GitHub has verified it, and a sign in fails if the email
addresses can't be retrieved.

The sign in button shows the GitHub logo by default. It can be replaced with the
[logo_uri](../../configuration/first-factor/external-identity.md#logo_uri) option.

## Linking Accounts

Each user links their GitHub account from _Settings_ → _Linked Accounts_ once. See
[Linking Accounts](introduction.md#linking-accounts).

Accounts are identified by the following attributes of the user returned by GitHub's
[authenticated user endpoint][GitHub Authenticated User].
Only the issuer and the subject identify the account. The other attributes are only displayed, so a user can change
their GitHub username or email address without affecting the link.

| Value        | Attribute                                                                                                       |
| :----------- | :-------------------------------------------------------------------------------------------------------------- |
| Issuer       | always `https://github.com/login/oauth`                                                                         |
| Subject      | `id`, recorded as its decimal representation                                                                    |
| Username     | `login`                                                                                                         |
| Display Name | `name`, or `login` when it is absent                                                                            |
| Email        | `email` of the entry from the [user emails endpoint][GitHub User Emails] which is both `primary` and `verified` |

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

The fields of the GitHub user are available to
[claims_mapping](../../configuration/first-factor/external-identity.md#claims_mapping) as the `claims.github` variable.
The `read:user` scope is required for the private fields of the user, such as
`claims.github.two_factor_authentication`, so test for them with `has()`. For example:

```yaml {title="configuration.yml"}
authentication_backend:
  external_identity:
    providers:
      - id: 'github'
        claims_mapping:
          amr: 'has(claims.github.two_factor_authentication) && claims.github.two_factor_authentication ? amr + ["otp"] : amr'
```

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
The `claims.github.two_factor_authentication` value shows the account has two-factor authentication enabled, not that
the user performed it for this sign in, so this example accepts that a session which never performed a second factor
satisfies the `two_factor` policy.
{{< /callout >}}

### Claims

The `claims.github` variable holds the response of GitHub's [authenticated user endpoint][GitHub Authenticated User]
unmodified, with the `emails`, `organizations`, and `teams` lists added by Authelia. The following tables describe the
layout for the version of the GitHub REST API Authelia requests, `2026-03-10`. GitHub determines the content of these
responses, so see its documentation for the meaning of each field. A field which may be absent must be tested with
`has()` before it is used.

#### User

| Field                                                                                                                                                                                     | Type            | Availability                               |
| :---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | :-------------- | :----------------------------------------- |
| `id`, `public_repos`, `public_gists`, `followers`, `following`                                                                                                                            | number          | Always                                     |
| `login`, `node_id`, `type`, `created_at`, `updated_at`                                                                                                                                    | string          | Always                                     |
| `avatar_url`, `url`, `html_url`, `followers_url`, `following_url`, `gists_url`, `starred_url`, `subscriptions_url`, `organizations_url`, `repos_url`, `events_url`, `received_events_url` | string          | Always                                     |
| `site_admin`                                                                                                                                                                              | boolean         | Always                                     |
| `gravatar_id`, `name`, `company`, `blog`, `location`, `email`, `bio`                                                                                                                      | string or null  | Always                                     |
| `hireable`                                                                                                                                                                                | boolean or null | Always                                     |
| `user_view_type`                                                                                                                                                                          | string          | May be absent                              |
| `notification_email`, `twitter_username`                                                                                                                                                  | string or null  | May be absent                              |
| `plan`, which has the `name` string, and the `space`, `collaborators`, and `private_repos` numbers                                                                                        | object          | May be absent                              |
| `private_gists`, `total_private_repos`, `owned_private_repos`, `disk_usage`, `collaborators`                                                                                              | number          | Private field: `read:user` or `user` scope |
| `two_factor_authentication`                                                                                                                                                               | boolean         | Private field: `read:user` or `user` scope |
| `business_plus`                                                                                                                                                                           | boolean         | Private field which may be absent          |
| `ldap_dn`                                                                                                                                                                                 | string          | Private field which may be absent          |
| `emails`                                                                                                                                                                                  | list of objects | `user:email` or `user` scope               |
| `organizations`, `teams`                                                                                                                                                                  | list of objects | `read:org` scope                           |

The `email` field is the public email address of the profile. It is not the email address shown for the linked account,
which is taken from the `emails` list as described in [Linking Accounts](#linking-accounts).

#### Emails

Each entry of `claims.github.emails` is an entry of the [user emails endpoint][GitHub User Emails]:

| Field        | Type           | Availability |
| :----------- | :------------- | :----------- |
| `email`      | string         | Always       |
| `primary`    | boolean        | Always       |
| `verified`   | boolean        | Always       |
| `visibility` | string or null | Always       |

#### Organizations

Each entry of `claims.github.organizations` is an entry of the
[user organizations endpoint][GitHub User Organizations]:

| Field                                                                                                                              | Type           | Availability |
| :--------------------------------------------------------------------------------------------------------------------------------- | :------------- | :----------- |
| `id`                                                                                                                               | number         | Always       |
| `login`, `node_id`, `url`, `repos_url`, `events_url`, `hooks_url`, `issues_url`, `members_url`, `public_members_url`, `avatar_url` | string         | Always       |
| `description`                                                                                                                      | string or null | Always       |

#### Teams

Each entry of `claims.github.teams` is an entry of the [user teams endpoint][GitHub User Teams]:

| Field                                                                                                                             | Type           | Availability  |
| :-------------------------------------------------------------------------------------------------------------------------------- | :------------- | :------------ |
| `id`, `members_count`, `repos_count`                                                                                              | number         | Always        |
| `node_id`, `url`, `html_url`, `name`, `slug`, `permission`, `members_url`, `repositories_url`, `created_at`, `updated_at`, `type` | string         | Always        |
| `description`                                                                                                                     | string or null | Always        |
| `organization`, which includes the `login` string and `id` number                                                                 | object         | Always        |
| `privacy`, `notification_setting`, `ldap_dn`                                                                                      | string         | May be absent |
| `organization_id`, `enterprise_id`                                                                                                | number         | May be absent |
| `parent`                                                                                                                          | object or null | May be absent |

See the [user teams endpoint][GitHub User Teams] for the complete `organization` and `parent` objects.

### Email Addresses

When the `user:email` or `user` scope is included in
[scopes](../../configuration/first-factor/external-identity.md#scopes), all of the email addresses of the user are
retrieved as the `claims.github.emails` list, including the addresses GitHub has not verified, so check the `verified`
field of an address before relying on it. The list is absent without one of these scopes, so test for it with `has()`.
With one of these scopes a sign in fails if the list can't be retrieved in full, including when the user reduced the
scopes they granted. Email addresses can contain uppercase characters, so normalize them with `lowerAscii()` before
comparing them. For example, the following only allows users with a verified address at a domain to sign in:

```yaml {title="configuration.yml"}
authentication_backend:
  external_identity:
    providers:
      - id: 'github'
        claims_mapping:
          disabled: '!has(claims.github.emails) || !claims.github.emails.exists(e, e.verified && e.email.lowerAscii().endsWith("@example.com"))'
```

### Organizations and Teams

When the `read:org` scope is included in [scopes](../../configuration/first-factor/external-identity.md#scopes), the
organizations and teams of the user are retrieved as the `claims.github.organizations` and `claims.github.teams` lists,
and a sign in fails if they can't be retrieved in full. They're absent without the scope. For example, the following
only allows the members of a team to sign in:

```yaml {title="configuration.yml"}
authentication_backend:
  external_identity:
    providers:
      - id: 'github'
        scopes:
          - 'read:user'
          - 'user:email'
          - 'read:org'
        claims_mapping:
          disabled: '!claims.github.teams.exists(t, t.organization.login == "example" && t.slug == "staff")'
```

An organization may restrict third party application access, in which case the owner of the organization must approve
the application for that organization to be included.

See [Authentication Level](../../configuration/first-factor/external-identity.md#authentication-level) for how these
values interact with access control.

## See Also

- [External Identity configuration reference][external identity configuration]
- [GitHub Creating an OAuth App Documentation][GitHub Creating an OAuth App]
- [GitHub Authorizing OAuth Apps Documentation](https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/authorizing-oauth-apps)
- [GitHub Get the Authenticated User API Documentation][GitHub Authenticated User]
- [GitHub List Email Addresses API Documentation][GitHub User Emails]
- [GitHub List Organizations API Documentation][GitHub User Organizations]
- [GitHub List Teams API Documentation][GitHub User Teams]

[GitHub]: https://github.com
[GitHub Creating an OAuth App]: https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/creating-an-oauth-app
[GitHub Authenticated User]: https://docs.github.com/en/rest/users/users?apiVersion=2026-03-10#get-the-authenticated-user
[GitHub User Emails]: https://docs.github.com/en/rest/users/emails?apiVersion=2026-03-10#list-email-addresses-for-the-authenticated-user
[GitHub User Organizations]: https://docs.github.com/en/rest/orgs/orgs?apiVersion=2026-03-10#list-organizations-for-the-authenticated-user
[GitHub User Teams]: https://docs.github.com/en/rest/teams/teams?apiVersion=2026-03-10#list-teams-for-the-authenticated-user
[external identity configuration]: ../../configuration/first-factor/external-identity.md#github
