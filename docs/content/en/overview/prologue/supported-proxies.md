---
title: "Supported Proxies"
description: "An introduction into the Authelia overview."
lead: "An introduction into the Authelia overview."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  overview:
    parent: "prologue"
weight: 130
toc: false
---

The following table is a support matrix for Authelia features and specific reverse proxies.

|         Proxy         |                                       Standard                                        |                                      Kubernetes                                      |             XHR Redirect             |            Request Method            |
|:---------------------:|:-------------------------------------------------------------------------------------:|:------------------------------------------------------------------------------------:|:------------------------------------:|:------------------------------------:|
|       [Traefik]       |       [<i class="icon-support-full"></i>](../../integration/proxies/traefik.md)       | [<i class="icon-support-full"></i>](../../integration/kubernetes/traefik-ingress.md) |  <i class="icon-support-full"></i>   |  <i class="icon-support-full"></i>   |
|        [NGINX]        |        [<i class="icon-support-full"></i>](../../integration/proxies/nginx.md)        |  [<i class="icon-support-full"></i>](../../integration/kubernetes/nginx-ingress.md)  |  <i class="icon-support-none"></i>   |  <i class="icon-support-full"></i>   |
| [NGINX Proxy Manager] | [<i class="icon-support-full"></i>](../../integration/proxies/nginx-proxy-manager.md) |                         <i class="icon-support-unknown"></i>                         |  <i class="icon-support-none"></i>   |  <i class="icon-support-full"></i>   |
|        [SWAG]         |        [<i class="icon-support-full"></i>](../../integration/proxies/swag.md)         |                         <i class="icon-support-unknown"></i>                         |  <i class="icon-support-none"></i>   |  <i class="icon-support-full"></i>   |
|       [HAProxy]       |       [<i class="icon-support-full"></i>](../../integration/proxies/haproxy.md)       |                         <i class="icon-support-unknown"></i>                         | <i class="icon-support-unknown"></i> |  <i class="icon-support-full"></i>   |
|        [Caddy]        |        [<i class="icon-support-full"></i>](../../integration/proxies/caddy.md)        |                         <i class="icon-support-unknown"></i>                         |  <i class="icon-support-full"></i>   |  <i class="icon-support-full"></i>   |
|     [Traefik] 1.x     |      [<i class="icon-support-full"></i>](../../integration/proxies/traefikv1.md)      |                         <i class="icon-support-unknown"></i>                         |  <i class="icon-support-full"></i>   |  <i class="icon-support-full"></i>   |
|        [Envoy]        |      [<i class="icon-support-unknown"></i>](../../integration/proxies/envoy.md)       |                         <i class="icon-support-unknown"></i>                         | <i class="icon-support-unknown"></i> | <i class="icon-support-unknown"></i> |
|       [Skipper]       |       [<i class="icon-support-full"></i>](../../integration/proxies/skipper.md)       |                         <i class="icon-support-unknown"></i>                         | <i class="icon-support-unknown"></i> | <i class="icon-support-unknown"></i> |
|       [Apache]        |                 <i class="icon-support-none" alt="Not Supported"></i>                 |                          <i class="icon-support-none"></i>                           |  <i class="icon-support-none"></i>   |  <i class="icon-support-none"></i>   |
|         [IIS]         |                           <i class="icon-support-none"></i>                           |                          <i class="icon-support-none"></i>                           |  <i class="icon-support-none"></i>   |  <i class="icon-support-none"></i>   |

Legend:

|                 Icon                 |       Meaning       |
|:------------------------------------:|:-------------------:|
|  <i class="icon-support-full"></i>   |      Supported      |
| <i class="icon-support-unknown"></i> |       Unknown       |
| <i class="icon-support-partial"></i> | Partially Supported |
|  <i class="icon-support-none"></i>   |    Not Supported    |

## More Information

For more comprehensive support information please see the
[Proxy Integration Support](../../integration/proxies/support.md) guide.

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
