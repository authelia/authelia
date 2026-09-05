---
title: "OpenID Connect 1.0 Relying Party"
description: "Authelia OpenID Connect 1.0 Relying Party Implementation"
summary: "The OpenID Connect 1.0 Relying Party role is a great addition to the existing authentication methods Authelia provides."
date: 2025-03-23T19:03:40+11:00
draft: false
images: []
weight: 330
toc: true
aliases:
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

See the [configuration documentation](../../configuration/first-factor/openid-connect.md) for how to configure it.

### Anchoring Implementation

{{< roadmap-status stage="complete" >}}

For the [OpenID Connect 1.0] Relying Party implementation to operate users must be able to anchor their Provider account
to their Relying Party account (i.e. Authelia).

Accounts are anchored on the pairwise `iss` and `sub` claims, i.e. the issuer and subject identifier respectively,
which are stored in the `user_openid_connect_links` table. Neither the `email` claim nor the `preferred_username` claim
participates in finding, matching, or creating a link.

Linking is a proposal and accept flow performed by a user who is already signed in. A user may start it from the
_Linked Accounts_ page in their settings, which offers a link action for each configured Provider they have not
already linked. When a validated external identity has no link it is held as a proposal against their session and they
are returned to that page to accept or decline it. Accepting a proposal, and removing an existing link, both require
an elevated session.

A flow started from the login page reaches the same place by a different route. Nothing describing the external
identity is carried across the sign in which must follow it: the validated identity is discarded and the user is
returned to the login page carrying only the Provider identifier, and the portal performs the flow again on their
behalf once the authentication their policy requires is complete. The Provider does not usually prompt a second time,
so this is not normally visible to the user.

See the
[FAW](../../integration/openid-connect/frequently-asked-questions.md#how-should-i-link-user-accounts-to-authelia-openid-connect-10-responses-in-the-application-im-designing)
for an explainer on why we've chosen these claims.

### Authorization Implementation

{{< roadmap-status stage="complete" >}}

The user is identified on an authorization attempt by the pairwise `iss` and `sub` claims, i.e. the issuer and subject
identifier respectively as previously [anchored](#anchoring-implementation).

This appears to the user as a button per configured Provider on the first factor login page, in the same place the
Login with Passkey button appears.

A sign in performed this way satisfies the `one_factor` policy. It does not satisfy the `two_factor` policy unless the
administrator opts in to trusting the Provider's asserted Authentication Method Reference values, as Authelia does not
observe the authentication the user performed at the Provider.

### Authentication Methods Reference Values

{{< roadmap-status stage="complete" >}}

The Authentication Methods Reference Values could theoretically be used to derive the effective authentication level of
the user in combination with [Granular Authorization](granular-authorization.md). This is implemented as an opt-in
setting per Provider which trusts the Provider's `amr` claim and merges its values into the session.

[OpenID Connect 1.0]: https://openid.net/connect/
