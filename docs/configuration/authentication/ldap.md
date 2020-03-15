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
    ldap:
        # The url to the ldap server. Scheme can be ldap:// or ldaps://
        url: ldap://127.0.0.1

        # Skip verifying the server certificate (to allow self-signed certificate).
        skip_verify: false

        # The base dn for every entries
        base_dn: dc=example,dc=com

        # The attribute holding the username of the user (introduced to handle
        # case insensitive search queries: #561).
        # Microsoft Active Directory usually uses 'sAMAccountName'
        # OpenLDAP usually uses 'uid'
        username_attribute: uid
        
        # An additional dn to define the scope to all users
        additional_users_dn: ou=users
        
        # This attribute is optional. The user filter used in the LDAP search queries
        # is a combination of this filter and the username attribute.
        # This filter is used to reduce the scope of users targeted by the LDAP search query.
        # For instance, if the username attribute is set to 'uid', the computed filter is
        # (&(uid=<username>)(&(objectCategory=person)(objectClass=user)))
        # Recommended settings are as follows:
        # Microsoft Active Directory '(&(objectCategory=person)(objectClass=user))'
        # OpenLDAP '(objectClass=person)' or '(objectClass=inetOrgPerson)'
        users_filter: (&(objectCategory=person)(objectClass=user))
        
        # An additional dn to define the scope of groups
        additional_groups_dn: ou=groups
        
        # The groups filter used for retrieving groups of a given user.
        # {0} is a matcher replaced by username (as provided in login portal).
        # {1} is a matcher replaced by username (as stored in LDAP).
        # {dn} is a matcher replaced by user DN.
        # 'member={dn}' by default.
        groups_filter: (&(member={dn})(objectclass=groupOfNames))
        
        # The attribute holding the name of the group
        group_name_attribute: cn
        
        # The attribute holding the mail address of the user
        mail_attribute: mail
        
        # The username and password of the admin user.
        user: cn=admin,dc=example,dc=com
        
        # This secret can also be set using the env variables AUTHELIA_AUTHENTICATION_BACKEND_LDAP_PASSWORD
        password: password
```

The user must have an email address in order for Authelia to perform
identity verification when password reset request is initiated or
when a second factor device is registered.