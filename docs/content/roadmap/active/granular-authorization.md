---
title: "Granular Authorization"
description: "Authelia Granular Authorization Implementation"
summary: "Implementation of a Granular Authorization framework will make the Authorization experience more tailored to complex requirements."
date: 2025-03-23T19:03:40+11:00
draft: false
images: []
weight: 325
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

While we already have a rich authentication and authorization experience, we plan to drastically improve the ability
for administrators to customize this. We plan to do this leveraging
[RFC8176: Authentication Method Reference Values] which is almost a universal standard implemented by major Identity
Provider protocols like [OpenID Connect 1.0] and [Security Assertion Markup Language (SAML) 2.0].

## Authentication Method Reference Values Explainer

Authentication Method Reference Values are standardized identifiers that indicate which authentication methods were
used during a user's authentication process.

Examples include `pwd` , `otp` , `mfa`, etc. A full list of meanings for each Authentication Method References Values
as it pertains to Authelia can be found in the
[Authentication Method References Values Reference Guide](../../reference/guides/authentication-method-references.md).
By recording and leveraging these values, Authelia can make more sophisticated authorization
decisions based on not just whether a user is authenticated, but specifically how they authenticated, enabling granular
access control policies that are customizable by administrators.

For example, an administrator could configure Authelia to:

  - Require `hwk` or `swk` for accessing internal company applications
  - Enforce `mfa` with specific combinations like `hwk` and `otp` for admin portals
    - Please note that any Authelia administration portal will require an absolute minimum of `mfa`
  - Allow `pwd` authentication for basic applications but require additional factors for sensitive resources

All at the same time as leveraging the already first-class
[Access Control Rules](../../configuration/security/access-control.md) or the emerging
[OpenID Connect 1.0 Authorization Polices](../../configuration/identity-providers/openid-connect/provider.md#authorization_policies)
to deliver an unparalleled authorization experience.

## Stages

This section represents the stages involved in implementation of this feature. The stages are either in order of
implementation due to there being an underlying requirement to implement them in this order, or in their likely order
due to how important or difficult to implement they are.

### Record Authentication Methods Reference Values

{{< roadmap-status stage="complete" version="v4.35.0" >}}

This stage is effectively the initial implementation. We implemented this for the sake of [OpenID Connect 1.0] initially
with the intention of expanding it's use to general authorization and [Security Assertion Markup Language (SAML) 2.0]
later.

### Derive Authorization Level from Authentication Methods Reference Values

{{< roadmap-status stage="complete" version="v4.39.0" >}}

This stage will leverage the Authorization Level entirely from the previously recorded
[RFC8176: Authentication Method Reference Values]. This will pave the way for the next stage and simplify important
logic.

### Implement Custom Authentication Methods Reference Values Policies

{{< roadmap-status stage="needs-design" version="v4.41.0" >}}

This stage will allow administrators to develop their own custom policies based on
[RFC8176: Authentication Method Reference Values]. This will enhance other features such as Passkeys allowing
administrators to decide for themselves what level is required for each rule. How we do this still requires a bit of
careful planning.

### Custom Policy Flows

{{< roadmap-status stage="needs-design" version="v4.41.0" >}}

To facilitate the
[Implement Custom Authentication Methods Reference Values Policies](#implement-custom-authentication-methods-reference-values-policies)
stage we will have to build a frontend flow that supports the configured policy. How we do this still requires a bit of
careful planning.

### Credential Registration

{{< roadmap-status stage="needs-design" >}}

There will likely need to be some adjustments of how we handle credential registration. In particular we probably need
to implement a more complex decision process on what to show and not show for registration, specifically for WebAuthn
since it can be used as a login method. How we do this still requires a bit of careful planning.

[OpenID Connect 1.0]: https://openid.net/specs/openid-connect-core-1_0.html
[Security Assertion Markup Language (SAML) 2.0]: https://docs.oasis-open.org/security/saml/Post2.0/sstc-saml-tech-overview-2.0.html
[RFC8176: Authentication Method Reference Values]: https://datatracker.ietf.org/doc/html/rfc8176
