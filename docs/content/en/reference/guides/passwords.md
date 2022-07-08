---
title: "Passwords"
description: "A reference guide on passwords and hashing etc"
lead: "This section contains reference documentation for Authelia."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  reference:
    parent: "guides"
weight: 220
toc: true
aliases:
  - /r/passwords
---

## User / Password File

This file should be set with read/write permissions as it could be updated by users resetting their passwords.

### YAML Format

The format of the [YAML] file is as follows:

```yaml
users:
  john:
    displayname: "John Doe"
    password: "$argon2id$v=19$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM"
    email: john.doe@authelia.com
    groups:
      - admins
      - dev
  harry:
    displayname: "Harry Potter"
    password: "$argon2id$v=19$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM"
    email: harry.potter@authelia.com
    groups: []
  bob:
    displayname: "Bob Dylan"
    password: "$argon2id$v=19$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM"
    email: bob.dylan@authelia.com
    groups:
      - dev
  james:
    displayname: "James Dean"
    password: "$argon2id$v=19$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM"
    email: james.dean@authelia.com
```

## Passwords

The file contains hashed passwords instead of plain text passwords for security reasons.

You can use Authelia binary or docker image to generate the hash of any password. The [hash-password] command has many
tunable options, you can view them with the `authelia hash-password --help` command. For example if you wanted to
improve the entropy you could generate a 16 byte salt and provide it with the `--salt` flag.

Example: `authelia hash-password --salt abcdefghijklhijl -- 'password'`.

Passwords passed to [hash-password] should be single quoted if using special characters to prevent parameter
substitution. In addition the password should be the last parameter, and should be after a `--`. For instance to
generate a hash with the docker image just run:

```bash
$ docker run authelia/authelia:latest authelia hash-password -- 'password'
Password hash: $argon2id$v=19$m=65536$3oc26byQuSkQqksq$zM1QiTvVPrMfV6BVLs2t4gM+af5IN7euO0VB6+Q8ZFs
```

You may also use the `--config` flag to point to your existing configuration. When used, the values defined in the
config will be used instead.

See the [full CLI reference documentation](../cli/authelia/authelia_hash-password.md).

### Cost

The most important part about choosing a password hashing function is the cost. It's generally recommended that the cost
takes roughly 500 milliseconds on your hardware to complete, however if you have very old hardware you may want to
consider more than 500 milliseconds, or if you have really high end hardware you may want to consider slightly less
depending on if you have a large quantity of users.

Ideally on average hardware the amount of time would be roughly 500 milliseconds at minimum.

In consideration of your cost you should take into account the fact some algorithms only support scaling the cost for
one factor and not others It's usually considered better to have a mix of cost types however this is not possible with
all algorithms. The main cost type measurements are:

* CPU
* Memory

*__Important Note:__ When using algorithms that use a memory cost like [Argon2] it should be noted that this memory is
released by Go after the hashing process completes, however the operating system may not reclaim the memory until a
later time such as when the system is experiencing memory pressure which may cause the appearance of more memory being
in use than Authelia is actually actively using. Authelia will typically reuse this memory if it has not be reclaimed as
long as another hashing calculation is not still utilizing it.*

To get a rough estimate of how much memory should be utilized with these algorithms you can utilize the following
command:

```bash
stress-ng --vm-bytes $(awk '/MemFree/{printf "%d\n", $2 * 0.9;}' < /proc/meminfo)k --vm-keep -m 1
```

If this is not desirable we recommend investigating the following options in order of most to least secure:

1. Use the [LDAP](../../configuration/first-factor/ldap.md) authentication provider instead
2. Adjusting the [memory](../../configuration/first-factor/file.md#memory) parameter
3. Changing the [algorithm](../../configuration/first-factor/file.md#algorithm)

### Algorithms

The default hash algorithm is the [Argon2] `id` variant version 19 with a salt. [Argon2] is at the time of this writing
widely considered to be the best hashing algorithm, and in 2015 won the [Password Hashing Competition]. It benefits from
customizable parameters including a memory parameter allowing the [cost](#cost) of computing a hash to scale into the
future with better hardware which makes it harder to brute-force.

For backwards compatibility and user choice support for the [SHA Crypt] algorithm (`SHA512` variant) is still available.
While it's a reasonable hashing function given high enough iterations, as hardware improves it has a higher chance of
being brute-forced since it only allows scaling the CPU [cost](#cost) whereas [Argon2] allows scaling both for CPU and
Memory [cost](#cost).

#### Identification

The algorithm that a hash is utilizing is identifiable by its prefix:

|  Algorithm  | Variant  |    Prefix    |
|:-----------:|:--------:|:------------:|
|  [Argon2]   |   `id`   | `$argon2id$` |
| [SHA Crypt] | `SHA512` |    `$6$`     |

See the [Crypt (C) Wiki page](https://en.wikipedia.org/wiki/Crypt_(C)) for more information.

#### Tuning

The configuration variables are unique to the file authentication provider, thus they all exist in a key under the file
authentication configuration key called [password](../../configuration/first-factor/file.md#password). The defaults are
considered as sane for a reasonable system however we still recommend taking time to figure out the best values to
adequately determine the [cost](#cost).

While there are recommended parameters for each algorithm it's your responsibility to tune these individually for your
particular system.

#### Recommended Parameters: Argon2id

This table adapts the [RFC9106 Parameter Choice] recommendations to our configuration options:

|  Situation  | Iterations (t) | Parallelism (p) | Memory (m) | Salt Size | Key Size |
|:-----------:|:--------------:|:---------------:|:----------:|:---------:|:--------:|
| Low Memory  |       3        |        4        |     64     |    16     |    32    |
| Recommended |       1        |        4        |    2048    |    16     |    32    |

#### Recommended Parameters: SHA512

This table suggests the parameters for the [SHA Crypt] (`SHA512` variant) algorithm:

|  Situation   | Iterations (rounds) | Salt Size |
|:------------:|:-------------------:|:---------:|
| Standard CPU |        50000        |    16     |
| High End CPU |       150000        |    16     |

[RFC9106 Parameter Choice]: https://www.rfc-editor.org/rfc/rfc9106.html#section-4
[YAML]: https://yaml.org/
[Argon2]: https://www.rfc-editor.org/rfc/rfc9106.html
[SHA Crypt]: https://www.akkadia.org/drepper/SHA-crypt.txt
[hash-password]: ../cli/authelia/authelia_hash-password.md
[Password Hashing Competition]: https://en.wikipedia.org/wiki/Password_Hashing_Competition
