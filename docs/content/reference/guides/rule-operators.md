---
title: "Access Control Rule Guide"
description: "A reference guide on access control rule operators"
summary: "This section contains a reference guide on access control rule operators."
date: 2024-03-14T06:00:14+11:00
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

## Operators

Rule operators are effectively words which alter the behavior of particular access control rules. The following table
is a guide on their use.

|   Operator    |                             Effect                             |
|:-------------:|:--------------------------------------------------------------:|
|    `equal`    |   Matches when the item value is equal to the provided value   |
|  `not equal`  | Matches when the item value is not equal to the provided value |
|   `present`   |        Matches when the item is present with any value         |
|   `absent`    |          Matches when the item is not present at all           |
|   `pattern`   |        Matches when the item matches the regex pattern         |
| `not pattern` |     Matches when the item doesn't match the regex pattern      |


## Multi-level Logical Criteria

Criteria which is described as multi-level logical criteria indicates that it is a list of lists. The first level i.e.
the list least indented to the right will be referred to the `OR-list`, and the list most indented to the right will be
referred to the `AND-list`.

The OR-list matches if any of the criteria from it's AND-list's matches; in other words, a *__logical OR__*. The
AND-list matches if all of it's criteria matches the given request; in other words, a *__logical AND__*.

In addition to these rules, if the AND-list only needs one item, it can be represented without the second level.

### Examples

#### List of Lists

The following examples show various abstract examples to express a rule that matches either c, or a AND b;
i.e `(a AND b) OR (c)`. In relation to access control rules all of these should be treated the same. This format should
not be used for the configuration item type `list(list(object))`, see [List of List Objects](#list-of-list-objects)
instead.

##### Fully Expressed

```yaml {title="configuration.yml"}
rule:
  - - 'a'
    - 'b'
  - - 'c'
```

##### Omitted Level

```yaml {title="configuration.yml"}
rule:
  - - 'a'
    - 'b'
  - 'c'
```

##### Compact

```yaml {title="configuration.yml"}
rule:
  - ['a', 'b']
  - ['c']
```

##### Compact with Omitted Level

```yaml {title="configuration.yml"}
rule:
  - ['a', 'b']
  - 'c'
```

##### Super Compact

```yaml {title="configuration.yml"}
rule: [['a', 'b'], ['c']]
```

#### List of List Objects

The following examples show various abstract examples that mirror the above rules however the AND-list is a list of
objects where the key is named `value`. This format should only be used for the configuration item type
`list(list(object))`, see [List of Lists](#list-of-lists) if you're not looking for a `list(list(object))`

##### Fully Expressed

```yaml {title="configuration.yml"}
rule:
  - - value: 'a'
    - value: 'b'
  - - value: 'c'
```

##### Omitted Level

```yaml {title="configuration.yml"}
rule:
  - - 'a'
    - 'b'
  - value: 'c'
```

##### Compact

```yaml {title="configuration.yml"}
rule:
  - ['a', 'b']
  - ['c']
```

##### Compact with Omitted Level

```yaml {title="configuration.yml"}
rule:
  - ['a', 'b']
  - 'c'
```

##### Super Compact

```yaml {title="configuration.yml"}
rule: [['a', 'b'], ['c']]
```
