---
title: "authelia crypto hash generate sha2crypt"
description: "Reference for the authelia crypto hash generate sha2crypt command."
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

## authelia crypto hash generate sha2crypt

Generate cryptographic SHA2 Crypt hash digests

### Synopsis

Generate cryptographic SHA2 Crypt hash digests.

This subcommand allows generating cryptographic SHA2 Crypt hash digests.

```
authelia crypto hash generate sha2crypt [flags]
```

### Examples

```
authelia crypto hash generate sha2crypt --help
```

### Options

```
  -h, --help             help for sha2crypt
  -i, --iterations int   number of iterations (default 50000)
  -s, --salt-size int    salt size in bytes (default 16)
  -v, --variant string   variant, options are sha256 and sha512 (default "sha512")
```

### Options inherited from parent commands

```
  -c, --config strings                        configuration files or directories to load, for more information run 'authelia -h authelia config' (default [configuration.yml])
      --config.experimental.filters strings   list of filters to apply to all configuration files, for more information run 'authelia -h authelia filters'
      --no-confirm                            skip the password confirmation prompt
      --password string                       manually supply the password rather than using the terminal prompt
      --random                                uses a randomly generated password
      --random.characters string              sets the explicit characters for the random string
      --random.charset string                 sets the charset for the random password, options are 'ascii', 'alphanumeric', 'alphabetic', 'numeric', 'numeric-hex', and 'rfc3986' (default "alphanumeric")
      --random.length int                     sets the character length for the random string (default 72)
```

### SEE ALSO

* [authelia crypto hash generate](authelia_crypto_hash_generate.md)	 - Generate cryptographic hash digests

