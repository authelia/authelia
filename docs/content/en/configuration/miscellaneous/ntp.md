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

Authelia has the ability to check the system time against an NTP server. Currently this only occurs at startup. This
section configures and tunes the settings for this check which is primarily used to ensure
[TOTP](../second-factor/time-based-one-time-password.md) can be accurately validated.

In the instance of inability to contact the NTP server Authelia will just log an error and will continue to run.

## Configuration

```yaml
ntp:
  address: "time.cloudflare.com:123"
  version: 3
  max_desync: 3s
  disable_startup_check: false
  disable_failure: false
```

## Options

### address

{{< confkey type="string" default="time.cloudflare.com:123" required="no" >}}

Determines the address of the NTP server to retrieve the time from. The format is `<host>:<port>`, and both of these are
required.

### version

{{< confkey type="integer" default="4" required="no" >}}

Determines the NTP version supported. Valid values are 3 or 4.

### max_desync

{{< confkey type="duration" default="3s" required="no" >}}

*__Note:__ This setting uses the [duration notation format](../prologue/common.md#duration-notation-format). Please see
the [common options](../prologue/common.md#duration-notation-format) documentation for information on this format.*

This is used to tune the acceptable desync from the time reported from the NTP server.

### disable_startup_check

{{< confkey type="boolean" default="false" required="no" >}}

Setting this to true will disable the startup check entirely.

### disable_failure

{{< confkey type="boolean" default="false" required="no" >}}

Setting this to true will allow Authelia to start and just log an error instead of exiting. The default is that if
Authelia can contact the NTP server successfully, and the time reported by the server is greater than what is configured
in [max_desync](#max_desync) that Authelia fails to start and logs a fatal error.
