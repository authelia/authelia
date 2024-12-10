---
title: "Attributes"
description: "This guide highlights information about attributes available via various methods"
summary: "This guide highlights information about attributes available via various methods."
date: 2022-06-20T10:05:55+10:00
draft: false
images: []
weight: 220
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

Authelia has three primary methods of deriving attributes:

1. Standard Attributes derived directly from the authentication backend.
2. Extra Attributes which are manually configured but still derived from the authentication backend.
3. Custom Attributes derived from the other available attribute sources using the [Common Expression Language].

## Standard Attributes

Standard Attributes are the ones that commonly available in most Authentication Backends directly. The
[LDAP Backend](../../configuration/first-factor/ldap.md#attributes) allows configuring the LDAP properties these values
come from, and the [File Backend](../../configuration/first-factor/file.md) directly supports all of them.

All of the standard attributes have a 1:1 mapping with the internal attribute name. For example with LDAP if you
configure the following then the LDAP property named `l` will be mapped to the Authelia attribute `locality`:

```yaml
authentication_backend:
  ldap:
    attributes:
      locality: 'l'
```

## Extra Attributes

Extra Attributes are special extra attributes where you have to define characteristics about them. Third-party
authentication backends like LDAP allow renaming these attributes, first-party authentication backends do
not.

Attributes can have the following types:

- `string`
- `integer` (validated)
- `boolean` (validated)

The following example loads the LDAP property `ldapAttributeName` into the Authelia attribute `autheliaAttributeName`,
treats it as a single-valued property, and considers it a string.

```yaml {title="configuration.yml"}
authentication_backend:
  ldap:
    attributes:
      extra:
        ldapAttributeName:
          name: 'autheliaAttributeName'
          multi_valued: false
          value_type: 'string'
```

The following example loads the YAML property `autheliaAttributeName` into the Authelia attribute of the same name,
treats it as a single-valued property, and considers it a string.


```yaml {title="configuration.yml"}
authentication_backend:
  file:
    extra_attributes:
      autheliaAttributeName:
        multi_valued: false
        value_type: 'string'
```

## Custom Attributes

Custom Attributes are one of the more exiting features introduced in 4.39 which allow you to configure an attribute
that's derived from other attributes. For example you may wish to provide a boolean value as to if a user is a member
of a specific group.

The following example creates a custom attribute named `is_admin` which returns a boolean if the user is in the group
`admin`.

```yaml {title="configuration.yml"}
definitions:
  user_attributes:
    is_admin:
      expression: '"admin" in groups'
```
