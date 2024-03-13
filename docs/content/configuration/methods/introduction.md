---
title: "Methods"
description: "Methods of Configuration."
lead: "Authelia has a layered configuration model. This section describes how to implement configuration."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  configuration:
    parent: "methods"
weight: 101100
toc: true
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
