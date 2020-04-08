---
layout: default
title: Regulation
parent: Configuration
nav_order: 7
---

# Regulation

**Authelia** can temporarily ban accounts when there are too many
authentication attempts. This helps prevent brute-force attacks.

## Configuration

```yaml
regulation:
    # The number of failed login attempts before user is banned.
    # Set it to 0 to disable regulation.
    max_retries: 3

    # The time range during which the user can attempt login before being banned.
    # The user is banned if the authentication failed `max_retries` times in a `find_time` seconds window.
    # Find Time accepts duration notation. See: https://docs.authelia.com/configuration/index.html#duration-notation-format
    find_time: 2m

    # The length of time before a banned user can sign in again.
    # Find Time accepts duration notation. See: https://docs.authelia.com/configuration/index.html#duration-notation-format
    ban_time: 5m
```

### Duration Notation

The configuration parameters find_time, and ban_time use duration notation. See the documentation
for [duration notation format](index.md#duration-notation-format) for more information.