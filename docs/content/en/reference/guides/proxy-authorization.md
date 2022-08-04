---
title: "Proxy Authorization"
description: "A reference guide on Proxy Authorization implementations"
lead: "This section contains reference guide on Proxy Authorization implementations Authelia supports."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
reference:
parent: "guides"
weight: 220
toc: true
aliases:
- /r/proxy-authz
---

Proxies can integrate with Authelia via several authorization endpoints. These endpoints are by default configured
appropriately for most use cases; however they can be individually configured, removed, added, etc.

They are currently divided into two sections:

- [Implementations](#implementations)
- [Authn Strategies](#authn-strategies)

These endpoints are meant to collect important information from these requests via headers to determine both
metadata about the request (such as the resource and IP address of the user) which is determined via the
[Implementations](#implementations), and the identity of the user which is determined via the
[Authn Strategies](#authn-strategies).

## Metadata

|   Name   |                   Description                    |
|:--------:|:------------------------------------------------:|
|  Scheme  |          The URI Scheme of the Request           |
| Hostname |         The URI Hostname of the Request          |
|   Path   |           The URI Path of the Request            |
|  Method  |          The Method Verb of the Request          |
|    IP    | The  IP address of the client making the Request |

## Default Endpoints

|     Name     |          Path           | [Implementation] |                   [Authn Strategies]                   |
|:------------:|:-----------------------:|:----------------:|:------------------------------------------------------:|
| forward-auth | /api/authz/forward-auth |  [ForwardAuth]   |      [HeaderProxyAuthorization], [CookieSession]       |
| auth-request | /api/authz/auth-request |  [AuthRequest]   | [HeaderAuthRequestProxyAuthorization], [CookieSession] |
|    legacy    |    /api/authz/legacy    |  [AuthRequest]   | [HeaderAuthRequestProxyAuthorization], [CookieSession] |

[Implementation]: #implementations
[Authn Strategies]: #authn-strategies
[ForwardAuth]: #forwardauth
[AuthRequest]: #authrequest
[Legacy]: #legacy
[HeaderProxyAuthorization]: #headerproxyauthorization
[HeaderAuthRequestProxyAuthorization]: #headerauthrequestproxyauthorization
[HeaderLegacy]: #headerlegacy
[CookieSession]: #cookiesession

## Implementations

### ForwardAuth

This is the implementation which supports Traefik's
[ForwardAuth middleware](https://doc.traefik.io/traefik/middlewares/http/forwardauth/), Caddy's
[forward_auth directive](https://caddyserver.com/docs/caddyfile/directives/forward_auth), and Skipper.

#### Metadata (ForwardAuth)

|       Header       | Metadata |
|:------------------:|:--------:|
| X-Forwarded-Proto  |  Scheme  |
|  X-Forwarded-Host  | Hostname |
|  X-Forwarded-Uri   |   Path   |
| X-Forwarded-Method |  Method  |
|  X-Forwarded-For   |    IP    |

####
### AuthRequest

This is the implementation which supports NGINX's
[auth_request HTTP module](https://nginx.org/en/docs/http/ngx_http_auth_request_module.html).

|      Header       |        Metadata        |
|:-----------------:|:----------------------:|
|  X-Original-URL   | Scheme, Hostname, Path |
| X-Original-Method |         Method         |
|  X-Forwarded-For  |           IP           |

### Legacy

This is the legacy implementation which used to operate similar to both the [ForwardAuth](#forwardauth) and
[AuthRequest](#authrequest) implementations.

*__Note:__ This implementation has duplicate entries for metadata. This is due to the fact this implementation used to
cater for the AuthRequest and ForwardAuth implementations. The table is in order of precedence where if a header higher
in the list exists it is used over those lower in the list.*

|       Header       |        Metadata        |
|:------------------:|:----------------------:|
|   X-Original-URL   | Scheme, Hostname, Path |
| X-Forwarded-Proto  |         Scheme         |
|  X-Forwarded-Host  |        Hostname        |
|  X-Forwarded-Uri   |          Path          |
| X-Forwarded-Method |         Method         |
| X-Original-Method  |         Method         |
|  X-Forwarded-For   |           IP           |

## Authn Strategies

### CookieSession

### HeaderAuthorization

### HeaderProxyAuthorization

### HeaderAuthRequestProxyAuthorization

### HeaderLegacy
