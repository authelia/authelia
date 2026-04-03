---
title: "Access Control"
description: "Access Control is the main authorization system in Authelia."
summary: "Access Control is the main authorization system in Authelia."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 310
toc: true
aliases:
  - /docs/features/access-control.html
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

__Authelia__ allows defining fine-grained rules-based access control policies. This list of rules is tested against
any requests protected by Authelia and defines the level of authentication the user must pass to get authorization to
the resource.

## Example

For instance a rule can look like this:

```yaml {title="configuration.yml"}
access_control:
  rules:
    - domain: 'dev.example.com'
      resources:
        - '^/groups/dev/.*$'
      subject: 'group:dev'
      policy: 'two_factor'
      methods:
        - 'GET'
        - 'POST'
      networks:
        - '192.168.1.0/24'
```

This rule matches when the request targets the domain `dev.example.com`, the path matches the regular expression
`^/groups/dev/.*$`, the user is a member of the `dev` group, the request comes from a client on the 192.168.1.0/24
subnet, and the HTTP method verb is GET or POST. In that case, a two-factor policy is applied requiring the user to
authenticate with two factors.

## Configuration

Please check the dedicated [documentation](../../configuration/security/access-control.md).
