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
  -c, --config strings                        configuration files to load (default [configuration.yml])
      --config.directory string               path to a directory with yml/yaml files to load as part of the configuration
      --config.experimental.filters strings   Applies filters in order to the configuration file before the YAML parser. Options are 'template', 'expand-env'
```

### SEE ALSO

* [authelia crypto](authelia_crypto.md)	 - Perform cryptographic operations
* [authelia crypto hash generate](authelia_crypto_hash_generate.md)	 - Generate cryptographic hash digests
* [authelia crypto hash validate](authelia_crypto_hash_validate.md)	 - Perform cryptographic hash validations

