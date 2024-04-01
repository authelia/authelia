---
title: "LDAP"
description: "A reference guide on the LDAP implementation specifics"
summary: "This section contains reference documentation for Authelia's LDAP implementation specifics."
date: 2022-06-17T21:03:47+10:00
draft: false
images: []
weight: 220
toc: true
aliases:
  - /r/ldap
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Binding

When it comes to LDAP there are several considerations for deciding how to bind to the LDAP server.

### Unauthenticated Binding

The most insecure method is unauthenticated binds. They are generally considered insecure due to the fact allowing them
at all ensures anyone with any level of network access can easily obtain objects and their attributes.

Authelia does support unauthenticated binds but it is not by default, you must configure the
[permit_unauthenticated_bind](../../configuration/first-factor/ldap.md#permit_unauthenticated_bind) configuration
option.

### End-User Binding

One method to bind to the server that is favored by a lot of people is binding to the LDAP server as the end user. While
this is more secure than methods such as [Unauthenticated Binding](#unauthenticated-binding) the drawback is that it can
only be used securely at the time the user enters their credentials. Storing a password in memory in general is not very
secure and prone to breakage due to outside influences (i.e. the user changes their password).

In addition, this method is not compatible with the password reset / forgot password flow at all (not to be confused
with a change password flow).

Authelia doesn't currently support such a binding method excluding for checking user passwords.

### Service-User Binding

This is the most common method of binding to LDAP. This involves setting up a special service user with a complex
password which has the minimum permissions required to do the tasks required.

Authelia primarily supports this method.

## Implementation Guide

The following implementations exist:

- `custom`:
  - Not specific to any particular LDAP provider
- `activedirectory`:
  - Specific configuration defaults for [Active Directory]
  - Special implementation details:
    - Includes a special encoding format required for changing passwords with [Active Directory]
- `rfc2307bis`:
  - Specific configuration defaults for [RFC2307bis]
  - No special implementation details
- `freeipa`:
  - Specific configuration defaults for [FreeIPA]
  - No special implementation details
- `lldap`:
  - Specific configuration defaults for [lldap]
  - No special implementation details
- `glauth`:
  - Specific configuration defaults for [GLAuth]
  - No special implementation details

[Active Directory]: https://learn.microsoft.com/en-us/windows-server/identity/ad-ds/active-directory-domain-services
[FreeIPA]: https://www.freeipa.org/
[lldap]: https://github.com/lldap/lldap
[GLAuth]: https://glauth.github.io/
[RFC2307bis]: https://datatracker.ietf.org/doc/html/draft-howard-rfc2307bis-02

### Group Search Modes

There are currently two group search modes that exist.

#### Search Mode: filter

The `filter` search mode is the default search mode. Generally this is recommended.

#### Search Mode: memberof

The `memberof` search mode is a special search mode. Generally this is discouraged and is currently experimental.

Some systems provide a `memberOf` attribute which may include additional groups that the user is a member of. This
search mode allows using this attribute as a method to determine their groups. How it works is the search is performed
against the base with the subtree scope and the groups filter must include one of the `{memberof:*}` replacements, and
the distinguished names of the results from the search are compared (case-insensitive) against the users `memberOf`
attribute to determine if they are members.

This means:

1. The groups still must be in the search base that you have configured.
2. The `memberOf` attribute *__MUST__* include the distinguished name of the group.
3. If the `{memberof:dn}` replacement is used:
   1. The distinguished name *__MUST__* be searchable by your directory server.
4. The first relative distinguished name of the distinguished name *__MUST__* be search

### Filter replacements

Various replacements occur in the user and groups filter. The replacements either occur at startup or upon an LDAP
search which is indicated by the phase column.

The phases exist to optimize performance. The replacements in the startup phase are replaced once before the connection
is ever established. In addition to this, during the startup phase we purposefully check the filters for which search
phase replacements exist so we only have to check if the replacement is necessary once, and we don't needlessly perform
every possible replacement on every search regardless of if it's needed or not.

#### General filter replacements

|          Placeholder           |  Phase  |                 Replacement                 |
|:------------------------------:|:-------:|:-------------------------------------------:|
| {distinguished_name_attribute} | startup | The configured distinguished name attribute |
|      {username_attribute}      | startup |      The configured username attribute      |
|        {mail_attribute}        | startup |        The configured mail attribute        |
|    {display_name_attribute}    | startup |    The configured display name attribute    |
|     {member_of_attribute}      | startup |     The configured member of attribute      |
|            {input}             | search  |      The input into the username field      |

#### Users filter replacements

|          Placeholder           |  Phase  |                                                   Replacement                                                    |
|:------------------------------:|:-------:|:----------------------------------------------------------------------------------------------------------------:|
|    {date-time:generalized}     | search  |          The current UTC time formatted as a LDAP generalized time in the format of `20060102150405.0Z`          |
|        {date-time:unix}        | search  |                                    The current time formatted as a Unix epoch                                    |
|    {date-time:microsoft-nt}    | search  | The current time formatted as a Microsoft NT epoch which is used by some Microsoft [Active Directory] attributes |

#### Groups filter replacements

|  Placeholder   | Phase  |                                                                     Replacement                                                                      |
|:--------------:|:------:|:----------------------------------------------------------------------------------------------------------------------------------------------------:|
|   {username}   | search |                                      The username from the profile lookup obtained from the username attribute                                       |
|      {dn}      | search |                                                    The distinguished name from the profile lookup                                                    |
| {memberof:dn}  | search |                                                            See the detailed section below                                                            |
| {memberof:rdn} | search | Only allowed with the `memberof` search method and contains the first relative distinguished name of every `memberOf` entry a use has in parenthesis |

##### memberof:dn

Requirements:

1. Must be using the `memberof` search mode.
2. Must have the distinguished name attribute configured in Authelia.
3. Directory server must support searching by the distinguished name attribute (many directory services *__DO NOT__*
   have a distinguished name attribute).

##### memberof:rdn

Requirements:

1. Must be using the `memberof` search mode.
2. Directory server must support searching by the first relative distinguished name as an attribute.

Splits every `memberOf` value to obtain th e first relative distinguished name and joins all of those after surrounding
them in parenthesis. This makes the general suggested filter pattern for this particular replacement
`(|{memberof:rdn})`. The format of this value is as follows:

```text
(<RDN>)
```

For example if the user has the following distinguished names in their object:

- CN=abc,OU=groups,DC=example,DC=com
- CN=xyz,OU=groups,DC=example,DC=com

The value will be replaced with `(CN=abc)(CN=xyz)` which using the suggested pattern for the filter becomes
`(|(CN=abc)(CN=xyz))` which will then return any user that as a `CN` of `abc` or `xyz`.

### Defaults

The below tables describes the current attribute defaults for each implementation.

#### Search Base defaults

The following set defaults for the `additional_users_dn` and `additional_groups_dn` values.

| Implementation |   Users   |  Groups   |
|:--------------:|:---------:|:---------:|
|     lldap      | OU=people | OU=groups |

#### Attribute defaults

This table describes the attribute defaults for each implementation. i.e. the username_attribute is described by the
Username column.

| Implementation  |    Username    | Display Name | Mail | Group Name | Distinguished Name | Member Of |
|:---------------:|:--------------:|:------------:|:----:|:----------:|:------------------:|:---------:|
|     custom      |      N/A       | displayName  | mail |     cn     |        N/A         |    N/A    |
| activedirectory | sAMAccountName | displayName  | mail |     cn     | distinguishedName  | memberOf  |
|   rfc2307bis    |      uid       | displayName  | mail |     cn     |        N/A         | memberOf  |
|     freeipa     |      uid       | displayName  | mail |     cn     |        N/A         | memberOf  |
|      lldap      |      uid       |      cn      | mail |     cn     |        N/A         | memberOf  |
|     glauth      |       cn       | description  | mail |     cn     |        N/A         | memberOf  |

#### Filter defaults

The filters are probably the most important part to get correct when setting up LDAP. You want to exclude accounts under
the following conditions:

- The account is disabled or locked:
  - The [Active Directory] implementation achieves this via the `(!(userAccountControl:1.2.840.113556.1.4.803:=2))` filter.
  - The [FreeIPA] implementation achieves this via the `(!(nsAccountLock=TRUE))` filter.
  - The [GLAuth] implementation achieves this via the `(!(accountStatus=inactive))` filter.
  - The following implementations have no suitable attribute for this as far as we're aware:
    - [RFC2307bis]
    - [lldap]
- Their password is expired:
  - The [Active Directory] implementation achieves this via the `(!(pwdLastSet=0))` filter.
  - The [FreeIPA] implementation achieves this via the `(krbPasswordExpiration>={date-time:generalized})` filter.
  - The [RFC2307bis] implementation achieves this via the `(!(pwdReset=TRUE))` filter.
  - The following implementations have no suitable attribute for this as far as we're aware:
    - [GLAuth]
    - [lldap]
- Their account is expired:
  - The [Active Directory] implementation achieves this via the `(|(!(accountExpires=*))(accountExpires=0)(accountExpires>={date-time:microsoft-nt}))` filter.
  - The [FreeIPA] implementation achieves this via the `(|(!(krbPrincipalExpiration=*))(krbPrincipalExpiration>={date-time:generalized}))` filter.
  - The following implementations have no suitable attribute for this as far as we're aware:
    - [RFC2307bis]
    - [GLAuth]
    - [lldap]

| Implementation  |                                                                                                                       Users Filter                                                                                                                       |                                                               Groups Filter                                                               |
|:---------------:|:--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------:|:-----------------------------------------------------------------------------------------------------------------------------------------:|
|     custom      |                                                                                                                           N/A                                                                                                                            |                                                                    N/A                                                                    |
| activedirectory | (&(&#124;({username_attribute}={input})({mail_attribute}={input}))(sAMAccountType=805306368)(!(userAccountControl:1.2.840.113556.1.4.803:=2))(!(pwdLastSet=0))(&#124;(!(accountExpires=*))(accountExpires=0)(accountExpires>={date-time:microsoft-nt}))) |                               (&(member={dn})(&#124;(sAMAccountType=268435456)(sAMAccountType=536870912)))                                |
|   rfc2307bis    |                                                (&(&#124;({username_attribute}={input})({mail_attribute}={input}))(&#124;(objectClass=inetOrgPerson)(objectClass=organizationalPerson))(!(pwdReset=TRUE)))                                                | (&(&#124;(member={dn})(uniqueMember={dn}))(&#124;(objectClass=groupOfNames)(objectClass=groupOfUniqueNames)(objectClass=groupOfMembers))) |
|     freeipa     |   (&(&#124;({username_attribute}={input})({mail_attribute}={input}))(objectClass=person)(!(nsAccountLock=TRUE))(krbPasswordExpiration>={date-time:generalized})(&#124;(!(krbPrincipalExpiration=*))(krbPrincipalExpiration>={date-time:generalized})))   |                                                (&(member={dn})(objectClass=groupOfNames))                                                 |
|      lldap      |                                                                                 (&(&#124;({username_attribute}={input})({mail_attribute}={input}))(objectClass=person))                                                                                  |                                                (&(member={dn})(objectClass=groupOfNames))                                                 |
|     glauth      |                                                                 (&(&#124;({username_attribute}={input})({mail_attribute}={input}))(objectClass=posixAccount)(!(accountStatus=inactive)))                                                                 |                                              (&(uniqueMember={dn})(objectClass=posixGroup))                                               |

##### Microsoft Active Directory sAMAccountType

| Account Type Value |               Description               |               Equivalent Filter                |
|:------------------:|:---------------------------------------:|:----------------------------------------------:|
|     268435456      | Global/Universal Security Group Objects |                      N/A                       |
|     536870912      |   Domain Local Security Group Objects   |                      N/A                       |
|     805306368      |          Normal User Accounts           | `(&(objectCategory=person)(objectClass=user))` |

*__References:__*
- Account Type Values: [Microsoft Learn](https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-samr/e742be45-665d-4576-b872-0bc99d1e1fbe).
- LDAP Syntax Filters: [Microsoft TechNet Wiki](https://social.technet.microsoft.com/wiki/contents/articles/5392.active-directory-ldap-syntax-filters.aspx)
