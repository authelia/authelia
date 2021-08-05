---
layout: default
title: Supported Proxies
parent: Home
nav_order: 2
---

The following table is a support matrix for Authelia features and specific reverse proxies.

|Proxy        |[Standard Support](#standard)                                                                          |[Kubernetes Support](#kubernetes)                                                               |[XHR Redirect](#xhr-redirect)                         |[Request Method](#request-method)                     |
|:-----------:|:-----------------------------------------------------------------------------------------------------:|:----------------------------------------------------------------------------------------------:|:----------------------------------------------------:|:----------------------------------------------------:|
|[NGINX]      |[<span class="material-icons green">check_circle</span>](../deployment/supported-proxies/nginx.md)     |[<span class="material-icons green">check_circle</span>](../deployment/deployment-kubernetes.md)|<span class="material-icons red">cancel</span>        |<span class="material-icons green">check_circle</span>|
|[Traefik] 1.x|[<span class="material-icons green">check_circle</span>](../deployment/supported-proxies/traefik1.x.md)|<span class="material-icons orange">error</span>                                                |<span class="material-icons green">check_circle</span>|<span class="material-icons green">check_circle</span>|
|[Traefik] 2.x|[<span class="material-icons green">check_circle</span>](../deployment/supported-proxies/traefik2.x.md)|[<span class="material-icons green">check_circle</span>](../deployment/deployment-kubernetes.md)|<span class="material-icons green">check_circle</span>|<span class="material-icons green">check_circle</span>|
|[HAProxy]    |[<span class="material-icons green">check_circle</span>](../deployment/supported-proxies/haproxy.md)   |<span class="material-icons red">cancel</span>                                                  |<span class="material-icons orange">error</span>      |<span class="material-icons green">check_circle</span>|
|[Envoy]      |<span class="material-icons orange">error</span>                                                       |<span class="material-icons orange">error</span>                                                |<span class="material-icons orange">error</span>      |<span class="material-icons orange">error</span>      |
|[Caddy] 2.x  |<span class="material-icons orange">error</span>                                                       |<span class="material-icons red">cancel</span>                                                  |<span class="material-icons orange">error</span>      |<span class="material-icons orange">error</span>      |
|[Apache]     |<span class="material-icons red">cancel</span>                                                         |<span class="material-icons red">cancel</span>                                                  |<span class="material-icons red">cancel</span>        |<span class="material-icons red">cancel</span>        |
|[IIS]        |<span class="material-icons red">cancel</span>                                                         |<span class="material-icons red">cancel</span>                                                  |<span class="material-icons red">cancel</span>        |<span class="material-icons red">cancel</span>        |

<span class="material-icons green">check_circle</span> *Support confirmed, additionally these icons are links to documentation for both the Standard and Kubernetes support columns*

<span class="material-icons orange">error</span> *Support is likely and being investigated*

<span class="material-icons red">cancel</span> *Either not supported or unlikely to be supported*

## Support

### Standard

Standard support includes the essential features in securing an application with Authelia such as:

- Redirecting users to the Authelia portal if they are not authenticated.
- Redirecting users to the target application after authentication has occurred successfully.

It does not include actually running Authelia as a service behind the proxy, any proxy should be compatible with serving
the Authelia portal itself. Standard support is only important for protected applications.

### Kubernetes

While proxies that generally support Authelia outside a [Kubernetes] cluster, there are a few situations where that does
not translate to being possible when used as an [Ingress Controller]. There are various reasons for this such as the
reverse proxy in question does not even support running as a [Kubernetes] [Ingress Controller], or the required modules
to perform authentication transparently to the user are not typically available inside a cluster.

More information about [Kubernetes] deployments of Authelia can be read in the 
[documentation](../deployment/deployment-kubernetes.md).

### XHR Redirect

XML HTTP Requests do not typically redirect browsers when returned 30x status codes. Instead, the standard method is to
return a 401 status code with a Location header. While this may seem trivial; currently there isn't wide support for it.
For example nginx's ngx_http_auth_request_module does not seem to support this in any way.

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

[Work](https://github.com/authelia/caddy-forwardauth) is being done to support Caddy 2.x, however this is a low 
priority. You can see the progress and try it for yourself if you're interested. Regular feedback would accelerate this
work.

### Apache

[Apache] has no module that supports this kind of authentication method. It's not certain this would even be possible,
however if anyone did something like this in the past we'd be interested in a contribution.

### IIS

Microsoft [IIS] not currently supported since no auth module exists for this purpose out-of-the-box or from any known
third party. It's likely possible but unlikely to be highly used so there is little to be gained by supporting this proxy.

[NGINX]: https://www.nginx.com/
[Traefik]: https://traefik.io/
[HAProxy]: https://www.haproxy.com/
[Envoy]: https://www.envoyproxy.io/
[Caddy]: https://caddyserver.com/
[Apache]: https://httpd.apache.org/
[IIS]: https://www.iis.net/
[Kubernetes]: https://kubernetes.io/
[Ingress Controller]: https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/