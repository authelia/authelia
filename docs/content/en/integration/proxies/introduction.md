---
title: "Proxies"
description: "An integration guide for Authelia and several supported reverse proxies"
lead: "An introduction into integrating Authelia with a reverse proxy."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  integration:
    parent: "proxies"
weight: 310
toc: true
aliases:
  - /i/proxies
---

__Authelia__ works in collaboration with several reverse proxies. In this section you will find the documentation of the
various tested proxies with examples of how you may configure them. We are eager for users to help us provide better
examples of already documented proxies, as well as provide us examples of undocumented proxies.

## Get Started

It's __*strongly recommended*__ that users setting up *Authelia* for the first time take a look at our
[Get Started](../prologue/get-started.md) guide. This takes you through various steps which are essential to
bootstrapping *Authelia*.

## Support

See [support](support.md) for support information.

## Integration Implementation

Authelia is capable of being integrated into many proxies due to the decisions regarding the implementation. We handle
requests to the `/api/verify` endpoint with specific headers and return standardized responses based on the headers and
the policy engines determination about what must be done.

### Destination Identification

The method to identify the destination of a request relies on metadata headers which need to be set by your reverse
proxy. The headers we rely on are as follows:

* [X-Forwarded-Proto](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-Proto)
* [X-Forwarded-Host](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-Host)
* X-Forwarded-Uri
* [X-Forwarded-For](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-For)
* X-Forwarded-Method

Alternatively we utilize `X-Original-URL` header which is expected to contain a fully formatted URL.

### User Identification

A logged in user must be identified via standard means. Users are identified by one of two methods:

* A session cookie with the HTTP only option set, and the secure option set meaning the cookie must only be sent over the
  [HTTPS scheme](https://developer.mozilla.org/en-US/docs/Glossary/https).
* The [Proxy-Authorization](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Proxy-Authorization) header
  utilizing the
  [basic authentication scheme](https://developer.mozilla.org/en-US/docs/Web/HTTP/Authentication#basic_authentication_scheme).

### Response Statuses

Authelia responds in various ways depending on the result of the authorization policies.

When the user is authenticated and authorized to access a resource we respond with a HTTP
[200 OK](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/200). When the user is not logged in and we need them
to authenticate with 1FA, or if they are already authenticated with only 1FA and they need to perform 2FA, the user is
redirected to the portal with:

* A HTTP [401 Unauthorized](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/401) status if the original request
  was an [XMLHTTPRequest](https://developer.mozilla.org/en-US/docs/Web/API/XMLHttpRequest) provided Authelia is able to
  detect it.
* A HTTP [302 Found](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/302) status if the original request had
  the [GET](https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/GET) or
  [OPTIONS](https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/OPTIONS) method verb.
* A HTTP [303 See Other](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/303) status if neither of the above
  conditions is met.

When the user is denied either by a default policy, or by an explicit policy we respond with a HTTP
[403 Forbidden](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/403) status.

### Response Headers

With the exception of the [403 Forbidden](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/403) and
[200 OK](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/200) status responses above,
Authelia responds with a [Location](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Location) header to
redirect the user to the authentication portal.

In the instance of a [200 OK](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/200) status response Authelia
also responds with various headers which can be forwarded by your reverse proxy to the backend application which are
potentially useful for SSO depending on if the backend application supports it.

See the [Trusted Header SSO](../trusted-header-sso/introduction.md) documentation for more information.
