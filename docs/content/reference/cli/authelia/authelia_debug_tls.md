---
title: "authelia debug tls"
description: "Reference for the authelia debug tls command."
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

## authelia debug tls

Perform a TLS debug operation

### Synopsis

Perform a TLS debug operation.

This subcommand allows checking a remote server's TLS configuration and the ability to validate the certificate.

```
authelia debug tls [address] [flags]
```

### Examples

```
authelia debug tls tcp://smtp.example.com:465
```

### Options

```
  -h, --help              help for tls
      --hostname string   overrides the hostname to use for the TLS connection which is usually extracted from the address
```

### Options inherited from parent commands

```
  -c, --config strings                        configuration files or directories to load, for more information run 'authelia -h authelia config' (default [configuration.yml])
      --config.experimental.filters strings   list of filters to apply to all configuration files, for more information run 'authelia -h authelia filters'
```

### SEE ALSO

* [authelia debug](authelia_debug.md)	 - Perform debug functions

