---
title: "File"
description: "File Session Configuration"
summary: "Configuring the File Session Storage."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 106300
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

This is a session provider. By default Authelia uses an in-memory provider. The file provider stores sessions as
individual files on the local filesystem. This is suitable for single-instance deployments where running Redis would be
unnecessary overhead. Like the memory provider, the file provider is
[stateful](../../overview/authorization/statelessness.md) and is not suitable for high availability scenarios.

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
session:
  secret: 'insecure_session_secret'
  file:
    path: '/config/sessions'
    cleanup_interval: '5m'
```

## Options

This section describes the individual configuration options.

### path

{{< confkey type="string" required="yes" >}}

The directory path where session files are stored. This must be an absolute path.

The directory will be created automatically if it does not exist. For security, the directory should have permissions
`0700` (owner read/write/execute only). A warning is logged if the directory has more permissive permissions.

### cleanup_interval

{{< confkey type="string,integer" syntax="duration" default="5 minutes" required="no" >}}

The interval between automatic cleanup runs that remove expired session files and orphaned temporary files.
