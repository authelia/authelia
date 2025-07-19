---
title: "LDAP"
description: "Configuring LDAP"
summary: "Authelia supports an LDAP server based first factor user provider. This section describes configuring this."
date: 2024-03-14T06:00:14+11:00
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

## Variables

Some of the values within this page can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

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
        -----BEGIN PRIVATE KEY-----
        ...
        -----END PRIVATE KEY-----
    pooling:
      enable: false
      count: 5
      retries: 2
      timeout: '10 seconds'
    base_dn: '{{< sitevar name="domain" format="dn" nojs="DC=example,DC=com" >}}'
    additional_users_dn: 'OU=users'
    users_filter: '(&({username_attribute}={input})(objectClass=person))'
    additional_groups_dn: 'OU=groups'
    groups_filter: '(&(member={dn})(objectClass=groupOfNames))'
    group_search_mode: 'filter'
    permit_referrals: false
    permit_unauthenticated_bind: false
    permit_feature_detection_failure: false
    user: 'CN=admin,{{< sitevar name="domain" format="dn" nojs="DC=example,DC=com" >}}'
    password: 'password'
    attributes:
      distinguished_name: 'distinguishedName'
      username: 'uid'
      display_name: 'displayName'
      family_name: 'sn'
      given_name: 'givenName'
      middle_name: 'middleName'
      nickname: ''
      gender: ''
      birthdate: ''
      website: 'wWWHomePage'
      profile: ''
      picture: ''
      zoneinfo: ''
      locale: ''
      phone_number: 'telephoneNumber'
      phone_extension: ''
      street_address: 'streetAddress'
      locality: 'l'
      region: 'st'
      postal_code: 'postalCode'
      country: 'c'
      mail: 'mail'
      member_of: 'memberOf'
      group_name: 'cn'
      extra:
        extra_example:
          name: ''
          multi_valued: false
          value_type: 'string'
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

See the [Implementation Guide](../../integration/ldap/introduction.md#implementation-guide) for information.

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

If defined this option controls the TLS connection verification parameters for the LDAP server.

By default Authelia uses the system certificate trust for TLS certificate verification of TLS connections and the
[certificates_directory](../miscellaneous/introduction.md#certificates_directory) global option can be used to augment
this.


### pooling

The connection pooling configuration.

#### enable

{{< confkey type="boolean" default="false" required="no" >}}

Enables the connection pooling functionality.

#### count

{{< confkey type="integer" default="5" required="no" >}}

The number of open connections to be available in the pool at any given time.

#### retries

{{< confkey type="integer" default="2" required="no" >}}

The number of attempts to obtain a free connecting that are made within the timeout period. This effectively splits the
timeout into chunks.

#### timeout

{{< confkey type="string,integer" syntax="duration" default="20 seconds" required="no" >}}

The amount of time that we wait for a connection to become free in the pool before giving up and failing with an error.

### base_dn

{{< confkey type="string" required="situational" >}}

Sets the base distinguished name container for all LDAP queries. If your LDAP domain is
`{{< sitevar name="domain" nojs="example.com" >}}` this is usually
`{{< sitevar name="domain" format="dn" nojs="DC=example,DC=com" >}}`, however you can fine tune this to be more specific
for example to only include objects inside the authelia OU:
`OU=authelia,{{< sitevar name="domain" format="dn" nojs="DC=example,DC=com" >}}`. This is prefixed with the
[additional_users_dn](#additional_users_dn) for user searches and [additional_groups_dn](#additional_groups_dn) for
groups searches.

### additional_users_dn

{{< confkey type="string" required="no" >}}

Additional LDAP path to append to the [base_dn](#base_dn) when searching for users. Useful if you want to restrict
exactly which OU to get users from for either security or performance reasons. For example setting it to
`OU=users,OU=people` with a base_dn set to `{{< sitevar name="domain" format="dn" nojs="DC=example,DC=com" >}}` will mean user searches will
occur in `OU=users,OU=people,{{< sitevar name="domain" format="dn" nojs="DC=example,DC=com" >}}`.

### users_filter

{{< confkey type="string" required="situational" >}}

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
This option is technically required however the [implementation](#implementation) option can implicitly set a
default negating this requirement. Refer to the [filter defaults](../../integration/ldap) of your implementation for
more information.
{{< /callout >}}

The LDAP filter to narrow down which users are valid. This is important to set correctly as to exclude disabled users.
The default value is dependent on the [implementation](#implementation), refer to the
[attribute defaults](../../integration/ldap) for more information.

### additional_groups_dn

{{< confkey type="string" required="no" >}}

Similar to [additional_users_dn](#additional_users_dn) but it applies to group searches.

### groups_filter

{{< confkey type="string" required="situational" >}}

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
This option is technically required however the [implementation](#implementation) option can implicitly set a
default negating this requirement. Refer to the [filter defaults](../../integration/ldap) of your implementation for
more information.
{{< /callout >}}

Similar to [users_filter](#users_filter) but it applies to group searches. In order to include groups the member is not
a direct member of, but is a member of another group that is a member of those (i.e. recursive groups), you may try
using the following filter which is currently only tested against Microsoft Active Directory:

`(&(member:1.2.840.113556.1.4.1941:={dn})(objectClass=group)(objectCategory=group))`

### group_search_mode

{{< confkey type="string" default="filter" required="no" >}}

The group search mode controls how user groups are discovered. The default of `filter` directly uses the filter to
determine the result. The `memberof` experimental mode does another special filtered search. See the
[Integration Documentation](../../integration/ldap/introduction.md#group-search-modes) for more information.

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

{{< confkey type="string" required="yes" secret="yes" secret="yes" >}}

The password paired with the [user](#user) used to bind to the LDAP server for lookup and password change operations.

It's __strongly recommended__ this is a
[Random Alphanumeric String](../../reference/guides/generating-secure-values.md#generating-a-random-alphanumeric-string) with 64 or more
characters and the user password is changed to this value.

### attributes

The following options configure The directory server attribute mappings. It's also recommended to check out the
[Attributes Reference Guide](../../reference/guides/attributes.md) for more information.

#### distinguished_name

{{< confkey type="string" required="situational" >}}

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
This option is technically not required however it is required when using the group search mode
`memberof` replacement `{memberof:dn}`.
{{< /callout >}}

The directory server attribute which contains the distinguished name, primarily used to perform filtered searches. There
is a clear distinction between the actual distinguished name and a distinguished name attribute, all directories have
distinguished names for objects, but not all have an attribute representing this that can be searched on.

The only known support at this time is with Active Directory.

#### username

{{< confkey type="string" required="situational" >}}

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
This option is technically required however the [implementation](#implementation) option can implicitly set a
default negating this requirement. Refer to the [attribute defaults](../../integration/ldap) of your implementation for more information.
{{< /callout >}}

The directory server attribute that maps to the username in *Authelia*. This must contain the `{username_attribute}` [placeholder].

#### display_name

{{< confkey type="string" required="situational" >}}

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
This option is technically required however the [implementation](#implementation) option can implicitly set a
default negating this requirement. Refer to the [attribute defaults](../../integration/ldap) of your implementation for more information.
{{< /callout >}}

The directory server attribute to retrieve which is shown on the Web UI to the user when they log in.

#### family_name

{{< confkey type="string" required="no" >}}

The directory server attribute which contains the users family name.

#### given_name

{{< confkey type="string" required="no" >}}

The directory server attribute which contains the users given name.

#### middle_name

{{< confkey type="string" required="no" >}}

The directory server attribute which contains the users middle name.

#### nickname

{{< confkey type="string" required="no" >}}

The directory server attribute which contains the users nickname.

#### gender

{{< confkey type="string" required="no" >}}

The directory server attribute which contains the users gender.

#### birthdate

{{< confkey type="string" required="no" >}}

The directory server attribute which contains the users birthdate.

#### website

{{< confkey type="string" required="no" >}}

The directory server attribute which contains the users website URL.

#### profile

{{< confkey type="string" required="no" >}}

The directory server attribute which contains the users profile URL.

#### picture

{{< confkey type="string" required="no" >}}

The directory server attribute which contains the users picture URL.

#### zoneinfo

{{< confkey type="string" required="no" >}}

The directory server attribute which contains the users timezone value from the
[IANA Time Zone Database](https://www.iana.org/time-zones).

#### locale

{{< confkey type="string" required="no" >}}

The directory server attribute which contains the users locale in the
[RFC5646 BCP 47](https://www.rfc-editor.org/rfc/rfc5646.html) format.

#### phone_number

{{< confkey type="string" required="no" >}}

The directory server attribute which contains the users phone number.

#### phone_extension

{{< confkey type="string" required="no" >}}

The directory server attribute which contains the users phone extension.

#### street_address

{{< confkey type="string" required="no" >}}

The directory server attribute which contains the users street address.

#### locality

{{< confkey type="string" required="no" >}}

The directory server attribute which contains the users locality i.e. city.

#### region

{{< confkey type="string" required="no" >}}

The directory server attribute which contains the users region i.e. state or province.

#### postal_code

{{< confkey type="string" required="no" >}}

The directory server attribute which contains the users postal code.

#### country

{{< confkey type="string" required="no" >}}

The directory server attribute which contains the users country.

#### mail

{{< confkey type="string" required="situational" >}}

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
This option is technically required however the [implementation](#implementation) option can implicitly set a
default negating this requirement. Refer to the [attribute defaults](../../integration/ldap) of your implementation for more information.
{{< /callout >}}

The directory server attribute to retrieve which contains the users email addresses. This is important for the device
registration and password reset processes. The user must have an email address in order for Authelia to perform
identity verification when a user attempts to reset their password or register a second factor device.

#### member_of

{{< confkey type="string" required="situational" >}}

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
This option is technically required however the [implementation](#implementation) option can implicitly set a
default negating this requirement. Refer to the [attribute defaults](../../integration/ldap) of your implementation for more information.
{{< /callout >}}

The directory server attribute which contains the groups a user is a member of. This is currently only used for the
`memberof` group search mode.

#### group_name

{{< confkey type="string" required="situational" >}}

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
This option is technically required however the [implementation](#implementation) option can implicitly set a
default negating this requirement. Refer to the [attribute defaults](../../integration/ldap) of your implementation for more information.
{{< /callout >}}

The directory server attribute that is used by Authelia to determine the group name.

#### extra

{{< confkey type="dictionary(object)" required="no" >}}

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
In addition to the extra attributes, you can configure custom attributes based on the values of existing attributes.
This is done via the [Definitions](../definitions/user-attributes.md) section.
{{< /callout >}}

The extra attributes to load from the directory server. These extra attributes can be used in other areas of Authelia
such as [OpenID Connect 1.0](../identity-providers/openid-connect/provider.md).

The key represents the backend attribute name, and by default is the name of the attribute within Authelia.

In the example below, we load the directory server attribute `exampleServerAttribute` into the Authelia attribute
`example_authelia_attribute`, treat it as a single valued attribute which has an underlying type of `integer`.

```yaml
authentication_backend:
  ldap:
    attributes:
      extra:
        exampleServerAttribute:
          name: 'example_authelia_attribute'
          multi_valued: false
          value_type: 'integer'
```

#### name

{{< confkey type="string" required="no" >}}

This option changes that attribute name used for internal references within Authelia.

#### value_type

{{< confkey type="string" required="yes" >}}

This defines the underlying type the attribute must be. This is required if an extra attribute is configured. The valid
values are `string`, `integer`, or `boolean`. When using the `integer` and `boolean` types, the directory attributes
must have parsable values.

#### multi_valued

{{< confkey type="boolean" required="no" >}}

This indicates the underlying type can have multiple values.

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

- [LDAP Integration Guide](../../integration/ldap/introduction.md)

[username attribute]: #username
[TechNet wiki]: https://social.technet.microsoft.com/wiki/contents/articles/5392.active-directory-ldap-syntax-filters.aspx
[RFC2307]: https://datatracker.ietf.org/doc/html/rfc2307
[attribute defaults]: ../../integration/ldap
[placeholder]: ../../integration/ldap/introduction.md#users-filter-replacements
