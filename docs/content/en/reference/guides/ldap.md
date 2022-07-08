---
title: "LDAP"
description: "A reference guide on the LDAP implementation specifics"
lead: "This section contains reference documentation for Authelia's LDAP implementation specifics."
date: 2022-06-17T21:03:47+10:00
draft: false
images: []
menu:
  reference:
    parent: "guides"
weight: 220
toc: true
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

There are currently two implementations, `custom` and `activedirectory`. The `activedirectory` implementation
must be used if you wish to allow users to change or reset their password as Active Directory
uses a custom attribute for this, and an input format other implementations do not use. The long term
intention of this is to have logical defaults for various RFC implementations of LDAP.

### Filter replacements

Various replacements occur in the user and groups filter. The replacements either occur at startup or upon an LDAP
search.

#### Users filter replacements

|       Placeholder        |  Phase  |              Replacement              |
|:------------------------:|:-------:|:-------------------------------------:|
|   {username_attribute}   | startup |   The configured username attribute   |
|     {mail_attribute}     | startup |     The configured mail attribute     |
| {display_name_attribute} | startup | The configured display name attribute |
|         {input}          | search  |   The input into the username field   |

#### Groups filter replacements

| Placeholder | Phase  |                                Replacement                                |
|:-----------:|:------:|:-------------------------------------------------------------------------:|
|   {input}   | search |                     The input into the username field                     |
| {username}  | search | The username from the profile lookup obtained from the username attribute |
|    {dn}     | search |              The distinguished name from the profile lookup               |

### Defaults

The below tables describes the current attribute defaults for each implementation.

#### Attribute defaults

This table describes the attribute defaults for each implementation. i.e. the username_attribute is described by the
Username column.

| Implementation  |    Username    | Display Name | Mail | Group Name |
|:---------------:|:--------------:|:------------:|:----:|:----------:|
|     custom      |      N/A       | displayName  | mail |     cn     |
| activedirectory | sAMAccountName | displayName  | mail |     cn     |

#### Filter defaults

The filters are probably the most important part to get correct when setting up LDAP. You want to exclude disabled
accounts. The active directory example has two attribute filters that accomplish this as an example (more examples would
be appreciated). The userAccountControl filter checks that the account is not disabled and the pwdLastSet makes sure that
value is not 0 which means the password requires changing at the next login.

| Implementation  |                                                                          Users Filter                                                                           |                       Groups Filter                       |
|:---------------:|:---------------------------------------------------------------------------------------------------------------------------------------------------------------:|:---------------------------------------------------------:|
|     custom      |                                                                               N/A                                                                               |                            N/A                            |
| activedirectory | (&(&#124;({username_attribute}={input})({mail_attribute}={input}))(sAMAccountType=805306368)(!(userAccountControl:1.2.840.113556.1.4.803:=2))(!(pwdLastSet=0))) | (&(member={dn})(objectClass=group)(objectCategory=group)) |

*__Note:__* The Active Directory filter `(sAMAccountType=805306368)` is exactly the same as
`(&(objectCategory=person)(objectClass=user))` except that the former is more performant, you can read more about this
and other Active Directory filters on the [TechNet wiki](https://social.technet.microsoft.com/wiki/contents/articles/5392.active-directory-ldap-syntax-filters.aspx).
