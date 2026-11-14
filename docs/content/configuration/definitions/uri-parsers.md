---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "URI Parsers"
description: "Configuring the Authelia URI parser definitions section which allows defining reusable named URI parsers for the authz endpoints."
summary: "Authelia allows configuring reusable URI parser definitions."
date: 2026-11-14T00:00:00+00:00
draft: false
images: []
weight: 199100
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

The URI parsers section configures named URI parsers which control how the request URI of an
[authz endpoint](../miscellaneous/server-endpoints-authz.md#uri_parser) is parsed. A URI containing a pattern which is
ambiguous, meaning different systems may interpret it differently, is rejected by default. Each definition changes the
default for individual ambiguity classes using the `allow` and `reject` options.

The `default` definition always exists, and it's used by every authz endpoint which doesn't configure the `uri_parser`
option. When it isn't configured it uses the default value of every ambiguity class, and configuring it changes the
rules for all of those endpoints.

The `authelia access-control check-policy` CLI command also uses the `default` definition to parse the URL, and the
`--uri-parser` flag selects another definition so the result matches an authz endpoint which uses it.

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
definitions:
  uri_parsers:
    default:
      allow: []
      reject: []
    legacy_application:
      allow: ['path_parameter']
      reject: ['path_encoded_percent']
```

## Options

This section describes the individual configuration options. The key of each definition is the name used by the
`uri_parser` option of the authz endpoints to reference it, and the value is the definition itself.

### key

The key is the name of the definition. In the example above, the keys are `default` and `legacy_application`, and these
are the values which must be used in other areas of the configuration to reference them.

### allow

{{< confkey type="list(string)" required="no" >}}

The ambiguity classes to allow which are otherwise rejected by default. A class must not be in both `allow` and
`reject`.

### reject

{{< confkey type="list(string)" required="no" >}}

The ambiguity classes to reject which are otherwise allowed by default. A class must not be in both `allow` and
`reject`.

## Ambiguity Classes

Access control rules match the decoded, cleaned path where `%2F` is `/`, redirections use the URL as requested, and the
canonical form of the request URI is used for matching the audiences and resources of a `Bearer` token.

Literal `.` and `..` dot segments, such as `/a/../b`, are normalized by default and are removed per RFC 3986. The
encoded forms are rejected by default as they are the ambiguous forms this option exists to stop.

Empty segments such as `//` are merged for the access control rules, so `//admin` is matched as `/admin`, and are
retained for matching the audiences and resources of a `Bearer` token, so `/a//b` doesn't match `/a/b`. Empty segments
can be rejected with the `path_empty_segment` class.

A dot segment which follows an empty segment, such as `/public//../admin`, is rejected by default with the
`path_empty_dot_segment` class. Per RFC 3986 the `..` only removes the empty segment so the path is `/public/admin`,
whereas implementations which merge empty segments first resolve the path to `/admin`. When this class is allowed the
path is resolved per RFC 3986 for both the access control rules and the comparison with a `Bearer` token.

The following table lists the ambiguity classes which are valid for both the `allow` and `reject` options, and the part
of the URI each one checks. Classes prefixed with `path_` only check the path, classes prefixed with `query_` only check
the query, and classes without a prefix check the part listed. A `path_` class never rejects a URI because of its query,
for example `?a=%2F` doesn't match `path_encoded_separator`.

| Name                              | Checks         | Meaning                                                                                                                                                                             | Default |
| --------------------------------- | -------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- |
| `user_info`                       | Authority      | A user part is present, including an empty one                                                                                                                                      | Reject  |
| `control_character`               | Path and Query | A control character, literal or encoded                                                                                                                                             | Reject  |
| `fragment`                        | Entire URI     | A literal `#` anywhere in the input                                                                                                                                                 | Reject  |
| `path_encoded_dot_segment`        | Path           | A dot segment formed with encoding, such as `%2e%2e`, `%2f..%2f`, `/%2e.%2f`, `..%2f`                                                                                               | Reject  |
| `path_parameter_dot_segment`      | Path           | A dot or empty segment with a parameter, such as `/..;/`, `/.;/`, `/;/`, `/%2e%2e;x/`                                                                                               | Reject  |
| `path_double_encoded_dot_segment` | Path           | The encoded dot segment check matches only after a second decode, such as `%252e%252e%252f`                                                                                         | Reject  |
| `path_empty_dot_segment`          | Path           | A dot segment following an empty segment, such as `/public//../admin`                                                                                                               | Reject  |
| `path_dot_segment`                | Path           | Any literal `.` or `..` segment                                                                                                                                                     | Allow   |
| `path_empty_segment`              | Path           | Any `//`                                                                                                                                                                            | Allow   |
| `path_backslash`                  | Path           | Literal `\` or `%5C`                                                                                                                                                                | Reject  |
| `path_invalid_utf8`               | Path           | The decoded path isn't valid UTF-8, such as `%C0%AE`                                                                                                                                | Reject  |
| `path_parameter`                  | Path           | A literal `;`, `%3B` falls under `path_encoded_non_unreserved`                                                                                                                      | Reject  |
| `path_encoded_separator`          | Path           | Any `%2F`                                                                                                                                                                           | Allow   |
| `path_encoded_percent`            | Path           | Any `%25`                                                                                                                                                                           | Allow   |
| `path_encoded_non_unreserved`     | Path           | Any percent-encoding of a character which isn't unreserved, such as `%2F`, `%25`, `%3B`, `%20`, `%C3%A9`. Blocks all spaces and non-ASCII paths, so it suits strict ASCII-only APIs | Allow   |
| `path_encoded_unreserved`         | Path           | Any percent-encoding of an unreserved character, such as `%41`, `%7E`, `%2E`                                                                                                        | Allow   |
| `query_invalid_encoding`          | Query          | A malformed percent-encoding                                                                                                                                                        | Allow   |

The `path_dot_segment` class rejects literal dot segments instead of normalizing them.
