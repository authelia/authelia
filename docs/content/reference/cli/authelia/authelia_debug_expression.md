---
title: "authelia debug expression"
description: "Reference for the authelia debug expression command."
lead: ""
date: 2025-08-01T16:23:47+10:00
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

## authelia debug expression

Perform a user attribute expression debug operation

### Synopsis

Perform a user attribute expression debug operation.

This subcommand allows checking a user attribute expression against a specific user.

```
authelia debug expression <username> <expression> [flags]
```

### Examples

```
authelia debug expression username "'abc' in groups"
```

### Options

```
  -h, --help   help for expression
```

### Options inherited from parent commands

```
  -c, --config strings                        configuration files or directories to load, for more information run 'authelia -h authelia config' (default [configuration.yml])
      --config.experimental.filters strings   list of filters to apply to all configuration files, for more information run 'authelia -h authelia filters'
```

### SEE ALSO

* [authelia debug](authelia_debug.md)	 - Perform debug functions

