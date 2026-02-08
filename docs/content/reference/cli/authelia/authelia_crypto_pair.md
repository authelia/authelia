---
title: "authelia crypto pair"
description: "Reference for the authelia crypto pair command."
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

## authelia crypto pair

Perform key pair cryptographic operations

### Synopsis

Perform key pair cryptographic operations.

This subcommand allows performing key pair cryptographic tasks.

### Examples

```
authelia crypto pair --help
```

### Options

```
  -h, --help   help for pair
```

### Options inherited from parent commands

```
  -c, --config strings                        configuration files or directories to load, for more information run 'authelia -h authelia config' (default [configuration.yml])
      --config.experimental.filters strings   list of filters to apply to all configuration files, for more information run 'authelia -h authelia filters'
```

### SEE ALSO

* [authelia crypto](authelia_crypto.md)	 - Perform cryptographic operations
* [authelia crypto pair ecdsa](authelia_crypto_pair_ecdsa.md)	 - Perform ECDSA key pair cryptographic operations
* [authelia crypto pair ed25519](authelia_crypto_pair_ed25519.md)	 - Perform Ed25519 key pair cryptographic operations
* [authelia crypto pair rsa](authelia_crypto_pair_rsa.md)	 - Perform RSA key pair cryptographic operations

