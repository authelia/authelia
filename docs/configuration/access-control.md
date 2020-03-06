---
layout: default
title: Access Control
parent: Configuration
nav_order: 2
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

* the policy to apply to the rule
* the criteria that must match the request presented to the reverse proxy

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


## Criteria

The criteria are:

* domain: domain targeted by the request.
* resources: list of patterns that the path should match (one is sufficient).
* subject: the user or group of users to define the policy for.
* networks: the network range from where should comes the request.

A rule is matched when all criteria of the rule match.


## Domain

The domain defined in rules must be either a subdomain of the domain
protected by Authelia or the protected domain itself. In order to match multiple
subdomains, the wildcard matcher character `*.` can be used as prefix of the domain.
For instance, to define a rule for all subdomains of *example.com*, one would use
`*.example.com` in the rule.


### Methods

A rule can define multiple methods used for matching a request. If any one of them matches,
the methodsThis is useful for 
configuring bypass rules for situations like in this 
[issue](https://github.com/authelia/authelia/issues/648). Available methods follow 
the methods documented by the 
[MDN web docs](https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods). The expected
format for these names is uppercase strings in a YAML list. 

If this feature is used it expects the X-Forwarded-Method header. This header is sent
by default with Traefik forward auth configurations, but other proxies will have to be 
configured for this. Please see the [nginx](../deployment/supported-proxies/nginx.md) and
[haproxy](../deployment/supported-proxies/haproxy.md) deployment guides for more information.
If you specify a rule with methods, and the X-Forwarded-Method header is not present or
does not match one of the available methods described by MDN, the rule will effectively
be completely disabled. If no method is defined the rule will match all methods. 


## Resources

A rule can define multiple regular expressions for matching the path of the resource. If
any one of them matches, the resources criteria of the rule matches.


## Subjects

A subject is a representation of a user or a group of user for who the rule should apply.

For a user with unique identifier `john`, the subject should be `user:john` and for a group
uniquely identified by `developers`, the subject should be `group:developers`. Unlike resources
there can be only one subject per rule. However, if multiple users or group must be matched by
a rule, one can just duplicate the rule as many times as there are subjects.

*Note: Any PR to make it a list instead of a single item is welcome.*

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

    access_control:
        default_policy: deny

        rules:
        - domain: public.example.com
          policy: bypass

        - domain: secure.example.com
          policy: one_factor
          networks:
          - 192.168.1.0/24
      
        - domain: secure.example.com
          policy: two_factor

        - domain: singlefactor.example.com
          policy: one_factor

        - domain: "mx2.mail.example.com"
          subject: "group:admins"
          policy: deny
        
        - domain: "*.example.com"
          subject: "group:admins"
          policy: two_factor

        - domain: dev.example.com
          resources:
          - "^/groups/dev/.*$"
          subject: "group:dev"
          policy: two_factor

        - domain: dev.example.com
          resources:
          - "^/users/john/.*$"
          subject: "user:john"
          policy: two_factor
          
        - domain: api.example.com
          policy: bypass
          methods:
          - OPTIONS
        - domain api.example.com
          policy: one_factor