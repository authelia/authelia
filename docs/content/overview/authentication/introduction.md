---
title: "Authentication"
description: "An overview of a authentication."
summary: "An overview of a authentication."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 210
toc: true
aliases:
  - /docs/features/2fa/
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

Multi-Factor Authentication or MFA as a concept is separated into three major categories. These categories are:

* something you know
* something you have
* something you are

Modern best security practice dictates that using multiple of these categories is necessary for security. Users are
unreliable and simple usernames and passwords are not sufficient for security.

__Authelia__ enables primarily two-factor authentication. These methods offered come in two forms:

* 1FA or first-factor authentication which is handled by a username and password. This falls into the
  *something you know* categorization.
* 2FA or second-factor authentication which is handled by several methods including Time-based One-Time Passwords,
  authentication keys, etc. This falls into the *something you have* categorization.

In addition to this Authelia can apply authorization policies to individual website resources which restrict which
identities can access which resources from a given remote address. These policies can require 1FA, 2FA, or outright deny
access depending on the criteria you configure.
