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

|         Proxy         |                                           Standard                                           |                                      Kubernetes                                       |           XHR Redirect            |          Request Method           |
|:---------------------:|:--------------------------------------------------------------------------------------------:|:-------------------------------------------------------------------------------------:|:---------------------------------:|:---------------------------------:|
|       [Traefik]       |          {{% support support="full" link="../../integration/proxies/traefik.md" %}}          | {{% support support="full" link="../../integration/kubernetes/traefik-ingress.md" %}} |  {{% support support="full" %}}   |  {{% support support="full" %}}   |
|        [Caddy]        |           {{% support support="full" link="../../integration/proxies/caddy.md" %}}           |                           {{% support support="unknown" %}}                           |  {{% support support="full" %}}   |  {{% support support="full" %}}   |
|        [Envoy]        |           {{% support support="full" link="../../integration/proxies/envoy.md" %}}           |      {{% support support="full" link="../../integration/kubernetes/istio.md" %}}      | {{% support support="unknown" %}} |  {{% support support="full" %}}   |
|        [NGINX]        |           {{% support support="full" link="../../integration/proxies/nginx.md" %}}           |  {{% support support="full" link="../../integration/kubernetes/nginx-ingress.md" %}}  |          {{% support %}}          |  {{% support support="full" %}}   |
| [NGINX Proxy Manager] | {{% support support="full" link="../../integration/proxies/nginx-proxy-manager/index.md" %}} |                                    {{% support %}}                                    |          {{% support %}}          |  {{% support support="full" %}}   |
|        [SWAG]         |           {{% support support="full" link="../../integration/proxies/swag.md" %}}            |                                    {{% support %}}                                    |          {{% support %}}          |  {{% support support="full" %}}   |
|       [HAProxy]       |          {{% support support="full" link="../../integration/proxies/haproxy.md" %}}          |                           {{% support support="unknown" %}}                           | {{% support support="unknown" %}} |  {{% support support="full" %}}   |
|     [Traefik] 1.x     |         {{% support support="full" link="../../integration/proxies/traefikv1.md" %}}         |                           {{% support support="unknown" %}}                           |  {{% support support="full" %}}   |  {{% support support="full" %}}   |
|       [Skipper]       |          {{% support support="full" link="../../integration/proxies/skipper.md" %}}          |                                    {{% support %}}                                    | {{% support support="unknown" %}} | {{% support support="unknown" %}} |
|       [Apache]        |                                       {{% support %}}                                        |                                    {{% support %}}                                    |          {{% support %}}          |          {{% support %}}          |
|         [IIS]         |                                       {{% support %}}                                        |                                    {{% support %}}                                    |          {{% support %}}          |          {{% support %}}          |

Legend:

|                Icon                |       Meaning       |
|:----------------------------------:|:-------------------:|
|   {{% support support="full" %}}   |      Supported      |
| {{% support support="unknown" %}}  |       Unknown       |
| {{% support support="partial" %}}  | Partially Supported |
|          {{% support %}}           |    Not Supported    |

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
