---
title: "Integrations"
description: "A collection of integration reference guides"
summary: "This section contains integration reference guides for Authelia."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 310
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

The integration guides in this section detail specific requirements when integrating Authelia with other products such
as supported versions, configurations, etc.

## General Rules

1. If the version or platform of the third party integration or combination thereof is not unsupported by the
   developer/vendor/etc of the third party integration we likely will not support it.
2. When we claim to support a product it is expressly the official releases of the product. It does not include
   versions that are heavily modified or drop in replacements (such as KeyDB which is a drop in replacement for redis
   that IS NOT supported).
