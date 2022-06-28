---
title: "LDAP"
description: "Configuring LDAP"
lead: "Authelia supports an LDAP server based first factor user provider. This section describes configuring this."
date: 2022-06-15T17:51:47+10:00
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
    permit_unauthenticated_bind: false
    user: CN=admin,DC=example,DC=com
    password: password
```

## Options

### implementation

{{< confkey type="string" default="custom" required="no" >}}

Configures the LDAP implementation used by Authelia.

See the [Implementation Guide](../../reference/guides/ldap.md#implementation-guide) for information.

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

### group_name_attribute

{{< confkey type="string" required="situational" >}}

*__Note:__ This option is technically required however the [implementation](#implementation) option can implicitly set a
default negating this requirement. Refer to the [attribute defaults](#attribute-defaults) for more
information.*

The LDAP attribute that is used by Authelia to determine the group name.

### permit_referrals

{{< confkey type="boolean" default="false" required="no" >}}

Permits following referrals. This is useful if you have read-only servers in your architecture and thus require
referrals to be followed when performing write operations.

### permit_unauthenticated_bind

{{< confkey type="boolean" default="false" required="no" >}}

*__WARNING:__ This option is strongly discouraged. Please consider disabling unauthenticated binding to your LDAP
server and utilizing a service account.*

Permits binding to the server without a password. For this option to be enabled both the [password](#password)
configuration option must be blank and the [password_reset disable](introduction.md#disable) option must be `true`.

### user

{{< confkey type="string" required="yes" >}}

The distinguished name of the user paired with the password to bind with for lookup and password change operations.

### password

{{< confkey type="string" required="yes" >}}

*__Important Note:__ This can also be defined using a [secret](../methods/secrets.md) which is __strongly recommended__
especially for containerized deployments.*

The password paired with the [user](#user) used to bind to the LDAP server for lookup and password change operations.

It's __strongly recommended__ this is a
[Random Alphanumeric String](../miscellaneous/guides.md#generating-a-random-alphanumeric-string) with 64 or more
characters and the user password is changed to this value.

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

## See Also

- [LDAP Reference Guide](../../reference/guides/ldap.md)

[username attribute]: #username_attribute
[TechNet wiki]: https://social.technet.microsoft.com/wiki/contents/articles/5392.active-directory-ldap-syntax-filters.aspx
[RFC2307]: https://www.rfc-editor.org/rfc/rfc2307.html
