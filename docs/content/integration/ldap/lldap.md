---
title: "LLDAP"
description: ""
summary: ""
date: 2022-10-20T15:27:09+11:00
draft: false
images: []
weight: 352
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

[LLDAP] is supported by __Authelia__.

*__Important:__ When using these guides, it's important to recognize that we cannot provide a guide for every possible
method of deploying an LDAP server. These guides show a suggested setup only, and you need to understand the LDAP
configuration and customize it to your needs. To-that-end, we include links to the official documentation specific to
the LDAP implementation throughout this documentation and in the [See Also](#see-also) section.*

*__Important:__ This guide makes use of a default configuration. Check the [Defaults](#defaults) section
and make adjustments according to your needs.*

## Assumptions and Adaptation

This guide makes a few assumptions. These assumptions may require adaptation in more advanced and complex scenarios. We
can not reasonably have examples for every advanced configuration option that exists. Some of these values can
automatically be replaced with documentation variables.

The following are the assumptions we make:

- The LDAP implementation to be used with authelia is fully setup and reachable by authelia.
- All services are part of the `example.com` domain:
  - This domain and the subdomains will have to be adapted in all examples to match your specific domains unless you're
    just testing or you want to use that specific domain

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [authentication backend configuration] for use with [lldap] which will operate with the application example:

```yaml {title="configuration.yml"}
authentication_backend:
  ldap:
    implementation: 'lldap'
    address: 'ldap://lldap:3890'
    base_dn: 'dc=example,dc=com'
    user: 'uid=authelia,ou=people,dc=example,dc=com'
    password: 'insecure_secret'
```

### Application

Create within [lldap], a basic user with a complex password, and add it to the group `lldap_password_manager`.

Using the `lldap_password_manager` group will prevent authelia from changing passwords of users in the `lldap_admin` group.

You can also create a group to use within Authelia if you would like granular control of who can login, and reference it
within the filters below.

### Defaults

The below tables describes the current attribute defaults for the [lldap] implementation.

#### Search Base defaults

The following set defaults for the `additional_users_dn` and `additional_groups_dn` values.

|   Users   |  Groups   |
|:---------:|:---------:|
| OU=people | OU=groups |

#### Attribute defaults

This table describes the attribute defaults for the [lldap] implementation. i.e. the username_attribute is described by the
Username column.

|    Username    | Display Name | Mail | Group Name | Distinguished Name | Member Of |
|:--------------:|:------------:|:----:|:----------:|:------------------:|:---------:|
|      uid       |      cn      | mail |     cn     |        N/A         | memberOf  |

#### Filter defaults

The filters are probably the most important part to get correct when setting up LDAP. You want to exclude accounts under
the following conditions:

- The account is disabled or locked:
  - The [lldap] implementation has no suitable attribute for this as far as we're aware.
- Their password is expired:
  - The [lldap] implementation has no suitable attribute for this as far as we're aware.
- Their account is expired:
  - The [lldap] implementation has no suitable attribute for this as far as we're aware.

##### Users Filter
```text
(&(&#124;({username_attribute}={input})({mail_attribute}={input}))(objectClass=person))
```

##### Groups Filter
```text
(&(member={dn})(objectClass=groupOfNames))
```

## See Also
- [LLDAP Client Configuration](https://github.com/lldap/lldap?tab=readme-ov-file#client-configuration)

[Authelia]: https://www.authelia.com
[lldap]: https://github.com/lldap/lldap
[authentication backend configuration]: ../../../configuration/first-factor/ldap.md
