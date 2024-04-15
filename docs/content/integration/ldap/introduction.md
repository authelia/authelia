---
title: "LDAP"
description: "An introduction into integrating Authelia with LDAP."
summary: "An introduction into integrating Authelia with LDAP."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
weight: 710
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## UNDER CONSTRUCTION

This section is still a work in progress.

## Configuration

### OpenLDAP

**Tested:**
* Version: [v2.5.13](https://www.openldap.org/software/release/announce_lts.html)
* Container `bitnami/openldap:2.5.13-debian-11-r7`

Create within OpenLDAP, either via CLI or with a GUI management application like
[phpLDAPadmin](http://phpldapadmin.sourceforge.net/wiki/index.php/Main_Page) or [LDAP Admin](http://www.ldapadmin.org/)
a basic user with a complex password.

*Make note of its CN.* You can also create a group to use within Authelia if you would like granular control of who can
login, and reference it within the filters below.

### Authelia

In your Authelia configuration you will need to enter and update the following variables -
* url `ldap://OpenLDAP:1389` - servers dns name & port.
  *tip: if you have Authelia on a container network that is routable, you can just use the container name*
* server_name `ldap01.example.com` - servers name
* base_dn `DC=example,DC=com` - common name of domain root.
* groups_filter `DC=example,DC=com` - replace relevant section with your own domain in common name format, same as base_dn.
* user `authelia` - username for Authelia service account
* password `SUPER_COMPLEX_PASSWORD` - password for Authelia service account

```yaml {title="configuration.yml"}
authentication_backend:
  ldap:
    address: 'ldap://OpenLDAP:1389'
    implementation: 'custom'
    timeout: '5s'
    start_tls: false
    tls:
      server_name: 'ldap01.example.com'
      skip_verify: true
      minimum_version: 'TLS1.2'
    base_dn: 'DC=example,DC=com'
    additional_users_dn: 'OU=users'
    users_filter: '(&(|({username_attribute}={input})({mail_attribute}={input}))(objectClass=person))'
    additional_groups_dn: 'OU=groups'
    groups_filter: '(&(member=UID={input},OU=users,DC=example,DC=com)(objectClass=groupOfNames))'
    user: 'UID=authelia,OU=service accounts,DC=example,DC=com'
    password: "SUPER_COMPLEX_PASSWORD"
    attributes:
      distinguished_name: 'distinguishedName'
      username: 'uid'
      mail: 'mail'
      member_of: 'memberOf'
      group_name: 'cn'
```
Following this, restart Authelia, and you should be able to begin using LDAP integration for your user logins, with
Authelia taking the email attribute for users straight from the 'mail' attribute within the LDAP object.

### FreeIPA

**Tested:**
* Version: [v4.9.9](https://www.freeipa.org/page/Releases/4.9.9)
* Container: `freeipa/freeipa-server:fedora-36-4.9.9`

Create within FreeIPA, either via CLI or within its GUI management application `https://server_ip` a basic user with a
complex password.

*Make note of its CN.* You can also create a group to use within Authelia if you would like granular control of who can
login, and reference it within the filters below.

### Authelia

In your Authelia configuration you will need to enter and update the following variables -
* url `ldap://ldap` - servers dns name. Port will assume 389 as standard. Specify custom port with `:port` if needed.
* server_name `ldap01.example.com` - servers name
* base_dn `DC=example,DC=com` - common name of domain root.
* groups_filter `DC=example,DC=com` - replace relevant section with your own domain in common name format, same as base_dn.
* user `authelia` - username for Authelia service account
* password `SUPER_COMPLEX_PASSWORD` - password for Authelia service account

```yaml {title="configuration.yml"}
authentication_backend:
 ldap:
    address: 'ldaps://ldap.example.com'
    implementation: 'custom'
    timeout: '5s'
    start_tls: false
    tls:
      server_name: 'ldap.example.com'
      skip_verify: true
      minimum_version: 'TLS1.2'
    base_dn: 'dc=example,DC=com'
    additional_users_dn: 'CN=users,CN=accounts'
    users_filter: '(&(|({username_attribute}={input})({mail_attribute}={input}))(objectClass=person))'
    additional_groups_dn: cn=groups,cn=accounts
    groups_filter: '(&(member=UID={input},CN=users,CN=accounts,DC=example,DC=com)(objectClass=groupOfNames))'
    user: 'UID=authelia,CN=users,CN=accounts,DC=example,DC=com'
    password: 'SUPER_COMPLEX_PASSWORD'
    attributes:
      distinguished_name: 'distinguishedName'
      username: 'uid'
      mail: 'mail'
      member_of: 'memberOf'
      group_name: 'cn'
```
Following this, restart Authelia, and you should be able to begin using LDAP integration for your user logins, with
Authelia taking the email attribute for users straight from the 'mail' attribute within the LDAP object.

### lldap

**Tested:**
* Version: [v0.4.0](https://github.com/nitnelave/lldap/releases/tag/v0.4.07)

Create within lldap, a basic user with a complex password, and add to the group "lldap_password_manager"
You can also create a group to use within Authelia if you would like granular control of who can login, and reference it
within the filters below.

### Authelia

In your Authelia configuration you will need to enter and update the following variables -
* url `ldap://OpenLDAP:1389` - servers dns name & port.
  *tip: if you have Authelia on a container network that is routable, you can just use the container name*
* base_dn `DC=example,DC=com` - common name of domain root.
* user `authelia` - username for Authelia service account.
* password `SUPER_COMPLEX_PASSWORD` - password for Authelia service account,

```yaml {title="configuration.yml"}
authentication_backend:
  ldap:
    address: 'ldap://lldap:3890'
    implementation: 'custom'
    timeout: '5s'
    start_tls: false
    base_dn: 'dc=example,DC=com'
    additional_users_dn: 'OU=people'
    # To allow sign in both with username and email, one can use a filter like
    # (&(|({username_attribute}={input})({mail_attribute}={input}))(objectClass=person))
    users_filter: '(&({username_attribute}={input})(objectClass=person))'
    additional_groups_dn: 'OU=groups'
    groups_filter: '(member={dn})'
    # The username and password of the admin or service user.
    user: 'UID=authelia,OU=people,DC=example,DC=com'
    password: 'SUPER_COMPLEX_PASSWORD'
    attributes:
      distinguished_name: 'distinguishedName'
      username: 'uid'
      mail: 'mail'
      member_of: 'memberOf'
      group_name: 'cn'
```
Following this, restart Authelia, and you should be able to begin using lldap integration for your user logins, with
Authelia taking the email attribute for users straight from the 'mail' attribute within the LDAP object.

## See Also

[Authelia]: https://www.authelia.com
[Bitnami OpenLDAP]: https://hub.docker.com/r/bitnami/openldap/
[FreeIPA]: https://www.freeipa.org/page/Main_Page
[lldap]: https://github.com/nitnelave/lldap
