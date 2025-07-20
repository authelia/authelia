---
title: "Seafile"
description: "Trusted Header SSO Integration for Seafile"
summary: ""
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
menu:
integration:
parent: "trusted-header-sso"
weight: 420
toc: true
support:
  level: community
  versions: true
  integration: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Introduction

This is a guide on integration of __Authelia__ and [Seafile] via the trusted header SSO authentication.

As with all guides in this section it's important you read the [introduction](../introduction.md) first.

## Tested Versions

* Authelia:
  * v4.35.5
* [Seafile] Server:
  * 9.0.4

## Before You Begin

This example makes the following assumptions:

* __Application Root URL:__ `https://seafile.{{< sitevar name="domain" nojs="example.com" >}}/`
* __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
* __User Email Domain:__ `@{{< sitevar name="domain" nojs="example.com" >}}`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

To configure [Seafile] to trust the `Remote-User` and `Remote-Email` header do the following:

1. Configure `seahub_settings.py` and adjust the following  settings:
```python
ENABLE_REMOTE_USER_AUTHENTICATION = True

# Optional, HTTP header, which is configured in your web server conf file,
# used for Seafile to get user's unique id, default value is 'HTTP_REMOTE_USER'.
REMOTE_USER_HEADER = 'HTTP_REMOTE_USER'

# Optional, when the value of HTTP_REMOTE_USER is not a valid email address，
# Seafile will build a email-like unique id from the value of 'REMOTE_USER_HEADER'
# and this domain, e.g. user1@{{< sitevar name="domain" nojs="example.com" >}}.
REMOTE_USER_DOMAIN = '{{< sitevar name="domain" nojs="example.com" >}}'

# Optional, whether to create new user in Seafile system, default value is True.
# If this setting is disabled, users doesn't preexist in the Seafile DB cannot login.
# The admin has to first import the users from external systems like LDAP.
REMOTE_USER_CREATE_UNKNOWN_USER = True

# Optional, whether to activate new user in Seafile system, default value is True.
# If this setting is disabled, user will be unable to login by default.
# the administrator needs to manually activate this user.
REMOTE_USER_ACTIVATE_USER_AFTER_CREATION = True

# Optional, map user attribute in HTTP header and Seafile's user attribute.
REMOTE_USER_ATTRIBUTE_MAP = {
    'HTTP_REMOTE_NAME': 'name',
    'HTTP_REMOTE_EMAIL': 'contact_email',
}
```

## See Also

* [Seafile Remote User Docs](https://manual.seafile.com/latest/config/remote_user/)

[Seafile]: https://www.seafile.com/
