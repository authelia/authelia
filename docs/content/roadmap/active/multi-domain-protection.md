---
title: "Multi-Domain Protection"
description: "Authelia Multi-Domain Protection Implementation"
summary: "Multi-Domain Protection is one of the most requested Authelia features."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 335
toc: true
aliases:
  - /r/multi-domain-protection
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

We have seen and heard the feedback from our users and we are acting on it. This feature is being prioritized. Allowing
administrators to protect more than one root domain utilizing a single Authelia instance is going to be a difficult
feature to implement but we'll actively take steps to implement it.

## Stages

This section represents the stages involved in implementation of this feature. The stages are either in order of
implementation due to there being an underlying requirement to implement them in this order, or in their likely order
due to how important or difficult to implement they are.

### Decide on a Method

{{< roadmap-status stage="complete" >}}

We need to decide on a method to implement this feature initially and how it will finally look to provide SSO between
root domains.

*__UPDATE:__* The [initial implementation](#initial-implementation) has been decided as well as the
[SSO implementation](#sso-implementation).

### Decide on a Session Library

{{< roadmap-status stage="complete" >}}

We've decided on moving away from using the current session library and plan on entirely implementing session logic
internally.

### Initial Implementation

{{< roadmap-status stage="complete" version="v4.38.0" >}}

This stage is waiting on the choice to handle sessions. Initial implementation will involve just a basic cookie
implementation where users will be required to sign in to each root domain and no inter-domain SSO functionality will
exist.

See the [SSO implementation](#sso-implementation) for how we plan to address the sign in limitation.

### SSO Implementation

{{< roadmap-status >}}

While the initial implementation will require users to sign in to each root domain and no SSO functionality will exist
as outlined in the [initial implementation](#initial-implementation), it's possible via Identity protocols / frameworks
like [OpenID Connect 1.0](openid-connect-1.0-provider.md) to perform Single-Sign On transparently for users.

This will very likely be implemented at the same time as
[OpenID Connect 1.0 Relying Party](../planning/openid-connect-1.0-relying-party.md) support.
