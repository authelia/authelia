---
layout: default
title: File
parent: Authentication backends
grand_parent: Configuration
nav_order: 1
---

# File

**Authelia** supports a file as a users database.

## Configuration

Configuring Authelia to use a file is done by specifying the path to the
file in the configuration file.

    authentication_backend:
        disable_reset_password: false
        file:
            path: /var/lib/authelia/users.yml
                password_hashing:
                    algorithm: argon2id
                    iterations: 1
                    salt_length: 16
                    parallelism: 8
                    memory: 1024


## Format

The format of the users file is as follows.

    users:
        john:
            password: "$argon2id$v=19$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM"
            email: john.doe@authelia.com
            groups:
                - admins
                - dev

        harry:
            password: "$argon2id$v=19$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM"
            email: harry.potter@authelia.com
            groups: []

        bob:
            password: "$argon2id$v=19$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM"
            email: bob.dylan@authelia.com
            groups:
                - dev

        james:
            password: "$argon2id$v=19$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM"
            email: james.dean@authelia.com


This file should be set with read/write permissions as it could be updated by users
resetting their passwords.


## Passwords

The file contains hashed passwords instead of plain text passwords for security reasons.

You can use Authelia binary or docker image to generate the hash of any password. The 
hash-password command has many tunable options, you can view them with the 
`authelia hash-password --help` command. For example if you wanted to improve the entropy
you could generate a 16 byte salt and provide it with the `--salt` flag. 
Example: `authelia hash-password --salt abcdefghijklhijl`. For argon2id the salt must 
always be valid for base64 decoding (characters a through z, A through Z, 0 through 9, and +/).

For instance to generate a hash with the docker image just run:

    $ docker run authelia/authelia:latest authelia hash-password yourpassword
    $ Password hash: $argon2id$v=19$m=65536$3oc26byQuSkQqksq$zM1QiTvVPrMfV6BVLs2t4gM+af5IN7euO0VB6+Q8ZFs

Full CLI Help Documentation:

```
Hash a password to be used in file-based users database. Default algorithm is argon2id.

Usage:
  authelia hash-password [password] [flags]

Flags:
  -h, --help              help for hash-password
  -i, --iterations int    set the number of hashing iterations (default 1)
  -k, --key-length int    [argon2id] set the key length param (default 32)
  -m, --memory int        [argon2id] set the amount of memory param (in MB) (default 1024)
  -p, --parallelism int   [argon2id] set the parallelism param (default 8)
  -s, --salt string       set the salt string
  -l, --salt-length int   set the auto-generated salt length (default 16)
  -z, --sha512            use sha512 as the algorithm (defaults iterations to 50000, change with -i)
```


## Password hash algorithm

The default hash algorithm is salted Argon2id version 19. Argon2id is currently considered 
the best hashing algorithm, and in 2015 won the 
[Password Hashing Competition](https://en.wikipedia.org/wiki/Password_Hashing_Competition).
It benefits from customizable parameters allowing the cost of computing a hash to scale 
into the future which makes it harder to brute-force. Argon2id was implemented due to community 
feedback as you can see in this closed [issue](https://github.com/authelia/authelia/issues/577).

Additionally SHA512 is supported for backwards compatibility and user choice. While it's a reasonable
hash function given high enough iterations, as hardware gets better it has a higher chance of being 
brute-forced.

Hashes are identifiable as argon2id or SHA512 by their prefix of either `$argon2id$` and `$6$` 
respectively,  as described in this [wiki page](https://en.wikipedia.org/wiki/Crypt_(C)).

### Password hash algorithm tuning
 
All algorithm tuning is supported for Argon2id. The only configuration variables that affect 
SHA512 are iterations and salt length. The configuration variables are unique to the file
authentication provider, thus they all exist in a key under the file authentication configuration
key called `password_hashing`. We have set what are considered as sane and recommended defaults
to cater for a reasonable system, if you're unsure about which settings to tune, please see the 
parameters below, or for a more in depth understanding see the referenced documentation in
[Argon2 links](./file.md#argon2-links).


### Password hashing configuration settings

#### algorithm
 - Value Type: String
 - Possible Value: `argon2id` or `sha512`
 - Recommended: `argon2id`
 - What it Does: Changes the hashing algorithm


#### iterations
   - Value Type: Int
   - Possible Value: `1` or higher for argon2id and `1000` or higher for sha512 
   (will automatically be set to `1000` on lower settings)
   - Recommended: `1` for the `argon2id` algorithm and `50000` for `sha512`
   - What it Does: Adjusts the number of times we run the password through the hashing algorithm


#### key_length
 - Value Type: Int
 - Possible Value: `16` or higher.
 - Recommended: `32` or higher.
 - What it Does: Adjusts the length of the actual hash


#### salt_length
  - Value Type: Int
  - Possible Value: between `2` and `16`
  - Recommended: `16`
  - What it Does: Adjusts the length of the random salt we add to the password, there
   is no reason not to set this to 16


#### parallelism
 - Value Type: Int
 - Possible Value: `1` or higher
 - Recommended: `8` or twice your CPU cores
 - What it Does: Sets the number of threads used for hashing


#### memory
 - Value Type: Int
 - Possible Value: at least `8` times the value of `parallelism`
 - Recommended: `1024‬‬` (1GB) or as much RAM as you can afford to give to hashing
 - What it Does: Sets the amount of RAM used in MB for hashing


#### Examples for specific systems

These examples have been tested against a single system to make sure they roughly take 
0.5 seconds each. Your results may vary depending on individual specification and
utilization, but they are a good guide to get started. You should however read the 
linked documents in [Argon2 links](./file.md#argon2-links).

|    System     |Iterations|Parallelism|Memory |
|:------------: |:--------:|:---------:|:-----:|
|Raspberry Pi 2 |    1     |     8     |    64 |
|Raspberry Pi 3 |    1     |     8     |   128 |
|Raspberry Pi 4 |    1     |     8     |   128 |
|Intel G5 i5 NUC|    1     |     8     |  1024 |


#### Argon2 Links
[How to choose the right parameters for Argon2](https://www.twelve21.io/how-to-choose-the-right-parameters-for-argon2/)

[Go Documentation](https://godoc.org/golang.org/x/crypto/argon2)

[IETF Draft](https://tools.ietf.org/id/draft-irtf-cfrg-argon2-09.html)