---
layout: default
title: Cache
parent: Authentication Backends
grand_parent: Configuration
nav_order: 1
---

The cached configuration is a special provider that sits in front of the other configured provider. It keeps an
in-memory cache of user details for the purposes of reducing the number of calls to your backend. The details cached
are simply profile information, it does not include the password.

In addition to the configured options causing the profile to be refreshed, the cache is forcibly refreshed for a user 
anytime a user performs 1FA.

## Configuration

```yaml
authentication_backend:
  disable_reset_password: false
  cache:
    disable: false
    ttl: 5m
```

## Options

### disable
<div markdown="1">
type: boolean
{: .label .label-config .label-purple }
default: false
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

This disables the cached provider entirely if set to `true`.


### ttl
<div markdown="1">
type: string/integer (duration)
{: .label .label-config .label-purple }
default: 5m
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

The time in [duration notation format](../index.md#duration-notation-format) before the cache will consider the 
information stale and will refresh the data from the backend.

