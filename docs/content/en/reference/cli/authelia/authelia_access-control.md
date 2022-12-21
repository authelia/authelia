---
title: "authelia access-control"
description: "Reference for the authelia access-control command."
lead: ""
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  reference:
    parent: "cli-authelia"
weight: 905
toc: true
---

## authelia access-control

Helpers for the access control system

### Synopsis

Helpers for the access control system.

### Examples

```
authelia access-control --help
```

### Options

```
  -h, --help   help for access-control
```

### Options inherited from parent commands

```
  -c, --config strings                        configuration files to load (default [configuration.yml])
      --config.experimental.filters strings   applies filters in order to the configuration file before the YAML parser, options are 'template', 'expand-env'
```

### SEE ALSO

* [authelia](authelia.md)	 - authelia untagged-unknown-dirty (master, unknown)
* [authelia access-control check-policy](authelia_access-control_check-policy.md)	 - Checks a request against the access control rules to determine what policy would be applied

