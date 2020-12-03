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

Configuration of the LDAP backend is done as follows

```yaml
# The authentication backend to use for verifying user passwords
# and retrieve information such as email address and groups
# users belong to.
#
# There are two supported backends: 'ldap' and 'file'.
authentication_backend:
  # Disable both the HTML element and the API for reset password functionality
  disable_reset_password: false

  # The amount of time to wait before we refresh data from the authentication backend. Uses duration notation.
  # To disable this feature set it to 'disable', this will slightly reduce security because for Authelia, users
  # will always belong to groups they belonged to at the time of login even if they have been removed from them in LDAP.
  # To force update on every request you can set this to '0' or 'always', this will increase processor demand.
  # See the below documentation for more information.
  # Duration Notation docs:  https://docs.authelia.com/configuration/index.html#duration-notation-format
  # Refresh Interval docs: https://docs.authelia.com/configuration/authentication/ldap.html#refresh-interval
  refresh_interval: 5m

  # LDAP backend configuration.
  #
  # This backend allows Authelia to be scaled to more
  # than one instance and therefore is recommended for
  # production.
  ldap:
    # The LDAP implementation, this affects elements like the attribute utilised for resetting a password.
    # Acceptable options are as follows:
    # - 'activedirectory' - For Microsoft Active Directory.
    # - 'custom' - For custom specifications of attributes and filters.
    # This currently defaults to 'custom' to maintain existing behaviour.
    #
    # Depending on the option here certain other values in this section have a default value, notably all
    # of the attribute mappings have a default value that this config overrides, you can read more
    # about these default values at https://docs.authelia.com/configuration/authentication/ldap.html#defaults
    implementation: custom

    # The url to the ldap server. Scheme can be ldap or ldaps in the format (port optional) <scheme>://<address>[:<port>].
    url: ldap://127.0.0.1
    
    # Skip verifying the server certificate (to allow a self-signed certificate).
    skip_verify: false

    # Use StartTLS with the LDAP connection.
    start_tls: false

    # Minimum TLS version for either Secure LDAP or LDAP StartTLS.
    minimum_tls_version: TLS1.2

    # The base dn for every entries.
    base_dn: dc=example,dc=com
    
    # The attribute holding the username of the user. This attribute is used to populate
    # the username in the session information. It was introduced due to #561 to handle case
    # insensitive search queries.
    # For you information, Microsoft Active Directory usually uses 'sAMAccountName' and OpenLDAP
    # usually uses 'uid'
    # Beware that this attribute holds the unique identifiers for the users binding the user and the configuration
    # stored in database. Therefore only single value attributes are allowed and the value
    # must never be changed once attributed to a user otherwise it would break the configuration
    # for that user. Technically, non-unique attributes like 'mail' can also be used but we don't recommend using
    # them, we instead advise to use the attributes mentioned above (sAMAccountName and uid) to follow
    # https://www.ietf.org/rfc/rfc2307.txt.
    # username_attribute: uid
    
    # An additional dn to define the scope to all users.
    additional_users_dn: ou=users

    # The users filter used in search queries to find the user profile based on input filled in login form.
    # Various placeholders are available to represent the user input and back reference other options of the configuration:
    # - {input} is a placeholder replaced by what the user inputs in the login form. 
    # - {username_attribute} is a mandatory placeholder replaced by what is configured in `username_attribute`.
    # - {mail_attribute} is a placeholder replaced by what is configured in `mail_attribute`.
    # - DON'T USE - {0} is an alias for {input} supported for backward compatibility but it will be deprecated in later versions, so please don't use it.
    #
    # Recommended settings are as follows:
    # - Microsoft Active Directory: (&({username_attribute}={input})(objectCategory=person)(objectClass=user))
    # - OpenLDAP: (&({username_attribute}={input})(objectClass=person))' or '(&({username_attribute}={input})(objectClass=inetOrgPerson))
    #
    # To allow sign in both with username and email, one can use a filter like
    # (&(|({username_attribute}={input})({mail_attribute}={input}))(objectClass=person))
    users_filter: (&({username_attribute}={input})(objectClass=person))

    # An additional dn to define the scope of groups.
    additional_groups_dn: ou=groups
    
    # The groups filter used in search queries to find the groups of the user.
    # - {input} is a placeholder replaced by what the user inputs in the login form.
    # - {username} is a placeholder replace by the username stored in LDAP (based on `username_attribute`).
    # - {dn} is a matcher replaced by the user distinguished name, aka, user DN.
    # - {username_attribute} is a placeholder replaced by what is configured in `username_attribute`.
    # - {mail_attribute} is a placeholder replaced by what is configured in `mail_attribute`.
    # - DON'T USE - {0} is an alias for {input} supported for backward compatibility but it will be deprecated in later versions, so please don't use it.
    # - DON'T USE - {1} is an alias for {username} supported for backward compatibility but it will be deprecated in later version, so please don't use it.
    groups_filter: (&(member={dn})(objectclass=groupOfNames))

    # The attribute holding the name of the group
    # group_name_attribute: cn

    # The attribute holding the mail address of the user. If multiple email addresses are defined for a user, only the first
    # one returned by the LDAP server is used.
    # mail_attribute: mail

    # The attribute holding the display name of the user. This will be used to greet an authenticated user.
    # display_name_attribute: displayname

    # The username and password of the admin user.
    user: cn=admin,dc=example,dc=com
    # Password can also be set using a secret: https://docs.authelia.com/configuration/secrets.html
    password: password
```

The user must have an email address in order for Authelia to perform
identity verification when a user attempts to reset their password or
register a second factor device.

## TLS Settings

### Skip Verify

The key `skip_verify` disables checking the authenticity of the TLS certificate. You should not disable this, instead
you should add the certificate that signed the certificate of your LDAP server to the machines certificate PKI trust.
For docker you can just add this to the hosts trusted store.

### Start TLS

The key `start_tls` enables use of the LDAP StartTLS process which is not commonly used. You should only configure this
if you know you need it. The initial connection will be over plain text, and Authelia will try to upgrade it with the
LDAP server. LDAPS URL's are slightly more secure.

### Minimum TLS Version

The key `minimum_tls_version` controls the minimum TLS version Authelia will use when opening LDAP connections.
The possible values are `TLS1.3`, `TLS1.2`, `TLS1.1`, `TLS1.0`. Anything other than `TLS1.3` or `TLS1.2`
are very old and deprecated. You should avoid using these and upgrade your LDAP solution instead of decreasing
this value. 

## Implementation

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

## Loading a password from a secret instead of inside the configuration

Password can also be defined using a [secret](../secrets.md).
