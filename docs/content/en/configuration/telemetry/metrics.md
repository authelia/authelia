---
title: "Metrics"
description: "Configuring the Metrics Telemetry settings"
lead: "Configuring the Metrics Telemetry settings."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  configuration:
    parent: "telemetry"
weight: 108200
toc: true
---

*Authelia* allows administrators to configure a [Prometheus] Metrics Exporter.

## Configuration

```yaml
telemetry:
  metrics:
    enabled: false
    address: "0.0.0.0:9959"
```

## Options

### enabled

{{< confkey type="boolean" default="false" required="no" >}}

Determines if the [Prometheus] HTTP Metrics Exporter is enabled.

### address

{{< confkey type="address" default="0.0.0.0:9959" required="no" >}}

Configures the listener address for the [Prometheus] HTTP Metrics Exporter. The address must be a IPv4 or IPv6 address
followed by the port in the `<address>:<port>` format.

## See More

- [Telemetry Reference Documentation](../../reference/guides/metrics.md)

[Prometheus]: https://prometheus.io/
