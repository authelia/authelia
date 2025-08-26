---
title: "authelia crypto certificate ed25519 request"
description: "Reference for the authelia crypto certificate ed25519 request command."
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

## authelia crypto certificate ed25519 request

Generate an Ed25519 private key and certificate signing request

### Synopsis

Generate an Ed25519 private key and certificate signing request.

This subcommand allows generating an Ed25519 private key and certificate signing request.

```
authelia crypto certificate ed25519 request [flags]
```

### Examples

```
authelia crypto certificate ed25519 request --help
```

### Options

```
  -n, --common-name string             certificate common name
      --country strings                certificate country
  -d, --directory string               directory where the generated keys, certificates, etc will be stored
      --duration string                duration of time the certificate is valid for (default "1y")
      --file.csr string                name of the file to export the certificate request data to (default "request.csr")
      --file.extension.legacy string   string to include before the actual extension as a sub-extension on the PKCS#1 and SECG1 legacy formats (default "legacy")
      --file.private-key string        name of the file to export the private key data to (default "private.pem")
  -h, --help                           help for request
      --legacy                         enables the output of the legacy PKCS#1 and SECG1 formats when enabled
  -l, --locality strings               certificate locality
      --not-after string               latest date and time the certificate is considered valid in various formats
      --not-before string              earliest date and time the certificate is considered valid in various formats (default is now)
  -o, --organization strings           certificate organization (default [Authelia])
      --organizational-unit strings    certificate organizational unit
  -p, --postcode strings               certificate postcode
      --province strings               certificate province
      --sans strings                   subject alternative names
      --signature string               signature algorithm for the certificate (default "SHA256")
  -s, --street-address strings         certificate street address
```

### Options inherited from parent commands

```
  -c, --config strings                        configuration files or directories to load, for more information run 'authelia -h authelia config' (default [configuration.yml])
      --config.experimental.filters strings   list of filters to apply to all configuration files, for more information run 'authelia -h authelia filters'
```

### SEE ALSO

* [authelia crypto certificate ed25519](authelia_crypto_certificate_ed25519.md)	 - Perform Ed25519 certificate cryptographic operations

