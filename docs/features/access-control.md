---
layout: default
title: Access Control
parent: Features
nav_order: 5
---

# Access Control

**Authelia** allows to define a fine-grained rule-based access control policy in
configuration. This list of rules is tested against any requests protected by
Authelia and defines the level of authentication the user must pass to get access
to the resource.

For instance a rule can look like this:

    - domain: dev.example.com
      resources:
        - "^/groups/dev/.*$"
      subject: "group:dev"
      policy: two_factor

This rule matches when the request targets the domain `dev.example.com` and the path
matches the regular expression `^/groups/dev/.*$`. In that case, a two-factor policy
is applied requiring the user to authenticate with two factors.

## Configuration

Please check the dedicated [documentation](./deployment/configuration/access-control.md)
