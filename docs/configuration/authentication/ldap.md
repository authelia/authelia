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
        
        # An additional dn to define the scope to all users
        additional_users_dn: ou=users
        
        # The users filter used to find the user DN
        # {0} is a matcher replaced by username.
        # 'cn={0}' by default.
        users_filter: (cn={0})
        
        # An additional dn to define the scope of groups
        additional_groups_dn: ou=groups
        
        # The groups filter used for retrieving groups of a given user.
        # {0} is a matcher replaced by username.
        # {dn} is a matcher replaced by user DN.
        # {uid} is a matcher replaced by user uid.
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