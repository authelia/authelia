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
        file:
            path: /var/lib/authelia/users.yml
                password_hashing:
                    algorithm: argon2id
                    iterations: 3
                    salt_length: 16
                    parallelism: 2
                    memory: ‭65536‬


### Password Hashing Configuration Settings

 #### algorithm
 - Value Type: String
 - Possible Value: `argon2id` and `sha512`
 - Recommended: `argon2id`
 - What it Does: Changes the Hashing Algorithm
 
 #### iterations
   - Value Type: Int
   - Possible Value: `1` or higher for argon2id and `1000` or higher for sha512 (will automatically be set to `1000` on lower settings)
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
  - What it Does: Adjusts the length of the random salt we add to the password, there is no reason not to set this to 16
 
 #### parallelism
 - Value Type: Int
 - Possible Value: `1` or higher
 - Recommended: `4`
 - What it Does: Sets the number of threads used for crypto
 
 #### memory
 - Value Type: Int
 - Possible Value: at least `8` times the value of `parallelism`
 - Recommended: `65535‬‬` (64MB)
 - What it Does: Sets the amount of RAM used in KB for crypto (1024 * MB desired)
 
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

You can use Authelia binary or docker image to generate the hash of any password. The hash-password command has many 
tunable options, you can view them with the `authelia hash-password --help` command. For example if you wanted to improve
the entropy you could generate a 16 byte salt and provide it with the `--salt` flag. 
Example: `authelia hash-password --salt abcdefghijklhijl`. For argon2id the salt must always be valid for base64
decoding (characters a through z, A through Z, 0 through 9, and +/).

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
  -m, --memory int        [argon2id] set the amount of memory param (in KB) (default 65536)
  -p, --parallelism int   [argon2id] set the parallelism param (default 4)
  -s, --salt string       set the salt string
  -l, --salt-length int   set the auto-generated salt length (default 16)
  -z, --sha512            use sha512 as the algorithm (defaults iterations to 50000, change with -i)
```

## Password Hash Function

The supported hash functions are salted Argon2id (default, version 19 only), and salted SHA512 for backwards compatibility.
This is determined by the prefix `$argon2id$` and `$6$` respectively, as described in this [wiki page](https://en.wikipedia.org/wiki/Crypt_(C)). 

Although SHA512 is supported default hashes are generated with Argon2id. This is because it is
not the best hash function, while it is a decent algorithm given the number of rounds is big enough the difficulty to 
crack the hash is not determined by the performance of the machine. The best current algorithm, 
[Argon2](https://en.wikipedia.org/wiki/Argon2) does though. It won the 
[Password Hashing Competition](https://en.wikipedia.org/wiki/Password_Hashing_Competition) in 2015 and is currently
considered the best hashing function. This was implemented due to community feedback as you can see in this closed
 [issue](https://github.com/authelia/authelia/issues/577).
 
 ### Password Hash Algorithm Tuning
 
 All algorithm tuning is supported for Argon2id. The only configuration variables that affect SHA512
 are iterations and salt length. The configuration variables are unique to the file authentication provider, thus they all
 exist in a key under the file authentication configuration key called `password_hashing`. We have set what are considered 
 as sane and recommended defaults, if you're unsure about which settings to tune, please see the parameters below, or 
 for a more in depth understanding see the referenced documentation.
 
 #### Argon2 Links
 [Go Documentation](https://godoc.org/golang.org/x/crypto/argon2)
 
 [IETF Draft](https://tools.ietf.org/id/draft-irtf-cfrg-argon2-05.html)
 