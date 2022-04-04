---
layout: default
title: Authentication Backends
parent: Configuration
nav_order: 2
has_children: true
---

# Authentication Backends

There are two ways to store the users along with their password:

* LDAP: users are stored in remote servers like OpenLDAP, OpenAM or Microsoft Active Directory.
* File: users are stored in YAML file with a hashed version of their password.

## Configuration

```yaml
authentication_backend:
  disable_reset_password: false
  password_reset:
    custom_url: ""
  file: {}
  ldap: {}
```

## Options

### disable_reset_password
<div markdown="1">
type: boolean
{: .label .label-config .label-purple } 
default: false
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

This setting controls if users can reset their password from the web frontend or not.

### password_reset

#### custom_url
<div markdown="1">
type: string
{: .label .label-config .label-purple } 
default: ""
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

The custom password reset URL. This replaces the inbuilt password reset functionality and disables the endpoints if
this is configured to anything other than nothing or an empty string.

### file

The [file](file.md) authentication provider.

### ldap

The [LDAP](ldap.md) authentication provider.
