---
title: "authelia debug notification"
description: "Reference for the authelia debug notification command."
lead: ""
date: 2026-04-30T14:44:59+10:00
draft: false
images: []
weight: 905
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## authelia debug notification

Perform a notifier debug operation

### Synopsis

Perform a notifier debug operation.

This subcommand loads the Authelia configuration, runs the notifier startup check, and dispatches a single test notification. It is useful for verifying that the SMTP server, filesystem path, or named-pipe consumer is reachable.

```
authelia debug notification [flags]
```

### Examples

```
authelia debug notification --recipient admin@example.com --subject "Test"
```

### Options

```
  -h, --help               help for notification
      --recipient string   recipient email address used for the test notification (default "test@example.com")
      --subject string     subject line for the test notification (default "Authelia notifier debug")
```

### Options inherited from parent commands

```
  -c, --config strings                        configuration files or directories to load, for more information run 'authelia -h authelia config' (default [configuration.yml])
      --config.experimental.filters strings   list of filters to apply to all configuration files, for more information run 'authelia -h authelia filters'
```

### SEE ALSO

* [authelia debug](authelia_debug.md)	 - Perform debug functions

