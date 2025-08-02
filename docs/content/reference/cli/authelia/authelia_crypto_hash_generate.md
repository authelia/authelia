---
title: "authelia crypto hash generate"
description: "Reference for the authelia crypto hash generate command."
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

## authelia crypto hash generate

Generate cryptographic hash digests

### Synopsis

Generate cryptographic hash digests.

This subcommand allows generating cryptographic hash digests.

See the help for the subcommands if you want to override the configuration or defaults.

```
authelia crypto hash generate [flags]
```

### Examples

```
authelia crypto hash generate --help
```

### Options

```
  -h, --help                       help for generate
      --no-confirm                 skip the password confirmation prompt
      --password string            manually supply the password rather than using the terminal prompt
      --random                     uses a randomly generated password
      --random.characters string   sets the explicit characters for the random string
      --random.charset string      sets the charset for the random password, options are 'ascii', 'alphanumeric', 'alphabetic', 'numeric', 'numeric-hex', and 'rfc3986' (default "alphanumeric")
      --random.length int          sets the character length for the random string (default 72)
```

### Options inherited from parent commands

```
  -c, --config strings                        configuration files or directories to load, for more information run 'authelia -h authelia config' (default [configuration.yml])
      --config.experimental.filters strings   list of filters to apply to all configuration files, for more information run 'authelia -h authelia filters'
```

### SEE ALSO

* [authelia crypto hash](authelia_crypto_hash.md)	 - Perform cryptographic hash operations
* [authelia crypto hash generate argon2](authelia_crypto_hash_generate_argon2.md)	 - Generate cryptographic Argon2 hash digests
* [authelia crypto hash generate bcrypt](authelia_crypto_hash_generate_bcrypt.md)	 - Generate cryptographic bcrypt hash digests
* [authelia crypto hash generate pbkdf2](authelia_crypto_hash_generate_pbkdf2.md)	 - Generate cryptographic PBKDF2 hash digests
* [authelia crypto hash generate scrypt](authelia_crypto_hash_generate_scrypt.md)	 - Generate cryptographic scrypt hash digests
* [authelia crypto hash generate sha2crypt](authelia_crypto_hash_generate_sha2crypt.md)	 - Generate cryptographic SHA2 Crypt hash digests

