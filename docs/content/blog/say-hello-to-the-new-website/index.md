---
title: "Say hello to the new website 👋"
description: "Introducing the new website"
summary: "Introducing the new website"
date: 2024-03-14T06:00:14+11:00
draft: false
weight: 50
categories: ["News"]
tags: ["website"]
contributors: ["James Elliott"]
pinned: false
homepage: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

We're pleased to have you take a look at our new website. It combines both the main landing site and the documentation
all in one neat package. The website is using a [Hugo] theme called [Doks].

This does change the development process for documentation slightly compared to what it was previously. However most of
the changes will make it easier for most documentation contributors. This process is documented in the
[Documentation Contributing] section. We will also be looking towards making all
documentation changes getting quickly published to a staging site so they can quickly be seen by the maintainers and
anyone else *too lazy* to follow the steps in the [Documentation Contributing] section command and check it out locally.

As part of this redesign we've taken the time to rewrite and reorganize key sections of the documentation. This may
result in some links not working, however we've aimed to temporarily redirect links that previously worked to reduce the
number of visitors being presented a 404.

In particular, you will see improvements in the following areas:

* The integration documentation is a new area which replaces the deployment documentation:
  * There are several additional proxies
  * There are several additional deployment scenarios
  * It's much better organized
  * There are lots of additional links to additional resources to help people find the configuration that suits them
  * Several of the proxy configurations have been refreshed to make them more modern
  * Some additional docs for k8s now exist
* The roadmap is now heavily documented with stages for areas which require it
* Many areas previously located in the configuration docs have moved into integration docs as that's a more appropriate
  area
* The contribution docs have been slightly tidied up

You may have also noticed we launched a new blog! This blog will be used to communicate key things about the future of
Authelia as well as key things we believe the Authelia community needs to know about. You can subscribe to this blog
via the [RSS Feed](https://www.authelia.com/blog/index.xml). We may introduce a mail list for the blog sometime in the
future.

[Hugo]: https://gohugo.io/
[Doks]: https://getdoks.org/
[Documentation Contributing]: ../../contributing/prologue/documentation-contributions.md
