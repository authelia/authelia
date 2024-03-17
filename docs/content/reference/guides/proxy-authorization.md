---
title: "Proxy Authorization"
description: "A reference guide on Proxy Authorization implementations"
summary: "This section contains reference guide on Proxy Authorization implementations Authelia supports."
date: 2023-01-25T20:36:40+11:00
draft: false
images: []
menu:
reference:
parent: "guides"
weight: 220
toc: true
aliases:
  - /r/proxy-authz
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
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

## Default Endpoints

|     Name     |          Path           | [Implementation] |                [Authn Strategies]                 |
|:------------:|:-----------------------:|:----------------:|:-------------------------------------------------:|
| forward-auth | /api/authz/forward-auth |  [ForwardAuth]   |      [HeaderAuthorization], [CookieSession]       |
|  ext-authz   |  /api/authz/ext-authz   |    [ExtAuthz]    |      [HeaderAuthorization], [CookieSession]       |
| auth-request | /api/authz/auth-request |  [AuthRequest]   | [HeaderAuthRequestAuthorization], [CookieSession] |
|    legacy    |       /api/verify       |     [Legacy]     |          [HeaderLegacy], [CookieSession]          |

## Metadata

Various metadata is collected from the request made to the Authelia authorization server. This table describes the
metadata collected. All of this metadata is utilized for the purpose of determining if the user is authorized to a
particular resource.

|     Name     |                   Description                   |
|:------------:|:-----------------------------------------------:|
|    Method    |         The Method Verb of the Request          |
|    Scheme    |          The URI Scheme of the Request          |
|   Hostname   |         The URI Hostname of the Request         |
|     Path     |           The URI Path of the Request           |
|      IP      | The IP address of the client making the Request |
| Authelia URL |         The URL of the Authelia Portal          |

Some values may have either fallbacks or override values. If they exist they will be in the alternatives table which
will be below the main metadata table.

The metadata table contains the recommended source of this information and this source is often times automatic
depending on the proxy implementation. The difference between an override and a fallback is an override values will
take precedence over the metadata values, and fallbacks only take effect if the override values or metadata values are
completely unset.

## Implementations

### ForwardAuth

This is the implementation which supports [Traefik] via the [ForwardAuth Middleware], [Caddy] via the
[forward_auth directive], [HAProxy] via the [auth-request lua plugin], and [Skipper] via the [webhook auth filter].

#### ForwardAuth Metadata

|   Metadata   |            Source            |         Key          |
|:------------:|:----------------------------:|:--------------------:|
|    Method    |           [Header]           | `X-Forwarded-Method` |
|    Scheme    |           [Header]           | [X-Forwarded-Proto]  |
|   Hostname   |           [Header]           |  [X-Forwarded-Host]  |
|     Path     |           [Header]           |  `X-Forwarded-URI`   |
|      IP      |           [Header]           |  [X-Forwarded-For]   |
| Authelia URL | Session Cookie Configuration |    `authelia_url`    |

#### ForwardAuth Metadata Alternatives

|   Metadata   | Alternative Type |     Source     |      Key       |
|:------------:|:----------------:|:--------------:|:--------------:|
|    Scheme    |     Fallback     |    [Header]    | Server Scheme  |
|      IP      |     Fallback     |   TCP Packet   |   Source IP    |
| Authelia URL |     Override     | Query Argument | `authelia_url` |

### ExtAuthz

This is the implementation which supports [Envoy] via the [HTTP ExtAuthz Filter].

#### ExtAuthz Metadata

|   Metadata   |            Source            |         Key         |
|:------------:|:----------------------------:|:-------------------:|
|    Method    |        _[Start Line]_        |    [HTTP Method]    |
|    Scheme    |           [Header]           | [X-Forwarded-Proto] |
|   Hostname   |           [Header]           |       [Host]        |
|     Path     |           [Header]           |  Endpoint Sub-Path  |
|      IP      |           [Header]           |  [X-Forwarded-For]  |
| Authelia URL | Session Cookie Configuration |   `authelia_url`    |

#### ExtAuthz Metadata Alternatives

|   Metadata   | Alternative Type |   Source   |        Key         |
|:------------:|:----------------:|:----------:|:------------------:|
|    Scheme    |     Fallback     |  [Header]  |   Server Scheme    |
|      IP      |     Fallback     | TCP Packet |     Source IP      |
| Authelia URL |     Override     |  [Header]  |  `X-Authelia-URL`  |

### AuthRequest

This is the implementation which supports [NGINX] via the [auth_request HTTP module], and can technically support
[HAProxy] via the [auth-request lua plugin].

#### AuthRequest Metadata

|   Metadata   |            Source            |         Key         |
|:------------:|:----------------------------:|:-------------------:|
|    Method    |           [Header]           | `X-Original-Method` |
|    Scheme    |           [Header]           |  `X-Original-URL`   |
|   Hostname   |           [Header]           |  `X-Original-URL`   |
|     Path     |           [Header]           |  `X-Original-URL`   |
|      IP      |           [Header]           |  [X-Forwarded-For]  |
| Authelia URL | Session Cookie Configuration |   `authelia_url`    |

_**Note:** This endpoint does not support automatic redirection. This is because there is no support on [NGINX]'s side
to achieve this with `ngx_http_auth_request_module` and the redirection must be performed within the [NGINX]
configuration. However we return the appropriate URL to redirect users to with the `Location` header which
simplifies this process especially for multi-cookie domain deployments._

#### AuthRequest Metadata Alternatives

|   Metadata   | Alternative Type |     Source     |      Key       |
|:------------:|:----------------:|:--------------:|:--------------:|
|      IP      |     Fallback     |   TCP Packet   |   Source IP    |
| Authelia URL |     Override     | Query Argument | `authelia_url` |

### Legacy

This is the legacy implementation which used to operate similar to both the [ForwardAuth](#forwardauth) and
[AuthRequest](#authrequest) implementations.

*__Note:__ This implementation has duplicate entries for metadata. This is due to the fact this implementation used to
cater for the AuthRequest and ForwardAuth implementations. The table is in order of precedence where if a header higher
in the list exists it is used over those lower in the list.*

|   Metadata   |     Source     |         Key          |
|:------------:|:--------------:|:--------------------:|
|    Method    |    [Header]    | `X-Original-Method`  |
|    Scheme    |    [Header]    |   `X-Original-URL`   |
|   Hostname   |    [Header]    |   `X-Original-URL`   |
|     Path     |    [Header]    |   `X-Original-URL`   |
|    Method    |    [Header]    | `X-Forwarded-Method` |
|    Scheme    |    [Header]    | [X-Forwarded-Proto]  |
|   Hostname   |    [Header]    |  [X-Forwarded-Host]  |
|     Path     |    [Header]    |  `X-Forwarded-URI`   |
|      IP      |    [Header]    |  [X-Forwarded-For]   |
| Authelia URL | Query Argument |         `rd`         |
| Authelia URL |    [Header]    |   `X-Authelia-URL`   |

## Authn Strategies

Authentication strategies are used to determine the users identity which is essential to determining if they are
authorized to visit a particular resource. Authentication strategies are executed in order, and have three potential
results.

1. Successful Authentication
2. No Authentication
3. Unsuccessful Authentication

Result 2 is the only result in which the next strategy is attempted, this occurs when there is not enough
information in the request to perform authentication. Both result 1 and 2 result in a short-circuit, i.e. no other
strategy will be attempted.

Result 1 occurs when the strategy requirements (i.e. a particular header) are present and the details are sufficient to
authenticate them and the details are correct. Result 2 occurs when the strategy requirements are present and either the
details are incomplete (i.e. malformed header) or the details are incorrect (i.e. bad password).

### CookieSession

This strategy uses a cookie which links the user to a session to determine the users identity. This is the default
strategy for end-users.

If this strategy if included in an endpoint will redirect the user the Authelia Authorization Portal on supported
proxies when they are not authorized and can potentially be authorized provided no other strategies have critical
errors.

### HeaderAuthorization

This strategy uses the [Authorization] header to determine the users' identity. If the user credentials are wrong, or
the header is malformed it will respond with the [WWW-Authenticate] header and a [401 Unauthorized] status code.

### HeaderProxyAuthorization

This strategy uses the [Proxy-Authorization] header to determine the users' identity. If the user credentials are wrong,
or the header is malformed it will respond with the [Proxy-Authenticate] header and a
[407 Proxy Authentication Required] status code.

### HeaderAuthRequestProxyAuthorization

This strategy uses the [Proxy-Authorization] header to determine the users' identity. If the user credentials are wrong,
or the header is malformed it will respond with the [WWW-Authenticate] header and a [401 Unauthorized] status code. It
is specifically intended for use with the [AuthRequest] implementation.

### HeaderLegacy

This strategy uses the [Proxy-Authorization] header to determine the users' identity. If the user credentials are wrong,
or the header is malformed it will respond with the [WWW-Authenticate] header.

[401 Unauthorized]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/401
[407 Proxy Authentication Required]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/407

[NGINX]: https://www.nginx.com/
[Traefik]: https://traefik.io/traefik/
[Envoy]: https://www.envoyproxy.io/
[Caddy]: https://caddyserver.com/
[Skipper]: https://opensource.zalando.com/skipper/
[HAProxy]: http://www.haproxy.org/

[HTTP ExtAuthz Filter]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/ext_authz/v3/ext_authz.proto#envoy-v3-api-msg-extensions-filters-http-ext-authz-v3-extauthz
[auth_request HTTP module]: https://nginx.org/en/docs/http/ngx_http_auth_request_module.html
[auth-request lua plugin]: https://github.com/TimWolla/haproxy-auth-request
[ForwardAuth Middleware]: https://doc.traefik.io/traefik/middlewares/http/forwardauth/
[forward_auth directive]: https://caddyserver.com/docs/caddyfile/directives/forward_auth
[webhook auth filter]: https://opensource.zalando.com/skipper/reference/filters/#webhook

[Implementation]: #implementations
[Authn Strategies]: #authn-strategies
[ForwardAuth]: #forwardauth
[ExtAuthz]: #extauthz
[AuthRequest]: #authrequest
[Legacy]: #legacy
[HeaderProxyAuthorization]: #headerproxyauthorization
[HeaderAuthRequestProxyAuthorization]: #headerauthrequestproxyauthorization
[HeaderLegacy]: #headerlegacy
[CookieSession]: #cookiesession

[Authorization]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Authorization
[WWW-Authenticate]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/WWW-Authenticate
[Proxy-Authorization]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Proxy-Authorization
[Proxy-Authenticate]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Proxy-Authenticate

[X-Forwarded-Proto]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-Proto
[X-Forwarded-Host]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-Host
[X-Forwarded-For]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-For
[Host]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Host

[HTTP Method]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods
[HTTP Method]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods
[Start Line]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Messages#start_line
[Header]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers
