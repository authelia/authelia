---
title: "Generating Secure Values"
description: "A reference guide on generating secure values such as password hashes, password strings, and cryptography keys"
summary: "This section contains reference documentation for generating secure values such as password hashes, password strings, and cryptography keys."
date: 2022-10-23T18:09:19+11:00
draft: false
images: []
weight: 220
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Generating a Random Password Hash

Often times it's required that a random password is generated. While you could randomly generate a string then hash it,
we provide a convenience layer for this purpose.

### authelia

The __Authelia__ docker container or CLI binary can be used to generate a random alphanumeric string and output the
string and the hash at the same time.

Use the `authelia crypto hash generate --help` command or see the [authelia crypto hash generate] reference guide for
more information on all available options and algorithms.

##### Using Docker

```bash
docker run authelia/authelia:latest authelia crypto hash generate argon2 --random --random.length 64 --random.charset alphanumeric
```

##### Using the Binary

```bash
authelia crypto hash generate argon2 --random --random.length 64 --random.charset alphanumeric
```

## Generating a Random Alphanumeric String

Some sections of the configuration recommend generating a random string. There are many ways to accomplish this and the
following methods are merely suggestions.

### authelia

The __Authelia__ docker container or CLI binary can be used to generate a random alphanumeric string.

Use the `authelia crypto rand --help` command or see the [authelia crypto rand] reference guide for more information on
all available options.

##### Using Docker

```bash
docker run authelia/authelia:latest authelia crypto rand --length 64 --charset alphanumeric
```

##### Using the Binary

```bash
authelia crypto rand --length 64 --charset alphanumeric
```

### openssl

The `openssl` command on Linux can be used to generate a random alphanumeric string:

```bash
openssl rand -hex 64
```

### Linux

Basic Linux commands can be used to generate a random alphanumeric string:

```bash
LENGTH=64
tr -cd '[:alnum:]' < /dev/urandom | fold -w "${LENGTH}" | head -n 1 | tr -d '\n' ; echo
```

## Generating an RSA Keypair

Some sections of the configuration need an RSA keypair. There are many ways to achieve this, this section explains two
such ways.

### authelia

The __Authelia__ docker container or CLI binary can be used to generate a RSA 4096 bit keypair.

Use the `authelia crypto pair --help` command or see the [authelia crypto pair] reference guide for more
information on all available options.

##### Using Docker

```bash
docker run -u "$(id -u):$(id -g)" -v "$(pwd)":/keys authelia/authelia:latest authelia crypto pair rsa generate --bits 4096 --directory /keys
```

##### Using the Binary

```bash
authelia crypto pair rsa generate --directory /path/to/keys
```

### openssl

The `openssl` command on Linux can be used to generate a RSA 4096 bit keypair:

```bash
openssl genrsa -out private.pem 4096
openssl rsa -in private.pem -outform PEM -pubout -out public.pem
```

## Generating an RSA Self-Signed Certificate

Some sections of the configuration need a certificate and it may be possible to use a self-signed certificate. There are
many ways to achieve this, this section explains two such ways.

### authelia

The __Authelia__ docker container or binary can be used to generate a RSA 4096 bit self-signed certificate for the
domain `example.com`.

Use the `authelia crypto certificate --help` command or see the [authelia crypto certificate] reference guide for more
information on all available options.

##### Using Docker

```bash
docker run -u "$(id -u):$(id -g)" -v "$(pwd)":/keys authelia/authelia authelia crypto certificate rsa generate --common-name example.com --directory /keys
```

##### Using the Binary

```bash
authelia crypto certificate rsa generate --common-name example.com --directory /path/to/keys
```

### openssl

The `openssl` command on Linux can be used to generate a RSA 4096 bit self-signed certificate for the domain
`example.com`:

```bash
openssl req -x509 -nodes -newkey rsa:4096 -keyout key.pem -out cert.pem -sha256 -days 365 -subj '/CN=example.com'
```

[authelia crypto hash generate]: ../cli/authelia/authelia_crypto_hash_generate.md
[authelia crypto rand]: ../cli/authelia/authelia_crypto_rand.md
[authelia crypto pair]: ../cli/authelia/authelia_crypto_pair.md
[authelia crypto certificate]: ../cli/authelia/authelia_crypto_certificate.md
