---
title: "Identity Validation"
description: "Identity Validation Configuration"
summary: "Authelia uses multiple methods to verify the identity of users to prevent a malicious user from performing actions on behalf of them. This section describes these methods."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 105100
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
identity_validation:
  elevated_session: {}
  reset_password: {}
```

## Options

The two areas protected by the validation methods are:

- [Elevated Session](elevated-session.md) which prevents a logged in user from performing privileged actions without
  first proving their identity.
- [Reset Password](reset-password.md) which prevents an anonymous user from performing the password reset for a user
  without first proving their identity.
