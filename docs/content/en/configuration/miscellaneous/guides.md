---
title: "Guides"
description: "Miscellaneous Guides for Configuration."
lead: "This section contains miscellaneous guides used in the configuration."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  configuration:
    parent: "miscellaneous"
weight: 199500
toc: true
---

## Generating a Random Alphanumeric String

Some sections of the configuration recommend generating a random string. There are many ways to accomplish this, one
possible way on Linux is utilizing the following command which prints a string with a length in characters of
`${LENGTH}` to `stdout`. The string will only contain alphanumeric characters.

```bash
LENGTH=64
tr -cd '[:alnum:]' < /dev/urandom | fold -w "${LENGTH}" | head -n 1 | tr -d '\n' ; echo
```

## Generating an RSA Keypair

Some sections of the configuration need an RSA keypair. There are many ways to achieve this, this section explains two
such ways.

### openssl

The `openssl` command on Linux can be used to generate a RSA 4096 bit keypair:

```bash
openssl genrsa -out private.pem 4096
openssl rsa -in private.pem -outform PEM -pubout -out public.pem
```

### authelia

The __Authelia__ docker container or CLI binary can be used to generate a RSA 4096 bit keypair:

```bash
docker run -u "$(id -u):$(id -g)" -v "$(pwd)":/keys authelia/authelia:latest authelia crypto pair rsa generate --bits 4096 --directory /keys
```

```bash
authelia crypto pair rsa generate --directory /path/to/keys
```

## Generating an RSA Self-Signed Certificate

Some sections of the configuration need a certificate and it may be possible to use a self-signed certificate. There are
many ways to achieve this, this section explains two such ways.

### openssl

The `openssl` command on Linux can be used to generate a RSA 4096 bit self-signed certificate for the domain
`example.com`:

```bash
openssl req -x509 -nodes -newkey rsa:4096 -keyout key.pem -out cert.pem -sha256 -days 365 -subj '/CN=example.com'
```

### authelia

The __Authelia__ docker container or binary can be used to generate a RSA 4096 bit self-signed certificate for the
domain `example.com`:

```bash
docker run -u "$(id -u):$(id -g)" -v "$(pwd)":/keys authelia/authelia authelia crypto certificate rsa generate --common-name example.com --directory /keys
```

```bash
authelia crypto certificate rsa generate --common-name example.com --directory /path/to/keys
```
