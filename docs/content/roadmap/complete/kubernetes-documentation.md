---
title: "Kubernetes Documentation"
description: "Add better Kubernetes documentation."
summary: "While there is some documentation for Kubernetes, and several people have it working, better documentation is needed."
date: 2025-03-23T19:03:40+11:00
draft: false
images: []
weight: 910
toc: true
aliases:
  - /r/k8s-docs
  - /roadmap/active/kubernetes-documentation
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

While this is generally complete we intend to invest effort to improve this as necessary based on community feedback.

### Integration Documentation

{{< roadmap-status stage="complete" >}}

Provide some generalized [integration documentation](../../integration/kubernetes/introduction.md) for [Kubernetes].

### Helm Chart

{{< roadmap-status stage="complete" >}}

Develop and release a [Helm] [Chart](https://helm.sh/docs/topics/charts/) which makes implementation on [Kubernetes]
easy.

While this is still in a pre-release considered relatively complete with only minor elements still outstanding for v1.

### Kustomize

{{< roadmap-status stage="abandoned" >}}

Implement a [Kustomize] bundle people can utilize with [Kubernetes].

This has mostly been abandoned as there seems to be little to no demand for this feature, and it's generally not popular
in the [Kubernetes] space. The remaining documentation will exist for the sake of an example for those interested.

[Helm]: https://helm.sh/
[Kubernetes]: https://kubernetes.io/
[Kustomize]: https://kustomize.io/
