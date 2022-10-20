---
title: "authelia crypto hash generate"
description: "Reference for the authelia crypto hash generate command."
lead: ""
date: 2022-10-17T21:51:59+11:00
draft: false
images: []
menu:
  reference:
    parent: "cli-authelia"
weight: 330
toc: true
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
  -c, --config strings             configuration files to load (default [configuration.yml])
  -h, --help                       help for generate
      --no-confirm                 skip the password confirmation prompt
      --password string            manually supply the password rather than using the terminal prompt
      --random                     uses a randomly generated password
      --random.characters string   sets the explicit characters for the random string
      --random.charset string      sets the charset for the random password, options are 'ascii', 'alphanumeric', 'alphabetic', 'numeric', and 'numeric-hex' (default "alphanumeric")
      --random.length int          when using a randomly generated password it configures the length (default 72)
```

### SEE ALSO

* [authelia crypto hash](authelia_crypto_hash.md)	 - Perform cryptographic hash operations
* [authelia crypto hash generate argon2](authelia_crypto_hash_generate_argon2.md)	 - Generate cryptographic Argon2 hash digests
* [authelia crypto hash generate bcrypt](authelia_crypto_hash_generate_bcrypt.md)	 - Generate cryptographic bcrypt hash digests
* [authelia crypto hash generate pbkdf2](authelia_crypto_hash_generate_pbkdf2.md)	 - Generate cryptographic PBKDF2 hash digests
* [authelia crypto hash generate scrypt](authelia_crypto_hash_generate_scrypt.md)	 - Generate cryptographic scrypt hash digests
* [authelia crypto hash generate sha2crypt](authelia_crypto_hash_generate_sha2crypt.md)	 - Generate cryptographic SHA2 Crypt hash digests

