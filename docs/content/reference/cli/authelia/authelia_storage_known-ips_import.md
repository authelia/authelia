---
title: "authelia storage known-ips import"
description: "Reference for the authelia storage known-ips import command."
lead: ""
date: 2026-04-02T15:48:21+11:00
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

## authelia storage known-ips import

Import known IP addresses from a YAML file

### Synopsis

Import known IP addresses from a YAML file.

This subcommand allows you to import known IP addresses from a YAML file. The YAML file can either be automatically
generated using the authelia storage known-ips export command, or manually provided the file is in the same format.

```
authelia storage known-ips import <filename> [flags]
```

### Examples

```
authelia storage known-ips import
authelia storage known-ips import authelia.export.known-ips.yml
```

### Options

```
  -h, --help   help for import
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

* [authelia storage known-ips](authelia_storage_known-ips.md)	 - Manages known IP addresses

