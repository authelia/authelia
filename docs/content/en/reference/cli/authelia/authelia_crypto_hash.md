---
title: "authelia crypto hash"
description: "Reference for the authelia crypto hash command."
lead: ""
date: 2022-10-17T21:51:59+11:00
draft: false
images: []
menu:
  reference:
    parent: "cli-authelia"
weight: 905
toc: true
---

## authelia crypto hash

Perform cryptographic hash operations

### Synopsis

Perform cryptographic hash operations.

This subcommand allows preforming hashing cryptographic tasks.

### Examples

```
authelia crypto hash --help
```

### Options

```
  -h, --help   help for hash
```

### Options inherited from parent commands

```
  -c, --config strings                        configuration files or directories to load (default [configuration.yml])
      --config.experimental.filters strings   list of filters to apply to all configuration files, the filters are applied after loading them from disk and before parsing their content, options are 'template', 'expand-env'
```

### SEE ALSO

* [authelia crypto](authelia_crypto.md)	 - Perform cryptographic operations
* [authelia crypto hash generate](authelia_crypto_hash_generate.md)	 - Generate cryptographic hash digests
* [authelia crypto hash validate](authelia_crypto_hash_validate.md)	 - Perform cryptographic hash validations

