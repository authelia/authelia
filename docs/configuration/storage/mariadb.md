---
layout: default
title: MariaDB
parent: Storage backends
grand_parent: Configuration
nav_order: 1
---

# MariaDB

```yaml
storage:
  mysql:
    host: 127.0.0.1
    port: 3306
    database: authelia
    username: authelia
    # Password can also be set using a secret: https://docs.authelia.com/configuration/secrets.html
    password: mypassword
```

## IPv6 Addresses

If utilising an IPv6 literal address it must be enclosed by square brackets and quoted:
```yaml
host: "[fd00:1111:2222:3333::1]"
```

## Loading a password from a secret instead of inside the configuration

Password can also be defined using a [secret](../secrets.md).
