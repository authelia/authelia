---
title: "Trusted Headers SSO"
description: "Trusted Headers SSO is a simple header authorization framework supported by Authelia."
lead: "Trusted Headers is a simple header authorization framework supported by Authelia."
date: 2022-11-27T16:07:08+11:00
draft: false
images: []
menu:
  overview:
    parent: "authorization"
weight: 340
toc: false
---

This mechanism is supported by proxies which inject certain response headers from Authelia into the protected
application. This is a very basic means that allows the target application to identify the user who is logged in
to Authelia. This like all single-sign on technologies requires support by the protected application.

You can read more about this in the [Trusted Header SSO Integration Guide](../../integration/trusted-header-sso/introduction.md).

