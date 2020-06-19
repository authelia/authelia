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

  ldap:
    # The url to the ldap server. Scheme can be ldap:// or ldaps://
    url: ldap://127.0.0.1

    # Skip verifying the server certificate (to allow self-signed certificate).
    skip_verify: false

    # The base dn for every entries
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
    username_attribute: uid
    
    # An additional dn to define the scope to all users
    additional_users_dn: ou=users
    
    # The users filter used in search queries to find the user profile based on input filled in login form.
    # Various placeholders are available to represent the user input and back reference other options of the configuration:
    # - {input} is a placeholder replaced by what the user inputs in the login form. 
    # - {username_attribute} is a placeholder replaced by what is configured in `username_attribute`.
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
    
    # An additional dn to define the scope of groups
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
    group_name_attribute: cn
    
    # The attribute holding the mail address of the user
    mail_attribute: mail
    
    # The attribute holding the display name of the user. This will be used to greet an authenticated user.
    display_name_attribute: displayname

    # The username and password of the admin user. If multiple email addresses are defined for a user, only the first
    # one returned by the LDAP server is used.
    user: cn=admin,dc=example,dc=com
    
    # Password can also be set using a secret: https://docs.authelia.com/configuration/secrets.html
    password: password
```

The user must have an email address in order for Authelia to perform
identity verification when a user attempts to reset their password or
register a second factor device.


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
`sAMAccountName` for Microsoft Active Directory and `uid` for other implementations as the attribute holding the
unique identifier for your users.

## Loading a password from a secret instead of inside the configuration

Password can also be defined using a [secret](../secrets.md).
