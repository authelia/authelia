---
title: "Metrics"
description: "Configuring the Metrics Telemetry settings"
summary: "Configuring the Metrics Telemetry settings."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 109200
toc: true
---

*Authelia* allows administrators to configure a [Prometheus] Metrics Exporter.

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
telemetry:
  metrics:
    enabled: false
    address: 'tcp://:9959/metrics'
    buffers:
      read: 4096
      write: 4096
    timeouts:
      read: '6s'
      write: '6s'
      idle: '30s'
```

## Options

This section describes the individual configuration options.

### enabled

{{< confkey type="boolean" default="false" required="no" >}}

Determines if the [Prometheus] HTTP Metrics Exporter is enabled.

### address

{{< confkey type="string" syntax="address" default="tcp://:9959/metrics" required="no" >}}

Configures the listener address for the [Prometheus] Metrics Exporter HTTP Server. The address itself is a listener and
the scheme must either be the `unix` scheme or one of the `tcp` schemes.

### buffers

{{< confkey type="structure" structure="server-buffers" required="no" >}}

Configures the server buffers.

### timeouts

{{< confkey type="structure" structure="server-timeouts" required="no" >}}

Configures the server timeouts.

## See More

- [Telemetry Reference Documentation](../../reference/guides/metrics.md)

[Prometheus]: https://prometheus.io/
