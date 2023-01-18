---
title: "4.38: Pre-Release Notes"
description: "Authelia 4.38 is just around the corner. This version has several additional features and improvements to existing features. In this blog post we'll discuss the new features and roughly what it means for users."
lead: "Pre-Release Notes for 4.38"
excerpt: "Authelia 4.38 is just around the corner. This version has several additional features and improvements to existing features. In this blog post we'll discuss the new features and roughly what it means for users."
date: 2023-01-18T19:47:09+10:00
draft: false
images: []
categories: ["News", "Release Notes"]
tags: ["releases", "pre-release-notes"]
contributors: ["James Elliott"]
pinned: false
homepage: false
---

Authelia [4.38](https://github.com/authelia/authelia/milestone/17) is just around the corner. This version has several
additional features and improvements to existing features. In this blog post we'll discuss the new features and roughly
what it means for users.

Overall this release adds several major roadmap items. It's quite a big release. We expect a few bugs here and there but
nothing major. It's one of our biggest releases to date, so while it's taken a longer time than usual it's for good
reason we think.

As this is a larger release we're probably going to ask users to help with some experimentation. If you're comfortable
backing up your database then please keep your eyes peeled in the [chat](../../information/contact.md#chat).

_**Note:** These features discussed in this blog post are still subject to change however they represent the most likely
outcome._

_**Important Note:** There are some changes in this release which deprecate older configurations. The changes should be
backwards compatible, however mistakes happen. In addition we advise making the adjustments to your configuration as
necessary as several new features will not be available or even possible without making the necessary adjustments. We
will be publishing some guides on making these adjustments on the blog in the near future, including an FAQ catered to
specific scenarios._

## OpenID Connect 1.0

As part of our ongoing effort for comprehensive support for OpenID Connect 1.0 we'll be introducing several important
features. Please see the [roadmap](../../roadmap/active/openid-connect.md) for more information.

### OAuth 2.0 Pushed Authorization Requests

Support for [RFC9126](https://datatracker.ietf.org/doc/html/rfc9126) known as `OAuth 2.0 Pushed Authorization Requests`
is one of the main features being added to our OpenID Connect 1.0 implementation in this release.

Pushed Authorization Requests allows for relying parties / clients to send the Authorization Request parameters over a
back-channel and receive an opaque URI to be used as the `redirect_uri` on the standard Authorization endpoint in place
of the standard Authorization Request parameters.

The endpoint used by this mechanism requires the relying party provides the Token Endpoint authentication parameters.

This means the actual Authorization Request parameters are never sent in the clear over the front-channel. This helps
mitigate a few things. In particular it's privacy focused, but it also reduces the attack surface by drastically
reducing the amount of information over the front-channel.

Even if an attacker gets the Authorization Code, they are unlikely to have the `client_id` for example, and this is
required to exchange the Authorization Code for an Access Token and ID Token.

This option can be enforced globally for users who only use relying parties which support PAR, or can be individually
enforced per client.

### Proof Key for Code Exchange by OAuth Public Clients

In addition to the PAR feature, users will now be able to enforce Proof Key for Code Exchange on individual clients. It
should also be noted, Proof Key for Code Exchange can be used with
[OAuth 2.0 Pushed Authorization Requests](#oauth-20-pushed-authorization-requests). These features combined with our
requirement for the HTTPS scheme are very powerful security measures.

## Multi-Domain Protection

In this release we are releasing the main implementation of the Multi-Domain Protection roadmap item.
Please see the [roadmap](../../roadmap/active/openid-connect.md) for more information.

### Initial Implementation

_**Important Note:** This feature at the time of this writing, will not work well with Webauthn. Steps are being taken
to address this however it will not specifically delay the release of this feature._

This release see's the initial implementation of multi-domain protection. Users will be able to configure more than a
single root domain for cookies provided none of them are a subdomain of another domain configured. In addition each
domain can have individual settings.

This does not allow single sign-on between these distinct domains. When surveyed users had very low interest in this
feature and technically speaking it's not trivial to implement such a feature as a lot of critical security
considerations need to be addressed.

In addition this feature will allow configuration based detection of the Authelia Portal URI on proxies other than
NGINX/NGINX Proxy Manager/SWAG/HAProxy with the use of the new
[Customizable Authorization Endpoints](#customizable-authorization-endpoints). This is important as it means you only
need to configure a single middleware or helper to perform automatic redirection.

## Webauthn

As part of our ongoing effort for comprehensive support for Webauthn we'll be introducing several important
features. Please see the [roadmap](../../roadmap/active/webauthn.md) for more information.

### Multiple Webauthn Credentials Per-User

In this release we see full support for multiple Webauthn credentials. This is a fairly basic feature but getting the
frontend experience right is important to us. This is going to be supported via the
[User Control Panel](#user-dashboard--control-panel).

## Customizable Authorization Endpoints

For the longest time we've managed to have the `/api/verify` endpoint perform all authorization verification. This has
served us well however we've been growing out of it. This endpoint is being deprecated in favor of new customizable
per-implementation endpoints. Each existing proxy we support uses one of these distinct implementations.

The old endpoint will still work, in fact you can technically configure an additional endpoint using the methodology of
it via the `Legacy` implementation. However this is strongly discouraged and will not intentionally have new features or
fixes (excluding security fixes) going forward.

In addition to being able to customize them you can create your own, and completely disable support for all other
implementations in the process. Use of these new endpoints will require reconfiguration of your proxy, we plan to
release a guide for each proxy.

## User Dashboard / Control Panel

As part of our ongoing effort for comprehensive support for a User Dashboard / Control Panel we'll be introducing
several important features. Please see the [roadmap](../../roadmap/active/dashboard-control-panel.md) for more
information.

### Device Registration OTP

Instead of the current link, in this release users will instead be sent a One Time Password, cryptographically randomly
generated by Authelia. This One Time Password will grant users a duration to perform security sensitive tasks.

The motivation for this is that it works in more situations, and is slightly less prone to phishing.

### TOTP Registration

Instead of just assuming that users have successfully registered their TOTP application, we will require users to enter
the TOTP code prior to it being saved to the database.

## Configuration

Several enhancements are landing for the configuration.

### Directories

Users will now be able to configure a directory where all `.yml` and `.yaml` files will be loaded in lexical order.
This will not allow combining lists of items, but it will allow you to split portions of the configuration easily.

### Configuration Discovery

Environment variables are being added to assist with configuration discovery, and this will be the default method for
our containers. The advantage is that since the variable will be available when execing into the container, even if
the configuration paths have changed or you've defined additional paths, the `authelia` command will know where the
files are if you properly use this variables.

### Templating

The file based configuration will have access to several experimental templating filters which will assist in creating
configuration templates. The initial one will just expand *most* environment variables into the configuration. The
second will use the go template engine in a very similar way to how Helm operates.

As these features are experimental they may break, be removed, or otherwise not operate as expected. However most of our
testing indicates they're incredibly solid.

### LDAP Implementation

Several new LDAP implementations which provide defaults are being introduced in this version to assist users in
integrating their LDAP server with Authelia.

## Miscellaneous

Some miscellaneous notes about this release.

### Email Notifications

Events triggered by users will generate new notifications sent to their inbox, for example adding a new 2FA device.

### Storage Import/Export

Utility functions to assist in exporting and subsequently importing the important values in Authelia are being added and
unified in this release.

### Privacy Policy

We'll be introducing a feature which allows administrators to more easily comply with the GDPR which optionally shows a
link to their individual privacy policy on the frontend, and optionally requires users to accept it before using
Authelia.
