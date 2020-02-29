---
layout: default
title: Duo Push Notifications
parent: Configuration
nav_order: 3
---

# Duo Push Notifications

Authelia supports mobile push notifications relying on [Duo].

Follow the instructions in the dedicated [documentation](../features/2fa/push-notifications.md)
to know how to set up push notifications in Authelia.

## Configuration

The configuration is as follows:

    duo_api:
        hostname: api-123456789.example.com
        integration_key: ABCDEF
        secret_key: 1234567890abcdefghifjkl

The secret key is shown as an example but you'd better set it using an environment
variable as described [here](./secrets.md).

[Duo]: https://duo.com/