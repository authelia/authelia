---
title: "Access Control"
description: "Configuring the Authelia access control rules including domain matching, user and group policies, network conditions, URI patterns, and CEL expressions."
summary: "Authelia supports a comprehensive access control system. This section describes configuring this."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 104200
toc: true
aliases:
  - /c/acl
  - /docs/configuration/access-control.html
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
This section does not apply to OpenID Connect 1.0. See the [Frequently Asked Questions](../../integration/openid-connect/frequently-asked-questions.md#why-doesnt-the-access-control-configuration-work-with-openid-connect-10) for more
information.
{{< /callout >}}

## Variables

Some of the values within this page can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
access_control:
  default_policy: 'deny'
  rules:
    - domain: 'private.{{< sitevar name="domain" nojs="example.com" >}}'
      domain_regex: '^(\d+\-)?priv-img\.{{< sitevar name="domain" format="regex" nojs="example\.com" >}}$'
      policy: 'one_factor'
      networks:
        - 'internal'
        - '1.1.1.1'
      subject:
        - ['user:adam']
        - ['user:fred']
        - ['group:admins']
      methods:
        - 'GET'
        - 'HEAD'
      resources:
        - '/api([/?].*)?$'
      query:
        - - operator: 'present'
            key: 'secure'
          - operator: 'absent'
            key: 'insecure'
        - - operator: 'pattern'
            key: 'token'
            value: '^(abc123|zyx789)$'
          - operator: 'not pattern'
            key: 'random'
            value: '^(1|2)$'
```

## Options

This section describes the individual configuration options.

### default_policy

{{< confkey type="string" default="deny" required="no" >}}

The default [policy](#policies) defines the policy applied if no [rules](#rules) section apply to the information known
about the request. It is recommended that this is configured to [deny] for security reasons. Sites which you do
not wish to secure at all with Authelia should not be configured in your reverse proxy to perform authentication with
Authelia at all for performance reasons.

See the [policies] section for more information.

### rules

{{< confkey type="list" required="no" >}}

The rules have many configuration options. A rule matches when all criteria of the rule match the request excluding the
[policy] which is the [policy](#policies) applied to the request.

A rule defines two primary things:

- the policy applied when all criteria match
- the matching criteria of the request presented to the reverse proxy

The criteria is broken into several parts:

- [domain]: domain or list of domains targeted by the request.
- [domain_regex]: regex form of [domain].
- [resources]: pattern or list of patterns that the path should match.
- [subject]: the user or group of users to define the policy for.
- [networks]: the network addresses, ranges (CIDR notation) or groups from where the request originates.
- [methods]: the http methods used in the request.

A rule is matched when all criteria of the rule match. Rules are evaluated in sequential order as per
[Rule Matching Concept 1]. It's _**strongly recommended**_ that individuals read the [Rule Matching](#rule-matching)
section.

[rules]: #rules

#### domain

{{< confkey type="list(string)" required="yes" >}}

_**Required:** This criteria and/or the [domain_regex] criteria are required._

This criteria matches the domain name and has two methods of configuration, either as a single string or as a list of
strings. When it's a list of strings the rule matches when **any** of the domains in the list match the request domain.
When used in conjunction with [domain_regex] the rule will match when either the [domain] or the [domain_regex] criteria
matches.

Rules may start with a few different wildcards:

- The standard wildcard is `*.`, which when in front of a domain means that any subdomain is effectively a match. For
  example `*.{{< sitevar name="domain" nojs="example.com" >}}` would match `abc.{{< sitevar name="domain" nojs="example.com" >}}` and `secure.{{< sitevar name="domain" nojs="example.com" >}}`.
  When using a wildcard like this the string **must** be quoted like `'*.{{< sitevar name="domain" nojs="example.com" >}}'`.
- There previously were user and group wildcards (`{user}.` and `{group}.`) which are officially deprecated as the
  [domain_regex] criteria completely replaces the functionality in a much more useful way. For backwards compatibility
  these old wildcards remain functional as they're automatically translated into the equivalent [domain_regex]; only
  specific invalid combinations (such as a `bypass` policy used together with one of these wildcards) are rejected during
  validation. It's strongly recommended to migrate to [domain_regex] instead.

  The translated patterns only match characters which are valid in a hostname, i.e. letters, digits, and the hyphen. The
  `{user}.` wildcard matches one or more labels as a username is permitted to contain periods, whereas the `{group}.`
  wildcard matches a single label. Users and groups containing any other character, such as an underscore, are not
  matched by these wildcards; use the [domain_regex] criteria with a `User` or `Group` named group if you need to match
  them.

Domains in this section must be the domain configured in the [session](../session/introduction.md#domain) configuration
or subdomains of that domain. This is because a website can only write cookies for a domain it is part of. It is
theoretically possible for us to do this with multiple domains however we would have to be security conscious in our
implementation, and it is not currently a priority.

[domain]: #domain

##### Examples

_Single domain of `*.{{< sitevar name="domain" nojs="example.com" >}}` matched. All rules in this list are effectively the same rule just expressed in
different ways._

```yaml {title="configuration.yml"}
access_control:
  rules:
    - domain: '*.{{< sitevar name="domain" nojs="example.com" >}}'
      policy: 'bypass'
    - domain:
        - '*.{{< sitevar name="domain" nojs="example.com" >}}'
      policy: 'bypass'
```

_Multiple domains matched. These rules will match either `apple.{{< sitevar name="domain" nojs="example.com" >}}` or `banana.{{< sitevar name="domain" nojs="example.com" >}}`. All rules in this
list are effectively the same rule just expressed in different ways._

```yaml {title="configuration.yml"}
access_control:
  rules:
    - domain: ['apple.{{< sitevar name="domain" nojs="example.com" >}}', 'banana.{{< sitevar name="domain" nojs="example.com" >}}']
      policy: 'bypass'
    - domain:
        - 'apple.{{< sitevar name="domain" nojs="example.com" >}}'
        - 'banana.{{< sitevar name="domain" nojs="example.com" >}}'
      policy: 'bypass'
```

_Multiple domains matched either via a static domain or via a [domain_regex]. This rule will match
either `apple.{{< sitevar name="domain" nojs="example.com" >}}`, `pub-data.{{< sitevar name="domain" nojs="example.com" >}}`, or `img-data.{{< sitevar name="domain" nojs="example.com" >}}`._

```yaml {title="configuration.yml"}
access_control:
  rules:
    - domain: 'apple.{{< sitevar name="domain" nojs="example.com" >}}'
      domain_regex: '^(pub|img)-data\.{{< sitevar name="domain" format="regex" nojs="example\.com" >}}$'
      policy: 'bypass'
```

#### domain_regex

{{< confkey type="list(string)" required="yes" >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
If you intend to use this criteria with a bypass rule please read [Rule Matching Concept 2](#rule-matching-concept-2-subject-criteria-requires-authentication).
{{< /callout >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
To utilize regex you must escape it properly. See
[regular expressions](../prologue/common.md#regular-expressions) for more information.
{{< /callout >}}

_**Required:** This criteria and/or the [domain] criteria are required._

This criteria matches the domain name and has two methods of configuration, either as a single string or as a list of
strings. When it's a list of strings the rule matches when **any** of the domains in the list match the request domain.
When used in conjunction with [domain] the rule will match when either the [domain] or the [domain_regex] criteria matches.

In addition to standard regex patterns this criteria can match some [Named Regex Groups].

[domain_regex]: #domain_regex

##### Examples

_An advanced multiple domain regex example with user/group matching. This will match the user `john` in the groups
`example` and `example1`, when the request is made to `user-john.{{< sitevar name="domain" nojs="example.com" >}}`,
`group-example.{{< sitevar name="domain" nojs="example.com" >}}`, or `group-example1.{{< sitevar name="domain" nojs="example.com" >}}`, it would not match when the
request is made to `user-fred.{{< sitevar name="domain" nojs="example.com" >}}` or `group-admin.{{< sitevar name="domain" nojs="example.com" >}}`._

```yaml {title="configuration.yml"}
access_control:
  rules:
    - domain_regex:
        - '^user-(?P<User>\w+)\.{{< sitevar name="domain" format="regex" nojs="example\.com" >}}$'
        - '^group-(?P<Group>\w+)\.{{< sitevar name="domain" format="regex" nojs="example\.com" >}}$'
      policy: 'one_factor'
```

_Multiple domains example, one with a static domain and one with a regex domain. This will match requests to
`protected.{{< sitevar name="domain" nojs="example.com" >}}`, `img-private.{{< sitevar name="domain" nojs="example.com" >}}`, or `data-private.{{< sitevar name="domain" nojs="example.com" >}}`._

```yaml {title="configuration.yml"}
access_control:
  rules:
    - domain: 'protected.{{< sitevar name="domain" nojs="example.com" >}}'
      domain_regex: '^(img|data)-private\.{{< sitevar name="domain" format="regex" nojs="example\.com" >}}'
      policy: 'one_factor'
```

#### policy

{{< confkey type="string" required="yes" >}}

The specific [policy](#policies) to apply to the selected rule. This is not criteria for a match, this is the action to
take when a match is made.

[policy]: #policy

#### subject

{{< confkey type="list(list(string))" required="no" >}}

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
This rule criteria **may not** be used for the [bypass](#bypass) policy the minimum required authentication level to
identify the subject is [one_factor](#one_factor). See [Rule Matching Concept 2](#rule-matching-concept-2-subject-criteria-requires-authentication) for more information.
{{< /callout >}}

This criteria matches identifying characteristics about the subject. Currently this is either user or groups the user
belongs to. This allows you to effectively control exactly what each user is authorized to access or to specifically
require two-factor authentication to specific users. Subjects must be prefixed with the following prefixes to
specifically match a specific part of a subject.

|   Subject Type   |      Prefix      |                                                                  Description                                                                   |
| :--------------: | :--------------: | :--------------------------------------------------------------------------------------------------------------------------------------------: |
|       User       |     `user:`      |                                                        Matches the username of a user.                                                         |
|      Group       |     `group:`     |                                                Matches if the user has a group with this name.                                                 |
| OAuth 2.0 Client | `oauth2:client:` | Matches if the request has been authorized via a token issued by a client with the specified id utilizing the `client_credentials` grant type. |

The format of this rule is unique in as much as it is a list of lists. The logic behind this format is to allow for both
`OR` and `AND` logic. The first level of the list defines the `OR` logic, and the second level defines the `AND` logic.
Additionally each level of these lists does not have to be explicitly defined.

[subject]: #subject

##### Examples

_Matches when the user has the username `john`, **or** the user is in the groups `admin` **and** `app-name`, **or** the
user is in the group `super-admin`. All rules in this list are effectively the same rule just expressed in different
ways._

```yaml {title="configuration.yml"}
access_control:
  rules:
    - domain: '{{< sitevar name="domain" nojs="example.com" >}}'
      policy: 'two_factor'
      subject:
        - 'user:john'
        - ['group:admin', 'group:app-name']
        - 'group:super-admin'
    - domain: '{{< sitevar name="domain" nojs="example.com" >}}'
      policy: 'two_factor'
      subject:
        - ['user:john']
        - ['group:admin', 'group:app-name']
        - ['group:super-admin']
```

_Matches when the user is in the `super-admin` group. All rules in this list are effectively the same rule just
expressed in different ways._

```yaml {title="configuration.yml"}
access_control:
  rules:
    - domain: '{{< sitevar name="domain" nojs="example.com" >}}'
      policy: 'one_factor'
      subject: 'group:super-admin'
    - domain: '{{< sitevar name="domain" nojs="example.com" >}}'
      policy: 'one_factor'
      subject:
        - 'group:super-admin'
    - domain: '{{< sitevar name="domain" nojs="example.com" >}}'
      policy: 'one_factor'
      subject:
        - ['group:super-admin']
```

#### methods

{{< confkey type="list(string)" required="no" >}}

This criteria matches the HTTP request method. This is primarily useful when trying to bypass authentication for specific
request types when those requests would prevent essential or public operation of the website. An example is when you
need to do CORS preflight requests you could apply the `bypass` policy to `OPTIONS` requests.

It's important to note that Authelia cannot preserve request data when redirecting the user. For example if the user had
permission to do GET requests, their authentication level was `one_factor`, and POST requests required them to do
`two_factor` authentication, they would lose the form data. Additionally it is sometimes not possible to redirect users
who have done requests other than HEAD or GET which means the user experience may suffer. These are the reasons it's
only recommended to use this to increase security where essential and for CORS preflight.

The accepted and valid methods for this configuration option are those specified in well known RFCs. The RFCs and the
relevant methods are listed in this table:

|    RFC    |                        Methods                        |                     Additional Documentation                     |
| :-------: | :---------------------------------------------------: | :--------------------------------------------------------------: |
| [RFC7231] | GET, HEAD, POST, PUT, DELETE, CONNECT, OPTIONS, TRACE | [MDN](https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods) |
| [RFC5789] |                         PATCH                         | [MDN](https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods) |
| [RFC4918] | PROPFIND, PROPPATCH, MKCOL, COPY, MOVE, LOCK, UNLOCK  |                                                                  |

[methods]: #methods

##### Examples

_Bypass `OPTIONS` requests to the `{{< sitevar name="domain" nojs="example.com" >}}` domain._

```yaml {title="configuration.yml"}
access_control:
  rules:
    - domain: '{{< sitevar name="domain" nojs="example.com" >}}'
      policy: 'bypass'
      methods:
        - 'OPTIONS'
```

#### networks

{{< confkey type="list(string)" syntax="network" required="no" >}}

These criteria consist of a list of values which can be an IP Address, network address range in CIDR notation, or a named
[Network Definition](../definitions/network.md). It matches against the first address in the `X-Forwarded-For` header,
or if there are none it will fall back to the IP address of the packet TCP source IP address. For this reason, it's
important for you to configure the proxy server correctly to accurately match requests with these criteria.
_**Note:** you may combine CIDR networks with the alias rules as you please._

The main use case for this criteria is adjust the security requirements of a resource based on the location of a user.
You can theoretically consider a specific network to be one of the factors involved in authentication, you can deny
specific networks, etc.

For example if you have an application exposed on both the local networks and the external networks, you are able to
distinguish between those requests and apply differing policies to each. Either denying access when the user is on the
external networks and allowing specific external clients to access it as well as internal clients, or by requiring less
privileges when a user is on the local networks.

There are a large number of scenarios regarding networks and the order of the rules. This provides a lot of flexibility
for administrators to tune the security to their specific needs if desired.

[networks]: #networks

##### Examples

_Require [two_factor](#two_factor) for all clients other than internal clients and `112.134.145.167`. The first two
rules in this list are effectively the same rule just expressed in different ways._

```yaml {title="configuration.yml"}
definitions:
  network:
    internal:
      - '10.0.0.0/8'
      - '172.16.0.0/12'
      - '192.168.0.0/18'
access_control:
  default_policy: 'two_factor'
  rules:
    - domain: 'secure.{{< sitevar name="domain" nojs="example.com" >}}'
      policy: 'one_factor'
      networks:
        - '10.0.0.0/8'
        - '172.16.0.0/12'
        - '192.168.0.0/18'
        - '112.134.145.167/32'
    - domain: 'secure.{{< sitevar name="domain" nojs="example.com" >}}'
      policy: 'one_factor'
      networks:
        - 'internal'
        - '112.134.145.167/32'
    - domain: 'secure.{{< sitevar name="domain" nojs="example.com" >}}'
      policy: 'two_factor'
```

#### resources

{{< confkey type="list(string)" required="no" >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
To utilize regex you must escape it properly. See
[regular expressions](../prologue/common.md#regular-expressions) for more information.
{{< /callout >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
This rule treats the path as case-sensitive as per the HTTP specification. It is important that you validate that the
proxy and the application handle the path case-sensitively too. See the [Case Sensitivity](#case-sensitivity) section
for more information.
{{< /callout >}}

This criteria matches the path and query of the request using regular expressions. The rule is expressed as a list of
strings. If any one of the regular expressions in the list matches the request it's considered a match. A useful tool
for debugging these regular expressions is called [Regex 101](https://regex101.com/) (ensure you pick the `Golang`
option).

Users should familiarize themselves with the various considerations of [Regular Expressions](#regular-expressions)
before configuring this criteria.

In addition to standard regular expression patterns this criteria can match some [Named Regex Groups](#named-regex-groups).

[resources]: #resources

##### Examples

_Applies the [bypass](#bypass) policy when the domain is `app.{{< sitevar name="domain" nojs="example.com" >}}` and the
URL is `/api`, or starts with either `/api/` or `/api?`, we recommend using the `([/?].*)?$` suffix to match the
trailing slash as well as the query string rather than just using `.*` as this can allow unintended paths to match._

```yaml {title="configuration.yml"}
access_control:
  rules:
    - domain: 'app.{{< sitevar name="domain" nojs="example.com" >}}'
      policy: 'bypass'
      resources:
        - '^/api([/?].*)?$'
```

#### query

{{< confkey type="list(list(object))" required="no" >}}

The query criteria is an advanced criteria which can allow configuration of rules that match specific query argument
keys against various rules. It's recommended to use [resources](#resources) rules instead for basic needs.

The format of this rule is unique in as much as it is a list of lists. The logic behind this format is to allow for both
`OR` and `AND` logic. The first level of the list defines the `OR` logic, and the second level defines the `AND` logic.
Additionally each level of these lists does not have to be explicitly defined.

##### key

{{< confkey type="string" required="yes" >}}

The query argument key to check.

##### value

{{< confkey type="string" required="situational" >}}

The value to match against. This is required unless the operator is `absent` or `present`. It's recommended this value
is always quoted as per the examples.

##### operator

{{< confkey type="string" required="situational" >}}

The rule operator for this rule. Valid operators can be found in the
[Rule Operators](../../reference/guides/rule-operators.md#operators) reference guide.

If [key](#key) and [value](#value) are specified this defaults to `equal`, otherwise if [key](#key) is specified it
defaults to `present`.

##### Examples

```yaml {title="configuration.yml"}
access_control:
  rules:
    - domain: 'app.{{< sitevar name="domain" nojs="example.com" >}}'
      policy: 'bypass'
      query:
        - - operator: 'present'
            key: 'secure'
          - operator: 'absent'
            key: 'insecure'
        - - operator: 'pattern'
            key: 'token'
            value: '^(abc123|zyx789)$'
          - operator: 'not pattern'
            key: 'random'
            value: '^(1|2)$'
```

## Policies

The policy of the first matching rule in the configured list decides the policy applied to the request, if no rule
matches the request the [default_policy](#default_policy) is applied.

[policies]: #policies

### deny

This is the policy applied by default, and is what we recommend as the default policy for all installs. Its effect
is literally to deny the user access to the resource. Additionally you can use this policy to conditionally deny
access in desired situations. Examples include denying access to an API that has no authentication mechanism built in.

[deny]: #deny

### bypass

This policy skips all authentication and allows anyone to use the resource. This policy is not available with a rule
that includes a [subject] restriction because the minimum authentication level required to obtain information
about the subject is [one_factor]. See [Rule Matching Concept 2] for more information.

[bypass]: #bypass

### one_factor

This policy requires the user at minimum complete 1FA successfully (username and password). This means if they have
performed 2FA then they will be allowed to access the resource.

[one_factor]: #one_factor

### two_factor

This policy requires the user to complete 2FA successfully. This is currently the highest level of authentication
policy available.

[two_factor]: #two_factor

## Rule Matching

There are two important concepts to understand when it comes to rule matching. This section covers these concepts.

You can easily evaluate if your access control rules section matches a given request, and why it doesn't match using the
[authelia access-control check-policy](../../reference/cli/authelia/authelia_access-control_check-policy.md) command.

### Rule Matching Concept 1: Sequential Order

Rules are matched in sequential order. The first entry in the list where all criteria match is the rule which is applied.
Some rule criteria additionally allow for a list of criteria, when one of these criteria in the list match a request that
criteria is considered a match for that specific rule.

This is particularly **important** for bypass rules. Bypass rules should generally appear near the top of the rules
list. However you need to carefully evaluate your rule list **in order** to see which rule matches a particular
scenario. A comprehensive understanding of how rules apply is also recommended.

For example the following rule will consider requests for either `{{< sitevar name="domain" nojs="example.com" >}}` or any subdomain of
`{{< sitevar name="domain" nojs="example.com" >}}` a match if they have a path of exactly `/api` or if they start with `/api/`. This means that
the second rule for `app.{{< sitevar name="domain" nojs="example.com" >}}` will not be considered if the request is to
`https://app.{{< sitevar name="domain" nojs="example.com" >}}/api` because the first rule is a match for that request.

```yaml {title="configuration.yml"}
- domain:
    - '{{< sitevar name="domain" nojs="example.com" >}}'
    - '*.{{< sitevar name="domain" nojs="example.com" >}}'
  policy: 'bypass'
  resources:
    - '^/api$'
    - '^/api/'
- domain:
    - 'app.{{< sitevar name="domain" nojs="example.com" >}}'
  policy: 'two_factor'
```

[Rule Matching Concept 1]: #rule-matching-concept-1-sequential-order

### Rule Matching Concept 2: Subject Criteria Requires Authentication

Rules that have subject reliant elements require authentication to determine if they match. Due to this these rules
must not be used with the [bypass] policy. The criteria which have subject reliant elements are:

- The [subject] criteria itself
- The [domain_regex] criteria when it contains the [Named Regex Groups].

In addition if the rule has a subject criteria but all other criteria match then the user will be immediately forwarded
for authentication if no prior rules match the request per [Rule Matching Concept 1]. This means if you have two
identical rules, and one of them has a subject based reliant criteria, and the other one is a [bypass] rule then the
[bypass] rule should generally come first.

[Rule Matching Concept 2]: #rule-matching-concept-2-subject-criteria-requires-authentication

## Regular Expressions

There are several important concepts to understand when it comes to regular expressions.

### Test Your Rules

It's important that you test your rules thoroughly before deploying them to production. It's recommended to use
[Regex 101](https://regex101.com/) to test your regular expressions, and to get advice about the specific way the rule
could be abused prior to using it.

### Escaping Special Characters

Regular Expressions use the `\` character to escape special characters or use special meta characters. It's important
that several characters are escaped when used to match literal strings. The most common special characters users should
make sure to escape are the period `.`, the asterisk `*`, the addition `+`, and the question mark `?`.

### Case Sensitivity

By default Authelia uses case-insensitive regular expressions in appropriate locations. This is different from the
default behavior of the Go standard library which is case-sensitive. This is achieved by prepending a `(?i)` to the
regular expression before parsing it. You can negate this behavior by adding `(?-i)` to the beginning of any regular
expression which explicitly disables case insensitivity.

The main area affected by this is currently:

- The [domain_regex](#domain_regex) criteria

The rationale behind this is that it hardens the security of the system by preventing users from incorrectly forming the
regular expression to match against.

We do not use case-insensitive regular expressions in the [resources](#resources) intentionally. This decision was
reviewed in 2026 as part of a defense-in-depth analysis and we believe it was the correct decision.

It should be noted that [RFC9110](https://www.rfc-editor.org/rfc/rfc9110.html#section-4.2.3) clearly states that only
the scheme and host components are case-insensitive and that every other component is case-sensitive. This is also
reflected in the [RFC3986](https://datatracker.ietf.org/doc/html/rfc3986#section-6.2.2.1) which mirrors this sentiment
more softly but still explicitly states that the path component is case-sensitive.

All regular expressions in Authelia can be modified to be case-insensitive by adding the `(?i)` to the beginning of the
regular expression and can be modified to be case-sensitive by adding the `(?-i)` to the beginning of the regular
expression. You can also adjust certain parts of the regular expression to match case-insensitively or case-sensitively
depending on your needs. Below is a table of examples of how to modify the case-sensitivity of a regular expression.

|       Description        |       Case-Sensitive       |     Case-Insensitive      |
| :----------------------: | :------------------------: | :-----------------------: |
| Adjust the whole pattern |   `(?-i)^/api([/?].*)?$`   |   `(?i)^/api([/?].*)?$`   |
| Adjust a single segment  |  `^/(?-i:api)([/?].*)?$`   |  `^/(?i:api)([/?].*)?$`   |
| Adjust multiple segments | `^/(?-i:api/v1)([/?].*)?$` | `^/(?i:api/v1)([/?].*)?$` |

#### Case Validation

While this decision is backed by the defense-in-depth analysis and the appropriate specifications, it is critically
important to note that a proxy or application may not respect the fact the path component is case-sensitive. As such
it's important to ensure that you test this and adjust the regular expressions accordingly.

For example if you have configured the `^/api([/?].*)?$` resources regular expression then you should validate that the
backend does not return a valid when you request a resource with a capitalized path i.e. `/API/example`.

Example configuration for testing (not to be used in production):

```yaml
access_control:
  rules:
    - domain: 'app.example.com'
      policy: 'one_factor'
      resources:
        - '^/api([/?].*)?$'
    - domain: 'app.example.com'
      policy: 'bypass'
```

Perform the check with credentials and the valid path:

```bash
curl -s -i -u "username:password" https://app.example.com/api/example
```

Output example below. Note the 200 status code indicating a successful response, and the body after the other headers.

```
HTTP/2 200
content-type: application/json; charset=utf-8
content-length: 15

{"status":"OK"}
```

Perform the check without credentials and the valid path:

```bash
curl -s -i https://app.example.com/api/example
```

Output example below. Note the 403 status code indicating a failed response. This may also be a 30x redirect depending
on the configuration.

```
HTTP/2 403
content-type: text/plain; charset=utf-8
content-length: 13

403 Forbidden
```

Perform the check without credentials and the invalid path:

```bash
curl -s -i https://app.example.com/API/example
```

At this point you should see a 404 in most situations. If you see a 200, validate the body of the response does not
match the response with the username and password above. If you see the same response then you need to adjust the
regular expression to match the path case-insensitively.

### Named Regex Groups

Some criteria allow matching named regex groups. These are the groups we accept:

| Group Name | Match Value | Match Type  |
| :--------: | :---------: | :---------: |
|    User    |  username   |   Equals    |
|   Group    |   groups    | Has (Equal) |

Named regex groups are represented with the syntax `(?P<User>\w+)` where `User` is the group name from the table above,
and `\w+` is the pattern for the area of the pattern that should be compared to the match value.

The match type `Equals` matches if the value extracted from the pattern is equal to the match value. The match type
`Has (Equal)` matches if the value extracted from the pattern is equal to one of the values in the match value (the
match value is a list/slice).

The regex groups are case-insensitive due to the fact that the regex groups are used in domain criteria and domain names
should not be compared in a case-sensitive way as per the [RFC4343](https://datatracker.ietf.org/doc/html/rfc4343)
abstract and [RFC3986 Section 3.2.2](https://datatracker.ietf.org/doc/html/rfc3986#section-3.2.2).

We do not currently apply any other normalization to usernames or groups when matching these groups. As such it's
generally _**not recommended**_ to use these patterns with usernames or groups which contain characters that are not
alphanumeric (including spaces).

[Named Regex Groups]: #named-regex-groups

## Detailed example

Here is a detailed example of an example access control section:

```yaml {title="configuration.yml"}
definitions:
  network:
    internal:
      - '10.10.0.0/16'
      - '192.168.2.0/24'
    vpn: '10.9.0.0/16'

access_control:
  default_policy: 'deny'
  rules:
    - domain: 'public.{{< sitevar name="domain" nojs="example.com" >}}'
      policy: 'bypass'

    - domain: '*.{{< sitevar name="domain" nojs="example.com" >}}'
      policy: 'bypass'
      methods:
        - 'OPTIONS'

    - domain: 'secure.{{< sitevar name="domain" nojs="example.com" >}}'
      policy: 'one_factor'
      networks:
        - 'internal'
        - 'vpn'
        - '192.168.1.0/24'
        - '10.0.0.1'

    - domain:
      - 'secure.{{< sitevar name="domain" nojs="example.com" >}}'
      - 'private.{{< sitevar name="domain" nojs="example.com" >}}'
      policy: 'two_factor'

    - domain: 'singlefactor.{{< sitevar name="domain" nojs="example.com" >}}'
      policy: 'one_factor'

    - domain: 'mx2.mail.{{< sitevar name="domain" nojs="example.com" >}}'
      subject: 'group:admins'
      policy: 'deny'

    - domain: '*.{{< sitevar name="domain" nojs="example.com" >}}'
      subject:
        - 'group:admins'
        - 'group:moderators'
      policy: 'two_factor'

    - domain: 'dev.{{< sitevar name="domain" nojs="example.com" >}}'
      resources:
      - '^/groups/dev([/?].*)?$'
      subject: 'group:dev'
      policy: 'two_factor'

    - domain: 'dev.{{< sitevar name="domain" nojs="example.com" >}}'
      resources:
      - '^/users/john([/?].*)?$'
      subject:
      - ['group:dev', 'user:john']
      - 'group:admins'
      policy: 'two_factor'
```

[RFC7231]: https://datatracker.ietf.org/doc/html/rfc7231
[RFC5789]: https://datatracker.ietf.org/doc/html/rfc5789
[RFC4918]: https://datatracker.ietf.org/doc/html/rfc4918
