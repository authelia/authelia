---
title: "Notifications"
description: "Configuring the Notifications Settings."
lead: "Authelia sends messages to users in order to verify their identity. This section describes how to configure this."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  configuration:
    parent: "notifications"
weight: 107100
toc: true
---

Authelia sends messages to users in order to verify their identity.

## Configuration

```yaml
notifier:
  disable_startup_check: false
  template_path: ''
  filesystem: {}
  smtp: {}
```

## Options

### disable_startup_check

{{< confkey type="boolean" default="false" required="no" >}}

The notifier has a startup check which validates the specified provider configuration is correct and will be able to
send emails. This can be disabled with the `disable_startup_check` option.

### template_path

{{< confkey type="string" required="no" >}}

*__Note:__ you may configure this directory and add only add the templates you wish to override, any templates not
supplied in this folder will utilize the default templates.*

This option allows the administrator to set a path to a directory where custom templates for notifications can be found.
The specifics are located in the
[Notification Templates Reference Guide](../../reference/guides/notification-templates.md).

### filesystem

The [filesystem](file.md) provider.

### smtp

The [smtp](smtp.md) provider.
