---
title: "authelia storage user totp export png"
description: "Reference for the authelia storage user totp export png command."
lead: ""
date: 2024-03-14T06:00:14+11:00
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

## authelia storage user totp export png

Perform exports of the TOTP configurations to QR code PNG images

### Synopsis

Perform exports of the TOTP configurations to QR code PNG images.

This subcommand allows exporting TOTP configurations to PNG images with QR codes which represent the appropriate URI so they can be scanned.

```
authelia storage user totp export png [flags]
```

### Examples

```
authelia storage user totp export png
authelia storage user totp export png --directory example/dir
authelia storage user totp export png --config config.yml
authelia storage user totp export png --encryption-key b3453fde-ecc2-4a1f-9422-2707ddbed495 --postgres.address tcp://postgres:5432 --postgres.password autheliapw
```

### Options

```
      --directory string   The directory where all exported png files will be saved to
  -h, --help               help for png
```

### Options inherited from parent commands

```
  -c, --config strings                        configuration files or directories to load, for more information run 'authelia -h authelia config' (default [configuration.yml])
      --config.experimental.filters strings   list of filters to apply to all configuration files, for more information run 'authelia -h authelia filters'
      --encryption-key string                 the storage encryption key to use
      --mysql.address string                  the MySQL server address (default "tcp://127.0.0.1:3306")
      --mysql.database string                 the MySQL database name (default "authelia")
      --mysql.password string                 the MySQL password
      --mysql.username string                 the MySQL username (default "authelia")
      --postgres.address string               the PostgreSQL server address (default "tcp://127.0.0.1:5432")
      --postgres.database string              the PostgreSQL database name (default "authelia")
      --postgres.password string              the PostgreSQL password
      --postgres.schema string                the PostgreSQL schema name (default "public")
      --postgres.username string              the PostgreSQL username (default "authelia")
      --sqlite.path string                    the SQLite database path
```

### SEE ALSO

* [authelia storage user totp export](authelia_storage_user_totp_export.md)	 - Perform exports of the TOTP configurations

