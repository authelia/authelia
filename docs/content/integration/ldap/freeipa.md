---
title: "FreeIPA"
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
* [FreeIPA]
  * [4.9.9](https://www.freeipa.org/page/Releases/4.9.9)

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

The following YAML configuration is an example __Authelia__ [authentication backend configuration] for use with [FreeIPA] which will operate with the application example:

```yaml {title="configuration.yml"}
authentication_backend:
  ldap:
    implementation: 'freeipa'
    address: 'ldaps://ldap.example'
    base_dn: 'dc=example,dc=com'
    user: 'uid=authelia,ou=people,dc=example,dc=com'
    password: 'insecure_secret'
```

### Application
Create within [FreeIPA], either via CLI or within its GUI management application `https://ldap.{{< sitevar name="domain" nojs="example.com" >}}` a basic user with a
complex password.

*Make note of its CN.* You can also create a group to use within Authelia if you would like granular control of who can
login, and reference it within the filters below.

### Defaults

The below tables describes the current attribute defaults for the [FreeIPA] implementation.

#### Attribute defaults

This table describes the attribute defaults for the [FreeIPA] implementation. i.e. the username_attribute is described by the
Username column.

|    Username    | Display Name | Mail | Group Name | Distinguished Name | Member Of |
|:--------------:|:------------:|:----:|:----------:|:------------------:|:---------:|
|      uid       | displayName  | mail |     cn     |        N/A         | memberOf  |


#### Filter defaults

The filters are probably the most important part to get correct when setting up LDAP. You want to exclude accounts under
the following conditions:

- The account is disabled or locked:
  - `(!(nsAccountLock=TRUE))`

- Their password is expired:
  - `(krbPasswordExpiration>={date-time:generalized})`

- Their account is expired:
  - `(|(!(krbPrincipalExpiration=*))(krbPrincipalExpiration>={date-time:generalized}))`

##### Users Filter
```text
(&(&#124;({username_attribute}={input})({mail_attribute}={input}))(objectClass=person)(!(nsAccountLock=TRUE))(krbPasswordExpiration>={date-time:generalized})(&#124;(!(krbPrincipalExpiration=*))(krbPrincipalExpiration>={date-time:generalized})))
```

##### Groups Filter
```text
(&(member={dn})(objectClass=groupOfNames))
```

[Authelia]: https://www.authelia.com
[FreeIPA]: https://www.freeipa.org/
[authentication backend configuration]: ../../../configuration/first-factor/ldap.md
