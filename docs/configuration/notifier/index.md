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

## Startup Check

The notifier has a startup check which validates the specified provider
configuration is correct and will be able to send emails. This can be
disabled with the `disable_startup_check` option:

```yaml
notifier:
  disable_startup_check: false
```
