---
title: "authelia crypto hash"
description: "Reference for the authelia crypto hash command."
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

## authelia crypto hash

Perform cryptographic hash operations

### Synopsis

Perform cryptographic hash operations.

This subcommand allows performing hashing cryptographic tasks.

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
  -c, --config strings                        configuration files or directories to load, for more information run 'authelia -h authelia config' (default [configuration.yml])
      --config.experimental.filters strings   list of filters to apply to all configuration files, for more information run 'authelia -h authelia filters'
```

### SEE ALSO

* [authelia crypto](authelia_crypto.md)	 - Perform cryptographic operations
* [authelia crypto hash generate](authelia_crypto_hash_generate.md)	 - Generate cryptographic hash digests
* [authelia crypto hash validate](authelia_crypto_hash_validate.md)	 - Perform cryptographic hash validations

