---
layout: default
title: Regulation
parent: Configuration
nav_order: 7
---

# Regulation

**Authelia** can temporarily ban accounts when there was too many
authentication attempts. This helps prevent brute force attacks.

## Configuration

```yaml
regulation:
    # The number of failed login attempts before user is banned.
    # Set it to 0 to disable regulation.
    max_retries: 3

    # The time range during which the user can attempt login before being banned.
    # The user is banned if the authentication failed `max_retries` times in a `find_time` seconds window.
    find_time: 120

    # The length of time before a banned user can sign in again.
    ban_time: 300
```