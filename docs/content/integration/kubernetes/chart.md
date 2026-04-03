---
title: "Chart"
description: "A guide to using the Authelia helm chart to integrate Authelia with Kubernetes"
summary: "A guide to using the Authelia helm chart to integrate Authelia with Kubernetes."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 520
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

Authelia offers a [Helm Chart] which can make integration with [Kubernetes] much easier. It's currently considered beta
status, and as such is subject to breaking changes.

## Get started

It's __*strongly recommended*__ that users setting up *Authelia* for the first time take a look at our
[Get started](../prologue/get-started.md) guide. This takes you through various steps which are essential to
bootstrapping *Authelia*.

## Repository

The [Helm Chart] repository for Authelia is `https://charts.authelia.com`. You can add it to your repository list with
the following [Helm] commands:

```bash
helm repo add authelia https://charts.authelia.com
helm repo update
```

## Authenticity Signature

The chart is signed as described by the
[Artifact Signing and Provenance Overview](../../overview/security/artifact-signing-and-provenance.md) where the
verification keys can also be found.

## Website

The [https://charts.authelia.com/](https://charts.authelia.com/) URL also serves a website with basic chart information.

## Source

The source for the [Helm Chart] is hosted on [GitHub](https://github.com/authelia/chartrepo). Please feel free to
[contribute](../../contributing/prologue/introduction.md).

[Kubernetes]: https://kubernetes.io/
[Helm]: https://helm.sh/
[Helm Chart]: https://helm.sh/docs/topics/charts/
