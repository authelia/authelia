---
title: "Proxies"
description: "An integration guide for Authelia and several supported reverse proxies"
summary: "An introduction into integrating Authelia with a reverse proxy."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 310
toc: true
aliases:
  - /i/proxies
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

__Authelia__ works in collaboration with several reverse proxies. In this section you will find the documentation of the
various tested proxies with examples of how you may configure them. We are eager for users to help us provide better
examples of already documented proxies, as well as provide us examples of undocumented proxies.

## Get started

It's __*strongly recommended*__ that users setting up *Authelia* for the first time take a look at our
[Get started](../prologue/get-started.md) guide. This takes you through various steps which are essential to
bootstrapping *Authelia*.

## Support

See [support](support.md) for support information.

### Required Headers

There are several required headers for Authelia to operate properly. These headers are considered part of the supported
configuration and they are assumed to be present for future development.

You may not be able to do several things that rely on these headers such as but not limited to:

  - Properly identify the request information for Access Control, such as client IP, hostname, request URL or Scheme,
    etc.
  - Properly identify the correct domain for session cookies.
  - Properly identify the public facing URL for redirection or email links making redirection or email based
    verifications inoperable.
  - Properly identify the OpenID Connect 1.0 Issuer or Endpoint URL's making OpenID Connect 1.0 inoperable.
  - Properly identify the WebAuthn Relying Party Identifier making WebAuthn inoperable.

In addition to the [Proxy Authorization Endpoint](../../reference/guides/proxy-authorization.md) implementations and the
headers required by those, __Authelia__ itself requires the following headers are set when secured behind a reverse
proxy i.e. the headers a reverse proxy must include for the __Authelia__ portal app itself:

* Scheme Detection:
  * Default: [X-Forwarded-Proto] (header)
  * Fallback: TLS (listening socket state)
* Host Detection:
  * Default: [X-Forwarded-Host] (header)
  * Fallback: [Host] (header)
* Path Detection:
  * Default: X-Forwarded-URI (header)
  * Fallback: [Start Line] Request Target (start line)
* Remote IP:
  * Default: [X-Forwarded-For]
  * Fallback: TCP source IP

[Host]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Host
[Start Line]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Messages#start_line
[X-Forwarded-For]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-For
[X-Forwarded-Proto]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-Proto
[X-Forwarded-Host]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-Host

## Important Notes

{{< callout context="danger" title="Important Notes" icon="outline/alert-octagon" >}}
The following section has important notes for integrating Authelia with your proxy.
{{< /callout >}}

- When configuring Authelia on a subpath either by the
  [server address](../../configuration/miscellaneous/server.md#address) or the deprecated server `path` option it's
  strongly recommended that when users are integrating the `/api/authz/*` or `/api/verify` endpoints do not include the
  configured path within those URLs. This is because the handler will listen on both the root path and the configured
  path and several misconfiguration issues can be avoided by doing this.

## Integration Implementation

Authelia is capable of being integrated into many proxies due to the decisions regarding the implementation. We handle
requests to the authz endpoints with specific headers and return standardized responses based on the headers and
the policy engines determination about what must be done.

### Destination Identification

Broadly speaking, the method to identify the destination of a request relies on metadata headers which need to be set by
your reverse proxy. The headers we rely on at the authz endpoints are as follows:

* [X-Forwarded-Proto](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-Proto)
* [X-Forwarded-Host](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-Host)
* X-Forwarded-URI
* [X-Forwarded-For](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-For)
* X-Forwarded-Method / X-Original-Method
* X-Original-URL

The specifics however are dictated by the specific
[Authorization Implementation](../../reference/guides/proxy-authorization.md) used. Please refer to the specific
implementation you're using.

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
