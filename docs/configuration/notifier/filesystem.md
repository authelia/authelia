---
layout: default
title: Filesystem
parent: Notifier
grand_parent: Configuration
nav_order: 1
---

# Filesystem

With this configuration, the message will be sent to a file. This option should only be used for testing purposes. 
This method will use the plain text email template for readability purposes.

## Configuration

```yaml
notifier:
  disable_startup_check: false
  filesystem:
    filename: /config/notification.txt
```

## Options

### filename

The file to write email text to. If it doesn't exist it will be created.
