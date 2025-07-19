---
title: "authelia debug oidc"
description: "Reference for the authelia debug oidc command."
lead: ""
date: 2025-05-10T14:52:22+10:00
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

## authelia debug oidc

Perform a OpenID Connect 1.0 debug operation

### Synopsis

Perform a OpenID Connect 1.0 debug operation.

This subcommand allows checking certain OpenID Connect 1.0 scenarios.

### Examples

```
authelia debug oidc --help
```

### Options

```
  -h, --help   help for oidc
```

### Options inherited from parent commands

```
  -c, --config strings                        configuration files or directories to load, for more information run 'authelia -h authelia config' (default [configuration.yml])
      --config.experimental.filters strings   list of filters to apply to all configuration files, for more information run 'authelia -h authelia filters'
```

### SEE ALSO

* [authelia debug](authelia_debug.md)	 - Perform debug functions
* [authelia debug oidc claims](authelia_debug_oidc_claims.md)	 - Perform a OpenID Connect 1.0 claims hydration debug operation

