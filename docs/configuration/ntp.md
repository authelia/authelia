---
layout: default
title: NTP
parent: Configuration
nav_order: 9
---

# NTP

Authelia has the ability to check the system time against an NTP server. Currently this only occurs at startup. This
section configures and tunes the settings for this check which is primarily used to ensure [TOTP](./one-time-password.md)
can be accurately validated.

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
<div markdown="1">
type: string
{: .label .label-config .label-purple } 
default: time.cloudflare.com:123
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Determines the address of the NTP server to retrieve the time from. The format is `<host>:<port>`, and both of these are
required.

### version
<div markdown="1">
type: integer
{: .label .label-config .label-purple } 
default: 4
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Determines the NTP verion supported. Valid values are 3 or 4.

### max_desync
<div markdown="1">
type: duration
{: .label .label-config .label-purple } 
default: 3s
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

This is used to tune the acceptable desync from the time reported from the NTP server. This uses our 
[duration notation](./index.md#duration-notation-format) format.

### disable_startup_check
<div markdown="1">
type: boolean
{: .label .label-config .label-purple } 
default: false
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Setting this to true will disable the startup check entirely.

### disable_failure
<div markdown="1">
type: boolean
{: .label .label-config .label-purple } 
default: false
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Setting this to true will allow Authelia to start and just log an error instead of exiting. The default is that if
Authelia can contact the NTP server successfully, and the time reported by the server is greater than what is configured
in [max_desync](#max_desync) that Authelia fails to start and logs a fatal error.