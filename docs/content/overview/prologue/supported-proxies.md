---
title: "Supported Proxies"
description: "An introduction into the Authelia overview."
summary: "An introduction into the Authelia overview."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 130
toc: true
aliases:
  - '/docs/deployment/supported-proxies/'
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

The following table is a support matrix for Authelia features and specific reverse proxies.

|                  Proxy                  |                       Standard                        |                                        Kubernetes                                        |           XHR Redirect            |          Request Method           |
|:---------------------------------------:|:-----------------------------------------------------:|:----------------------------------------------------------------------------------------:|:---------------------------------:|:---------------------------------:|
|     [Traefik] ([guide](/i/traefik))     |   {{% support support="full" link="/i/traefik" %}}    |  {{% support support="full" link="../../integration/kubernetes/traefik-ingress.md" %}}   |  {{% support support="full" %}}   |  {{% support support="full" %}}   |
|       [Caddy] ([guide](/i/caddy))       |    {{% support support="full" link="/i/caddy" %}}     |                            {{% support support="unknown" %}}                             |  {{% support support="full" %}}   |  {{% support support="full" %}}   |
|       [Envoy] ([guide](/i/envoy))       |    {{% support support="full" link="/i/envoy" %}}     | {{% support support="full" link="../../integration/kubernetes/envoy/introduction.md" %}} | {{% support support="unknown" %}} |  {{% support support="full" %}}   |
|       [NGINX] ([guide](/i/nginx))       |    {{% support support="full" link="/i/nginx" %}}     |   {{% support support="full" link="../../integration/kubernetes/nginx-ingress.md" %}}    |          {{% support %}}          |  {{% support support="full" %}}   |
| [NGINX Proxy Manager] ([guide](/i/npm)) |     {{% support support="full" link="/i/npm" %}}      |                            {{% support support="unknown" %}}                             |          {{% support %}}          |  {{% support support="full" %}}   |
|        [SWAG] ([guide](/i/swag))        |     {{% support support="full" link="/i/swag" %}}     |                            {{% support support="unknown" %}}                             |          {{% support %}}          |  {{% support support="full" %}}   |
|     [HAProxy] ([guide](/i/haproxy))     |   {{% support support="full" link="/i/haproxy" %}}    |                            {{% support support="unknown" %}}                             | {{% support support="unknown" %}} |  {{% support support="full" %}}   |
|     [Skipper] ([guide](/i/skipper))     |   {{% support support="full" link="/i/skipper" %}}    |                            {{% support support="unknown" %}}                             | {{% support support="unknown" %}} | {{% support support="unknown" %}} |
| [Traefik] 1.x ([guide](/i/traefik/v1))  | {{% support support="legacy" link="/i/traefik/v1" %}} |                            {{% support support="unknown" %}}                             | {{% support support="legacy" %}}  | {{% support support="legacy" %}}  |
|                [Apache]                 |                    {{% support %}}                    |                                     {{% support %}}                                      |          {{% support %}}          |          {{% support %}}          |
|                  [IIS]                  |                    {{% support %}}                    |                                     {{% support %}}                                      |          {{% support %}}          |          {{% support %}}          |

Legend:

|               Icon                |                 Meaning                  |
|:---------------------------------:|:----------------------------------------:|
|  {{% support support="full" %}}   |                Supported                 |
| {{% support support="unknown" %}} |                 Unknown                  |
| {{% support support="partial" %}} |           Partially Supported            |
| {{% support support="legacy" %}}  | Previously / Formally Supported (Legacy) |
|          {{% support %}}          |              Not Supported               |

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
