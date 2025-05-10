---
title: "Zoraxy"
description: "An integration guide for Authelia and the Zoraxy reverse proxy"
summary: "A guide on integrating Authelia with the Zoraxy reverse proxy."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
weight: 380
toc: true
aliases:
  - /i/zoraxy
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

[Zoraxy] is a reverse proxy that will soon be supported by __Authelia__.

*__Important:__ When using these guides, it's important to recognize that we cannot provide a guide for every possible
method of deploying a proxy. These guides show a suggested setup only, and you need to understand the proxy
configuration and customize it to your needs. To-that-end, we include links to the official proxy documentation
throughout this documentation and in the [See Also](#see-also) section.*

## Requirements

Authelia by default only generally provides support for versions of products that are also supported by their respective
developer. As such we only support the versions [Zoraxy] officially provides support for. The versions and lifetime
of support for [Zoraxy] is unknown at this time and we're using a development version. As such it can be assumed we
do not yet support an official build, and when we do it will only be the latest version.

It should be noted that while these are the listed versions that are supported you may have luck with older versions.

We can officially guarantee the following versions of [Zoraxy] as these are the versions we perform integration testing
with at the current time:

## Get started

It's __*strongly recommended*__ that users setting up *Authelia* for the first time take a look at our
[Get started](../prologue/get-started.md) guide. This takes you through various steps which are essential to
bootstrapping *Authelia*.

[Zoraxy]: https://zoraxy.aroz.org/
