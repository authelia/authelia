---
layout: default
title: LDAP
parent: Authentication backends
grand_parent: Configuration
nav_order: 2
---

# LDAP

**Authelia** supports using a LDAP server as the users database.

## Configuration

```yaml
authentication_backend:
  disable_reset_password: false
  refresh_interval: 5m
  ldap:
    implementation: custom
    url: ldap://127.0.0.1
    start_tls: false
    tls:
      server_name: ldap.example.com
      skip_verify: false
      minimum_version: TLS1.2
    base_dn: dc=example,dc=com
    username_attribute: uid
    additional_users_dn: ou=users
    users_filter: (&({username_attribute}={input})(objectClass=person))
    additional_groups_dn: ou=groups
    groups_filter: (&(member={dn})(objectclass=groupOfNames))
    group_name_attribute: cn
    mail_attribute: mail
    display_name_attribute: displayname
    user: cn=admin,dc=example,dc=com
    password: password
```

## Options

### implementation

See the [Implementation Guide](#implementation-guide) for information.

The user must have an email address in order for Authelia to perform
identity verification when a user attempts to reset their password or
register a second factor device.

### url

The LDAP url which consists of a scheme, address, and port. Format is `<scheme>://<address>:<port>` or 
`<scheme>://<address>` where scheme is either `ldap` or `ldaps`.

If utilising an IPv6 literal address it must be enclosed by square brackets:
```yaml
url: ldap://[fd00:1111:2222:3333::1]
```

### start_tls

The key `start_tls` enables use of the LDAP StartTLS process which is not commonly used. You should only configure this
if you know you need it. The initial connection will be over plain text, and Authelia will try to upgrade it with the
LDAP server. LDAPS URL's are slightly more secure.

### tls

Controls the TLS connection validation process. You can see how to configure the tls 
section [here](../index.md#tls-configuration).

### base_dn

Sets the base distinguished name container for all LDAP queries. If your LDAP domain is example.com this is usually 
`dc=example,dc=com`, however you can fine tune this to be more specific for example to only include objects inside the
authelia OU: `ou=authelia,dc=example,dc=com`. This is prefixed with the [additional_users_dn](#additional_users_dn) for
user searches and [additional_groups_dn](#additional_groups_dn) for groups searches.

### username_attribute

The LDAP attribute that maps to the username in Authelia.

### additional_users_dn

Additional LDAP path to append to the [base_dn](#base_dn) when searching for users. Useful if you want to restrict 
exactly which OU to get users from for either security or performance reasons. For example setting it to 
`ou=users,ou=people` with a base_dn set to `dc=example,dc=com` will mean user searches will occur in 
`ou=users,ou=people,dc=example,dc=com`.

### users_filter

The LDAP filter to narrow down which users are valid. This is important to set correctly as to exclude disabled users.

### additional_groups_dn

Similar to [additional_users_dn](#additional_users_dn) but it applies to group searches.

### groups_filter

Similar to [users_filter](#users_filter) but it applies to group searches.

### mail_attribute

The attribute to retrieve which contains the users email addresses. This is important for the device registration and
password reset processes.

### display_name_attribute

The attribute to retrieve which is shown on the Web UI to the user when they log in.

### user

The distinguished name of the user paired with the password to bind with for lookup and password change operations.

### password

The password of the user paired with the user to bind with for lookup and password change operations.
Can also be defined using a [secret](../secrets.md) which is the recommended for containerized deployments. 

## Implementation Guide

There are currently two implementations, `custom` and `activedirectory`. The `activedirectory` implementation
must be used if you wish to allow users to change or reset their password as Active Directory
uses a custom attribute for this, and an input format other implementations do not use. The long term 
intention of this is to have logical defaults for various RFC implementations of LDAP. 

### Defaults

The below tables describes the current attribute defaults for each implementation.

#### Attributes
This table describes the attribute defaults for each implementation. i.e. the username_attribute is
described by the Username column.

|Implementation |Username      |Display Name|Mail|Group Name|
|:-------------:|:------------:|:----------:|:--:|:--------:|
|custom         |n/a           |displayname |mail|cn        |
|activedirectory|sAMAccountName|displayname |mail|cn        |

#### Filters

The filters are probably the most important part to get correct when setting up LDAP. 
You want to exclude disabled accounts. The active directory example has two attribute 
filters that accomplish this as an example (more examples would be appreciated). The 
userAccountControl filter checks that the account is not disabled and the pwdLastSet 
makes sure that value is not 0 which means the password requires changing at the next login.

|Implementation |Users Filter  |Groups Filter|
|:-------------:|:------------:|:-----------:|
|custom         |n/a           |n/a       |
|activedirectory|(&(&#124;({username_attribute}={input})({mail_attribute}={input}))(objectCategory=person)(objectClass=user)(!userAccountControl:1.2.840.113556.1.4.803:=2)(!pwdLastSet=0))|(&(member={dn})(objectClass=group)(objectCategory=group))|


## Refresh Interval

This setting takes a [duration notation](../index.md#duration-notation-format) that sets the max frequency
for how often Authelia contacts the backend to verify the user still exists and that the groups stored 
in the session are up to date. This allows us to destroy sessions when the user no longer matches the
user_filter, or deny access to resources as they are removed from groups.

In addition to the duration notation, you may provide the value `always` or `disable`. Setting to `always`
is the same as setting it to 0 which will refresh on every request, `disable` turns the feature off, which is 
not recommended. This completely prevents Authelia from refreshing this information, and it would only be
refreshed when the user session gets destroyed by other means like inactivity, session expiration or logging 
out and in.

This value can be any value including 0, setting it to 0 would automatically refresh the session on
every single request. This means Authelia will have to contact the LDAP backend every time an element
on a page loads which could be substantially costly. It's a trade-off between load and security that 
you should adapt according to your own security policy.

## Important notes

Users must be uniquely identified by an attribute, this attribute must obviously contain a single value and
be guaranteed by the administrator to be unique. If multiple users have the same value, Authelia will simply
fail authenticating the user and display an error message in the logs.

In order to avoid such problems, we highly recommended you follow https://www.ietf.org/rfc/rfc2307.txt by using
`sAMAccountName` for Active Directory and `uid` for other implementations as the attribute holding the
unique identifier for your users.

As of versions > `4.24.0` the `users_filter` must include the `username_attribute` placeholder, not including this will
result in Authelia throwing an error.
In versions <= `4.24.0` not including the `username_attribute` placeholder will cause issues with the session refresh
and will result in session resets when the refresh interval has expired, default of 5 minutes.