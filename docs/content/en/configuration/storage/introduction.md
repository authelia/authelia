---
title: "Storage"
description: "Storage Configuration"
lead: "Configuring the SQL Storage."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  configuration:
    parent: "storage"
weight: 106100
toc: true
aliases:
  - /docs/configuration/storage/
---

__Authelia__ supports multiple storage backends. The backend is used to store user preferences, 2FA device handles and
secrets, authentication logs, etc...

The available storage backends are listed in the table of contents below.

## Configuration

```yaml
storage:
  encryption_key: a_very_important_secret
  local: {}
  mysql: {}
  postgres: {}
```

## Options

### encryption_key

{{< confkey type="string" required="yes" >}}

*__Important Note:__ This can also be defined using a [secret](../methods/secrets.md) which is __strongly recommended__
especially for containerized deployments.*

The encryption key used to encrypt data in the database. We encrypt data by creating a sha256 checksum of the provided
value, and use that to encrypt the data with the AES-GCM 256bit algorithm.

The minimum length of this key is 20 characters.

It's __strongly recommended__ this is a
[Random Alphanumeric String](../miscellaneous/guides.md#generating-a-random-alphanumeric-string) with 64 or more
characters.

See [securty measures](../../overview/security/measures.md#storage-security-measures) for more information.

### postgres

See [PostgreSQL](postgres.md).

### local

See [SQLite](sqlite.md).

### mysql

See [MySQL](mysql.md).
