---
title: "authelia crypto pair ed25519 generate"
description: "Reference for the authelia crypto pair ed25519 generate command."
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

## authelia crypto pair ed25519 generate

Generate a cryptographic Ed25519 key pair

### Synopsis

Generate a cryptographic Ed25519 key pair.

This subcommand allows generating an Ed25519 key pair.

```
authelia crypto pair ed25519 generate [flags]
```

### Examples

```
authelia crypto pair ed25519 generate --help
```

### Options

```
  -d, --directory string               directory where the generated keys, certificates, etc will be stored
      --file.extension.legacy string   string to include before the actual extension as a sub-extension on the PKCS#1 and SECG1 legacy formats (default "legacy")
      --file.private-key string        name of the file to export the private key data to (default "private.pem")
      --file.public-key string         name of the file to export the public key data to (default "public.pem")
  -h, --help                           help for generate
      --legacy                         enables the output of the legacy PKCS#1 and SECG1 formats when enabled
```

### Options inherited from parent commands

```
  -c, --config strings                        configuration files or directories to load, for more information run 'authelia -h authelia config' (default [configuration.yml])
      --config.experimental.filters strings   list of filters to apply to all configuration files, for more information run 'authelia -h authelia filters'
```

### SEE ALSO

* [authelia crypto pair ed25519](authelia_crypto_pair_ed25519.md)	 - Perform Ed25519 key pair cryptographic operations

