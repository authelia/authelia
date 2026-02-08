---
title: "Active Directory"
description: ""
summary: ""
date: 2025-05-22T10:12:47+00:00
draft: false
images: []
weight: 752
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

[Active Directory] is supported by __Authelia__.

*__Important:__ When using these guides, it's important to recognize that we cannot provide a guide for every possible
method of deploying an LDAP server. These guides show a suggested setup only, and you need to understand the LDAP
configuration and customize it to your needs. To-that-end, we include links to the official documentation specific to
the LDAP implementation throughout this documentation and in the [See Also](#see-also) section.*

*__Important:__ This guide makes use of a default configuration. Check the [Defaults](#defaults) section
and make adjustments according to your needs.*

## Assumptions and Adaptation

This guide makes a few assumptions. These assumptions may require adaptation in more advanced and complex scenarios. We
can not reasonably have examples for every advanced configuration option that exists. Some of these values can
automatically be replaced with documentation variables.

The following are the assumptions we make:

- The LDAP implementation to be used with Authelia is fully setup and reachable by Authelia.
- All services are part of the `example.com` domain:
  - This domain and the subdomains will have to be adapted in all examples to match your specific domains unless you're
    just testing or you want to use that specific domain.

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [authentication backend configuration] for use with
[Active Directory] which will operate with the application example:

```yaml {title="configuration.yml"}
authentication_backend:
  ldap:
    implementation: 'activedirectory'
    address: 'ldaps://ldap.example.com'
    base_dn: 'DC=example,DC=com'
    user: 'CN=authelia,OU=people,DC=example,DC=com'
    password: 'insecure_secret'
```

### Application

Create a service user within the application with a complex password. Use the users Distinguished Name as a username,
and make sure the user has the appropriate permissions to perform the following actions:

- Read the attributes of users and groups that are meant to be able to use Authelia.
- Change the password of users provided the functionality to reset passwords is desired.

See the Microsoft [Active Directory] documentation on how to configure permissions for the newly created user.

### Defaults

The below tables describes the current attribute defaults for the [Active Directory] implementation.

#### Attribute defaults

This table describes the attribute defaults for each implementation. i.e. the username_attribute is described by the
Username column.

|    Username    | Display Name | Mail | Group Name | Distinguished Name | Member Of |
|:--------------:|:------------:|:----:|:----------:|:------------------:|:---------:|
| sAMAccountName | displayName  | mail |     cn     | distinguishedName  | memberOf  |

#### Filter defaults

The filters are probably the most important part to get correct when setting up LDAP. You want to exclude accounts under
the following conditions:

- The account is disabled or locked:
  - `(!(userAccountControl:1.2.840.113556.1.4.803:=2))`
- Their password is expired:
  - `(!(pwdLastSet=0))`
- Their account is expired:
  - `(|(!(accountExpires=*))(accountExpires=0)(accountExpires>={date-time:microsoft-nt}))`

##### Users Filter

```text
(&(|({username_attribute}={input})({mail_attribute}={input}))(sAMAccountType=805306368)(!(userAccountControl:1.2.840.113556.1.4.803:=2))(!(pwdLastSet=0))(|(!(accountExpires=*))(accountExpires=0)(accountExpires>={date-time:microsoft-nt})))
```

##### Groups Filter

```text
(&(member={dn})(|(sAMAccountType=268435456)(sAMAccountType=536870912)))
```

##### Microsoft Active Directory sAMAccountType

| Account Type Value |               Description               |               Equivalent Filter                |
|:------------------:|:---------------------------------------:|:----------------------------------------------:|
|     268435456      | Global/Universal Security Group Objects |                      N/A                       |
|     536870912      |   Domain Local Security Group Objects   |                      N/A                       |
|     805306368      |          Normal User Accounts           | `(&(objectCategory=person)(objectClass=user))` |

## See Also

- Account Type Values: [Microsoft Learn](https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-samr/e742be45-665d-4576-b872-0bc99d1e1fbe)
- LDAP Syntax Filters: [Microsoft TechNet Wiki](https://social.technet.microsoft.com/wiki/contents/articles/5392.active-directory-ldap-syntax-filters.aspx)

[Authelia]: https://www.authelia.com
[Active Directory]: https://learn.microsoft.com/en-us/windows-server/identity/ad-ds/active-directory-domain-services
[authentication backend configuration]: ../../../configuration/first-factor/ldap.md
