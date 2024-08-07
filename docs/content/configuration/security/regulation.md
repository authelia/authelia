---
title: "Regulation"
description: "Regulation Configuration"
summary: "Configuring the Regulation system."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
weight: 104300
toc: true
aliases:
  - /docs/configuration/regulation.html
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---


__Authelia__ can temporarily ban accounts when there are too many
authentication attempts. This helps prevent brute-force attacks.

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
regulation:
  mode: 'user'
  max_retries: 3
  find_time: '2m'
  ban_time: '5m'
```

## Options

This section describes the individual configuration options.

### mode

{{< confkey type="string" default="user" required="no" >}}

The mode for regulation. The table below describes each option. The recommended mode is `ip`. It should be noted that
regardless of the ban mode if bans exist in the database the user or ip will be denied access. See the
[authelia storage bans](../../reference/cli/authelia/authelia_storage_bans.md) command for information on managing
ban entries.

| Mode |                             Description                             |
|:----:|:-------------------------------------------------------------------:|
| user |        The user account is the subject of any automatic bans        |
|  ip  |         The remote ip is the subject of any automatic bans          |
| both | The user account and remote ip is the subject of any automatic bans |

### max_retries

{{< confkey type="integer" default="3" required="no" >}}

The number of failed login attempts before a user may be banned. Setting this option to 0 disables regulation entirely.

### find_time

{{< confkey type="string,integer" syntax="duration" default="2 minutes" required="no" >}}

The period of time analyzed for failed attempts. For
example if you set `max_retries` to 3 and `find_time` to `2m` this means the user must have 3 failed logins in
2 minutes.

### ban_time

{{< confkey type="string,integer" syntax="duration" default="5 minutes" required="no" >}}

The period of time the user is banned for after meeting the `max_retries` and `find_time` configuration. After this
duration the account will be able to login again.
