---
title: "authelia storage known-ips"
description: "Reference for the authelia storage known-ips command."
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

## authelia storage known-ips

Manages known IP addresses

### Synopsis

Manages known IP addresses.

This subcommand allows listing, adding, deleting, exporting, importing, and pruning known IP addresses used for the
login from a new IP notification feature.

### Examples

```
authelia storage known-ips --help
```

### Options

```
  -h, --help   help for known-ips
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

* [authelia storage](authelia_storage.md)	 - Manage the Authelia storage
* [authelia storage known-ips add](authelia_storage_known-ips_add.md)	 - Add a known IP address
* [authelia storage known-ips delete](authelia_storage_known-ips_delete.md)	 - Delete a known IP address
* [authelia storage known-ips export](authelia_storage_known-ips_export.md)	 - Export known IP addresses to a YAML file
* [authelia storage known-ips import](authelia_storage_known-ips_import.md)	 - Import known IP addresses from a YAML file
* [authelia storage known-ips list](authelia_storage_known-ips_list.md)	 - List known IP addresses
* [authelia storage known-ips prune](authelia_storage_known-ips_prune.md)	 - Prune expired known IP addresses

