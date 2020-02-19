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

A secret can be configured using an environment variable with name
starting with AUTHELIA_ and followed by the path of the option capitalized
and with dots replaced by underscores.

For instance the LDAP password is identified by the path
**authentication_backend.ldap.password**, so this password could
alternatively be set using the environment variable called
**AUTHELIA_AUTHENTICATION_BACKEND_LDAP_PASSWORD**.

Here is the list of the environment variables which are considered
secrets and can be defined. Any other option defined using an
environment variable will not be replaced.

* AUTHELIA_JWT_SECRET
* AUTHELIA_DUO_API_SECRET_KEY
* AUTHELIA_SESSION_SECRET
* AUTHELIA_AUTHENTICATION_BACKEND_LDAP_PASSWORD
* AUTHELIA_NOTIFIER_SMTP_PASSWORD
* AUTHELIA_SESSION_REDIS_PASSWORD
* AUTHELIA_STORAGE_MYSQL_PASSWORD
* AUTHELIA_STORAGE_POSTGRES_PASSWORD

## Secrets in configuration file

If for some reason you prefer keeping the secrets in the configuration
file, be sure to apply the right permissions to the file in order to
prevent secret leaks if an another application gets compromised on your
server. The UNIX permissions should probably be something like 600.
