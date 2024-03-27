---
title: "File"
description: "File"
summary: "Authelia supports a file based first factor user provider. This section describes configuring this."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
weight: 102300
toc: true
aliases:
  - /docs/configuration/authentication/file.html
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
authentication_backend:
  file:
    path: '/config/users.yml'
    watch: false
    search:
      email: false
      case_insensitive: false
    password:
      algorithm: 'argon2'
      argon2:
        variant: 'argon2id'
        iterations: 3
        memory: 65536
        parallelism: 4
        key_length: 32
        salt_length: 16
      scrypt:
        iterations: 16
        block_size: 8
        parallelism: 1
        key_length: 32
        salt_length: 16
      pbkdf2:
        variant: 'sha512'
        iterations: 310000
        salt_length: 16
      sha2crypt:
        variant: 'sha512'
        iterations: 50000
        salt_length: 16
      bcrypt:
        variant: 'standard'
        cost: 12
```

## Options

This section describes the individual configuration options.

### path

{{< confkey type="string" required="yes" >}}

The path to the file with the user details list. Supported file types are:

* [YAML File](../../reference/guides/passwords.md#yaml-format)

### watch

{{< confkey type="boolean" default="false" required="no" >}}

Enables reloading the database by watching it for changes.

### search {#config-search}

Username searching functionality options.

*__Important Note:__ This functionality is experimental.*

#### email

{{< confkey type="boolean" default="false" required="no" >}}

Allows users to login using their email address. If enabled two users must not have the same emails and their usernames
must not be an email.

*__Note:__ Emails are always checked using case-insensitive lookup.*

#### case_insensitive

{{< confkey type="boolean" default="false" required="no" >}}

Enabling this search option allows users to login with their username regardless of case. If enabled users must only
have lowercase usernames.

*__Note:__ Emails are always checked using case-insensitive lookup.*

## Password Options

A [reference guide](../../reference/guides/passwords.md) exists specifically for choosing password hashing values. This
section contains far more information than is practical to include in this configuration document. See the
[Passwords Reference Guide](../../reference/guides/passwords.md) for more information.

This guide contains examples such as the [User / Password File](../../reference/guides/passwords.md#user--password-file).

### algorithm

{{< confkey type="string" default="argon2" required="no" >}}

Controls the hashing algorithm used for hashing new passwords. Value must be one of:

* `argon2` for the [Argon2](#argon2) algorithm
* `scrypt` for the [Scrypt](#scrypt) algorithm
* `pbkdf2` for the [PBKDF2](#pbkdf2) algorithm
* `sha2crypt` for the [SHA2Crypt](#sha2crypt) algorithm
* `bcrypt` for the [Bcrypt](#bcrypt) algorithm

### argon2

The [Argon2] algorithm implementation. This is one of the only algorithms that was designed purely with password hashing
in mind and is subsequently one of the best algorithms to date for security.

#### variant

{{< confkey type="string" default="argon2id" required="no" >}}

Controls the variant when hashing passwords using [Argon2]. Recommended `argon2id`.
Permitted values `argon2id`, `argon2i`, `argon2d`.

#### iterations

{{< confkey type="integer" default="3" required="no" >}}

Controls the number of iterations when hashing passwords using [Argon2].

#### memory

{{< confkey type="integer" default="65536" required="no" >}}

Controls the amount of memory in kibibytes when hashing passwords using [Argon2].

#### parallelism

{{< confkey type="integer" default="4" required="no" >}}

Controls the parallelism factor when hashing passwords using [Argon2].

#### key_length

{{< confkey type="integer" default="32" required="no" >}}

Controls the output key length when hashing passwords using [Argon2].

#### salt_length

{{< confkey type="integer" default="16" required="no" >}}

Controls the output salt length when hashing passwords using [Argon2].

### scrypt

The [Scrypt] algorithm implementation.

#### iterations

{{< confkey type="integer" default="16" required="no" >}}

Controls the number of iterations when hashing passwords using [Scrypt].

#### block_size

{{< confkey type="integer" default="8" required="no" >}}

Controls the block size when hashing passwords using [Scrypt].

#### parallelism

{{< confkey type="integer" default="1" required="no" >}}

Controls the parallelism factor when hashing passwords using [Scrypt].

#### key_length

{{< confkey type="integer" default="32" required="no" >}}

Controls the output key length when hashing passwords using [Scrypt].

#### salt_length

{{< confkey type="integer" default="16" required="no" >}}

Controls the output salt length when hashing passwords using [Scrypt].

### pbkdf2

The [PBKDF2] algorithm implementation.

#### variant

{{< confkey type="string" default="sha512" required="no" >}}

Controls the variant when hashing passwords using [PBKDF2]. Recommended `sha512`.
Permitted values `sha1`, `sha224`, `sha256`, `sha384`, `sha512`.

#### iterations

{{< confkey type="integer" default="310000" required="no" >}}

Controls the number of iterations when hashing passwords using [PBKDF2].

#### salt_length

{{< confkey type="integer" default="16" required="no" >}}

Controls the output salt length when hashing passwords using [PBKDF2].

### sha2crypt

The [SHA2 Crypt] algorithm implementation.

#### variant

{{< confkey type="string" default="sha512" required="no" >}}

Controls the variant when hashing passwords using [SHA2 Crypt]. Recommended `sha512`.
Permitted values `sha256`, `sha512`.

#### iterations

{{< confkey type="integer" default="50000" required="no" >}}

Controls the number of iterations when hashing passwords using [SHA2 Crypt].

#### salt_length

{{< confkey type="integer" default="16" required="no" >}}

Controls the output salt length when hashing passwords using [SHA2 Crypt].

### bcrypt

The [Bcrypt] algorithm implementation.

#### variant

{{< confkey type="string" default="standard" required="no" >}}

Controls the variant when hashing passwords using [Bcrypt]. Recommended `standard`.
Permitted values `standard`, `sha256`.

*__Important Note:__ The `sha256` variant is a special variant designed by
[Passlib](https://passlib.readthedocs.io/en/stable/lib/passlib.hash.bcrypt_sha256.html). This variant passes the
password through a SHA256 HMAC before passing it to the [Bcrypt] algorithm, effectively bypassing the 72 byte password
truncation that [Bcrypt] does. It is not supported by many other systems.*

#### cost

{{< confkey type="integer" default="12" required="no" >}}

Controls the hashing cost when hashing passwords using [Bcrypt].

[Argon2]: https://datatracker.ietf.org/doc/html/rfc9106
[Scrypt]: https://en.wikipedia.org/wiki/Scrypt
[PBKDF2]: https://datatracker.ietf.org/doc/html/rfc2898
[SHA2 Crypt]: https://www.akkadia.org/drepper/SHA-crypt.txt
[Bcrypt]: https://en.wikipedia.org/wiki/Bcrypt
