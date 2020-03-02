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

You can use Authelia binary or docker image to generate the hash of any password.

For instance, with the docker image, just run

    $ docker run authelia/authelia:latest authelia hash-password yourpassword
    $argon2id$v=19$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$RNAy4ppk5taziCtvH48b6PadEz7r88vZV5n7WmU7yGk


## Password Hash Function

The only supported hash functions are salted Argon2id (default, version 19 only), and salted SHA512 for backwards compatibility; determined
by the prefix `$argon2id$` and `$6$` respectively, as described in this [wiki](https://en.wikipedia.org/wiki/Crypt_(C)) page. 

Although SHA512 is supported, we do not provide a method currently to generate these hashes. This is because it is
not the best hash function; while it is a decent algorithm given the number of rounds is big enough, the difficulty to 
crack the hash is not determined by the performance of the machine. The best current algorithm, 
[Argon2](https://en.wikipedia.org/wiki/Argon2) does though. It won the 
[Password Hashing Competition](https://en.wikipedia.org/wiki/Password_Hashing_Competition) in 2015 and is currently
considered the best hashing function. This was implemented due to community feedback as you can see in this closed
 [issue](https://github.com/authelia/authelia/issues/577).