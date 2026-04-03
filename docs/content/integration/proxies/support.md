---
title: "Support"
description: "An support matrix for Authelia and several supported reverse proxies"
summary: "This documentation details a support matrix for Authelia features and specific reverse proxies as well as several caveats etc."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 311
toc: true
aliases:
  - /i/proxy
  - /docs/home/supported-proxies.html
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

|                  Proxy                  | [Implementation] |                 [Standard](#standard)                 |                                [Kubernetes](#kubernetes)                                 |   [XHR Redirect](#xhr-redirect)   | [Request Method](#request-method) |
|:---------------------------------------:|:----------------:|:-----------------------------------------------------:|:----------------------------------------------------------------------------------------:|:---------------------------------:|:---------------------------------:|
|     [Traefik] ([guide](/i/traefik))     |  [ForwardAuth]   |   {{% support support="full" link="/i/traefik" %}}    |  {{% support support="full" link="../../integration/kubernetes/traefik-ingress.md" %}}   |  {{% support support="full" %}}   |  {{% support support="full" %}}   |
|       [Caddy] ([guide](/i/caddy))       |  [ForwardAuth]   |    {{% support support="full" link="/i/caddy" %}}     |                            {{% support support="unknown" %}}                             |  {{% support support="full" %}}   |  {{% support support="full" %}}   |
|       [Envoy] ([guide](/i/envoy))       |    [ExtAuthz]    |    {{% support support="full" link="/i/envoy" %}}     | {{% support support="full" link="../../integration/kubernetes/envoy/introduction.md" %}} | {{% support support="unknown" %}} |  {{% support support="full" %}}   |
|       [NGINX] ([guide](/i/nginx))       |  [AuthRequest]   |    {{% support support="full" link="/i/nginx" %}}     |   {{% support support="full" link="../../integration/kubernetes/nginx-ingress.md" %}}    |          {{% support %}}          |  {{% support support="full" %}}   |
| [NGINX Proxy Manager] ([guide](/i/npm)) |  [AuthRequest]   |     {{% support support="full" link="/i/npm" %}}      |                            {{% support support="unknown" %}}                             |          {{% support %}}          |  {{% support support="full" %}}   |
|        [SWAG] ([guide](/i/swag))        |  [AuthRequest]   |     {{% support support="full" link="/i/swag" %}}     |                            {{% support support="unknown" %}}                             |          {{% support %}}          |  {{% support support="full" %}}   |
|     [HAProxy] ([guide](/i/haproxy))     |  [ForwardAuth]   |   {{% support support="full" link="/i/haproxy" %}}    |                            {{% support support="unknown" %}}                             | {{% support support="unknown" %}} |  {{% support support="full" %}}   |
|     [Skipper] ([guide](/i/skipper))     |  [ForwardAuth]   |   {{% support support="full" link="/i/skipper" %}}    |                            {{% support support="unknown" %}}                             | {{% support support="unknown" %}} | {{% support support="unknown" %}} |
| [Traefik] 1.x ([guide](/i/traefik/v1))  |  [ForwardAuth]   | {{% support support="legacy" link="/i/traefik/v1" %}} |                            {{% support support="unknown" %}}                             | {{% support support="legacy" %}}  | {{% support support="legacy" %}}  |
|                [Apache]                 |       N/A        |            {{% support link="#apache" %}}             |                                     {{% support %}}                                      |          {{% support %}}          |          {{% support %}}          |
|                  [IIS]                  |       N/A        |              {{% support link="#iis" %}}              |                                     {{% support %}}                                      |          {{% support %}}          |          {{% support %}}          |

[ForwardAuth]: ../../reference/guides/proxy-authorization.md#forwardauth
[AuthRequest]: ../../reference/guides/proxy-authorization.md#authrequest
[ExtAuthz]: ../../reference/guides/proxy-authorization.md#extauthz
[Implementation]: ../../reference/guides/proxy-authorization.md#implementations

Legend:

|               Icon                |                  Meaning                  |
|:---------------------------------:|:-----------------------------------------:|
|  {{% support support="full" %}}   |                 Supported                 |
| {{% support support="unknown" %}} |                  Unknown                  |
| {{% support support="partial" %}} | Partially Supported and/or Legacy Support |
| {{% support support="legacy" %}}  | Previously / Formally Supported (Legacy)  |
|  {{% support support="none" %}}   |               Not Supported               |

## Support

### Standard

Standard support includes the essential features in securing an application with Authelia such as:

* Redirecting users to the Authelia portal if they are not authenticated.
* Redirecting users to the target application after authentication has occurred successfully.

It does not include actually running Authelia as a service behind the proxy, any proxy should be compatible with serving
the Authelia portal itself. Standard support is only important for protected applications.

### Kubernetes

While proxies that generally support Authelia outside a [Kubernetes] cluster, there are a few situations where that does
not translate to being possible when used as an [Ingress Controller]. There are various reasons for this such as the
reverse proxy in question does not even support running as a [Kubernetes] [Ingress Controller], or the required modules
to perform authentication transparently to the user are not typically available inside a cluster.

More information about [Kubernetes] deployments of Authelia can be read in the
[documentation](../../integration/kubernetes/introduction.md).

### XHR Redirect

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
The XHR is a deprecated web feature and applications should be using the new [Fetch API](https://developer.mozilla.org/en-US/docs/Web/API/Fetch_API) which does not have
the same issues regarding redirects (the [Fetch API](https://developer.mozilla.org/en-US/docs/Web/API/Fetch_API) allows developers to
[control how to handle them](https://developer.mozilla.org/en-US/docs/Web/API/Request/redirect)). As such the fact
a proxy does not support it should only be seen as a means to communicate a feature not that the proxy should not be
used.
{{< /callout >}}

XML HTTP Requests do not typically redirect browsers when returned 30x status codes. Instead, the standard method is to
return a 401 status code with a Location header. While this may seem trivial; currently there isn't wide support for it.
For example the nginx ngx_http_auth_request_module does not seem to support this in any way.

### Request Method

Authelia detects the upstream request method using the X-Forwarded-Method header. Some proxies set this out of the box,
some require you to configure this manually. At the present time all proxies that have
[Standard Support](#standard) do support this.

## Specific Proxy Notes

### HAProxy

[HAProxy] is only supported via a lua [module](https://github.com/haproxytech/haproxy-lua-http). Lua is typically not
available in [Kubernetes]. You would likely have to build your own [HAProxy] image.

### Envoy

[Envoy] is supported with Authelia v4.37.0 and higher via the [Envoy] proxy [external authorization] filter.

[external authorization]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/ext_authz/v3/ext_authz.proto.html#extensions-filters-http-ext-authz-v3-extauthz

### Caddy

[Caddy] needs to be version 2.5.1 or greater.

### Apache

[Apache] is not supported as it has no module that supports this kind of authentication method. It's not certain this
would even be possible, however if anyone did something like this in the future we'd be interested in a contribution.

### IIS

Microsoft [IIS] is not supported as it has no module that supports this kind of authentication method. It's not certain
this would even be possible, however if anyone did something like this in the future we'd be interested in a
contribution.

[NGINX]: https://www.nginx.com/
[NGINX Proxy Manager]: https://nginxproxymanager.com/
[SWAG]: https://docs.linuxserver.io/general/swag
[Traefik]: https://traefik.io/
[Caddy]: https://caddyserver.com/
[HAProxy]: https://www.haproxy.com/
[Envoy]: https://www.envoyproxy.io/
[Skipper]: https://opensource.zalando.com/skipper/
[Caddy]: https://caddyserver.com/
[Apache]: https://httpd.apache.org/
[IIS]: https://www.iis.net/
[Kubernetes]: https://kubernetes.io/
[Ingress Controller]: https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/
