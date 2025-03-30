---
title: "authelia debug"
description: "Reference for the authelia debug command."
lead: ""
date: 2026-09-12T18:52:11+10:00
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

## authelia debug

Perform debug functions

### Synopsis

Perform debug related functions.

This subcommand contains other subcommands related to debugging.

### Examples

```
authelia debug --help
```

### Options

```
  -h, --help   help for debug
```

### Options inherited from parent commands

```
  -c, --config strings                        configuration files or directories to load, for more information run 'authelia -h authelia config' (default [configuration.yml])
      --config.experimental.filters strings   list of filters to apply to all configuration files, for more information run 'authelia -h authelia filters'
      --config.filters.values strings         file paths of values files (.yml, .yaml, .json, .toml) to utilize with configuration file filters; files are loaded in order with later files deep-merged on top, for more information run 'authelia -h authelia filters'
```

### SEE ALSO

* [authelia](authelia.md)	 - authelia untagged-unknown-dirty (master, unknown)
* [authelia debug expression](authelia_debug_expression.md)	 - Perform a user attribute expression debug operation
* [authelia debug oidc](authelia_debug_oidc.md)	 - Perform a OpenID Connect 1.0 debug operation
* [authelia debug tls](authelia_debug_tls.md)	 - Perform a TLS debug operation

