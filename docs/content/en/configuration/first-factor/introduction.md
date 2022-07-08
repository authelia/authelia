---
title: "First Factor"
description: "Configuring Authelia First Factor Authentication."
lead: "Authelia uses a username and password for a first factor method. This section describes configuring this."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  configuration:
    parent: "first-factor"
weight: 102100
toc: true
aliases:
  - /c/1fa
  - /docs/configuration/authentication/
---

There are two ways to integrate *Authelia* with an authentication backend:

* [LDAP](ldap.md): users are stored in remote servers like [OpenLDAP], [OpenDJ], [FreeIPA], or
  [Microsoft Active Directory].
* [File](file.md): users are stored in [YAML] file with a hashed version of their password.

## Configuration

```yaml
authentication_backend:
  refresh_interval: 5m
  password_reset:
    disable: false
    custom_url: ""
```

## Options

### refresh_interval

{{< confkey type="duration" default="5m" required="no" >}}

This setting controls the interval at which details are refreshed from the backend. Particularly useful for
[LDAP](#ldap).

### password_reset

#### disable

{{< confkey type="boolean" default="false" required="no" >}}

This setting controls if users can reset their password from the web frontend or not.

#### custom_url

{{< confkey type="string" required="no" >}}

The custom password reset URL. This replaces the inbuilt password reset functionality and disables the endpoints if
this is configured to anything other than nothing or an empty string.

### file

The [file](file.md) authentication provider.

### ldap

The [LDAP](ldap.md) authentication provider.

[OpenLDAP]: https://www.openldap.org/
[OpenDJ]: https://www.openidentityplatform.org/opendj
[FreeIPA]: https://www.freeipa.org/
[Microsoft Active Directory]: https://docs.microsoft.com/en-us/windows-server/identity/ad-ds/ad-ds-getting-started
[YAML]: https://yaml.org/
