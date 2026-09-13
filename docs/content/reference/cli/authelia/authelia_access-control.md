---
title: "authelia access-control"
description: "Reference for the authelia access-control command."
lead: ""
date: 2026-09-12T18:33:16+10:00
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
  -c, --config strings                                   configuration files or directories to load, for more information run 'authelia -h authelia config' (default [configuration.yml])
      --config.filters strings                           list of filters to apply to all configuration files, for more information run 'authelia -h authelia filters'
      --config.filters.template.delimiter.left string    sets the left delimiter for the 'template' filter
      --config.filters.template.delimiter.right string   sets the right delimiter for the 'template' filter
```

### SEE ALSO

* [authelia](authelia.md)	 - authelia untagged-unknown-dirty (master, unknown)
* [authelia access-control check-policy](authelia_access-control_check-policy.md)	 - Checks a request against the access control rules to determine what policy would be applied

