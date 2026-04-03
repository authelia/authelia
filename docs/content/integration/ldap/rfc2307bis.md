---
title: "RFC2307bis"
description: ""
summary: ""
date: 2025-05-22T10:12:47+00:00
draft: false
images: []
weight: 752
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

[RFC2307bis] is supported by __Authelia__.

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

- The LDAP implementation to be used with Authelia is fully setup and reachable by Authelia.
- All services are part of the `example.com` domain:
  - This domain and the subdomains will have to be adapted in all examples to match your specific domains unless you're
    just testing or you want to use that specific domain.

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [authentication backend configuration] for use with [RFC2307bis] which will operate with the application example:

```yaml {title="configuration.yml"}
authentication_backend:
  ldap:
    implementation: 'lldap'
    address: 'ldaps://ldap.example.com'
    base_dn: 'DC=example,DC=com'
    user: 'UID=authelia,OU=people,DC=example,DC=com'
    password: 'insecure_secret'
```

### Application

Create a service user within the application with a complex password. Use the users Distinguished Name as a username,
and make sure the user has the appropriate permissions to perform the following actions:

- Read the attributes of users and groups that are meant to be able to use Authelia.
- Change the password of users provided the functionality to reset passwords is desired.

See the documentation from the maintainer or vendor of the RFC2307bis LDAP server on how to configure permissions for
the newly created user.

### Defaults

The below tables describes the current attribute defaults for each implementation.

#### Attribute defaults

This table describes the attribute defaults for each implementation. i.e. the username_attribute is described by the
Username column.

|    Username    | Display Name | Mail | Group Name | Distinguished Name | Member Of |
|:--------------:|:------------:|:----:|:----------:|:------------------:|:---------:|
|      uid       | displayName  | mail |     cn     |        N/A         | memberOf  |


#### Filter defaults

The filters are probably the most important part to get correct when setting up LDAP. You want to exclude accounts under
the following conditions:

- The account is disabled or locked:
  - The [RFC2307bis] implementation has no suitable attribute for this as far as we're aware.

- Their password is expired:
  - `(!(pwdReset=TRUE))`

- Their account is expired:
  - The [RFC2307bis] implementation has no suitable attribute for this as far as we're aware.

##### Users Filter
```text
(&(&#124;({username_attribute}={input})({mail_attribute}={input}))(&#124;(objectClass=inetOrgPerson)(objectClass=organizationalPerson))(!(pwdReset=TRUE)))
```

##### Groups Filter
```text
(&(&#124;(member={dn})(uniqueMember={dn}))(&#124;(objectClass=groupOfNames)(objectClass=groupOfUniqueNames)(objectClass=groupOfMembers)))
```

## See Also

[Authelia]: https://www.authelia.com
[RFC2307bis]: https://datatracker.ietf.org/doc/html/draft-howard-rfc2307bis-02
[authentication backend configuration]: ../../../configuration/first-factor/ldap.md
