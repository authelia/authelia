---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "External Identity"
description: "An introduction into integrating external identity providers users can sign in to Authelia with"
summary: "An introduction into integrating external identity providers users can sign in to Authelia with."
date: 2026-09-13T18:00:00+10:00
draft: false
images: []
weight: 651
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

Authelia can sign users in with an account they hold at an external identity provider, such as an
[OpenID Connect 1.0] Provider like GitLab, Discord, GitHub, or Plex. This section contains guides for integrating
specific external identity providers. See the [External Identity configuration] for a reference of every option and
how signing in with an external identity provider works.

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
Authelia never creates users from an external identity provider. Every user who signs in this way must already exist in
the configured user provider, and must have linked their external account to their Authelia account.
{{< /callout >}}

## Guides

- [Discord](discord.md)
- [GitHub](github.md)
- [GitLab](gitlab.md)
- [Plex](plex.md)

## Common Steps

Every integration follows the same steps.

### Redirect URI

Most external identity providers require the Redirect URI to be registered with them. The Redirect URI is fixed, where
`<id>` is the [id](../../configuration/first-factor/external-identity.md#id) of the provider:

```text
https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/firstfactor/external-identity/<id>/callback
```

See the [Redirect URI](../../configuration/first-factor/external-identity.md#redirect-uri) reference for how it is
derived when the portal is served under a path or reached over more than one name.

Some of the values presented in these guides can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

### Linking Accounts

Once the provider is configured, each user links their external account to their Authelia account:

1. Sign in to Authelia.
2. Open _Settings_ and select _Linked Accounts_.
3. Select the provider, and sign in to it when prompted.
4. Accept the proposed link, and complete the identity verification Authelia asks for.

Once linked, the user can sign in by selecting the provider on the login page. A user who selects the provider on the
login page before linking is asked to sign in to Authelia first, and the link is proposed once they have. See
[Linking](../../configuration/first-factor/external-identity.md#linking) for the details.

[OpenID Connect 1.0]: https://openid.net/specs/openid-connect-core-1_0.html
[External Identity configuration]: ../../configuration/first-factor/external-identity.md
