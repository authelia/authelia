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


## Format


The format of the file is as follows.

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
Example: `authelia hash-password --salt abcdefghijklhijl`. For argon2id the salt must always be a valid for base64
decoding (characters a through z, A through Z, 0 through 9, and +/).

For instance, with the docker image, just run

    $ docker run authelia/authelia:latest authelia hash-password yourpassword
    $argon2id$v=19$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$RNAy4ppk5taziCtvH48b6PadEz7r88vZV5n7WmU7yGk


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
 
 All algorithm tuning is supported by Argon2id except Key length. The only configuration variables that affects SHA512
 are rounds and salt length. The configuration variables are unique to the file authentication provider, thus they all
 exist in a key under the file authentication configuration called `password_hashing`.
 
 Example:
 ```
file:
  path: /var/authelia.users.yml
  password_hashing:
    algorithm: argon2id
    iterations: 3
    salt_length: 16
    parallelism: 2
    memory: ‭65536‬
```
 
 #### algorithm
 - Value Type: String
 - Possible Value: `argon2id` and `sha512`
 - Recommended: `argon2id`
 - What it Does: Changes the Hashing Algorithm
 
 #### iterations
   - Value Type: Int
   - Possible Value: `1` or higher for argon2id and `1000` or higher for sha512 (will automatically be set to `1000` on lower settings)
   - Recommended: `3` for the `argon2id` algorithm and `50000` for `sha512`
   - What it Does: Adjusts the number of times we run the password through the hashing algorithm
   
 #### salt_length
  - Value Type: Int
  - Possible Value: between `2` and `16`
  - Recommended: `16`
  - What it Does: Adjusts the length of the random salt we add to the password, there is no reason not to set this to 16
 
 #### parallelism
 - Value Type: Int
 - Possible Value: `1` or higher
 - Recommended: `2`
 
 #### memory
 - Value Type: Int
 - Possible Value: at least `8` times the value of `parallelism`
 - Recommended: `‭65536‬`