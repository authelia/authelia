---
title: "NTP"
description: "Configuring the NTP Settings."
lead: "Authelia checks the system time is in sync with an NTP server. This section describes how to configure and tune this."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  configuration:
    parent: "miscellaneous"
weight: 199300
toc: true
aliases:
  - /docs/configuration/ntp.html
---

Authelia has the ability to check the system time against an NTP server, which at the present time is checked only
during startup. This section configures and tunes the settings for this check which is primarily used to ensure
[TOTP](../second-factor/time-based-one-time-password.md) can be accurately validated.

In the instance of inability to contact the NTP server or an issue with the synchronization Authelia will fail to start
unless configured otherwise.

## Configuration

{{< config-alert-example >}}

```yaml
ntp:
  address: 'udp://time.cloudflare.com:123'
  version: 3
  max_desync: '3s'
  disable_startup_check: false
  disable_failure: false
```

## Options

This section describes the individual configuration options.

### address

{{< confkey type="string" default="time.cloudflare.com:123" required="no" >}}

Determines the address of the NTP server to retrieve the time from. The format is `<host>:<port>`, and both of these are
required.

### address

{{< confkey type="address" default="udp://time.cloudflare.com:123" required="no" >}}

*__Reference Note:__ This configuration option uses the [address common syntax](../prologue/common.md#address). Please
see the [documentation](../prologue/common.md#address) on this format for more information.*

Configures the address for the NTP Server. The address itself is a connector and the scheme must be `udp`,
`udp4`, or `udp6`.

__Examples:__

```yaml
ntp:
  address: 'udp://127.0.0.1:123'
```

```yaml
ntp:
  address: 'udp6://[fd00:1111:2222:3333::1]:123'
```

### version

{{< confkey type="integer" default="4" required="no" >}}

Determines the NTP version supported. Valid values are 3 or 4.

### max_desync

{{< confkey type="duration" default="3s" required="no" >}}

*__Reference Note:__ This configuration option uses the [duration common syntax](../prologue/common.md#duration).
Please see the [documentation](../prologue/common.md#duration) on this format for more information.*

This is used to tune the acceptable desync from the time reported from the NTP server.

### disable_startup_check

{{< confkey type="boolean" default="false" required="no" >}}

Setting this to true will disable the startup check entirely.

### disable_failure

{{< confkey type="boolean" default="false" required="no" >}}

Setting this to true will allow Authelia to start and just log an error instead of exiting. The default is that if
Authelia can contact the NTP server successfully, and the time reported by the server is greater than what is configured
in [max_desync](#maxdesync) that Authelia fails to start and logs a fatal error.
