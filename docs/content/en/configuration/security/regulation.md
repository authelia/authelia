---
title: "Regulation"
description: "Regulation Configuration"
lead: "Configuring the Regulation system."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  configuration:
    parent: "security"
weight: 104300
toc: true
aliases:
  - /docs/configuration/regulation.html
---


__Authelia__ can temporarily ban accounts when there are too many
authentication attempts. This helps prevent brute-force attacks.

## Configuration

```yaml
regulation:
  max_retries: 3
  find_time: 2m
  ban_time: 5m
```

## Options

### max_retries

{{< confkey type="integer " default="3" required="no" >}}

The number of failed login attempts before a user may be banned. Setting this option to 0 disables regulation entirely.

### find_time

{{< confkey type="duration " default="2m" required="no" >}}

*__Note:__ This setting uses the [duration notation format](../prologue/common.md#duration-notation-format). Please see
the [common options](../prologue/common.md#duration-notation-format) documentation for information on this format.*

The period of time analyzed for failed attempts. For
example if you set `max_retries` to 3 and `find_time` to `2m` this means the user must have 3 failed logins in
2 minutes.

### ban_time

{{< confkey type="duration" default="5m" required="no" >}}

*__Note:__ This setting uses the [duration notation format](../prologue/common.md#duration-notation-format). Please see
the [common options](../prologue/common.md#duration-notation-format) documentation for information on this format.*

The period of time the user is banned for after meeting the `max_retries` and `find_time` configuration. After this
duration the account will be able to login again.
