---
title: "Kubernetes"
description: "An introduction into integrating Authelia with Kubernetes."
summary: "An introduction into integrating Authelia with Kubernetes."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 510
toc: true
aliases:
  - /i/k8s
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

{{< figure src="images/logos/kubernetes.png" alt="Kubernetes" width="100" style="padding-right: 10px" ignoreStaticImages="false" >}}

## Get started

It's __*strongly recommended*__ that users setting up *Authelia* for the first time take a look at our
[Get started](../prologue/get-started.md) guide. This takes you through various steps which are essential to
bootstrapping *Authelia*.

## Important Notes

The following section has special notes regarding utilizing Authelia with Kubernetes.

### External Traffic Policy

Authelia (and all of your other applications) may receive an invalid remote IP if the service handling traffic to the
Kubernetes Ingress of your choice doesn't have the `externalTrafficPolicy` setting configured to `local` as per the
Kubernetes [preserving the client source ip] documentation.

### Enable Service Links

Authelia's configuration management system conflicts with the `enableServiceLinks` option when it's set to `true` which
is the default. This should be changed to `false`. See the
[PodSpec v1 core](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#podspec-v1-core) documentation
for more details.

#### Pod Example

```yaml {title="pod.yml"}
---
apiVersion: v1
kind: Pod
metadata:
  name: authelia
spec:
  enableServiceLinks: false
...
```

## Frequently Asked Questions

### RAM usage

If using file-based authentication, the argon2id provider will by default use 1GB of RAM for password generation. This
means you should allow for at least this amount in your deployment/daemonset spec and have this much available on your
node, alternatively you can
[tweak the providers settings](../../configuration/first-factor/file.md#memory). Otherwise,
your Authelia may OOM during login. See [here](https://github.com/authelia/authelia/issues/1234#issuecomment-663910799)
for more info.

[preserving the client source ip]: https://kubernetes.io/docs/tasks/access-application-cluster/create-external-load-balancer/#preserving-the-client-source-ip
