---
title: "authelia crypto pair ecdsa"
description: "Reference for the authelia crypto pair ecdsa command."
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

## authelia crypto pair ecdsa

Perform ECDSA key pair cryptographic operations

### Synopsis

Perform ECDSA key pair cryptographic operations.

This subcommand allows performing ECDSA key pair cryptographic tasks.

```
authelia crypto pair ecdsa [flags]
```

### Examples

```
authelia crypto pair ecdsa --help
```

### Options

```
  -h, --help   help for ecdsa
```

### Options inherited from parent commands

```
  -c, --config strings                                   configuration files or directories to load, for more information run 'authelia -h authelia config' (default [configuration.yml])
      --config.filters strings                           list of filters to apply to all configuration files, for more information run 'authelia -h authelia filters'
      --config.filters.template.delimiter.left string    sets the left delimiter for the 'template' filter
      --config.filters.template.delimiter.right string   sets the right delimiter for the 'template' filter
```

### SEE ALSO

* [authelia crypto pair](authelia_crypto_pair.md)	 - Perform key pair cryptographic operations
* [authelia crypto pair ecdsa generate](authelia_crypto_pair_ecdsa_generate.md)	 - Generate a cryptographic ECDSA key pair

