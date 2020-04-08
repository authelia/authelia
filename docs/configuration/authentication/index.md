---
layout: default
title: Authentication backends
parent: Configuration
nav_order: 1
has_children: true
---

# Authentication Backends

There are two ways to store the users along with their password:

* LDAP: users are stored in remote servers like OpenLDAP, OpenAM or Microsoft Active Directory.
* File: users are stored in YAML file with a hashed version of their password.

## Disabling Reset Password

You can disable the reset password functionality for additional security as per this configuration:

```yaml
# The authentication backend to use for verifying user passwords
# and retrieve information such as email address and groups
# users belong to.
#
# There are two supported backends: 'ldap' and 'file'.
authentication_backend:
  # Disable both the HTML element and the API for reset password functionality
  disable_reset_password: true
```