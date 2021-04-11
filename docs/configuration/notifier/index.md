---
layout: default
title: Notifier
parent: Configuration
nav_order: 6
has_children: true
---

# Notifier

**Authelia** sometimes needs to send messages to users in order to
verify their identity.

## Configuration

```yaml
notifier:
  disable_startup_check: false
  filesystem: {}
  smtp: {}
```

## Options

### disable_startup_check
<div markdown="1">
type: boolean
{: .label .label-config .label-purple }
default: false
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

The notifier has a startup check which validates the specified provider
configuration is correct and will be able to send emails. This can be
disabled with the `disable_startup_check` option:

### filesystem

The [filesystem](filesystem.md) provider.

### smtp

The [smtp](smtp.md) provider.
