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

{{< config-alert-example >}}

```yaml
telemetry:
  metrics:
    enabled: false
    address: "tcp://0.0.0.0:9959"
    buffers:
      read: 4096
      write: 4096
    timeouts:
      read: 6s
      write: 6s
      idle: 30s
```

## Options

This section describes the individual configuration options.

### enabled

{{< confkey type="boolean" default="false" required="no" >}}

Determines if the [Prometheus] HTTP Metrics Exporter is enabled.

### address

{{< confkey type="address" default="tcp://0.0.0.0:9959" required="no" >}}

Configures the listener address for the [Prometheus] HTTP Metrics Exporter. This configuration key uses the
[Address](../prologue/common.md#address) format. The scheme must be `tcp://` or empty.

### buffers

Configures the server buffers. See the [Server Buffers](../prologue/common.md#server-buffers) documentation for more
information.

### timeouts

Configures the server timeouts. See the [Server Timeouts](../prologue/common.md#server-timeouts) documentation for more
information.

## See More

- [Telemetry Reference Documentation](../../reference/guides/metrics.md)

[Prometheus]: https://prometheus.io/
