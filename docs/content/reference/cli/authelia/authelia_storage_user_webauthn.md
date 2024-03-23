---
title: "authelia storage user webauthn"
description: "Reference for the authelia storage user webauthn command."
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

## authelia storage user webauthn

Manage WebAuthn credentials

### Synopsis

Manage WebAuthn credentials.

This subcommand allows interacting with WebAuthn credentials.

### Examples

```
authelia storage user webauthn --help
```

### Options

```
  -h, --help   help for webauthn
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

* [authelia storage user](authelia_storage_user.md)	 - Manages user settings
* [authelia storage user webauthn delete](authelia_storage_user_webauthn_delete.md)	 - Delete a WebAuthn credential
* [authelia storage user webauthn export](authelia_storage_user_webauthn_export.md)	 - Perform exports of the WebAuthn credentials
* [authelia storage user webauthn import](authelia_storage_user_webauthn_import.md)	 - Perform imports of the WebAuthn credentials
* [authelia storage user webauthn list](authelia_storage_user_webauthn_list.md)	 - List WebAuthn credentials

