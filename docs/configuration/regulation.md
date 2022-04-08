---
layout: default
title: Regulation
parent: Configuration
nav_order: 10
---

# Regulation

**Authelia** can temporarily ban accounts when there are too many
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
<div markdown="1">
type: integer 
{: .label .label-config .label-purple } 
default: 3
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

The number of failed login attempts before a user may be banned. Setting this option to 0 disables regulation entirely.

### find_time
<div markdown="1">
type: string (duration) 
{: .label .label-config .label-purple } 
default: 2m
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

The period of time in [duration notation format](index.md#duration-notation-format) analyzed for failed attempts. For
example if you set `max_retries` to 3 and `find_time` to `2m` this means the user must have 3 failed logins in
2 minutes.

### ban_time
<div markdown="1">
type: string (duration) 
{: .label .label-config .label-purple } 
default: 5m
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

The period of time in [duration notation format](index.md#duration-notation-format) the user is banned for after meeting
the `max_retries` and `find_time` configuration. After this duration the account will be able to login again.
