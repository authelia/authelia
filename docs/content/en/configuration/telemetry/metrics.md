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
    address: 'tcp://:9959'
    umask: 0022
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

{{< confkey type="address" default="tcp://:9959" required="no" >}}

*__Reference Note:__ This configuration option uses the [address common syntax](../prologue/common.md#address). Please
see the [documentation](../prologue/common.md#address) on this format for more information.*

Configures the listener address for the [Prometheus] Metrics Exporter HTTP Server. The address itself is a listener and
the scheme must either be the `unix` scheme or one of the `tcp` schemes.

### umask

{{< confkey type="int" required="no" >}}

If set temporarily changes the umask during the creation of the unix domain socket if configured as such in the
[address](#address). Typically this should be set before the process is actually running and users should not use this
option, however it's recognized in various specific scenarios this may not be completely adequate.

One such example is when you want the proxy to have permission to the socket but not the files, in which case running a
umask of `0077` by default is good, and running a umask of `0027` so that the group Authelia is running as has
permission to the socket.

This value should typically be prefixed with a `0` to ensure the relevant parsers handle it correctly.

### buffers

*__Reference Note:__ This configuration option uses the
[Server buffers common structure](../prologue/common.md#server-buffers). Please see the
[documentation](../prologue/common.md#server-buffers) on this structure for more information.*

Configures the server buffers.

### timeouts

*__Reference Note:__ This configuration option uses the
[Server timeouts common structure](../prologue/common.md#server-timeouts). Please see the
[documentation](../prologue/common.md#server-timeouts) on this structure for more information.*

Configures the server timeouts.

## See More

- [Telemetry Reference Documentation](../../reference/guides/metrics.md)

[Prometheus]: https://prometheus.io/
