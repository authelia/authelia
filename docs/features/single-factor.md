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

Authelia also supports machine to machine authentication. The authentication
backed stay the same as UI driven authentication. But since there is not 
user to input the credentials, those credentials will try to be extracted 
from either the `Proxy-Authorization` header or the `Authorization` header.

## Proxy-Authorization header

By default, Authelia reads credentials from the header `Proxy-Authorization` 
instead of the usual `Authorization` header. This header is considered a 
[Hop header](https://tools.ietf.org/html/rfc7235#section-4.4) and many proxies 
will remove the header before sending the authentication request to Authelia. 

## Authorization header

If the Proxy-Authorization header is not found, but the Authorization header is present, Authelia will try to use that header to extract username and password for authentication. This could cause issues when the `Authentication` header is not directed to Authelia. Still, to the destination application, in that case, the proxy in front of Authelia should be instructed to explicitly remove the headers before sending an authentication request to Authelia. 