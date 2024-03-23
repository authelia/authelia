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
      server_name: 'ldap.example.com'
      skip_verify: false
      minimum_version: 'TLS1.2'
      maximum_version: 'TLS1.3'
      certificate_chain: |
        -----BEGIN CERTIFICATE-----
        MIIC5jCCAc6gAwIBAgIRAK4Sj7FiN6PXo/urPfO4E7owDQYJKoZIhvcNAQELBQAw
        EzERMA8GA1UEChMIQXV0aGVsaWEwHhcNNzAwMTAxMDAwMDAwWhcNNzEwMTAxMDAw
        MDAwWjATMREwDwYDVQQKEwhBdXRoZWxpYTCCASIwDQYJKoZIhvcNAQEBBQADggEP
        ADCCAQoCggEBAPKv3pSyP4ozGEiVLJ14dIWFCEGEgq7WUMI0SZZqQA2ID0L59U/Q
        /Usyy7uC9gfMUzODTpANtkOjFQcQAsxlR1FOjVBrX5QgjSvXwbQn3DtwMA7XWSl6
        LuYx2rBYSlMSN5UZQm/RxMtXfLK2b51WgEEYDFi+nECSqKzR4R54eOPkBEWRfvuY
        91AMjlhpivg8e4JWkq4LVQUKbmiFYwIdK8XQiN4blY9WwXwJFYs5sQ/UYMwBFi0H
        kWOh7GEjfxgoUOPauIueZSMSlQp7zqAH39N0ZSYb6cS0Npj57QoWZSY3ak87ebcR
        Nf4rCvZLby7LoN7qYCKxmCaDD3x2+NYpWH8CAwEAAaM1MDMwDgYDVR0PAQH/BAQD
        AgWgMBMGA1UdJQQMMAoGCCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwDQYJKoZIhvcN
        AQELBQADggEBAHSITqIQSNzonFl3DzxHPEzr2hp6peo45buAAtu8FZHoA+U7Icfh
        /ZXjPg7Xz+hgFwM/DTNGXkMWacQA/PaNWvZspgRJf2AXvNbMSs2UQODr7Tbv+Fb4
        lyblmMUNYFMCFVAMU0eIxXAFq2qcwv8UMcQFT0Z/35s6PVOakYnAGGQjTfp5Ljuq
        wsdc/xWmM0cHWube6sdRRUD7SY20KU/kWzl8iFO0VbSSrDf1AlEhnLEkp1SPaxXg
        OdBnl98MeoramNiJ7NT6Jnyb3zZ578fjaWfThiBpagItI8GZmG4s4Ovh2JbheN8i
        ZsjNr9jqHTjhyLVbDRlmJzcqoj4JhbKs6/I^invalid DO NOT USE=
        -----END CERTIFICATE-----
        -----BEGIN CERTIFICATE-----
        MIIDBDCCAeygAwIBAgIRALJsPg21kA0zY4F1wUCIuoMwDQYJKoZIhvcNAQELBQAw
        EzERMA8GA1UEChMIQXV0aGVsaWEwHhcNNzAwMTAxMDAwMDAwWhcNNzEwMTAxMDAw
        MDAwWjATMREwDwYDVQQKEwhBdXRoZWxpYTCCASIwDQYJKoZIhvcNAQEBBQADggEP
        ADCCAQoCggEBAMXHBvVxUzYk0u34/DINMSF+uiOekKOAjOrC6Mi9Ww8ytPVO7t2S
        zfTvM+XnEJqkFQFgimERfG/eGhjF9XIEY6LtnXe8ATvOK4nTwdufzBaoeQu3Gd50
        5VXr6OHRo//ErrGvFXwP3g8xLePABsi/fkH3oDN+ztewOBMDzpd+KgTrk8ysv2ou
        kNRMKFZZqASvCgv0LD5KWvUCnL6wgf1oTXG7aztduA4oSkUP321GpOmBC5+5ElU7
        ysoRzvD12o9QJ/IfEaulIX06w9yVMo60C/h6A3U6GdkT1SiyTIqR7v7KU/IWd/Qi
        Lfftcj91VhCmJ73Meff2e2S2PrpjdXbG5FMCAwEAAaNTMFEwDgYDVR0PAQH/BAQD
        AgKkMA8GA1UdJQQIMAYGBFUdJQAwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQU
        Z7AtA3mzFc0InSBA5fiMfeLXA3owDQYJKoZIhvcNAQELBQADggEBAEE5hm1mtlk/
        kviCoHH4evbpw7rxPxDftIQlqYTtvMM4eWY/6icFoSZ4fUHEWYyps8SsPu/8f2tf
        71LGgZn0FdHi1QU2H8m0HHK7TFw+5Q6RLrLdSyk0PItJ71s9en7r8pX820nAFEHZ
        HkOSfJZ7B5hFgUDkMtVM6bardXAhoqcMk4YCU96e9d4PB4eI+xGc+mNuYvov3RbB
        D0s8ICyojeyPVLerz4wHjZu68Z5frAzhZ68YbzNs8j2fIBKKHkHyLG1iQyF+LJVj
        2PjCP+auJsj6fQQpMGoyGtpLcSDh+ptcTngUD8JsWipzTCjmaNqdPHAOYmcgtf4b
        qocikt3WAdU^invalid DO NOT USE=
        -----END CERTIFICATE-----
      private_key: |
        -----BEGIN RSA PRIVATE KEY-----
        MIIEpAIBAAKCAQEA8q/elLI/ijMYSJUsnXh0hYUIQYSCrtZQwjRJlmpADYgPQvn1
        T9D9SzLLu4L2B8xTM4NOkA22Q6MVBxACzGVHUU6NUGtflCCNK9fBtCfcO3AwDtdZ
        KXou5jHasFhKUxI3lRlCb9HEy1d8srZvnVaAQRgMWL6cQJKorNHhHnh44+QERZF+
        +5j3UAyOWGmK+Dx7glaSrgtVBQpuaIVjAh0rxdCI3huVj1bBfAkVizmxD9RgzAEW
        LQeRY6HsYSN/GChQ49q4i55lIxKVCnvOoAff03RlJhvpxLQ2mPntChZlJjdqTzt5
        txE1/isK9ktvLsug3upgIrGYJoMPfHb41ilYfwIDAQABAoIBAQDTOdFf2JjHH1um
        aPgRAvNf9v7Nj5jytaRKs5nM6iNf46ls4QPreXnMhqSeSwj6lpNgBYxOgzC9Q+cc
        Y4ob/paJJPaIJTxmP8K/gyWcOQlNToL1l+eJ20eQoZm23NGr5fIsunSBwLEpTrdB
        ENqqtcwhW937K8Pxy/Q1nuLyU2bc6Tn/ivLozc8n27dpQWWKh8537VY7ancIaACr
        LJJLYxKqhQpjtBWAyCDvZQirnAOm9KnvIHaGXIswCZ4Xbsu0Y9NL+woARPyRVQvG
        jfxy4EmO9s1s6y7OObSukwKDSNihAKHx/VIbvVWx8g2Lv5fGOa+J2Y7o9Qurs8t5
        BQwMTt0BAoGBAPUw5Z32EszNepAeV3E2mPFUc5CLiqAxagZJuNDO2pKtyN29ETTR
        Ma4O1cWtGb6RqcNNN/Iukfkdk27Q5nC9VJSUUPYelOLc1WYOoUf6oKRzE72dkMQV
        R4bf6TkjD+OVR17fAfkswkGahZ5XA7j48KIQ+YC4jbnYKSxZTYyKPjH/AoGBAP1i
        tqXt36OVlP+y84wWqZSjMelBIVa9phDVGJmmhz3i1cMni8eLpJzWecA3pfnG6Tm9
        ze5M4whASleEt+M00gEvNaU9ND+z0wBfi+/DwJYIbv8PQdGrBiZFrPhTPjGQUldR
        lXccV2meeLZv7TagVxSi3DO6dSJfSEHyemd5j9mBAoGAX8Hv+0gOQZQCSOTAq8Nx
        6dZcp9gHlNaXnMsP9eTDckOSzh636JPGvj6m+GPJSSbkURUIQ3oyokMNwFqvlNos
        fTaLhAOfjBZI9WnDTTQxpugWjphJ4HqbC67JC/qIiw5S6FdaEvGLEEoD4zoChywZ
        9oGAn+fz2d/0/JAH/FpFPgsCgYEAp/ipZgPzziiZ9ov1wbdAQcWRj7RaWnssPFpX
        jXwEiXT3CgEMO4MJ4+KWIWOChrti3qFBg6i6lDyyS6Qyls7sLFbUdC7HlTcrOEMe
        rBoTcCI1GqZNlqWOVQ65ZIEiaI7o1vPBZo2GMQEZuq8mDKFsOMThvvTrM5cAep84
        n6HJR4ECgYABWcbsSnr0MKvVth/inxjbKapbZnp2HUCuw87Ie5zK2Of/tbC20wwk
        yKw3vrGoE3O1t1g2m2tn8UGGASeZ842jZWjIODdSi5+icysQGuULKt86h/woz2SQ
        27GoE2i5mh6Yez6VAYbUuns3FcwIsMyWLq043Tu2DNkx9ijOOAuQzw^invalid..
        DO NOT USE==
        -----END RSA PRIVATE KEY-----
    base_dn: 'DC=example,DC=com'
    additional_users_dn: 'OU=users'
    users_filter: '(&({username_attribute}={input})(objectClass=person))'
    additional_groups_dn: 'OU=groups'
    groups_filter: '(&(member={dn})(objectClass=groupOfNames))'
    group_search_mode: 'filter'
    permit_referrals: false
    permit_unauthenticated_bind: false
    user: 'CN=admin,DC=example,DC=com'
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

```yaml
authentication_backend:
  ldap:
    address: 'ldaps://dc1.example.com'
```

```yaml
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
