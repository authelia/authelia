---
title: "authelia storage user totp generate"
description: "Reference for the authelia storage user totp generate command."
lead: ""
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
weight: 905
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## authelia storage user totp generate

Generate a TOTP configuration for a user

### Synopsis

Generate a TOTP configuration for a user.

This subcommand allows generating a new TOTP configuration for a user,
and overwriting the existing configuration if applicable.

```
authelia storage user totp generate <username> [flags]
```

### Examples

```
authelia storage user totp generate john
authelia storage user totp generate john --period 90
authelia storage user totp generate john --digits 8
authelia storage user totp generate john --algorithm SHA512
authelia storage user totp generate john --algorithm SHA512 --config config.yml
authelia storage user totp generate john --algorithm SHA512 --config config.yml --path john.png
```

### Options

```
      --algorithm string   set the algorithm to either SHA1 (supported by most applications), SHA256, or SHA512 (default "SHA1")
      --digits uint        set the number of digits (default 6)
  -f, --force              forces the configuration to be generated regardless if it exists or not
  -h, --help               help for generate
      --issuer string      set the issuer description (default "Authelia")
  -p, --path string        path to a file to create a PNG file with the QR code (optional)
      --period uint        set the period between rotations (default 30)
      --secret string      set the shared secret as base32 encoded bytes (no padding), it's recommended that you do not use this option unless you're restoring a configuration
      --secret-size uint   set the secret size (default 32)
```

### Options inherited from parent commands

```
  -c, --config strings                         configuration files or directories to load, for more information run 'authelia -h authelia config' (default [configuration.yml])
      --config.experimental.filters strings    list of filters to apply to all configuration files, for more information run 'authelia -h authelia filters'
      --encryption-key string                  the storage encryption key to use
      --mysql.database string                  the MySQL database name (default "authelia")
      --mysql.host string                      the MySQL hostname
      --mysql.password string                  the MySQL password
      --mysql.port int                         the MySQL port (default 3306)
      --mysql.username string                  the MySQL username (default "authelia")
      --postgres.database string               the PostgreSQL database name (default "authelia")
      --postgres.host string                   the PostgreSQL hostname
      --postgres.password string               the PostgreSQL password
      --postgres.port int                      the PostgreSQL port (default 5432)
      --postgres.schema string                 the PostgreSQL schema name (default "public")
      --postgres.ssl.certificate string        the PostgreSQL ssl certificate file location
      --postgres.ssl.key string                the PostgreSQL ssl key file location
      --postgres.ssl.mode string               the PostgreSQL ssl mode (default "disable")
      --postgres.ssl.root_certificate string   the PostgreSQL ssl root certificate file location
      --postgres.username string               the PostgreSQL username (default "authelia")
      --sqlite.path string                     the SQLite database path
```

### SEE ALSO

* [authelia storage user totp](authelia_storage_user_totp.md)	 - Manage TOTP configurations

