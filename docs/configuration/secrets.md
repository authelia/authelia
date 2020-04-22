---
layout: default
title: Secrets
parent: Configuration
nav_order: 8
---

# Secrets

Configuration of Authelia requires some secrets and passwords.
Even if they can be set in the configuration file, the recommended
way to set secrets is to use environment variables as described
below.

## Environment variables

A secret can be configured using an environment variable with the
prefix AUTHELIA_ followed by the path of the option capitalized
and with dots replaced by underscores followed by the suffix _FILE.

The contents of the environment variable must be a path to a file
containing the secret data. This file must be readable by the
user the Authelia daemon is running as.

For instance the LDAP password can be defined in the configuration
at the path **authentication_backend.ldap.password**, so this password 
could alternatively be set using the environment variable called
**AUTHELIA_AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE**.

Here is the list of the environment variables which are considered
secrets and can be defined. Any other option defined using an
environment variable will not be replaced.

|Configuration Key|Environment Variable|
|:---------------:|:------:|
|jwt_secret       |AUTHELIA_JWT_SECRET_FILE|
|duo_api.secret_key|AUTHELIA_DUO_API_SECRET_KEY_FILE|
|session.secret|AUTHELIA_SESSION_SECRET_FILE|
|session.redis.password|AUTHELIA_SESSION_REDIS_PASSWORD_FILE|
|storage.mysql.password|AUTHELIA_STORAGE_MYSQL_PASSWORD_FILE|
|storage.postgres.password|AUTHELIA_STORAGE_POSTGRES_PASSWORD_FILE|
|notifier.smtp.password|AUTHELIA_NOTIFIER_SMTP_PASSWORD|
|authentication_backend.ldap.password|AUTHELIA_AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE|


## Secrets exposed in an environment variable

Prior to implementing file secrets you were able to define the
values of secrets in the environment variables themselves
in plain text instead of referencing a file. This is still
supported but discouraged. If you still want to do this
just remove _FILE from the environment variable name
and define the value in insecure plain text.


## Secrets in configuration file

If for some reason you prefer keeping the secrets in the configuration
file, be sure to apply the right permissions to the file in order to
prevent secret leaks if an another application gets compromised on your
server. The UNIX permissions should probably be something like 600.
