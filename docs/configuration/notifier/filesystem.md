---
layout: default
title: Filesystem
parent: Notifier
grand_parent: Configuration
nav_order: 1
---

# Filesystem

With this configuration, the message will be sent to a file. This option
should only be used for testing purposes.

```yaml
# Configuration of the notification system.
#
# Notifications are sent to users when they require a password reset, a u2f
# registration or a TOTP registration.
# Use only an available configuration: filesystem, gmail
notifier:
  # You can disable the notifier startup check by setting this to true
  disable_startup_check: false

  # For testing purpose, notifications can be sent in a file
  filesystem:
    filename: /tmp/authelia/notification.txt
```