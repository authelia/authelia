---
title: "authelia crypto pair ed25519"
description: "Reference for the authelia crypto pair ed25519 command."
lead: ""
date: 2024-03-14T06:00:14+11:00
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

## authelia crypto pair ed25519

Perform Ed25519 key pair cryptographic operations

### Synopsis

Perform Ed25519 key pair cryptographic operations.

This subcommand allows performing Ed25519 key pair cryptographic tasks.

```
authelia crypto pair ed25519 [flags]
```

### Examples

```
authelia crypto pair ed25519 --help
```

### Options

```
  -h, --help   help for ed25519
```

### Options inherited from parent commands

```
  -c, --config strings                        configuration files or directories to load, for more information run 'authelia -h authelia config' (default [configuration.yml])
      --config.experimental.filters strings   list of filters to apply to all configuration files, for more information run 'authelia -h authelia filters'
```

### SEE ALSO

* [authelia crypto pair](authelia_crypto_pair.md)	 - Perform key pair cryptographic operations
* [authelia crypto pair ed25519 generate](authelia_crypto_pair_ed25519_generate.md)	 - Generate a cryptographic Ed25519 key pair

