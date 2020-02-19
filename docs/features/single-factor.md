---
layout: default
title: Single Factor
parent: Features
nav_order: 3
---

# Single Factor

**Authelia** supports single factor authentication to let applications
send authenticated requests to other applications.

Single or two-factor authentication can be configured per resource of an
application for flexibility.

For instance, you can configure Authelia to grant access to all resources
matching `app1.example.com/api/(.*)` with only a single factor and all
resources matching `app1.example.com/admin` with two factors.

To know more about the configuration of the feature, please visit the
documentation about the [configuration](../deployment/configuration.md).


## Proxy-Authorization header

Authelia reads credentials from the header `Proxy-Authorization` instead of
the usual `Authorization` header. This is because in some circumstances both Authelia
and the application could require authentication in order to provide specific
authorizations at the level of the application.
