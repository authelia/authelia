---
title: "User Attributes"
description: "User Attributes Definitions Configuration"
summary: "Authelia allows configuring reusable user attribute definitions."
date: 2024-10-17T21:43:20+11:00
draft: false
images: []
weight: 199100
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

The user attributes section configures custom user attributes.

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
definitions:
  user_attributes:
    custom_name:
      expression: '"admin" in groups'
```

## Options

This section describes the individual configuration options. Currently these attribute definitions are only used in the
[OpenID Connect 1.0 Provider](../identity-providers/openid-connect/provider.md#claims_policies)
