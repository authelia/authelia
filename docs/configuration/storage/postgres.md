---
layout: default
title: PostgreSQL
parent: Storage backends
grand_parent: Configuration
nav_order: 2
---

# PostgreSQL

```yaml
storage:
    postgres:
        host: 127.0.0.1
        port: 3306
        database: authelia
        username: authelia
        # This secret can also be set using the env variables AUTHELIA_STORAGE_POSTGRES_PASSWORD
        password: mypassword
```
