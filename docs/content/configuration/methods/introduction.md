---
title: "Methods"
description: "Methods of Configuration."
summary: "Authelia has a layered configuration model. This section describes how to implement configuration."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 101100
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Layers

Authelia has several methods of configuration available to it. The order of precedence is as follows:

1. [Secrets](secrets.md)
2. [Environment Variables](environment.md)
3. [Files](files.md) (in order of them being specified)

This order of precedence puts higher weight on things higher in the list. This means anything specified in the
[files](files.md) is overridden by [environment variables](environment.md) if specified, and anything specified by
[environment variables](environment.md) is overridden by [secrets](secrets.md) if specified.

[YAML]: https://yaml.org/
