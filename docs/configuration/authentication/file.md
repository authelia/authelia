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
            password: "$6$rounds=500000$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/"
            email: john.doe@authelia.com
            groups:
                - admins
                - dev

        harry:
            password: "$6$rounds=500000$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/"
            email: harry.potter@authelia.com
            groups: []

        bob:
            password: "$6$rounds=500000$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/"
            email: bob.dylan@authelia.com
            groups:
                - dev

        james:
            password: "$6$rounds=500000$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/"
            email: james.dean@authelia.com

This file should be set with read/write permissions as it could be updated by users
resetting their passwords.

## Passwords

The file contains hash of passwords instead of plain text passwords for security reasons.

You can use authelia binary or docker image to generate the hash of any password.

For instance, with the docker image, just run

    $ docker run authelia/authelia:latest authelia hash-password yourpassword
    $6$rounds=50000$BpLnfgDsc2WD8F2q$be7OyobnQ8K09dyDiGjY.cULh4yDePMh6CUMpLwF4WHTJmLcPE2ijM2ZsqZL.hVAANojEfDu3sU9u9uD7AeBJ/


## Password Hash Function

The only supported hash function is salted sha512 determined by the prefix `$6$` as described
in this [wiki](https://en.wikipedia.org/wiki/Crypt_(C)) page. 

Although not the best hash function, Salted SHA512 is a decent algorithm given the number of rounds is big
enough. It's not the best because the difficulty to crack the hash does not on the performance of the machine.
The best algorithm, [Argon2](https://en.wikipedia.org/wiki/Argon2) does though. It won the
[Password Hashing Competition](https://en.wikipedia.org/wiki/Password_Hashing_Competition) in 2015 and is now
considered the best hashing function. There is an open [issue](https://github.com/authelia/authelia/issues/577)
to add support for this hashing function.

