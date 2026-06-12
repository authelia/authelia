---
title: "File System"
description: "Configuring the File Notifications Settings."
summary: "Authelia can save notifications to a file. This section describes how to configure this."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 108300
toc: true
aliases:
  - /docs/configuration/notifier/filesystem.html
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

It is recommended in a production environment that you do not use the file notification system, and that it should only
be used for testing purposes. See one of [the other methods](introduction.md) for a production ready solution.

This method will use the plain text email template for readability purposes.

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
notifier:
  disable_startup_check: false
  filesystem:
    filename: '/config/notification.txt'
```

## Options

This section describes the individual configuration options.

### filename

{{< confkey type="string" required="yes" >}}

The file to add email text to. If it doesn't exist it will be created.

If `filename` points to an existing FIFO (named pipe), Authelia detects this from the file mode and writes notifications
to the pipe instead of a regular file. The startup check leaves the FIFO untouched and the per-notification write skips
the disk sync that is unsupported on pipes. The byte stream written to the pipe is identical to the regular-file format,
so any tooling that already parses notifications continues to work.

The FIFO must be created by the operator with `mkfifo` before Authelia starts; Authelia will not create one. A reader
must be attached on the other end of the pipe. If no reader is connected when a notification is generated, the write
blocks until one is.

Use the `authelia debug notification` subcommand to verify the configured notifier (file, FIFO, or SMTP) without
exercising the full identity verification flow.
