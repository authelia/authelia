---
title: "Storage"
description: "Storage Configuration"
summary: "Configuring the SQL Storage."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 107100
toc: true
aliases:
  - /docs/configuration/storage/
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

__Authelia__ supports multiple storage backends. The backend is used to store user preferences, 2FA device handles and
secrets, authentication logs, etc...

The available storage backends are listed in the table of contents below.

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
storage:
  encryption_key: 'a_very_important_secret'
  local: {}
  mysql: {}
  postgres: {}
```

## Options

This section describes the individual configuration options.

### encryption_key

{{< confkey type="string" required="yes" secret="yes" >}}

The encryption key used to encrypt data in the database. We encrypt data by creating a sha256 checksum of the provided
value, and use that to encrypt the data with the AES-GCM 256bit algorithm.

The minimum length of this key is 20 characters.

It's __strongly recommended__ this is a
[Random Alphanumeric String](../../reference/guides/generating-secure-values.md#generating-a-random-alphanumeric-string) with 64 or more
characters.

See [security measures](../../overview/security/measures.md#storage-security-measures) for more information.

### postgres

See [PostgreSQL](postgres.md).

### local

See [SQLite](sqlite.md).

### mysql

See [MySQL](mysql.md).
