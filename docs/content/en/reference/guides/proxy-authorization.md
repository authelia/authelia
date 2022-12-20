---
title: "Proxy Authorization"
description: "A reference guide on Proxy Authorization implementations"
lead: "This section contains reference guide on Proxy Authorization implementations Authelia supports."
date: 2022-10-31T09:33:39+11:00
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

## Default Endpoints

|     Name     |          Path           | [Implementation] |                   [Authn Strategies]                   |
|:------------:|:-----------------------:|:----------------:|:------------------------------------------------------:|
| forward-auth | /api/authz/forward-auth |  [ForwardAuth]   |      [HeaderProxyAuthorization], [CookieSession]       |
|  ext-authz   |  /api/authz/ext-authz   |    [ExtAuthz]    |      [HeaderProxyAuthorization], [CookieSession]       |
| auth-request | /api/authz/auth-request |  [AuthRequest]   | [HeaderAuthRequestProxyAuthorization], [CookieSession] |
|    legacy    |    /api/authz/legacy    |  [AuthRequest]   | [HeaderAuthRequestProxyAuthorization], [CookieSession] |

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

## Metadata

Various metadata is collected from the request made to the Authelia authorization server. This table describes the
metadata collected. All of this metadata is utilized for the purpose of determining if the user is authorized to a
particular resource.

|    Name    |                   Description                   |
|:----------:|:-----------------------------------------------:|
|   Method   |         The Method Verb of the Request          |
|   Scheme   |          The URI Scheme of the Request          |
|  Hostname  |         The URI Hostname of the Request         |
|    Path    |           The URI Path of the Request           |
|     IP     | The IP address of the client making the Request |
| Portal URL |         The URL of the Authelia Portal          |

## Implementations

### ForwardAuth

This is the implementation which supports Traefik's
[ForwardAuth middleware](https://doc.traefik.io/traefik/middlewares/http/forwardauth/), Caddy's
[forward_auth directive](https://caddyserver.com/docs/caddyfile/directives/forward_auth), and Skipper.

#### ForwardAuth Metadata

|  Metadata  |     Source     |         Key          |
|:----------:|:--------------:|:--------------------:|
|   Method   |    [Header]    | `X-Forwarded-Method` |
|   Scheme   |    [Header]    | [X-Forwarded-Proto]  |
|  Hostname  |    [Header]    |  [X-Forwarded-Host]  |
|    Path    |    [Header]    |  `X-Forwarded-URI`   |
|     IP     |    [Header]    |  [X-Forwarded-For]   |
| Portal URL | Query Argument |         `rd`         |

### ExtAuthz

This is the implementation which supports Envoy's [ExtAuthz Protocol].

[ExtAuthz Protocol]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/ext_authz/v3/ext_authz.proto#envoy-v3-api-msg-extensions-filters-http-ext-authz-v3-extauthz

#### ExtAuthz Metadata

|  Metadata  |     Source     |         Key         |
|:----------:|:--------------:|:-------------------:|
|   Method   | _[Start Line]_ |    [HTTP Method]    |
|   Scheme   |    [Header]    | [X-Forwarded-Proto] |
|  Hostname  |    [Header]    | [X-Forwarded-Host]  |
|    Path    |    [Header]    |  `X-Forwarded-URI`  |
|     IP     |    [Header]    |  [X-Forwarded-For]  |
| Portal URL |    [Header]    |  `X-Authelia-URL`   |

### AuthRequest

This is the implementation which supports NGINX's
[auth_request HTTP module](https://nginx.org/en/docs/http/ngx_http_auth_request_module.html).

|  Metadata  |  Source  |         Key         |
|:----------:|:--------:|:-------------------:|
|   Method   | [Header] | `X-Original-Method` |
|   Scheme   | [Header] |  `X-Original-URL`   |
|  Hostname  | [Header] |  `X-Original-URL`   |
|    Path    | [Header] |  `X-Original-URL`   |
|     IP     | [Header] |  [X-Forwarded-For]  |
| Portal URL |   N/A    |         N/A         |

### Legacy

This is the legacy implementation which used to operate similar to both the [ForwardAuth](#forwardauth) and
[AuthRequest](#authrequest) implementations.

*__Note:__ This implementation has duplicate entries for metadata. This is due to the fact this implementation used to
cater for the AuthRequest and ForwardAuth implementations. The table is in order of precedence where if a header higher
in the list exists it is used over those lower in the list.*

|  Metadata  |     Source     |         Key          |
|:----------:|:--------------:|:--------------------:|
|   Method   |    [Header]    | `X-Original-Method`  |
|   Scheme   |    [Header]    |   `X-Original-URL`   |
|  Hostname  |    [Header]    |   `X-Original-URL`   |
|    Path    |    [Header]    |   `X-Original-URL`   |
|   Method   |    [Header]    | `X-Forwarded-Method` |
|   Scheme   |    [Header]    | [X-Forwarded-Proto]  |
|  Hostname  |    [Header]    |  [X-Forwarded-Host]  |
|    Path    |    [Header]    |  `X-Forwarded-URI`   |
|     IP     |    [Header]    |  [X-Forwarded-For]   |
| Portal URL | Query Argument |         `rd`         |

## Authn Strategies

Authentication strategies are used to determine the users identity which is essential to determining if they are
authorized to visit a particular resource. Authentication strategies are executed in order, and have three potential
results.

1. Successful Authentication
2. No Authentication
3. Unsuccessful Authentication

Both result 2 is the only result in which the next strategy is attempted, this occurs when there is not enough
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
the header is malformed it will respond with the [WWW-Authenticate] header.

### HeaderProxyAuthorization

This strategy uses the [Proxy-Authorization] header to determine the users' identity. If the user credentials are wrong,
or the header is malformed it will respond with the [Proxy-Authenticate] header.

### HeaderAuthRequestProxyAuthorization

TODO:

### HeaderLegacy

This strategy uses the [Proxy-Authorization] header to determine the users' identity. If the user credentials are wrong,
or the header is malformed it will respond with the [WWW-Authenticate] header.

[Authorization]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Authorization
[WWW-Authenticate]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/WWW-Authenticate
[Proxy-Authorization]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Proxy-Authorization
[Proxy-Authenticate]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Proxy-Authenticate

[X-Forwarded-Proto]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-Proto
[X-Forwarded-Host]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-Host
[X-Forwarded-For]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-For
[HTTP Method]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods
[HTTP Method]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods
[Start Line]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Messages#start_line
[Header]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers
