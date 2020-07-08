---
layout: default
title: Access Control
parent: Configuration
nav_order: 1
---

# Access Control
{: .no_toc }

## Access Control List

With **Authelia** you can define a list of rules that are going to be evaluated in
sequential order when authorization is delegated to Authelia.

The first matching rule of the list defines the policy applied to the resource, if
no rule matches the resource a customizable default policy is applied.


## Access Control Rule

A rule defines two things:

* the matching criteria of the request presented to the reverse proxy
* the policy applied when all criteria match.

The criteria are:

* domain: domain targeted by the request.
* resources: list of patterns that the path should match (one is sufficient).
* subject: the user or group of users to define the policy for.
* networks: the network range from where should comes the request.

A rule is matched when all criteria of the rule match.


## Policies

A policy represents the level of authentication the user needs to pass before
being authorized to request the resource.

There exist 4 policies:

* bypass: the resource is public as the user does not need any authentication to
get access to it.
* one_factor: the user needs to pass at least the first factor to get access to
the resource.
* two_factor: the user needs to pass two factors to get access to the resource.
* deny: the user does not have access to the resource.

## Domains

The domains defined in rules must obviously be either a subdomain of the domain
protected by Authelia or the protected domain itself. In order to match multiple
subdomains, the wildcard matcher character `*.` can be used as prefix of the domain.
For instance, to define a rule for all subdomains of *example.com*, one would use
`*.example.com` in the rule. A single rule can define multiple domains for matching.

## Resources

A rule can define multiple regular expressions for matching the path of the resource. If
any one of them matches, the resource criteria of the rule matches.


## Subjects

A subject is a representation of a user or a group of user for who the rule should apply.

For a user with unique identifier `john`, the subject should be `user:john` and for a group
uniquely identified by `developers`, the subject should be `group:developers`. Similar to resources
and domains you can define multiple subjects in a single rule.

If you want a combination of subjects to be matched at once, you can specify a list of subjects like
`- ["group:developers", "group:admins"]`. Make sure to preceed it by a list key `-`.
In summary, the first level of subjects are evaluated using a logical `OR`, whereas the second level 
by a logical `AND`.

## Networks

A list of network ranges can be specified in a rule in order to apply different policies when
requests come from different networks.

The main use case is when, lets say a resource should be exposed both on the Internet and from an
authenticated VPN for instance. Passing a second factor a first time to get access to the VPN and
a second time to get access to the application can sometimes be cumbersome if the endpoint is not
considered overly sensitive.

Even if Authelia provides this flexibility, you might prefer a higher level of security and avoid
this option entirely. You and only you can define your security policy and it's up to you to
configure Authelia accordingly.


## Complete example

Here is a complete example of complex access control list that can be defined in Authelia.

```yaml
access_control:
  default_policy: deny
  rules:
    - domain: public.example.com
      policy: bypass

    - domain: secure.example.com
      policy: one_factor
      networks:
      - 192.168.1.0/24

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
      policy: two_factor
```
