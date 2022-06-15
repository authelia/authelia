---
title: "LDAP"
description: "Configuring LDAP"
lead: "Authelia supports an LDAP server based first factor user provider. This section describes configuring this."
date: 2022-03-20T12:52:27+11:00
draft: false
images: []
menu:
  configuration:
    parent: "first-factor"
weight: 102200
toc: true
aliases:
  - /c/ldap
  - /docs/configuration/authentication/ldap.html
---

## Configuration

```yaml
authentication_backend:
  ldap:
    implementation: custom
    url: ldap://127.0.0.1
    timeout: 5s
    start_tls: false
    tls:
      server_name: ldap.example.com
      skip_verify: false
      minimum_version: TLS1.2
    base_dn: DC=example,DC=com
    additional_users_dn: ou=users
    users_filter: (&({username_attribute}={input})(objectClass=person))
    username_attribute: uid
    mail_attribute: mail
    display_name_attribute: displayName
    additional_groups_dn: ou=groups
    groups_filter: (&(member={dn})(objectClass=groupOfNames))
    group_name_attribute: cn
    permit_referrals: false
    user: CN=admin,DC=example,DC=com
    password: password
```

## Options

### implementation

{{< confkey type="string" default="custom" required="no" >}}

Configures the LDAP implementation used by Authelia.

See the [Implementation Guide](#implementation-guide) for information.

### url

{{< confkey type="string" required="yes" >}}

The LDAP URL which consists of a scheme, address, and port. Format is `<scheme>://<address>:<port>` or
`<scheme>://<address>` where scheme is either `ldap` or `ldaps`.

```yaml
authentication_backend:
  ldap:
    url: ldaps://dc1.example.com
```

If utilising an IPv6 literal address it must be enclosed by square brackets:

```yaml
authentication_backend:
  ldap:
    url: ldap://[fd00:1111:2222:3333::1]
```

### timeout

{{< confkey type="duration" default="5s" required="no" >}}

The timeout for dialing an LDAP connection.

### start_tls

{{< confkey type="boolean" default="false" required="no" >}}

Enables use of the LDAP StartTLS process which is not commonly used. You should only configure this if you know you need
it. The initial connection will be over plain text, and *Authelia* will try to upgrade it with the LDAP server. LDAPS
URL's are slightly more secure.

### tls

Controls the TLS connection validation process. You can see how to configure the tls
section [here](../prologue/common.md#tls-configuration).

### base_dn

{{< confkey type="string" required="yes" >}}

Sets the base distinguished name container for all LDAP queries. If your LDAP domain is example.com this is usually
`DC=example,DC=com`, however you can fine tune this to be more specific for example to only include objects inside the
authelia OU: `OU=authelia,DC=example,DC=com`. This is prefixed with the [additional_users_dn](#additional_users_dn) for
user searches and [additional_groups_dn](#additional_groups_dn) for groups searches.

### additional_users_dn

{{< confkey type="string" required="no" >}}

Additional LDAP path to append to the [base_dn](#base_dn) when searching for users. Useful if you want to restrict
exactly which OU to get users from for either security or performance reasons. For example setting it to
`OU=users,OU=people` with a base_dn set to `DC=example,DC=com` will mean user searches will occur in
`OU=users,OU=people,DC=example,DC=com`.

### users_filter

{{< confkey type="string" required="situational" >}}

*__Note:__ This option is technically required however the [implementation](#implementation) option can implicitly set a
default negating this requirement. Refer to the [filter defaults](#filter-defaults) for more information.*

The LDAP filter to narrow down which users are valid. This is important to set correctly as to exclude disabled users.
The default value is dependent on the [implementation](#implementation), refer to the
[attribute defaults](#attribute-defaults) for more information.

### username_attribute

{{< confkey type="string" required="situational" >}}

*__Note:__ This option is technically required however the [implementation](#implementation) option can implicitly set a
default negating this requirement. Refer to the [attribute defaults](#attribute-defaults) for more information.*

The LDAP attribute that maps to the username in *Authelia*. This must contain the `{username_attribute}`
[placeholder](#users-filter-replacements).

### mail_attribute

{{< confkey type="string" required="situational" >}}

*__Note:__ This option is technically required however the [implementation](#implementation) option can implicitly set a
default negating this requirement. Refer to the [attribute defaults](#attribute-defaults) for more information.*

The attribute to retrieve which contains the users email addresses. This is important for the device registration and
password reset processes. The user must have an email address in order for Authelia to perform identity verification
when a user attempts to reset their password or register a second factor device.

### display_name_attribute

{{< confkey type="string" required="situational" >}}

*__Note:__ This option is technically required however the [implementation](#implementation) option can implicitly set a
default negating this requirement. Refer to the [attribute defaults](#attribute-defaults) for more information.*

The attribute to retrieve which is shown on the Web UI to the user when they log in.

### additional_groups_dn

{{< confkey type="string" required="no" >}}

Similar to [additional_users_dn](#additional_users_dn) but it applies to group searches.

### groups_filter

{{< confkey type="string" required="situational" >}}

*__Note:__ This option is technically required however the [implementation](#implementation) option can implicitly set a
default negating this requirement. Refer to the [filter defaults](#filter-defaults) for more information.*

Similar to [users_filter](#users_filter) but it applies to group searches. In order to include groups the member is not
a direct member of, but is a member of another group that is a member of those (i.e. recursive groups), you may try
using the following filter which is currently only tested against Microsoft Active Directory:

`(&(member:1.2.840.113556.1.4.1941:={dn})(objectClass=group)(objectCategory=group))`

## group_name_attribute

{{< confkey type="string" required="situational" >}}

*__Note:__ This option is technically required however the [implementation](#implementation) option can implicitly set a
default negating this requirement. Refer to the [attribute defaults](#attribute-defaults) for more
information.*

The LDAP attribute that is used by Authelia to determine the group name.

## permit_referrals

{{< confkey type="boolean" default="false" required="no" >}}

Permits following referrals. This is useful if you have read-only servers in your architecture and thus require
referrals to be followed when performing write operations.

### user

{{< confkey type="string" required="yes" >}}

The distinguished name of the user paired with the password to bind with for lookup and password change operations.

### password

{{< confkey type="string" required="yes" >}}

The password of the user paired with the user to bind with for lookup and password change operations.
Can also be defined using a [secret](../methods/secrets.md) which is the recommended for containerized deployments.

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

## Refresh Interval

It's recommended you either use the default [refresh interval](./introduction.md#refresh_interval) or configure this to
a value low enough to refresh the user groups and status (deleted, disabled, etc) to adequately secure your environment.

## Important notes

Users must be uniquely identified by an attribute, this attribute must obviously contain a single value and be guaranteed
by the administrator to be unique. If multiple users have the same value, Authelia will simply fail authenticating the
user and display an error message in the logs.

In order to avoid such problems, we highly recommended you follow [RFC2307] by using `sAMAccountName` for Active
Directory and `uid` for other implementations as the attribute holding the unique identifier
for your users.

[username attribute]: #username_attribute
[TechNet wiki]: https://social.technet.microsoft.com/wiki/contents/articles/5392.active-directory-ldap-syntax-filters.aspx
[RFC2307]: https://www.rfc-editor.org/rfc/rfc2307.html
