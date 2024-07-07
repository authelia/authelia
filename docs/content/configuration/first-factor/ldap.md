---
title: "LDAP"
description: "Configuring LDAP"
summary: "Authelia supports an LDAP server based first factor user provider. This section describes configuring this."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
weight: 102200
toc: true
aliases:
  - /c/ldap
  - /docs/configuration/authentication/ldap.html
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
authentication_backend:
  ldap:
    address: 'ldap://127.0.0.1'
    implementation: 'custom'
    timeout: '5s'
    start_tls: false
    tls:
      server_name: 'ldap.{{< sitevar name="domain" nojs="example.com" >}}'
      skip_verify: false
      minimum_version: 'TLS1.2'
      maximum_version: 'TLS1.3'
      certificate_chain: |
        -----BEGIN CERTIFICATE-----
        ...
        -----END CERTIFICATE-----
        -----BEGIN CERTIFICATE-----
        ...
        -----END CERTIFICATE-----
      private_key: |
        -----BEGIN RSA PRIVATE KEY-----
        ...
        -----END RSA PRIVATE KEY-----
    base_dn: '{{< sitevar name="domain" format="dn" nojs="DC=example,DC=com" >}}'
    additional_users_dn: 'OU=users'
    users_filter: '(&({username_attribute}={input})(objectClass=person))'
    additional_groups_dn: 'OU=groups'
    groups_filter: '(&(member={dn})(objectClass=groupOfNames))'
    group_search_mode: 'filter'
    permit_referrals: false
    permit_unauthenticated_bind: false
    user: 'CN=admin,{{< sitevar name="domain" format="dn" nojs="DC=example,DC=com" >}}'
    password: 'password'
    attributes:
      distinguished_name: 'distinguishedName'
      username: 'uid'
      display_name: 'displayName'
      mail: 'mail'
      member_of: 'memberOf'
      group_name: 'cn'
```

## Options

This section describes the individual configuration options.

### address

{{< confkey type="string" syntax="address" required="yes" >}}

The LDAP URL which consists of a scheme, hostname, and port. Format is `[<scheme>://]<hostname>[:<port>]`. The default
scheme is `ldapi` if the path is absolute otherwise it's `ldaps`, and the permitted schemes are `ldap`, `ldaps`, or
`ldapi` (a unix domain socket).

If the scheme is `ldapi` it must be followed by an absolute path to an existing unix domain socket that the
user/group the Authelia process is running as has the appropriate permissions to access. For example if the socket is
located at `/var/run/slapd.sock` the address should be `ldapi:///var/run/slapd.sock`.

__Examples:__

```yaml {title="configuration.yml"}
authentication_backend:
  ldap:
    address: 'ldaps://dc1.{{< sitevar name="domain" nojs="example.com" >}}'
```

```yaml {title="configuration.yml"}
authentication_backend:
  ldap:
    address: 'ldap://[fd00:1111:2222:3333::1]'
```

### implementation

{{< confkey type="string" default="custom" required="no" >}}

Configures the LDAP implementation used by Authelia.

See the [Implementation Guide](../../reference/guides/ldap.md#implementation-guide) for information.

### timeout

{{< confkey type="string,integer" syntax="duration" default="5 seconds" required="no" >}}

The timeout for dialing an LDAP connection.

### start_tls

{{< confkey type="boolean" default="false" required="no" >}}

Enables use of the LDAP StartTLS process which is not commonly used. You should only configure this if you know you need
it. The initial connection will be over plain text, and *Authelia* will try to upgrade it with the LDAP server. LDAPS
URL's are slightly more secure.

### tls

{{< confkey type="structure" structure="tls" required="no" >}}

Controls the TLS connection validation parameters for either StartTLS or the TLS socket.

### base_dn

{{< confkey type="string" required="yes" >}}

Sets the base distinguished name container for all LDAP queries. If your LDAP domain is `{{< sitevar name="domain" nojs="example.com" >}}`
this is usually `{{< sitevar name="domain" format="dn" nojs="DC=example,DC=com" >}}`, however you can fine tune this to be more specific for
example to only include objects inside the authelia OU: `OU=authelia,{{< sitevar name="domain" format="dn" nojs="DC=example,DC=com" >}}`. This
is prefixed with the [additional_users_dn](#additional_users_dn) for user searches and [additional_groups_dn](#additional_groups_dn) for groups searches.

### additional_users_dn

{{< confkey type="string" required="no" >}}

Additional LDAP path to append to the [base_dn](#base_dn) when searching for users. Useful if you want to restrict
exactly which OU to get users from for either security or performance reasons. For example setting it to
`OU=users,OU=people` with a base_dn set to `{{< sitevar name="domain" format="dn" nojs="DC=example,DC=com" >}}` will mean user searches will
occur in `OU=users,OU=people,{{< sitevar name="domain" format="dn" nojs="DC=example,DC=com" >}}`.

### users_filter

{{< confkey type="string" required="situational" >}}

*__Note:__ This option is technically required however the [implementation](#implementation) option can implicitly set a
default negating this requirement. Refer to the [filter defaults](../../reference/guides/ldap.md#filter-defaults) for
more information.*

The LDAP filter to narrow down which users are valid. This is important to set correctly as to exclude disabled users.
The default value is dependent on the [implementation](#implementation), refer to the
[attribute defaults](../../reference/guides/ldap.md#attribute-defaults) for more information.

### additional_groups_dn

{{< confkey type="string" required="no" >}}

Similar to [additional_users_dn](#additional_users_dn) but it applies to group searches.

### groups_filter

{{< confkey type="string" required="situational" >}}

*__Note:__ This option is technically required however the [implementation](#implementation) option can implicitly set a
default negating this requirement. Refer to the [filter defaults](../../reference/guides/ldap.md#filter-defaults) for
more information.*

Similar to [users_filter](#users_filter) but it applies to group searches. In order to include groups the member is not
a direct member of, but is a member of another group that is a member of those (i.e. recursive groups), you may try
using the following filter which is currently only tested against Microsoft Active Directory:

`(&(member:1.2.840.113556.1.4.1941:={dn})(objectClass=group)(objectCategory=group))`

### group_search_mode

{{< confkey type="string" default="filter" required="no" >}}

The group search mode controls how user groups are discovered. The default of `filter` directly uses the filter to
determine the result. The `memberof` experimental mode does another special filtered search. See the
[Reference Documentation](../../reference/guides/ldap.md#group-search-modes) for more information.

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

### permit_feature_detection_failure

{{< confkey type="boolean" default="false" required="no" >}}

Authelia searches for the RootDSE to discover supported controls and extensions. This option is a compatibility option
which *__should not__* be enabled unless the LDAP server returns an error when searching for the RootDSE.

### user

{{< confkey type="string" required="yes" >}}

The distinguished name of the user paired with the password to bind with for lookup and password change operations.

### password

{{< confkey type="string" required="yes" >}}

*__Important Note:__ This can also be defined using a [secret](../methods/secrets.md) which is __strongly recommended__
especially for containerized deployments.*

The password paired with the [user](#user) used to bind to the LDAP server for lookup and password change operations.

It's __strongly recommended__ this is a
[Random Alphanumeric String](../../reference/guides/generating-secure-values.md#generating-a-random-alphanumeric-string) with 64 or more
characters and the user password is changed to this value.

### attributes

The following options configure The directory server attribute mappings.

#### distinguished_name

{{< confkey type="string" required="situational" >}}

*__Note:__ This option is technically not required however it is required when using the group search mode
`memberof` replacement `{memberof:dn}`.*

The directory server attribute which contains the distinguished name, primarily used to perform filtered searches. There
is a clear distinction between the actual distinguished name and a distinguished name attribute, all directories have
distinguished names for objects, but not all have an attribute representing this that can be searched on.

The only known support at this time is with Active Directory.

#### username

{{< confkey type="string" required="situational" >}}

*__Note:__ This option is technically required however the [implementation](#implementation) option can implicitly set a
default negating this requirement. Refer to the [attribute defaults] for more information.*

The directory server attribute that maps to the username in *Authelia*. This must contain the `{username_attribute}` [placeholder].

#### display_name

{{< confkey type="string" required="situational" >}}

*__Note:__ This option is technically required however the [implementation](#implementation) option can implicitly set a
default negating this requirement. Refer to the [attribute defaults] for more information.*

The directory server attribute to retrieve which is shown on the Web UI to the user when they log in.

#### mail

{{< confkey type="string" required="situational" >}}

*__Note:__ This option is technically required however the [implementation](#implementation) option can implicitly set a
default negating this requirement. Refer to the [attribute defaults] for more information.*

The directory server attribute to retrieve which contains the users email addresses. This is important for the device
registration and password reset processes. The user must have an email address in order for Authelia to perform
identity verification when a user attempts to reset their password or register a second factor device.

#### member_of

{{< confkey type="string" required="situational" >}}

*__Note:__ This option is technically required however the [implementation](#implementation) option can implicitly set a
default negating this requirement. Refer to the [attribute defaults] for more information.*

The directory server attribute which contains the groups a user is a member of. This is currently only used for the
`memberof` group search mode.

#### group_name

{{< confkey type="string" required="situational" >}}

*__Note:__ This option is technically required however the [implementation](#implementation) option can implicitly set a
default negating this requirement. Refer to the [attribute defaults] for more information.*

The directory server attribute that is used by Authelia to determine the group name.

## Refresh Interval

It's recommended you either use the default [refresh interval](introduction.md#refresh_interval) or configure this to
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

[username attribute]: #username
[TechNet wiki]: https://social.technet.microsoft.com/wiki/contents/articles/5392.active-directory-ldap-syntax-filters.aspx
[RFC2307]: https://datatracker.ietf.org/doc/html/rfc2307
[attribute defaults]: ../../reference/guides/ldap.md#attribute-defaults
[placeholder]: ../../reference/guides/ldap.md#users-filter-replacements
