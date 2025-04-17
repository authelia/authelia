---
title: "OpenID Connect 1.0 Relying Party"
description: "Authelia OpenID Connect 1.0 Relying Party Implementation"
summary: "The OpenID Connect 1.0 Relying Party role is a great addition to the existing authentication methods Authelia provides."
date: 2025-03-23T19:03:40+11:00
draft: false
images: []
weight: 220
toc: true
  - /r/openid-connect-rp
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

The [OpenID Connect 1.0] Relying Party role is often described as the client. Effectively the relying party relies on an
[OpenID Connect 1.0] Provider for authentication and authorization information, delegating the validation to the
Provider.

### Anchoring Implementation

{{< roadmap-status stage="needs-design" >}}

For the [OpenID Connect 1.0] Relying Party implementation to operate users must be able to anchor their Provider account
to their Relying Party account (i.e. Authelia). To do this the likely implementation will require the user to already be
authenticated to a level that satisfies the `two_factor` policy and for them to click a link to onboard them.

The accounts will then likely be linked using the pairwise `iss` and `sub` claims, i.e. issuer and subject identifier
respectively.

See the
[FAW](../../integration/openid-connect/frequently-asked-questions.md#how-should-i-link-user-accounts-to-authelia-openid-connect-10-responses-in-the-application-im-designing)
for an explainer on why we've chosen these claims.

### Authorization Implementation

{{< roadmap-status stage="needs-design" >}}

To identify the user on an authorization attempt will likely be completed using the pairwise `iss` and `sub` claims,
i.e. issuer and subject identifier respectively as previously [anchored](#anchoring-implementation).

This will very likely appear to the user similar to how the Login with Passkey button appears. There may be some
customization provided as to how the layout occurs, i.e. either individually, in a menu, or a combination of the two.

### Authentication Methods Reference Values

{{< roadmap-status stage="needs-design" >}}

The Authentication Methods Reference Values could theoretically be used to derive the effective authentication level of
the user in combination with [Granular Authorization](../active/granular-authorization.md). We will likely implement
this with a opt-in setting to trust the Provider.

[OpenID Connect 1.0]: https://openid.net/connect/
