---
title: "4.39: Release Notes"
description: "Authelia 4.39 release notes."
summary: "Authelia 4.39 has been released and the following is a guide on all the massive changes."
date: 2025-03-08T15:54:02+11:00
draft: false
weight: 50
categories: ["News", "Release Notes"]
tags: ["releases", "release-notes"]
contributors: ["James Elliott", "Brynn Crowley"]
pinned: false
homepage: false
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

Authelia [4.39](https://github.com/authelia/authelia/releases/tag/v4.39.0) is released! This version has several
additional features and improvements to existing features. In this blog post we'll discuss the new features and roughly
what it means for users.

Overall this release adds several major roadmap items. It's quite a big release. We expect a few bugs here and there but
nothing major.

## Foreword

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
This section is important to read for all users who are upgrading, especially those who are automatically upgrading.
{{< /callout >}}

There are some changes in this release which deprecate older configurations, you will get a warning about these
deprecations as it's likely in version v5.0.0 we'll remove support for them, however if a log message for
a configuration is a warning then it's just a warning, and can fairly safely be ignored for now. These changes should be
backwards compatible, however mistakes happen. If you find a mistake we kindly request that you let us know.

By far the most important change to be aware of is a change to OpenID Connect 1.0 and the default claims in issued ID
Tokens. As we get closer to practical completion and removing OpenID Connect 1.0 beta status we're adding additional
features which dictate the need to make an adjustment to how we handle some things like this.

## On This Page

This blog article is rather large so this serves as an index for the section headings and relevant important changes.

- Key Sections:
  - [Docker](#docker)
  - [OpenID Connect 1.0](#openid-connect-10)
  - [WebAuthn](#webauthn)
- [Other Improvements](#other-improvements)
  - [Password Change](#password-change)
  - [Log File Reopening](#log-file-reopening)
  - [Basic Authorization Caching](#basic-authorization-caching)


---

## Docker

As an intentional improvement to both the compatibility and security of the Authelia container we've made a number of
important changes to our container image.

The first change which is most impactful to security in as much as it hardens the Authelia container is we've moved
away from the Alpine Linux base image and developed our own base image using
[chisel](https://github.com/canonical/chisel). This base image is a glibc minimal image that only has the essentials
for running the Authelia binary and the healthcheck, there is no package manager, and some unnecessary but common tools
have been removed. This container is rebuilt daily and on every tagged release.

The second change which is most impactful to end users is the removal of the `VOLUME` directive from our images. This
directive is fairly useless overall, the most impactful thing it does is leaves dangling docker volumes that get
forgotten about and lose their association with the original container, in effect making the volume data seem deleted.
Most users will not see an impact from this, and those who've used the `volumes` directive in a compose to manually map
volumes will not.

---

## WebAuthn

A number of exiting features have been added to our WebAuthn implementation.

### Passkeys and Passwordless Authentication

This release adds support for Passkeys including the ability to perform Passwordless Authentication with them. The
feature has been implemented to count as non-MFA, and by default users will have to enter their password to perform
full MFA.

A configuration option exists to change this behaviour. It should be noted we have future plans
to make this experience more customizable which will remove this configuration option in favor of one that uses
[Authentication Method Reference](#authentication-method-reference).

### Authentication Method Reference

We've adjusted the security flow because of the introduction of Passwordless Authentication to support
[RFC8176: Authentication Method Reference Values](https://www.rfc-editor.org/rfc/rfc8176.html) to determine the
authentication level. This will not only exactly map to OpenID Connect 1.0 allowing us to communicate the users
authentication level to third parties in a machine understandable way but also in the future allow us to make very
granular custom access control policies to complement `one_factor` and `two_factor`.

### FIDO Alliance Metadata Service

This release allows administrators to enable validation of authenticators via the FIDO Alliance MDS3. This includes
comprehensive checks that can be customized. This is generally considered a business feature, but it's something we'd
generally recommend users enable since it has little downsides. See the
[configuration](../../configuration/second-factor/webauthn.md#metadata) documentation for more information.

### Credential Filtering

We've added several filters that administrators can customize that validate the authenticators used. This is useful
usually for company policy where employees are expected to use a specific set of authenticators. See the
[configuration](../../configuration/second-factor/webauthn.md#filtering) documentation for more information.

### Attachment Modality

This release allows support for the platform attachment modality whereas previously we only specifically allowed the
cross-platform attachment. This should allow services such as Windows Hello to register a credential.

---

## OpenID Connect 1.0

As part of our ongoing effort for comprehensive support for [OpenID Connect 1.0] we'll be introducing several important
features. Please see the [roadmap](../../roadmap/active/openid-connect.md) for more information.

### ID Token Changes

The default claims for ID Tokens now mirrors the standard claims from the specification. This is in an effort to improve
security, improve privacy, and properly support the claims authorization parameter which is the correct means to request
additional claims.

This may affect some clients in unexpected ways, however we've included
[a guide](../../integration/openid-connect/openid-connect-1.0-claims.md#restore-functionality-prior-to-claims-parameter)
on working around this issue.

### Claims Policies

We have introduced a concept of
[claims policies](../../configuration/identity-providers/openid-connect/provider.md#claims_policies) which allows
controlling the default claims for ID Tokens and Access Tokens where access is applicable as well as custom claims and
claim scopes.

### Custom Attributes and Mapping

We've introduced a heavily requested feature of custom attributes. These custom attributes can either be directly
configured in your relevant backend such as [LDAP](../../configuration/first-factor/ldap.md#extra) or
[File](../../configuration/first-factor/file.md#extra_attributes), or they can be derived from existing attributes using
[Common Expression Language](../../configuration/definitions/user-attributes.md).

These attributes currently can only be used with OpenID Connect 1.0 but expect this to flow through to other things
at a later time.

### JSON Web Encryption

Prior to this release the only option for users was to use signed JSON Web Tokens. In this release we allow the use of
the JSON Web Encryption Nested JSON Web Tokens. This allows superior privacy in transmission of JSON Web Tokens as well
as some security when using alternative keys for signing and encryption.

This feature requires specific support by a client, and it is rare to see clients support it, but it's a feature that
exists within the scope of where we intend Authelia to sit within the ecosystem.

### OAuth 2.0 Device Code Flow

We now support the Device Code Flow which is the last major flow we did not support. This flow is the experience some
may be familiar with where they either scan a QR code on a TV-like device and sign in on a separate device like a mobile
phone, or visit a URL and enter a code.

---

## Other Improvements

This section contains all the other improvements that don't fit well into any particular grouping.

### Password Change

For a long time we've supported the ability to reset passwords. This is an exceptionally useful feature for users
who have forgotten their passwords provided the admin is agreeable to allowing this. However it's quite reasonable to
also allow users to change their known password.

In this release, we have added the ability for users to accomplish exactly this directly from the settings interface.
This means that should a user want to change a password they already know they are easily able to. This feature requires
the user perform session elevation in addition to knowing their current password.

Additionally, administrators can disable this functionality using the
[disable](../../configuration/first-factor/introduction.md#disable-1) option.

This also offers a technically more secure way for users to change their passwords, so it's quite reasonable to assume
that this may offer an alternative for administrators who had previously disabled or wanted to disable the reset
password functionality due to some of these concerns.

### Log File Reopening

Sending the `SIGHUP` signal in this release will instruct Authelia to reopen any log files. This facilitates the ability
to use a external log rotation tool like [logrotate](https://linux.die.net/man/8/logrotate) to rotate the log file while
Authelia is running. It could also realistically be used with the available replacement options the
[file_path](../../configuration/miscellaneous/logging.md#file_path) configuration option has.

When Authelia receives a SIGHUP signal, it will:

1. Safely reopen its log file handle
2. Create the log file if it doesn't exist

### Basic Authorization Caching

While we generally at this time recommend using the
[Bearer Scheme via OAuth 2.0 Bearer Token Usage](../../integration/openid-connect/oauth-2.0-bearer-token-usage.md) the
Basic Scheme is still widely used. This scheme can take some time to perform validation due to the backing password hash
which is good for security but bad for some performance requirements.

For this reason we've added an optional cache system for the Basic Scheme. This is only available on the new
[Server Authz Endpoints](../../configuration/miscellaneous/server-endpoints-authz.md) not the deprecated `/api/verify`
endpoint. The cache mechanism is in-memory and is activated by configuring the
[scheme_basic_cache_lifespan](../../configuration/miscellaneous/server-endpoints-authz.md#scheme_basic_cache_lifespan).

The lifespan configures how long each cached credential exists. The credentials are cached in a dictionary where the key
is the username, and he value is a data structure that contains the expiration and a comparison value. The comparison
value is a HMAC-SHA256 digest of the password and username, i.e. `HMAC-SHA256(password+username)`. The secret key for
the HMAC-SHA256 algorithm is cryptographically randomly generated for each
[Server Authz Endpoint](../../configuration/miscellaneous/server-endpoints-authz.md) on startup.

In the event the cached value does not yet exist or does not match the password is revalidated and the cache is updated
if the newly provided password is correct.

We do not recommend enabling this if you have the ability to utilize a more appropriate and modern scheme such as the
Bearer Scheme.
