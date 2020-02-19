---
layout: default
title: Filesystem
parent: Notifier
grand_parent: Configuration
nav_order: 1
---

# Filesystem

With this configuration, the message will be sent to a file. This option
should be used only for testing purpose.

```yaml
notifier:
    filesystem:
        filename: /tmp/authelia/notification.txt
```