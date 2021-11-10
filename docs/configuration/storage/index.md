---
layout: default
title: Storage Backends
parent: Configuration
nav_order: 14
has_children: true
---

**Authelia** supports multiple storage backends. The backend is used to store user preferences, 2FA device handles and 
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
<div markdown="1">
type: string
{: .label .label-config .label-purple }
required: yes
{: .label .label-config .label-red }
</div>

The encryption key used to encrypt data in the database. It has a minimum length of 20 and must be provided.

### local
See [SQLite](./sqlite.md).

### mysql
See [MySQL](./mysql.md).

### postgres
See [PostgreSQL](./postgres.md).