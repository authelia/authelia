---
title: "Cache"
description: "Cache Configuration"
summary: "Configuring the Cache settings."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 106100
toc: true
aliases: []
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

__Authelia__ uses a caching provider with publish / subscribe capabilities to store several things in a highly
available way. This is required to be able to deploy Authelia with high availability. Examples of things stored here
that at the time of writing are:

1. [Sessions](../session/introduction.md)

## Variables

Some of the values within this page can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
cache:
  redis: {}
  redis_cluster: {}
  redis_sentinel: {}
```

## Providers

There are currently three providers that can be configured for a cache:

* [Redis](redis.md).
* [Redis Cluster](redis-cluster.md) (additional high availability support).
* [Redis Sentinel](redis-sentinel.md) (additional high availability support).

Only one may be selected at a time and configuring more than one is not supported.

## Options

There are no individual options at this time. Each option is configured directly in one of the above providers.
