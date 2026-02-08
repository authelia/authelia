---
title: "Passwords"
description: "A reference guide on passwords and hashing etc"
summary: "This section contains reference documentation for Authelia."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 220
toc: true
aliases:
  - /r/passwords
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## User / Password File

This file should be set with read/write permissions as it could be updated by users resetting their passwords.

### YAML Format

The format of the [YAML] file is documented via the [JSONSchema](schemas.md#json-schema). An example of this is as
follows:

```yaml {title="users-database.yml"}
# yaml-language-server: $schema=https://www.authelia.com/schemas/latest/json-schema/user-database.json
users:
  john:
    disabled: false
    displayname: 'John Doe'
    password: '$argon2id$v=19$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM'
    email: 'john.doe@authelia.com'
    groups:
      - 'admins'
      - 'dev'
  harry:
    disabled: false
    displayname: 'Harry Potter'
    password: '$argon2id$v=19$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM'
    email: 'harry.potter@authelia.com'
    groups: []
  james:
    disabled: false
    displayname: 'James Dean'
    password: '$argon2id$v=19$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM'
    email: 'james.dean@authelia.com'
    groups: []
  bob:
    disabled: false
    displayname: 'Bob Dylan'
    password: '$argon2id$v=19$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM'
    email: 'bob.dylan@authelia.com'
    groups:
      - 'dev'
    given_name: 'Robert'
    family_name: 'Zimmerman'
    middle_name: 'Allen'
    nickname: 'Bob'
    profile: 'https://en.wikipedia.org/wiki/Bob_Dylan'
    picture: 'https://kelvinokaforart.com/wp-content/uploads/2023/01/Bob-Dylan.jpg'
    website: 'https://www.bobdylan.com/'
    gender: 'male'
    birthdate: '1941-05-24'
    zoneinfo: 'America/Chicago'
    locale: 'en-US'
    phone_number: '+1 (425) 555-1212'
    phone_extension: '1000'
    address:
      street_address: '2-3 Kitanomarukoen'
      locality: 'Chiyoda City'
      region: 'Tokyo'
      postal_code: '102-8321'
      country: 'Japan'
    extra:
      example: 'value'
```

It's recommended to check out the [Attributes Reference Guide](../../reference/guides/attributes.md) for more
information on all of the attribute specifics, and it should be noted that all of the attributes are validated
including the extra attributes which may not exist unless they are configured.

## Passwords

The file contains hashed passwords instead of plain text passwords for security reasons.

You can use Authelia binary or docker image to generate the hash of any password. The [crypt hash generate] command has
many supported algorithms. To view them run the `authelia crypto hash generate --help` command. To see the tunable
options for an algorithm subcommand include that command before `--help`. For example for the [Argon2] algorithm use the
`authelia crypto hash generate argon2 --help` command to see the available options.

Passwords passed to [crypt hash generate] should be single quoted if using the `--password` parameter instead of the
console prompt, especially if it has  special characters to prevent parameter substitution.

To generate an [Argon2] hash with the docker image interactively just run:

{{< envTabs "Generate Password (Interactive)" >}}
{{< envTab "Docker" >}}
```bash
docker run --rm -it authelia/authelia:latest authelia crypto hash generate argon2
```
{{< /envTab >}}
{{< envTab "Bare-Metal" >}}
```bash
authelia crypto hash generate argon2
```
{{< /envTab >}}
{{< /envTabs >}}

To generate an [Argon2] hash with the docker image without a prompt you can run:

{{< envTabs "Generate Password" >}}
{{< envTab "Docker" >}}
```bash
docker run --rm authelia/authelia:latest authelia crypto hash generate argon2 --password 'password'
```
{{< /envTab >}}
{{< envTab "Bare-Metal" >}}
```bash
authelia crypto hash generate argon2 --password 'password'
```
{{< /envTab >}}
{{< /envTabs >}}

Output Example:
```bash
Digest: $argon2id$v=19$m=65536,t=3,p=4$Hjc8e7WYcBFcJmEDUOsS9A$ozM7RyZR1EyDR8cuyVpDDfmLrGPGFgo5E2NNqRumui4
```

You may also use the `--config` flag to point to your existing configuration. When used, the values defined in the
config will be used instead. For example to generate the password with a configuration file named `configuration.yml`
in the current directory:

{{< envTabs "Generate Password (Interactive)" >}}
{{< envTab "Docker" >}}
```bash
docker run --rm -it -v ./configuration.yml:/configuration.yml authelia/authelia:latest authelia crypto hash generate --config /configuration.yml
```
{{< /envTab >}}
{{< envTab "Bare Metal" >}}
```bash
authelia crypto hash generate --config /configuration.yml
```
{{< /envTab >}}
{{< /envTabs >}}

Output Example:

```bash
Enter Password:
Confirm Password:

Digest: $argon2id$v=19$m=65536,t=3,p=4$Hjc8e7WYcBFcJmEDUOsS9A$ozM7RyZR1EyDR8cuyVpDDfmLrGPGFgo5E2NNqRumui4
```

See the [full CLI reference documentation](../cli/authelia/authelia_crypto_hash_generate.md).

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

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
When using algorithms that use a memory cost like [Argon2](https://datatracker.ietf.org/doc/html/rfc9106) and [Scrypt](https://en.wikipedia.org/wiki/Scrypt) it should be noted that
this memory is released by Go after the hashing process completes, however the operating system may not reclaim the
memory until a later time such as when the system is experiencing memory pressure which may cause the appearance of more
memory being in use than Authelia is actually actively using. Authelia will typically reuse this memory if it has not be
reclaimed as long as another hashing calculation is not still utilizing it.
{{< /callout >}}

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

For backwards compatibility and user choice support for the [SHA2 Crypt] algorithm (`SHA512` variant) is still available.
While it's a reasonable hashing function given high enough iterations, as hardware improves it has a higher chance of
being brute-forced since it only allows scaling the CPU [cost](#cost) whereas [Argon2] allows scaling both for CPU and
Memory [cost](#cost).

#### Identification

The algorithm that a hash is utilizing is identifiable by its prefix:

|  Algorithm   |  Variant   |      Prefix       |
|:------------:|:----------:|:-----------------:|
|   [Argon2]   | `argon2id` |   `$argon2id$`    |
|   [Argon2]   | `argon2i`  |    `$argon2i$`    |
|   [Argon2]   | `argon2d`  |    `$argon2d$`    |
|   [Scrypt]   |  `scrypt`  |    `$scrypt$`     |
|   [Scrypt]   | `yescrypt` |       `$y$`       |
|   [PBKDF2]   |   `sha1`   |    `$pbkdf2$`     |
|   [PBKDF2]   |  `sha224`  | `$pbkdf2-sha224$` |
|   [PBKDF2]   |  `sha256`  | `$pbkdf2-sha256$` |
|   [PBKDF2]   |  `sha384`  | `$pbkdf2-sha384$` |
|   [PBKDF2]   |  `sha512`  | `$pbkdf2-sha512$` |
| [SHA2 Crypt] |  `SHA256`  |       `$5$`       |
| [SHA2 Crypt] |  `SHA512`  |       `$6$`       |
|   [Bcrypt]   | `standard` |      `$2b$`       |
|   [Bcrypt]   |  `sha256`  | `$bcrypt-sha256$` |

See the [Crypt (C) Wiki page](https://en.wikipedia.org/wiki/Crypt_(C)) for more information.

#### Tuning

The configuration variables are unique to the file authentication provider, thus they all exist in a key under the file
authentication configuration key called [password](../../configuration/first-factor/file.md#password-options). The defaults are
considered as sane for a reasonable system however we still recommend taking time to figure out the best values to
adequately determine the [cost](#cost).

While there are recommended parameters for each algorithm it's your responsibility to tune these individually for your
particular system. We strongly recommend reading other sources such as the
[OWASP Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html) when tuning these
algorithms.

#### Algorithm Choice

We generally discourage [Bcrypt] except when needed for interoperability with legacy systems. The `argon2id` variant of
the [Argon2] algorithm is the best choice of the algorithms available, but it's important to note that the `argon2id`
variant is the most resilient variant, followed by the `argon2d` variant and the `argon2i` variant not being recommended.
It's strongly recommended if you're unsure that you use `argon2id`. [Scrypt] is a likely second best algorithm. [PBKDF2]
is practically the only choice when it comes to [FIPS-140 compliance]. The `sha512` variant of the [SHA2 Crypt]
algorithm is also a reasonable option, but is mainly available for backwards compatibility.

All other algorithms and variants available exist only for interoperability and we discourage their use if a better
algorithm is available in your scenario.

#### Recommended Parameters: Argon2

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
The memory parameter assumes you're utilizing the new configuration with the explicit names
detailed in the [Argon2 configuration](../../configuration/first-factor/file.md#argon2) documentation.
{{< /callout >}}

This table adapts the [RFC9106 Parameter Choice] recommendations to our configuration options:

|  Situation  | Variant  | Iterations (t) | Parallelism (p) | Memory (m) | Salt Size | Key Size |
|:-----------:|:--------:|:--------------:|:---------------:|:----------:|:---------:|:--------:|
| Low Memory  | argon2id |       3        |        4        |   65536    |    16     |    32    |
| Recommended | argon2id |       1        |        4        |  2097152   |    16     |    32    |

#### Recommended Parameters: SHA2 Crypt

This table suggests the parameters for the [SHA2 Crypt] algorithm:

|  Situation   | Variant | Iterations (rounds) | Salt Size |
|:------------:|:-------:|:-------------------:|:---------:|
| Standard CPU | sha512  |        50000        |    16     |
| High End CPU | sha512  |       150000        |    16     |

[Argon2]: https://datatracker.ietf.org/doc/html/rfc9106
[Scrypt]: https://en.wikipedia.org/wiki/Scrypt
[PBKDF2]: https://datatracker.ietf.org/doc/html/rfc2898
[SHA2 Crypt]: https://www.akkadia.org/drepper/SHA-crypt.txt
[Bcrypt]: https://en.wikipedia.org/wiki/Bcrypt
[FIPS-140 compliance]: https://csrc.nist.gov/publications/detail/fips/140/2/final

[RFC9106 Parameter Choice]: https://datatracker.ietf.org/doc/html/rfc9106#section-4
[YAML]: https://yaml.org/
[crypt hash generate]: ../cli/authelia/authelia_crypto_hash_generate.md
[Password Hashing Competition]: https://en.wikipedia.org/wiki/Password_Hashing_Competition
