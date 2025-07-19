---
title: "authelia crypto pair ecdsa generate"
description: "Reference for the authelia crypto pair ecdsa generate command."
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

## authelia crypto pair ecdsa generate

Generate a cryptographic ECDSA key pair

### Synopsis

Generate a cryptographic ECDSA key pair.

This subcommand allows generating an ECDSA key pair.

```
authelia crypto pair ecdsa generate [flags]
```

### Examples

```
authelia crypto pair ecdsa generate --help
```

### Options

```
  -b, --curve string                   Sets the elliptic curve which can be P224, P256, P384, or P521 (default "P256")
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

* [authelia crypto pair ecdsa](authelia_crypto_pair_ecdsa.md)	 - Perform ECDSA key pair cryptographic operations

