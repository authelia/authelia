---
title: "Support"
description: "An support matrix for Authelia and several supported reverse proxies"
lead: "This documentation details a support matrix for Authelia features and specific reverse proxies as well as several caveats etc."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  integration:
    parent: "proxies"
weight: 311
toc: true
aliases:
  - /i/proxy
  - /docs/home/supported-proxies.html
---

|         Proxy         |                      [Standard](#standard)                       |                              [Kubernetes](#kubernetes)                               |    [XHR Redirect](#xhr-redirect)     |  [Request Method](#request-method)   |
|:---------------------:|:----------------------------------------------------------------:|:------------------------------------------------------------------------------------:|:------------------------------------:|:------------------------------------:|
|       [Traefik]       |         [<i class="icon-support-full"></i>](traefik.md)          | [<i class="icon-support-full"></i>](../../integration/kubernetes/traefik-ingress.md) |  <i class="icon-support-full"></i>   |  <i class="icon-support-full"></i>   |
|        [NGINX]        |          [<i class="icon-support-full"></i>](nginx.md)           |  [<i class="icon-support-full"></i>](../../integration/kubernetes/nginx-ingress.md)  |  <i class="icon-support-none"></i>   |  <i class="icon-support-full"></i>   |
| [NGINX Proxy Manager] |   [<i class="icon-support-full"></i>](nginx-proxy-manager.md)    |                         <i class="icon-support-unknown"></i>                         |  <i class="icon-support-none"></i>   |  <i class="icon-support-full"></i>   |
|        [SWAG]         |           [<i class="icon-support-full"></i>](swag.md)           |                         <i class="icon-support-unknown"></i>                         |  <i class="icon-support-none"></i>   |  <i class="icon-support-full"></i>   |
|       [HAProxy]       |         [<i class="icon-support-full"></i>](haproxy.md)          |                         <i class="icon-support-unknown"></i>                         | <i class="icon-support-unknown"></i> |  <i class="icon-support-full"></i>   |
|        [Caddy]        |          [<i class="icon-support-full"></i>](caddy.md)           |                         <i class="icon-support-unknown"></i>                         |  <i class="icon-support-full"></i>   |  <i class="icon-support-full"></i>   |
|     [Traefik] 1.x     |        [<i class="icon-support-full"></i>](traefikv1.md)         |                         <i class="icon-support-unknown"></i>                         |  <i class="icon-support-full"></i>   |  <i class="icon-support-full"></i>   |
|        [Envoy]        |         [<i class="icon-support-unknown"></i>](envoy.md)         |                         <i class="icon-support-unknown"></i>                         | <i class="icon-support-unknown"></i> | <i class="icon-support-unknown"></i> |
|       [Skipper]       |         [<i class="icon-support-full"></i>](skipper.md)          |                         <i class="icon-support-unknown"></i>                         | <i class="icon-support-unknown"></i> | <i class="icon-support-unknown"></i> |
|       [Apache]        | [<i class="icon-support-none" alt="Not Supported"></i>](#apache) |                          <i class="icon-support-none"></i>                           |  <i class="icon-support-none"></i>   |  <i class="icon-support-none"></i>   |
|         [IIS]         |            [<i class="icon-support-none"></i>](#iis)             |                          <i class="icon-support-none"></i>                           |  <i class="icon-support-none"></i>   |  <i class="icon-support-none"></i>   |

Legend:

|                 Icon                 |       Meaning       |
|:------------------------------------:|:-------------------:|
|  <i class="icon-support-full"></i>   |      Supported      |
| <i class="icon-support-unknown"></i> |       Unknown       |
| <i class="icon-support-partial"></i> | Partially Supported |
|  <i class="icon-support-none"></i>   |    Not Supported    |

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

*__Note:__ The XHR is a deprecated web feature and applications should be using the new [Fetch API] which does not have
the same issues regarding redirects (the [Fetch API] allows developers to
[control how to handle them](https://developer.mozilla.org/en-US/docs/Web/API/Request/redirect)). As such the fact
a proxy does not support it should only be seen as a means to communicate a feature not that the proxy should not be
used.*

XML HTTP Requests do not typically redirect browsers when returned 30x status codes. Instead, the standard method is to
return a 401 status code with a Location header. While this may seem trivial; currently there isn't wide support for it.
For example the nginx ngx_http_auth_request_module does not seem to support this in any way.

### Request Method

Authelia detects the upstream request method using the X-Forwarded-Method header. Some proxies set this out of the box,
some require you to configure this manually. At the present time all proxies that have
[Standard Support](#standard-support) do support this.

## Specific proxy notes

### HAProxy

[HAProxy] is only supported via a lua [module](https://github.com/haproxytech/haproxy-lua-http). Lua is typically not
available in [Kubernetes]. You would likely have to build your own [HAProxy] image.

### Envoy

[Envoy] is currently not documented however we believe it is likely to be technically supported. This should be possible
via [Envoy]'s [external authorization](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/ext_authz/v3/ext_authz.proto.html#extensions-filters-http-ext-authz-v3-extauthz).

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

[Fetch API]: https://developer.mozilla.org/en-US/docs/Web/API/Fetch_API
