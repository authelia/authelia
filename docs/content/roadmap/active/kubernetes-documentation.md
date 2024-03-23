---
title: "Kubernetes Documentation"
description: "Add better Kubernetes documentation."
summary: "While there is some documentation for Kubernetes, and several people have it working, better documentation is needed."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
weight: 250
toc: true
aliases:
  - /r/k8s-docs
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Stages

This section represents the stages involved in implementation of this feature. The stages are either in order of
implementation due to there being an underlying requirement to implement them in this order, or in their likely order
due to how important or difficult to implement they are.

### Integration Documentation

{{< roadmap-status stage="in-progress" >}}

Provide some generalized integration documentation for [Kubernetes].

### Helm Chart

{{< roadmap-status stage="in-progress" >}}

Develop and release a [Helm] [Chart](https://helm.sh/docs/topics/charts/) which makes implementation on [Kubernetes]
easy.

This is currently in progress and there is a [Helm Chart Repository](https://charts.authelia.com). This is considered
beta and the chart itself has a lot of work to go.

### Kustomize

{{< roadmap-status >}}

Implement a [Kustomize] bundle people can utilize with [Kubernetes].

[Helm]: https://helm.sh/
[Kubernetes]: https://kubernetes.io/
[Kustomize]: https://kustomize.io/
