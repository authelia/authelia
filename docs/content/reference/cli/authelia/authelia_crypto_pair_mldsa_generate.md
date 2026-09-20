---
title: "authelia crypto pair mldsa generate"
description: "Reference for the authelia crypto pair mldsa generate command."
lead: ""
date: 2026-08-29T18:29:41+10:00
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

## authelia crypto pair mldsa generate

Generate a cryptographic ML-DSA key pair

### Synopsis

Generate a cryptographic ML-DSA key pair.

This subcommand allows generating an ML-DSA key pair.

```
authelia crypto pair mldsa generate [flags]
```

### Examples

```
authelia crypto pair mldsa generate --help
```

### Options

```
  -d, --directory string               directory where the generated keys, certificates, etc will be stored
      --file.extension.legacy string   string to include before the actual extension as a sub-extension on the PKCS#1 and SECG1 legacy formats (default "legacy")
      --file.private-key string        name of the file to export the private key data to (default "private.pem")
      --file.public-key string         name of the file to export the public key data to (default "public.pem")
  -h, --help                           help for generate
      --legacy                         enables the output of the legacy PKCS#1 and SECG1 formats when enabled
  -b, --parameters string              Sets the ML-DSA parameter set which can be ML-DSA-44, ML-DSA-65, or ML-DSA-87 (default "ML-DSA-65")
```

### Options inherited from parent commands

```
  -c, --config strings                        configuration files or directories to load, for more information run 'authelia -h authelia config' (default [configuration.yml])
      --config.experimental.filters strings   list of filters to apply to all configuration files, for more information run 'authelia -h authelia filters'
```

### SEE ALSO

* [authelia crypto pair mldsa](authelia_crypto_pair_mldsa.md)	 - Perform ML-DSA key pair cryptographic operations

