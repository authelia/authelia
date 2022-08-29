---
title: "LDAP"
description: "An introduction into integrating Authelia with LDAP."
lead: "An introduction into integrating Authelia with LDAP."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  integration:
    parent: "ldap"
weight: 710
toc: true
---

## Tested Versions

* [Authelia]
  * [v4.36.5](https://github.com/authelia/authelia/releases/tag/v4.36.5)

## Configuration

### OpenLDAP
#### Tested Version: [Bitnami OpenLDAP - 2.5.13](https://github.com/bitnami/bitnami-docker-openldap/releases/tag/2.5.13-debian-11-r7)  
Create within OpenLDAP, either via CLI or with a GUI management application like [phpLDAPadmin](http://phpldapadmin.sourceforge.net/wiki/index.php/Main_Page) or [LDAP Admin](http://www.ldapadmin.org/) a basic user with a complex password.
*Make note of its CN.*
You can also create a group to use within Authelia if you would like granular control of who can login, and reference it within the filters below.

### Authelia

In your Authelia configuration you will need to enter and update the following variables - 
* url `ldap://OpenLDAP:1389` - servers dns name & port.  
  *tip: if you have Authelia on a container network that is routable, you can just use the container name*
* server_name `ldap01.example.com` - servers name
* base_dn `dc=example,dc=com` - common name of domain root.
* groups_filter `dc=example,dc=com` - replace relevant section with your own domain in common name format, same as base_dn.
* user `authelia` - username for Authelia service account
* password `SUPER_COMPLEX_PASSWORD` - password for Authelia service account

```yaml
  ldap:
    implementation: custom
    url: ldap://OpenLDAP:1389
    timeout: 5s
    start_tls: false
    tls:
      server_name: ldap01.example.com
      skip_verify: true
      minimum_version: TLS1.2
    base_dn: dc=example,dc=com
    additional_users_dn: ou=users
    users_filter: (&(|({username_attribute}={input})({mail_attribute}={input}))(objectClass=person))
    username_attribute: uid
    mail_attribute: mail
    display_name_attribute: displayName
    additional_groups_dn: ou=groups
    groups_filter: (&(member=uid={input},ou=users,dc=example,dc=com)(objectclass=groupofnames))
    group_name_attribute: cn
    user: uid=authelia,ou=service accounts,dc=example,dc=com
    password: "SUPER_COMPLEX_PASSWORD"
```
Following this, restart Authelia, and you should be able to begin using LDAP integration for your user logins, with Authelia taking the email attribute for users straight from the 'mail' attribute within the LDAP object.  

### FreeIPA
#### Tested Version: [FreeIPA - 4.9.9/CentOS]([https://github.com/bitnami/bitnami-docker-openldap/releases/tag/2.5.13-debian-11-r7](https://www.freeipa.org/page/Releases/4.9.9))  
Create within FreeIPA, either via CLI or within its GUI management application `https://server_ip` a basic user with a complex password.
*Make note of its CN.*
You can also create a group to use within Authelia if you would like granular control of who can login, and reference it within the filters below.

### Authelia

In your Authelia configuration you will need to enter and update the following variables - 
* url `ldap://ldap` - servers dns name. Port will assume 389 as standard. Specify custom port with `:port` if needed.  
* server_name `ldap01.example.com` - servers name
* base_dn `dc=example,dc=com` - common name of domain root.
* groups_filter `dc=example,dc=com` - replace relevant section with your own domain in common name format, same as base_dn.
* user `authelia` - username for Authelia service account
* password `SUPER_COMPLEX_PASSWORD` - password for Authelia service account

```yaml
    ldap:
    implementation: custom
    url: ldaps://ldap.example.com
    timeout: 5s
    start_tls: false
    tls:
      server_name: ldap.example.com
      skip_verify: true
      minimum_version: TLS1.2
    base_dn: dc=example,dc=com
    username_attribute: uid
    additional_users_dn: cn=users,cn=accounts
    users_filter: (&(|({username_attribute}={input})({mail_attribute}={input}))(objectClass=person))
    additional_groups_dn: ou=groups
    groups_filter: (&(member=uid={input},cn=users,cn=accounts,dc=example,dc=com)(objectclass=groupofnames))
    group_name_attribute: cn
    mail_attribute: mail
    display_name_attribute: displayName
    user: uid=authelia,cn=users,cn=accounts,dc=example,dc=com
    password: "SUPER_COMPLEX_PASSWORD"
```
Following this, restart Authelia, and you should be able to begin using LDAP integration for your user logins, with Authelia taking the email attribute for users straight from the 'mail' attribute within the LDAP object.  

## See Also
[Authelia]: https://www.authelia.com
[Bitname OpenLDAP]: [https://www.bookstackapp.com/](https://hub.docker.com/r/bitnami/openldap/)
[FreeIPA]: https://www.freeipa.org/page/Main_Page
