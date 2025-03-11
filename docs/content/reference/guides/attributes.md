---
title: "Attributes"
description: "This guide highlights information about attributes available via various methods"
summary: "This guide highlights information about attributes available via various methods."
date: 2025-02-22T06:40:08+00:00
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
3. Custom Attributes derived from the other available attribute sources using the [Common Expression Language](https://github.com/google/cel-spec).

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

### Validation

The standard user attributes are validated against several constraints. This table describes the constraints, the
attribute must satisfy all the constrains not marked as `N/A`.

|    Attribute    | Constraint: Type | Constraint: Multi-Value |   Constraint: Syntax    |
|:---------------:|:----------------:|:-----------------------:|:-----------------------:|
|    username     |      string      |      Single Valued      |           N/A           |
|  display_name   |      string      |      Single Valued      |           N/A           |
|   family_name   |      string      |      Single Valued      |           N/A           |
|   given_name    |      string      |      Single Valued      |           N/A           |
|   middle_name   |      string      |      Single Valued      |           N/A           |
|    nickname     |      string      |      Single Valued      |           N/A           |
|     gender      |      string      |      Single Valued      |           N/A           |
|    birthdate    |      string      |      Single Valued      |           N/A           |
|     website     |      string      |      Single Valued      | [RFC3986: Absolute URI] |
|     profile     |      string      |      Single Valued      | [RFC3986: Absolute URI] |
|     picture     |      string      |      Single Valued      | [RFC3986: Absolute URI] |
|    zoneinfo     |      string      |      Single Valued      |           N/A           |
|     locale      |      string      |      Single Valued      |        [BCP 47]         |
|  phone_number   |      string      |      Single Valued      |           N/A           |
| phone_extension |      string      |      Single Valued      |           N/A           |
| street_address  |      string      |      Single Valued      |           N/A           |
|    locality     |      string      |      Single Valued      |           N/A           |
|     region      |      string      |      Single Valued      |           N/A           |
|   postal_code   |      string      |      Single Valued      |           N/A           |
|     country     |      string      |      Single Valued      |           N/A           |
|      mail       |      string      |           N/A           |     [RFC5322: Addr]     |

[BCP 47]: https://www.rfc-editor.org/info/bcp47
[RFC3986: Absolute URI]: https://datatracker.ietf.org/doc/html/rfc3986#section-4.3
[RFC5322: Addr]: https://datatracker.ietf.org/doc/html/rfc5322#section-3.4.1

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
