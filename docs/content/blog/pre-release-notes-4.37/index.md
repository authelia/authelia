---
title: "4.37: Pre-Release Notes"
description: "Authelia 4.37 is just around the corner. This version has several additional features and improvements to existing features. In this blog post we'll discuss the new features and roughly what it means for users."
summary: "Authelia 4.37 is just around the corner. This version has several additional features and improvements to existing features. In this blog post we'll discuss the new features and roughly what it means for users."
date: 2024-03-14T06:00:14+11:00
draft: false
weight: 50
categories: ["News", "Release Notes"]
tags: ["releases", "pre-release-notes"]
contributors: ["James Elliott"]
pinned: false
homepage: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

Authelia [4.37](https://github.com/authelia/authelia/milestone/16) is just around the corner. This version has several
additional features and improvements to existing features. In this blog post we'll discuss the new features and roughly
what it means for users.

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
These features are still subject to change; however, it represents the most likely features.
{{< /callout >}}

This blog entry (and technically the blog itself) is part of a new effort I'm making for which I'm not entirely sure how
useful it'll be but I'd love to hear your feedback regardless. We don't use any analytics or interactive components to
gauge the consumption or reception of the website so it is invaluable to get this feedback.

## Envoy Support

We'll be supporting [Envoy] and [Istio] in this release. Support for this proxy mostly completes our proxy support status
with all major proxies supported excluding Microsoft IIS.

[Envoy]: https://www.envoyproxy.io/
[Istio]: https://istio.io/

## OpenID Connect 1.0 Improvements

Several items from the [OpenID Connect 1.0 Roadmap](../../roadmap/active/openid-connect.md) are being checked off in this
release.

### Hashed Client Secrets

We'll be supporting hashed OpenID Connect 1.0 client secrets in this release. People will still be able to use plaintext
secrets if they wish however we'll be recommending people utilize PBKDF2, Bcrypt or SHA512 SHA2CRYPT (see
[Password Algorithms](#password-algorithms) for a full compatibility list). This doesn't change anything for OpenID
Connect Relying Parties, it only requires a change in the Authelia configuration.

### Consent Modes

Currently we support an explicit consent mode, and a pre-configured consent mode if the pre-configured duration is set.
In this release we're planning to support an implicit consent mode which will never ask users for any consent. In
addition it will make the consent mode configuration slightly more explicit.

### JWKS Certificate Chain

Currently we do not support JWKS certificates, we only support private keys. We will support advertising the Certificate
Chain via the JWKS endpoint in this release. This means when provided with a Certificate Chain will be able to
theoretically validate the level of trust associated with the JWKS.

Some applications theoretically require this, most probably don't support it at all. However the beauty of this change
is that if it's not supported by the other party they can just ignore it. We've yet to have users request this but it's
likely inevitable that someone will ask or some third party will require it at some point, so we're preemptively
implementing it.

## Container Annotations / Labels

In this release we're going to start adding the [OCI Image Format Specification]'s set of [Annotations] to all of our
images.

For the time being we will also add the [Annotations] as container labels. This is because [Annotations] are a
relatively unsupported specification at this stage. A majority of use cases for the [Annotations] either actually use
labels or fallback to labels.

[OCI Image Format Specification]: https://github.com/opencontainers/image-spec
[Annotations]: https://github.com/opencontainers/image-spec/blob/main/annotations.md

## Password Algorithms

Several new password hashing algorithms will be supported in this release. The list of supported algorithms will become:

* Argon2:
  * Argon2id (previously supported)
  * Argon2i
  * Argon2d
* PBKDF2:
  * SHA1
  * SHA224
  * SHA256
  * SHA384
  * SHA512
* Scrypt:
  * Scrypt (standard variation)
  * Yescrypt
* Bcrypt
* SHA2 CRYPT:
  * SHA256
  * SHA512 (previously supported)

## Users YAML File Authentication Backend

In addition to the [Password Algorithms](#password-algorithms) changes we'll also be adding a few major features to the
Users YAML File Authentication Backend.

### Automatic Reload

Administrators will be able to allow automatic reload the YAML file with the users database for deployments of the YAML
File backend. This change will not extend to the main configuration file at this time.

### Email Lookup

Administrators will be able to allow users to use their email or their username to login similar to how this can be done
with an LDAP filter already, bringing feature parity to the YAML File backend.

## Mutual TLS Support

This release will add support for Mutual TLS for Redis, LDAP, and SMTP. This improves compatibility with these systems
when password authentication is not desired.

## Query Parameter Authorization Criteria

We'll be adding a very specific query parameter matcher to the access control rules in this release. It will allow
individually targeting specific query arguments and testing if they exist/don't exist, equal/don't equal a value, or if
they match/don't match a specific Regular Expression.

This rule type takes a performance hit when compared to the resources rule type, so the resources rule type is generally
encouraged. However for complex matching of query parameters a Regular Expression is hard to get exactly right. This
feature alleviates this issue.

## Compatibility Features

Compatibility Features are toggle on features which allows operation with a third party. Usually they're implemented to
allow ignoring a specific specification the third party has not implemented correctly.

We'll be adding the following compatibility features in this release:

* LDAP Servers which do not support querying the RootDSE for supported controls or extensions.
* SMTP Servers which advertise support for STARTTLS but do not actually support it.
