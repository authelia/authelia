---
title: "Chart"
description: "A guide to using the Authelia helm chart to integrate Authelia with Kubernetes"
lead: "A guide to using the Authelia helm chart to integrate Authelia with Kubernetes."
date: 2022-06-22T22:58:23+10:00
draft: false
images: []
menu:
  integration:
    parent: "kubernetes"
weight: 520
toc: true
---

Authelia offers a [Helm Chart] which can make integration with [Kubernetes] much easier. It's currently considered beta
status, and as such is subject to breaking changes.

## Get Started

It's __*strongly recommended*__ that users setting up *Authelia* for the first time take a look at our
[Get Started](../prologue/get-started.md) guide. This takes you through various steps which are essential to
bootstrapping *Authelia*.

## Repository

The [Helm Chart] repository for Authelia is `https://charts.authelia.com`. You can add it to your repository list with
the following [Helm] commands:

```bash
helm repo add authelia https://charts.authelia.com
helm repo update
```

## Website

The [https://charts.authelia.com/](https://charts.authelia.com/) URL also serves a website with basic chart information.

## Source

The source for the [Helm Chart] is hosted on [GitHub](https://github.com/authelia/chartrepo). Please feel free to
[contribute](../../contributing/prologue/introduction.md).

[Kubernetes]: https://kubernetes.io/
[Helm]: https://helm.sh/
[Helm Chart]: https://helm.sh/docs/topics/charts/
