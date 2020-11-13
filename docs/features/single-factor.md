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
documentation about the [configuration](../configuration/access-control.md).


## Proxy-Authorization header

Authelia reads credentials from the header `Proxy-Authorization` instead of
the usual `Authorization` header. This is because in some circumstances both Authelia
and the application could require authentication in order to provide specific
authorizations at the level of the application.


## Session-Username header

Authelia by default only verifies the cookie and the associated user with that cookie can
access a protected resource. The client browser does not know the username and does not send
this to Authelia, it's stored by Authelia for security reasons.
 
The Session-Username header has been implemented as a means
to use Authelia with non-web services such as PAM. Basically how it works is if the
Session-Username header is sent in the request to the /api/verify endpoint it will
only respond with a sucess message if the cookie username and the header username
match. 