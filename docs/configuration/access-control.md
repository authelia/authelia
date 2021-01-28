---
layout: default
title: Access Control
parent: Configuration
nav_order: 1
---

# Access Control
{: .no_toc }

## Policies

With **Authelia** you can define a list of rules that are going to be evaluated in
sequential order when authorization is delegated to Authelia.

The first matching rule of the list defines the policy applied to the resource, if
no rule matches the resource a customizable default policy is applied.

### deny

This is the policy applied by default, and is what we recommend as the default policy for all installs. Its effect
is literally to deny the user access to the resource. Additionally you can use this policy to conditionally deny 
access in desired situations. Examples include denying access to an API that has no authentication mechanism built in.

### bypass

This policy skips all authentication and allows anyone to use the resource. This policy has little to no effect
when used with a rule that also has a subject defined, just because the minimum authentication level required
to obtain information about the subject is [one_factor](#one_factor)

### one_factor

This policy requires the user at minimum complete 1FA successfully (username and password). This means if they have done 2FA then
they will be allowed to access the resource.

### two_factor

This policy requires the user complete 2FA successfully. This is currently the highest level of authentication
policy available.

## Default Policy

The default policy is the policy applied when no other rule matches. It is recommended that this is configured to 
[deny](#deny) for security reasons. Sites which you do not wish to secure with Authelia should not be configured to do 
authentication with Authelia at all.

See [Policies](#policies) for more information.

## Network Aliases

The main networks section defines a list of network aliases, where the name matches a list of networks. These names can
be used in any [rule](#rules) instead of a literal network. This makes it easier to define a group of networks multiple
times.

You can combine both literal networks and these aliases inside the [networks](#networks) section of a rule. See this
section for more details.

## Rules

A rule defines two things:

* the matching criteria of the request presented to the reverse proxy
* the policy applied when all criteria match.

The criteria are:

* domain: domain or list of domains targeted by the request.
* resources: pattern or list of patterns that the path should match.
* subject: the user or group of users to define the policy for.
* networks: the network addresses, ranges (CIDR notation) or groups from where the request originates.
* methods: the http methods used in the request.

A rule is matched when all criteria of the rule match. Rules are evaluated in sequential order, and this is
particularly **important** for bypass rules. Bypass rules should generally 


### Policy

A policy represents the level of authentication the user needs to pass before
being authorized to request the resource.

See [Policies](#policies) for more information.

### Domains

The domains defined in rules must obviously be either a subdomain of the domain
protected by Authelia or the protected domain itself. In order to match multiple
subdomains, the wildcard matcher character `*.` can be used as prefix of the domain.
For instance, to define a rule for all subdomains of *example.com*, one would use
`*.example.com` in the rule. A single rule can define multiple domains for matching.
These domains can be either listed in YAML-short form `["example1.com", "example2.com"]`
or in YAML long-form as dashed list.

### Resources

A rule can define multiple regular expressions for matching the path of the resource
similar to the list of domains. If any one of them matches, the resource criteria of
the rule matches.

Note that regular expressions can be used to match a given path. However, they do not match
the query parameters in the URL, only the path.

You might also face some escaping issues preventing Authelia to start. Please make sure that
when you are using regular expressions, you enclose them between quotes. It's optional but
it will likely save you a lot of debugging time.


### Subjects

A subject is a representation of a user or a group of user for who the rule should apply.

For a user with unique identifier `john`, the subject should be `user:john` and for a group
uniquely identified by `developers`, the subject should be `group:developers`. Similar to resources
and domains you can define multiple subjects in a single rule.

If you want a combination of subjects to be matched at once using a logical `AND`, you can
specify a nested list of subjects like `- ["group:developers", "group:admins"]`.
In summary, the first list level of subjects are evaluated using a logical `OR`, whereas the
second level by a logical `AND`. The last example below reads as: the group is `dev` AND the
username is `john` OR the group is `admins`.

### Networks

A list of network addresses, ranges (CIDR notation) or groups can be specified in a rule in order to apply different
policies when requests originate from different networks. This list can contain both literal definitions of networks
and [network aliases](#network-aliases).

Main use cases for this rule option is to adjust the security requirements of a resource based on the location of
the user. For example lets say a resource should be exposed both on the Internet and from an
authenticated VPN for instance. Passing a second factor a first time to get access to the VPN and
a second time to get access to the application can sometimes be cumbersome if the endpoint is not
considered overly sensitive.

An additional situation where this may be useful is if there is a specific network you wish to deny access
or require a higher level of authentication for (like a public machine network vs a company device network, or a 
BYOD network). 

Even if Authelia provides this flexibility, you might prefer a higher level of security and avoid
this option entirely. You and only you can define your security policy and it's up to you to
configure Authelia accordingly.

### Methods

A list of HTTP request methods to apply the rule to. Valid values are GET, HEAD, POST, PUT, DELETE, 
CONNECT, OPTIONS, and TRACE. Additional information about HTTP request methods can be found on the 
[MDN](https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods).

It's important to note this policy type is primarily intended for use when you wish to bypass authentication for
a specific request method. This is because there are several key limitations in what is possible to accomplish
without Authelia being a reverse proxy server. This rule type is discouraged unless you really know what you're
doing or you wish to setup a rule to bypass CORS preflight requests by bypassing for the OPTIONS method.

For example, if you require authentication only for write events (POST, PATCH, DELETE, PUT), when a user who is not
currently authenticated tries to do one of these actions, they will be redirected to Authelia. Authelia will decide
what level is required for authentication, and then after the user authenticates it will redirect them to the original
URL where Authelia decided they needed to authenticate. So if the endpoint they are redirected to originally had
data sent as part of the request, this data is completely lost. Further if the endpoint expects the data or doesn't allow
GET request types, the user may be presented with an error leading to a bad user experience.

## Complete example

Here is a complete example of complex access control list that can be defined in Authelia.

```yaml
access_control:
  default_policy: deny
  networks:
    - name: internal
      networks:
        - 10.10.0.0/16
        - 192.168.2.0/24
    - name: VPN
      networks: 10.9.0.0/16
  rules:
    - domain: public.example.com
      policy: bypass

    - domain: "*.example.com"
      policy: bypass
      methods:
        - OPTIONS

    - domain: secure.example.com
      policy: one_factor
      networks:
        - internal
        - VPN
        - 192.168.1.0/24
        - 10.0.0.1

    - domain:
      - secure.example.com
      - private.example.com
      policy: two_factor

    - domain: singlefactor.example.com
      policy: one_factor

    - domain: "mx2.mail.example.com"
      subject: "group:admins"
      policy: deny

    - domain: "*.example.com"
      subject:
        - "group:admins"
        - "group:moderators"
      policy: two_factor

    - domain: dev.example.com
      resources:
      - "^/groups/dev/.*$"
      subject: "group:dev"
      policy: two_factor

    - domain: dev.example.com
      resources:
      - "^/users/john/.*$"
      subject: 
      - ["group:dev", "user:john"]
      - "group:admins"
      policy: two_factor
```
