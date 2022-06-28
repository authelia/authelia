---
title: "File"
description: "File"
lead: "Authelia supports a file based first factor user provider. This section describes configuring this."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  configuration:
    parent: "first-factor"
weight: 102300
toc: true
aliases:
  - /docs/configuration/authentication/file.html
---

## Configuration

```yaml
authentication_backend:
  file:
    path: /config/users.yml
    password:
      algorithm: argon2id
      iterations: 3
      key_length: 32
      salt_length: 16
      parallelism: 4
      memory: 64
```

## Options

### path

{{< confkey type="string" required="yes" >}}

The path to the file with the user details list. Supported file types are:

* [YAML File](../../reference/guides/passwords.md#yaml-format)

### password

#### algorithm

{{< confkey type="string" default="argon2id" required="no" >}}

Controls the hashing algorithm used for hashing new passwords. Value must be one of:

* `argon2id` for the [Argon2] `id` variant
* `sha512` for the [SHA Crypt] `SHA512` variant

#### iterations

{{< confkey type="integer" required="no" >}}

Controls the number of hashing iterations done by the other hashing settings ([Argon2] parameter `t`, [SHA Crypt]
parameter `rounds`). This affects the effective cost of hashing.

| Algorithm | Minimum | Default |                                        Recommended                                         |
|:---------:|:-------:|:-------:|:------------------------------------------------------------------------------------------:|
| argon2id  |    1    |    3    | [See Recommendations](../../reference/guides/passwords.md#recommended-parameters-argon2id) |
|  sha512   |  1000   |  50000  |  [See Recommendations](../../reference/guides/passwords.md#recommended-parameters-sha512)  |

#### key_length

{{< confkey type="integer" default="32" required="no" >}}

*__Important:__ This setting is specific to the `argon2id` algorithm and unused with the `sha512` algorithm.*

Sets the key length of the [Argon2] hash output. The minimum value is `16` with the recommended value of `32` being set
as the default.

#### salt_length

{{< confkey type="integer" default="16" required="no" >}}

Controls the length of the random salt added to each password before hashing. There is not a compelling reason to have
this set to anything other than `16`, however the minimum is `8` with the recommended value of `16` being set as the
default.

#### parallelism

{{< confkey type="integer" default="4" required="no" >}}

*__Important:__ This setting is specific to the `argon2id` algorithm and unused with the `sha512` algorithm.*

Sets the number of threads used by [Argon2] when hashing passwords ([Argon2] parameter `p`). The minimum value is `1`
with the recommended value of `4` being set as the default. This affects the effective cost of hashing.

#### memory

{{< confkey type="integer" default="64" required="no" >}}

*__Important:__ This setting is specific to the `argon2id` algorithm and unused with the `sha512` algorithm.*

Sets the amount of memory in megabytes allocated to a single password hashing calculation ([Argon2] parameter `m`). This
affects the effective cost of hashing.

This memory is released by go after the hashing process completes, however the operating system may not reclaim the
memory until a later time such as when the system is experiencing memory pressure which may cause the appearance of more
memory being in use than Authelia is actually actively using. Authelia will typically reuse this memory if it has not be
reclaimed as long as another hashing calculation is not still utilizing it.

## Reference

A [reference guide](../../reference/guides/passwords.md) exists specifically for choosing password hashing values. This
section contains far more information than is practical to include in this configuration document. See the
[Passwords Reference Guide](../../reference/guides/passwords.md) for more information.

This guide contains examples such as the [User / Password File](../../reference/guides/passwords.md#user--password-file).

[Argon2]: https://www.rfc-editor.org/rfc/rfc9106.html
[SHA Crypt]: https://www.akkadia.org/drepper/SHA-crypt.txt
