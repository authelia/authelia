---
title: "authelia crypto certificate ecdsa request"
description: "Reference for the authelia crypto certificate ecdsa request command."
lead: ""
date: 2022-06-27T18:27:57+10:00
draft: false
images: []
menu:
  reference:
    parent: "cli-authelia"
weight: 905
toc: true
---

## authelia crypto certificate ecdsa request

Generate an ECDSA private key and certificate signing request

### Synopsis

Generate an ECDSA private key and certificate signing request.

This subcommand allows generating an ECDSA private key and certificate signing request.

```
authelia crypto certificate ecdsa request [flags]
```

### Examples

```
authelia crypto certificate ecdsa request --help
```

### Options

```
  -n, --common-name string            certificate common name
      --country strings               certificate country
  -b, --curve string                  Sets the elliptic curve which can be P224, P256, P384, or P521 (default "P256")
  -d, --directory string              directory where the generated keys, certificates, etc will be stored
      --duration string               duration of time the certificate is valid for (default "1y")
      --file.csr string               name of the file to export the certificate request data to (default "request.csr")
      --file.private-key string       name of the file to export the private key data to (default "private.pem")
  -h, --help                          help for request
  -l, --locality strings              certificate locality
      --not-after string              latest date and time the certificate is considered valid in various formats
      --not-before string             earliest date and time the certificate is considered valid in various formats (default is now)
  -o, --organization strings          certificate organization (default [Authelia])
      --organizational-unit strings   certificate organizational unit
      --pkcs8                         force PKCS #8 ASN.1 format
  -p, --postcode strings              certificate postcode
      --province strings              certificate province
      --sans strings                  subject alternative names
      --signature string              signature algorithm for the certificate (default "SHA256")
  -s, --street-address strings        certificate street address
```

### Options inherited from parent commands

```
  -c, --config strings                        configuration files or directories to load, for more information run 'authelia -h authelia config' (default [configuration.yml])
      --config.experimental.filters strings   list of filters to apply to all configuration files, for more information run 'authelia -h authelia filters'
```

### SEE ALSO

* [authelia crypto certificate ecdsa](authelia_crypto_certificate_ecdsa.md)	 - Perform ECDSA certificate cryptographic operations

