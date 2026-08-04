---
title: "authelia storage known-ips export"
description: "Reference for the authelia storage known-ips export command."
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

## authelia storage known-ips export

Export known IP addresses to a YAML file

### Synopsis

Export known IP addresses to a YAML file.

This subcommand allows exporting known IP addresses in order to back them up.

```
authelia storage known-ips export [flags]
```

### Examples

```
authelia storage known-ips export
authelia storage known-ips export --file export.yml
```

### Options

```
  -f, --file string   The file name for the YAML export (default "authelia.export.known-ips.yml")
  -h, --help          help for export
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

