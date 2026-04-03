---
title: "Regulation"
description: "Regulation of failed attempts is an important function of an IAM system."
summary: "Regulation of failed attempts is an important function of an IAM system."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 320
toc: true
aliases:
  - /docs/features/regulation.html
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

__Authelia__ takes the security of users very seriously and comes with a way to avoid brute-forcing the first factor
credentials by regulating the authentication attempts and temporarily banning an account when too many attempts have
been made.

## Configuration

Please check the dedicated [documentation](../../configuration/security/regulation.md).
