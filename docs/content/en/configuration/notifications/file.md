---
title: "File System"
description: "Configuring the File Notifications Settings."
lead: "Authelia can save notifications to a file. This section describes how to configure this."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  configuration:
    parent: "notifications"
weight: 107300
toc: true
aliases:
  - /docs/configuration/notifier/filesystem.html
---

It is recommended in a production environment that you do not use the file notification system, and that it should only
be used for testing purposes. See one of [the other methods](introduction.md) for a production ready solution.

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

{{< confkey type="string" required="yes" >}}

The file to add email text to. If it doesn't exist it will be created.
