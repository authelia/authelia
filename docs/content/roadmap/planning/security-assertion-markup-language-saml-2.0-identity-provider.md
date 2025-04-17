---
title: "Security Assertion Markup Language (SAML) 2.0 Provider"
description: "Authelia Security Assertion Markup Language (SAML) 2.0 Provider Implementation"
summary: "The Security Assertion Markup Language (SAML) 2.0 Provider role is a very useful but complex feature to enhance interoperability of Authelia with other products."
date: 2025-03-23T19:03:40+11:00
draft: false
images: []
weight: 225
toc: true
aliases:
  - /r/saml
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

The [Security Assertion Markup Language (SAML) 2.0] Provider implementation has many of the same benefits of the
[OpenID Connect 1.0 Provider](../active/openid-connect-1.0-provider.md) implementation but obviously supports clients
support it instead of [OpenID Connect 1.0].

### Decide On a Library

{{< roadmap-status stage="in-progress" >}}

While there are not many effectively operation libraries for this, it's important we consider all our options. There are
many pitfalls with [Security Assertion Markup Language (SAML) 2.0] due to the choice to base the protocol on XML and we
need to be very mindful of these.

[Security Assertion Markup Language (SAML) 2.0]: https://docs.oasis-open.org/security/saml/Post2.0/sstc-saml-tech-overview-2.0.html
[OpenID Connect 1.0]: https://openid.net/connect/
