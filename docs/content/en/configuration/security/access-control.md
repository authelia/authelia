---
title: "Access Control"
description: "Configuring the Access Control or RBAC settings."
lead: "Authelia supports a comprehensive access control system. This section describes configuring this."
date: 2020-02-29T01:43:59+01:00
draft: false
images: []
menu:
  configuration:
    parent: "security"
weight: 104200
toc: true
aliases:
  - /c/acl
  - /docs/configuration/access-control.html
---

## Configuration

```yaml
access_control:
  default_policy: deny
  networks:
  - name: internal
    networks:
    - '10.0.0.0/8'
    - '172.16.0.0/12'
    - '192.168.0.0/18'
  rules:
  - domain: 'private.example.com'
    domain_regex: '^(\d+\-)?priv-img.example.com$'
    policy: one_factor
    networks:
    - 'internal'
    - '1.1.1.1'
    subject:
    - ['user:adam']
    - ['user:fred']
    - ['group:admins']
    methods:
    - GET
    - HEAD
    resources:
    - '^/api.*'
```

## Options

### default_policy

{{< confkey type="string" default="deny" required="no" >}}

The default [policy](#policies) defines the policy applied if no [rules](#rules) section apply to the information known
about the request. It is recommended that this is configured to [deny](#deny) for security reasons. Sites which you do
not wish to secure at all with Authelia should not be configured in your reverse proxy to perform authentication with
Authelia at all for performance reasons.

See [Policies](#policies) for more information.

### networks (global)

{{< confkey type="list" required="no" >}}

The main/global networks section contains a list of networks with a name label that can be reused in the
[rules](#networks) section instead of redefining the same networks over and over again. This additionally makes
complicated network related configuration a lot cleaner and easier to read.

This section has two options, `name` and `networks`. Where the `networks` section is a list of IP addresses in CIDR
notation and where `name` is a friendly name to label the collection of networks for reuse in the [networks](#networks)
section of the [rules](#rules) section below.

This configuration option *does nothing* by itself, it's only useful if you use these aliases in the [rules](#networks)
section below.

### rules

{{< confkey type="list" required="no" >}}

The rules have many configuration options. A rule matches when all criteria of the rule match the request excluding the
`policy` which is the [policy](#policies) applied to the request.

A rule defines two primary things:

* the policy applied when all criteria match
* the matching criteria of the request presented to the reverse proxy

The criteria is broken into several parts:

* [domain](#domain): domain or list of domains targeted by the request.
* [domain_regex](#domain_regex): regex form of [domain](#domain).
* [resources](#resources): pattern or list of patterns that the path should match.
* [subject](#subject): the user or group of users to define the policy for.
* [networks](#networks): the network addresses, ranges (CIDR notation) or groups from where the request originates.
* [methods](#methods): the http methods used in the request.

A rule is matched when all criteria of the rule match. Rules are evaluated in sequential order, and the first rule that
is a match for a given request is the rule applied; subsequent rules have *no effect*. This is particularly
__important__ for bypass rules. Bypass rules should generally appear near the top of the rules list. However you need to
carefully evaluate your rule list __in order__ to see which rule matches a particular scenario. A comprehensive
understanding of how rules apply is also recommended.

#### domain

{{< confkey type="list(string)" required="yes" >}}

*__Required:__ This criteria and/or the [domain_regex](#domain_regex) criteria are required.*

This criteria matches the domain name and has two methods of configuration, either as a single string or as a list of
strings. When it's a list of strings the rule matches when __any__ of the domains in the list match the request domain.
When used in conjunction with [domain_regex](#domain_regex) the rule will match when either the [domain](#domain) or the
[domain_regex](#domain_regex) criteria matches.

Rules may start with a few different wildcards:

* The standard wildcard is `*.`, which when in front of a domain means that any subdomain is effectively a match. For
  example `*.example.com` would match `abc.example.com` and `secure.example.com`. When using a wildcard like this the
  string __must__ be quoted like `"*.example.com"`.
* The user wildcard is `{user}.`, which when in front of a domain dynamically matches the username of the user. For
  example `{user}.example.com` would match `fred.example.com` if the user logged in was named `fred`. *__Warning:__ this is
  officially deprecated as the [domain_regex](#domain_regex) criteria completely replaces the functionality in a much
  more useful way. It is strongly recommended you do not use this as it will be removed in a future version, most likely
  v5.0.0.*
* The group wildcard is `{group}.`, which when in front of a domain dynamically matches if the logged in user has the
  group in that location. For example `{group}.example.com` would match `admins.example.com` if the user logged in was
  in the following groups `admins,users,people` because `admins` is in the list.

Domains in this section must be the domain configured in the [session](../session/introduction.md#domain) configuration
or subdomains of that domain. This is because a website can only write cookies for a domain it is part of. It is
theoretically possible for us to do this with multiple domains however we would have to be security conscious in our
implementation, and it is not currently a priority.

##### Examples

*Single domain of `*.example.com` matched. All rules in this list are effectively the same rule just expressed in
different ways.*

```yaml
access_control:
  rules:
  - domain: '*.example.com'
    policy: bypass
  - domain:
    - '*.example.com'
    policy: bypass
```

*Multiple domains matched. These rules will match either `apple.example.com` or `orange.example.com`. All rules in this
list are effectively the same rule just expressed in different ways.*

```yaml
access_control:
  rules:
  - domain: ['apple.example.com', 'banana.example.com']
    policy: bypass
  - domain:
    - apple.example.com
    - banana.example.com
    policy: bypass
```

*Multiple domains matched either via a static domain or via a [domain_regex](#domain_regex). This rule will match
either `apple.example.com`, `pub-data.example.com`, or `img-data.example.com`.*

```yaml
access_control:
  rules:
  - domain: 'apple.example.com'
    domain_regex: '^(pub|img)-data\.example\.com$'
    policy: bypass
```

### domain_regex

{{< confkey type="list(string)" required="yes" >}}

*__Required:__ This criteria and/or the [domain](#domain) criteria are required.*

*__Important Note:__ If you intend to use this criteria with a bypass rule please read
[bypass and subjects](#bypass-and-user-identity) before doing so.*

*__Important Note:__ to utilize regex you must escape it properly. See
[regular expressions](../prologue/common.md#regular-expressions) for more information.*

This criteria matches the domain name and has two methods of configuration, either as a single string or as a list of
strings. When it's a list of strings the rule matches when __any__ of the domains in the list match the request domain.
When used in conjunction with [domain](#domain) the rule will match when either the [domain](#domain) or the
[domain_regex](#domain_regex) criteria matches.

In addition to standard regex patterns this criteria can match some [Named Regex Groups](#named-regex-groups).

#### Examples

*An advanced multiple domain regex example with user/group matching. This will match the user `john` in the groups
`example` and `example1`, when the request is made to `user-john.example.com`, `group-example.example.com`, or
`group-example1.example.com`, it would not match when the request is made to `user-fred.example.com` or
`group-admin.example.com`.*

```yaml
access_control:
  rules:
  - domain_regex:
    - '^user-(?P<User>\w+)\.example\.com$'
    - '^group-(?P<Group>\w+)\.example\.com$'
    policy: one_factor
```

*Multiple domains example, one with a static domain and one with a regex domain. This will match requests to
`protected.example.com`, `img-private.example.com`, or `data-private.example.com`.*

```yaml
access_control:
  rules:
  - domain: 'protected.example.com'
  - domain_regex: '^(img|data)-private\.example\.com'
    policy: one_factor
```

#### policy

{{< confkey type="string" required="yes" >}}

The specific [policy](#policies) to apply to the selected rule. This is not criteria for a match, this is the action to
take when a match is made.

#### subject

{{< confkey type="list(list(string))" required="no" >}}

*__Note:__ this rule criteria __may not__ be used for the `bypass` policy the minimum required authentication level to
identify the subject is `one_factor`. We have taken an opinionated stance on preventing this configuration as it could
result in problematic security scenarios with badly thought out configurations and cannot see a likely configuration
scenario that would require users to do this. If you have a scenario in mind please open an
[issue](https://github.com/authelia/authelia/issues/new) on GitHub.*

This criteria matches identifying characteristics about the subject. Currently this is either user or groups the user
belongs to. This allows you to effectively control exactly what each user is authorized to access or to specifically
require two-factor authentication to specific users. Subjects are prefixed with either `user:` or `group:` to identify
which part of the identity to check.

The format of this rule is unique in as much as it is a list of lists. The logic behind this format is to allow for both
`OR` and `AND` logic. The first level of the list defines the `OR` logic, and the second level defines the `AND` logic.
Additionally each level of these lists does not have to be explicitly defined.

##### Examples

*Matches when the user has the username `john`, __or__ the user is in the groups `admin` __and__ `app-name`, __or__ the
user is in the group `super-admin`. All rules in this list are effectively the same rule just expressed in different
ways.*

```yaml
access_control:
  rules:
  - domain: example.com
    policy: two_factor
    subject:
    - 'user:john'
    - ['group:admin', 'group:app-name']
    - 'group:super-admin'
  - domain: example.com
    policy: two_factor
    subject:
    - ['user:john']
    - ['group:admin', 'group:app-name']
    - ['group:super-admin']
```

*Matches when the user is in the `super-admin` group. All rules in this list are effectively the same rule just
expressed in different ways.*

```yaml
access_control:
  rules:
  - domain: example.com
    policy: one_factor
    subject: 'group:super-admin'
  - domain: example.com
    policy: one_factor
    subject:
    - 'group:super-admin'
  - domain: example.com
    policy: one_factor
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

The accepted and valid methods for this configuration option are those specified in well known RFC's. The RFC's and the
relevant methods are listed in this table:

|    RFC    |                        Methods                        |                     Additional Documentation                     |
|:---------:|:-----------------------------------------------------:|:----------------------------------------------------------------:|
| [RFC7231] | GET, HEAD, POST, PUT, DELETE, CONNECT, OPTIONS, TRACE | [MDN](https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods) |
| [RFC5789] |                         PATCH                         | [MDN](https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods) |
| [RFC4918] | PROPFIND, PROPPATCH, MKCOL, COPY, MOVE, LOCK, UNLOCK  |                                                                  |

##### Examples

*Bypass `OPTIONS` requests to the `example.com` domain.*

```yaml
access_control:
  rules:
  - domain: example.com
    policy: bypass
    methods:
    - OPTIONS
```

#### networks

{{< confkey type="list(string)" required="no" >}}

This criteria is a list of values which can be an IP Address, network address range in CIDR notation, or an alias from
the [global](#networks-global) section. It matches against the first address in the `X-Forwarded-For` header, or if there
are none it will fall back to the IP address of the packet TCP source IP address. For this reason it's important for you
to configure the proxy server correctly in order to accurately match requests with this criteria. *__Note:__ you may
combine CIDR networks with the alias rules as you please.*

The main use case for this criteria is adjust the security requirements of a resource based on the location of a user.
You can theoretically consider a specific network to be one of the factors involved in authentiation, you can deny
specific networks, etc.

For example if you have an application exposed on both the local networks and the external networks, you are able to
distinguish between those requests and apply differing policies to each. Either denying access when the user is on the
external networks and allowing specific external clients to access it as well as internal clients, or by requiring less
privileges when a user is on the local networks.

There are a large number of scenarios regarding networks and the order of the rules. This provides a lot of flexibility
for administrators to tune the security to their specific needs if desired.

##### Examples

*Require [two_factor](#two_factor) for all clients other than internal clients and `112.134.145.167`. The first two
rules in this list are effectively the same rule just expressed in different ways.*

```yaml
access_control:
  default_policy: two_factor
  networks:
  - name: internal
    networks:
      - '10.0.0.0/8'
      - '172.16.0.0/12'
      - '192.168.0.0/18'
  rules:
  - domain: secure.example.com
    policy: one_factor
    networks:
    - '10.0.0.0/8'
    - '172.16.0.0/12'
    - '192.168.0.0/18'
    - '112.134.145.167/32'
  - domain: secure.example.com
    policy: one_factor
    networks:
    - 'internal'
    - '112.134.145.167/32'
  - domain: secure.example.com
    policy: two_factor
```

#### resources

{{< confkey type="list(string)" required="no" >}}

*__Important Note:__ to utilize regex you must escape it properly. See
[regular expressions](../prologue/common.md#regular-expressions) for more information.*

This criteria matches the path and query of the request using regular expressions. The rule is expressed as a list of
strings. If any one of the regular expressions in the list matches the request it's considered a match. A useful tool
for debugging these regular expressions is called [Regex 101](https://regex101.com/) (ensure you pick the `Golang`
option).

In addition to standard regex patterns this criteria can match some [Named Regex Groups](#named-regex-groups).

*__Note:__ Prior to 4.27.0 the regular expressions only matched the path excluding the query parameters. After 4.27.0
they match the entire path including the query parameters. When upgrading you may be required to alter some of your
resource rules to get them to operate as they previously did.*

It's important when configuring resource rules that you enclose them in quotes otherwise you may run into some issues
with escaping the expressions. Failure to do so may prevent Authelia from starting. It's technically optional but will
likely save you a lot of time if you do it for all resource rules.

##### Examples

*Applies the [bypass](#bypass) policy when the domain is `app.example.com` and the url is `/api`, or starts with either
`/api/` or `/api?`.*

```yaml
access_control:
  rules:
  - domain: app.example.com
    policy: bypass
    resources:
    - '^/api([/?].*)?$'
```

## Policies

The policy of the first matching rule in the configured list decides the policy applied to the request, if no rule
matches the request the [default_policy](#default_policy) is applied.

### deny

This is the policy applied by default, and is what we recommend as the default policy for all installs. Its effect
is literally to deny the user access to the resource. Additionally you can use this policy to conditionally deny
access in desired situations. Examples include denying access to an API that has no authentication mechanism built in.

### bypass

This policy skips all authentication and allows anyone to use the resource. This policy is not available with a rule
that includes a [subject](#subject) restriction because the minimum authentication level required to obtain information
about the subject is [one_factor](#one_factor).

#### bypass and user identity

The [bypass](#bypass) policy cannot be used when the rule uses a criteria that requires we know the users identity. This
means:

* If the rule defines [subjects](#subject) criteria
* If the rule defines [domain regex](#domain_regex) criteria which contains either the user or group named match groups

This is because these criteria types require knowing who the user is in order to determine if their identity matches the
request. This information can only be known after 1FA, which means the minimum policy that can be used logically is
[one_factor](#one_factor).

### one_factor

This policy requires the user at minimum complete 1FA successfully (username and password). This means if they have
performed 2FA then they will be allowed to access the resource.

### two_factor

This policy requires the user to complete 2FA successfully. This is currently the highest level of authentication
policy available.

## Named Regex Groups

Some criteria allow matching named regex groups. These are the groups we accept:

| Group Name |    Match Value    |
|:----------:|:-----------------:|
|    User    |     username      |
|   Group    | groups (contains) |

For the group name `Group` the regex pattern matches if the user has the specific group name matching the pattern. Both
regex groups are case-insensitive due to the fact that the regex groups are used in domain criteria and domain names
should not be compared in a case-sensitive way as per the [RFC4343](https://www.rfc-editor.org/rfc/rfc4343.html)
abstract and [RFC3986 Section 3.2.2](https://www.rfc-editor.org/rfc/rfc3986#section-3.2.2).

## Detailed example

Here is a detailed example of an example access control section:

```yaml
access_control:
  default_policy: deny
  networks:
    - name: internal
      networks:
        - '10.10.0.0/16'
        - '192.168.2.0/24'
    - name: VPN
      networks: 10.9.0.0/16
  rules:
    - domain: 'public.example.com'
      policy: bypass

    - domain: '*.example.com'
      policy: bypass
      methods:
        - OPTIONS

    - domain: 'secure.example.com'
      policy: one_factor
      networks:
        - 'internal'
        - 'VPN'
        - '192.168.1.0/24'
        - '10.0.0.1'

    - domain:
      - 'secure.example.com'
      - 'private.example.com'
      policy: two_factor

    - domain: 'singlefactor.example.com'
      policy: one_factor

    - domain: 'mx2.mail.example.com'
      subject: 'group:admins'
      policy: deny

    - domain: '*.example.com'
      subject:
        - 'group:admins'
        - 'group:moderators'
      policy: two_factor

    - domain: dev.example.com
      resources:
      - '^/groups/dev/.*$'
      subject: 'group:dev'
      policy: two_factor

    - domain: dev.example.com
      resources:
      - '^/users/john/.*$'
      subject:
      - ['group:dev', 'user:john']
      - 'group:admins'
      policy: two_factor

    - domain: '{user}.example.com'
      policy: bypass
```

[RFC7231]: https://www.rfc-editor.org/rfc/rfc7231.html
[RFC5789]: https://www.rfc-editor.org/rfc/rfc5789.html
[RFC4918]: https://www.rfc-editor.org/rfc/rfc4918.html
