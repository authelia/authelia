---
title: "NTP"
description: "Configuring the NTP Settings."
summary: "Authelia checks the system time is in sync with an NTP server. This section describes how to configure and tune this."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 199300
toc: true
aliases:
  - /docs/configuration/ntp.html
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

Authelia has the ability to check the system time against an NTP server, which at the present time is checked only
during startup. This section configures and tunes the settings for this check.

In the instance of inability to contact the NTP server or an issue with the synchronization Authelia will fail to start
unless configured otherwise. It should however be noted that disabling this check is not a supported configuration and
instead administrators should correct the underlying time issue. i.e. if this check is disabled and a service reliant on
the time being accurate has a failure, it's very unlikely we will produce/accept a fix in this scenario without
additional benefits.

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
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

{{< confkey type="string" syntax="address" default="udp://time.cloudflare.com:123" required="no" >}}

Configures the address for the NTP Server. The address itself is a connector and the scheme must be `udp`,
`udp4`, or `udp6`.

__Examples:__

```yaml {title="configuration.yml"}
ntp:
  address: 'udp://127.0.0.1:123'
```

```yaml {title="configuration.yml"}
ntp:
  address: 'udp6://[fd00:1111:2222:3333::1]:123'
```

### version

{{< confkey type="integer" default="4" required="no" >}}

Determines the NTP version supported. Valid values are 3 or 4.

### max_desync

{{< confkey type="string,integer" syntax="duration" default="3 seconds" required="no" >}}

This is used to tune the acceptable desync from the time reported from the NTP server.

### disable_startup_check

{{< confkey type="boolean" default="false" required="no" >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
Administrators are strongly urged to fix the underlying time issue instead of utilizing this
option. See the [FAQ](#why-should-this-check-not-be-disabled) for more information.
{{< /callout >}}

Setting this to true will disable the startup check entirely.

### disable_failure

{{< confkey type="boolean" default="false" required="no" >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
Administrators are strongly urged to fix the underlying time issue instead of utilizing this
option. See the [FAQ](#why-should-this-check-not-be-disabled) for more information.
{{< /callout >}}

Setting this to true will allow Authelia to start and just log an error instead of exiting. The default is that if
Authelia can contact the NTP server successfully, and the time reported by the server is greater than what is configured
in [max_desync](#max_desync) that Authelia fails to start and logs a fatal error.


## Frequently Asked Questions

This section acts as a frequently asked questions for the NTP behavior and configuration.

### Why is this check important and enabled by default?

This check is essential to validate the system time is accurate which ensures the following:

- The [Session](../session/introduction.md) cookie expiration times are accurately set which is important because:
  - If the time is too far in the past sessions could:
    - Be considered already expired by browsers leading to strange redirect issues.
    - Be considered expired by browsers much sooner than intended.
  - If the time is too far into the future sessions could:
    - Be considered expired by browsers much later than intended.
- The [OpenID Connect](../identity-providers/openid-connect/provider.md) JWT issued at/not before/expiration times are
  set correctly which is important because:
  - If the time is too far in the past the OpenID Connect issued JWT's could:
    - Be considered already expired by relying parties at the time of issue.
    - Be considered expired by relying parties much sooner than intended.
  - If the time is too far into the future OpenID Connect issued JWT's could:
    - Be considered invalid by correctly configured relying parties as the issue time is too far in the future.
    - Be considered invalid by badly configured relying parties much later than intended.
- The [TOTP](../second-factor/time-based-one-time-password.md) verification codes could:
  - Be considered invalid when they are technically correct.

### Why should this check not be disabled?

Due to the fact this can affect elements such as the JWT validity and session validity it's important for security this
check is operational.
