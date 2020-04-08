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
    disable_reset_password: false
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
        
        # The username and password of the admin user.
        user: cn=admin,dc=example,dc=com
        
        # This secret can also be set using the env variables AUTHELIA_AUTHENTICATION_BACKEND_LDAP_PASSWORD
        password: password
```

The user must have an email address in order for Authelia to perform
identity verification when password reset request is initiated or
when a second factor device is registered.