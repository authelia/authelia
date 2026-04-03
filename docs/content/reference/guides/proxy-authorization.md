---
title: "Proxy Authorization"
description: "A reference guide on Proxy Authorization implementations"
summary: "This section contains reference guide on Proxy Authorization implementations Authelia supports."
date: 2024-03-14T06:00:14+11:00
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

|     Name     |          Path           | [Implementation] |           [Authn Strategies]           |
|:------------:|:-----------------------:|:----------------:|:--------------------------------------:|
| forward-auth | /api/authz/forward-auth |  [ForwardAuth]   | [HeaderAuthorization], [CookieSession] |
|  ext-authz   |  /api/authz/ext-authz   |    [ExtAuthz]    | [HeaderAuthorization], [CookieSession] |
| auth-request | /api/authz/auth-request |  [AuthRequest]   | [HeaderAuthorization], [CookieSession] |
|    legacy    |       /api/verify       |     [Legacy]     |    [HeaderLegacy], [CookieSession]     |

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

|     Metadata      |            Source            |            Key            |
|:-----------------:|:----------------------------:|:-------------------------:|
|    Method [^1]    |           [Header]           | `X-Forwarded-Method` [^2] |
|    Scheme [^1]    |           [Header]           | [X-Forwarded-Proto] [^2]  |
|   Hostname [^1]   |           [Header]           |  [X-Forwarded-Host] [^2]  |
|     Path [^1]     |           [Header]           |  `X-Forwarded-URI` [^2]   |
|      IP [^1]      |           [Header]           |  [X-Forwarded-For] [^3]   |
| Authelia URL [^1] | Session Cookie Configuration |      `authelia_url`       |

#### ForwardAuth Metadata Alternatives

|     Metadata      | Alternative Type |     Source     |      Key       |
|:-----------------:|:----------------:|:--------------:|:--------------:|
|    Scheme [^1]    |     Fallback     |    [Header]    | Server Scheme  |
|      IP [^1]      |     Fallback     |   TCP Packet   |   Source IP    |
| Authelia URL [^1] |     Override     | Query Argument | `authelia_url` |

### ExtAuthz

This is the implementation which supports [Envoy] via the [HTTP ExtAuthz Filter].

#### ExtAuthz Metadata

|     Metadata      |            Source            |           Key            |
|:-----------------:|:----------------------------:|:------------------------:|
|    Method [^1]    |        _[Start Line]_        |    [HTTP Method] [^2]    |
|    Scheme [^1]    |           [Header]           | [X-Forwarded-Proto] [^2] |
|   Hostname [^1]   |           [Header]           |       [Host] [^2]        |
|     Path [^1]     |           [Header]           |  Endpoint Sub-Path [^2]  |
|      IP [^1]      |           [Header]           |  [X-Forwarded-For] [^2]  |
| Authelia URL [^1] | Session Cookie Configuration |      `authelia_url`      |

#### ExtAuthz Metadata Alternatives

|     Metadata      | Alternative Type |   Source   |       Key        |
|:-----------------:|:----------------:|:----------:|:----------------:|
|    Scheme [^1]    |     Fallback     |  [Header]  |  Server Scheme   |
|      IP [^1]      |     Fallback     | TCP Packet |    Source IP     |
| Authelia URL [^1] |     Override     |  [Header]  | `X-Authelia-URL` |

### AuthRequest

This is the implementation which supports [NGINX] via the [auth_request HTTP module], and can technically support
[HAProxy] via the [auth-request lua plugin].

#### AuthRequest Metadata

|     Metadata      |            Source            |           Key            |
|:-----------------:|:----------------------------:|:------------------------:|
|    Method [^1]    |           [Header]           | `X-Original-Method` [^2] |
|    Scheme [^1]    |           [Header]           |  `X-Original-URL` [^2]   |
|   Hostname [^1]   |           [Header]           |  `X-Original-URL` [^2]   |
|     Path [^1]     |           [Header]           |  `X-Original-URL` [^2]   |
|      IP [^1]      |           [Header]           |  [X-Forwarded-For] [^3]  |
| Authelia URL [^1] | Session Cookie Configuration |      `authelia_url`      |

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
This endpoint does not support automatic redirection. This is because there is no support on [NGINX](https://www.nginx.com/)'s side
to achieve this with `ngx_http_auth_request_module` and the redirection must be performed within the [NGINX](https://www.nginx.com/)
configuration. However, we return the appropriate URL to redirect users to with the `Location` header which
simplifies this process especially for multi-cookie domain deployments.
{{< /callout >}}

#### AuthRequest Metadata Alternatives

|     Metadata      | Alternative Type |     Source     |      Key       |
|:-----------------:|:----------------:|:--------------:|:--------------:|
|      IP [^1]      |     Fallback     |   TCP Packet   |   Source IP    |
| Authelia URL [^1] |     Override     | Query Argument | `authelia_url` |

### Legacy

This is the legacy implementation which used to operate similar to both the [ForwardAuth](#forwardauth) and
[AuthRequest](#authrequest) implementations.

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
This implementation has duplicate entries for metadata. This is due to the fact this implementation used to
cater for the AuthRequest and ForwardAuth implementations. The table is in order of precedence where if a header higher
in the list exists it is used over those lower in the list.
{{< /callout >}}

|     Metadata      |     Source     |         Key          |
|:-----------------:|:--------------:|:--------------------:|
|    Method [^1]    |    [Header]    | `X-Original-Method`  |
|    Scheme [^1]    |    [Header]    |   `X-Original-URL`   |
|   Hostname [^1]   |    [Header]    |   `X-Original-URL`   |
|     Path [^1]     |    [Header]    |   `X-Original-URL`   |
|    Method [^1]    |    [Header]    | `X-Forwarded-Method` |
|    Scheme [^1]    |    [Header]    | [X-Forwarded-Proto]  |
|   Hostname [^1]   |    [Header]    |  [X-Forwarded-Host]  |
|     Path [^1]     |    [Header]    |  `X-Forwarded-URI`   |
|      IP [^1]      |    [Header]    |  [X-Forwarded-For]   |
| Authelia URL [^1] | Query Argument |         `rd`         |
| Authelia URL [^1] |    [Header]    |   `X-Authelia-URL`   |

## Authn Strategies

Authentication strategies are used to determine the users identity which is essential to determining if they are
authorized to visit a particular resource. Authentication strategies are executed in order, and have three potential
results.

1. Successful Authentication
   - This result occurs when the required metadata i.e. headers are in the request for the strategy and they can be
     validated.
   - This result causes a short-circuit which generally results in a [200 OK].
2. Unsuccessful Authentication
   - This result occurs when the required metadata i.e. headers are in the request for the strategy and they can not be
     validated as they are either explicitly invalid or the means of validation could not be attempted due to an error.
   - This result causes a short-circuit applying the failure action and no other strategies will be attempted.
3. No Authentication
   - This result occurs when the required metadata i.e. headers are absent from the request for the strategy.
   - This result does not cause a short-circuit and:
      1. The next strategy will be attempted.
      2. If there is no next strategy the failure action will be applied.

### CookieSession

**Failure Action:** Redirect the user for authentication.

**Metadata:** [Cookie] header value, considered absent when the configured cookie key is absent from this header or the
header is absent.

This strategy uses a cookie which links the user to a session to determine the users identity. This is the default
strategy for end-users.

### HeaderAuthorization

**Failure Action:** Responds with the [WWW-Authenticate] header and a [401 Unauthorized] status code.

**Metadata:** [Authorization] header, considered absent when the header is absent.

This strategy uses the [Authorization] header to determine the users' identity.

### HeaderProxyAuthorization

**Failure Action:** Responds with the [Proxy-Authenticate] header and a [407 Proxy Authentication Required] status code.

**Metadata:** [Proxy-Authorization] header, considered absent when the header is absent.

This strategy uses the [Proxy-Authorization] header to determine the users' identity.

### HeaderAuthRequestProxyAuthorization

**Failure Action:** Responds with the [WWW-Authenticate] header and a [401 Unauthorized] status code.

**Metadata:** [Proxy-Authorization] header, considered absent when the header is absent.

This strategy uses the [Proxy-Authorization] header to determine the users' identity. It is specifically intended for
use with the [AuthRequest] implementation.

### HeaderLegacy

**Failure Action:** Responds with the [WWW-Authenticate] header and a [401 Unauthorized] status code.

**Metadata:** [Proxy-Authorization] header, considered absent when the header is absent.

This strategy uses the [Proxy-Authorization] header to determine the users' identity.

## Footnotes

  [^1]: This is considered required metadata, and must either be provided via the primary metadata source or the
        alternative source for the request to be considered valid.
  [^2]: This is considered a required header. If an alternative or fallback source is described this is very likely to
        be incorrect and cannot be supported.
  [^3]: This header is not required but the fallback is likely desirable in most scenarios.

[200 OK]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/200
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
[HeaderAuthorization]: #headerauthorization
[CookieSession]: #cookiesession

[Cookie]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cookie
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
