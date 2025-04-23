---
title: "RFC2307bis"
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

## Tested Versions

* [Authelia]
  * [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)

## Assumptions and Adaptation

This guide makes a few assumptions. These assumptions may require adaptation in more advanced and complex scenarios. We
can not reasonably have examples for every advanced configuration option that exists. Some of these values can
automatically be replaced with documentation variables.

The following are the assumptions we make:

* All services are part of the `example.com` domain:
  * This domain and the subdomains will have to be adapted in all examples to match your specific domains unless you're
    just testing or you want to use that specific domain

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [authentication backend configuration] for use with [RFC2307bis] which will operate with the application example:

```yaml {title="configuration.yml"}
authentication_backend:
  ldap:
    implementation: 'lldap'
    address: 'ldaps://ldap.example.com'
    base_dn: 'dc=example,dc=com'
    user: 'uid=authelia,ou=people,dc=example,dc=com'
    password: 'insecure_secret'
```

### Application

This integration guide is application agnostic, refer to the documentation of your LDAP server.

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

[Authelia]: https://www.authelia.com
[RFC2307bis]: https://datatracker.ietf.org/doc/html/draft-howard-rfc2307bis-02
[authentication backend configuration]: ../../../configuration/first-factor/ldap.md
