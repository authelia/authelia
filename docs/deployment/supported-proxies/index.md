---
layout: default
title: Proxy Integration
parent: Deployment
nav_order: 4
has_children: true
---

# Integration with proxies

**Authelia** works in collaboration with reverse proxies. In the sub-pages you
can find the documentation of the configuration required for every supported
proxy.

If you are not aware of the workflow of an authentication request, reading this
[documentation](../../home/architecture.md) first is highly recommended.


## How Authelia integrates with proxies?

Authelia takes authentication requests coming from the proxy and targeting the 
`/api/verify` endpoint exposed by Authelia. Two pieces of information are required for
Authelia to be able to authenticate the user request:

* The session cookie or a `Proxy-Authorization` header (see [single factor authentication](../../features/single-factor.md)).
* The target URL of the user request (used primarily for [access control](../../features/access-control.md)).

The target URL can be provided using one of the following ways:

* With `X-Original-URL` header containing the complete URL of the initial request.
* With a combination of `X-Forwarded-Proto`, `X-Forwarded-Host` and `X-Forwarded-URI` headers.

In the case of Traefik, these headers are automatically provided and therefore don't
appear in the configuration examples.

## Redirection to the login portal

The endpoint `/api/verify` has different behaviors depending on whether
the `rd` (for redirection) query parameter is provided.

If redirection parameter is provided and contains the URL to the login portal
served by Authelia, the request will either generate a 200 response
if the request is authenticated or perform a redirection (302 response) to the
login portal if not authenticated yet.

If no redirection parameter is provided, the response code is either 200 or 401. The
redirection must then be handled by the proxy when an error is detected
(see [nginx](./nginx.md) example).
