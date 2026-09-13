---
title: "Administration"
description: "Administration Configuration."
summary: "This describes a section of the configuration for enabling administrator user and group management features."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 199500
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
administration:
  enabled: false
  admin_group: 'authelia-admin'
  enable_user_management: false
  allow_admins_to_add_admins: false
```

## Options

This section describes the individual configuration options.

### enabled

{{< confkey type="boolean" default="false" required="no" >}}

Enables administration interface features.

### admin_group

{{< confkey type="string" default="authelia-admin" required="no" >}}

The group that will allow a user to access Authelia administration features.

### enable_user_management

{{< confkey type="boolean" default="false" required="no" >}}

Enables the user management interface for users with the [admin_group](#admin_group) group.

### allow_admins_to_add_admins

{{< confkey type="boolean" default="false" required="no" >}}

Allows admins with the [admin_group](#admin_group) group to give the admin group to other users.
