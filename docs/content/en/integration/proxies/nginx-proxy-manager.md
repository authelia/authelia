---
title: "NGINX Proxy Manager"
description: "An integration guide for Authelia and the NGINX Proxy Manager reverse proxy"
lead: "A guide on integrating Authelia with NGINX Proxy Manager."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  integration:
    parent: "proxies"
weight: 352
toc: true
aliases:
  - /i/npm
---

[NGINX Proxy Manager] is supported by __Authelia__. It's a [NGINX] proxy with a configuration UI.

*__Important:__ When using these guides it's important to recognize that we cannot provide a guide for every possible
method of deploying a proxy. These guides show a suggested setup only and you need to understand the proxy
configuration and customize it to your needs. To-that-end we include links to the official proxy documentation
throughout this documentation and in the [See Also](#see-also) section.*

## UNDER CONSTRUCTION

While this proxy is supported we don't have any specific documentation for it at the present time. Please see the
[NGINX integration documentation](nginx.md) for hints on how to configure this.

## Get Started

It's __*strongly recommended*__ that users setting up *Authelia* for the first time take a look at our
[Get Started](../prologue/get-started.md) guide. This takes you through various steps which are essential to
bootstrapping *Authelia*.

## Requirements

[NGINX Proxy Manager] supports the required [NGINX](nginx.md#requirements) requirements for __Authelia__ out-of-the-box.

## Trusted Proxies

*__Important:__ You should read the [Forwarded Headers] section and this section as part of any proxy configuration.
Especially if you have never read it before.*

To configure trusted proxies for [NGINX Proxy Manager] see the [NGINX] section on
[Trusted Proxies](nginx.md#trusted-proxies). Adapting this to [NGINX Proxy Manager] is beyond the scope of
this documentation.

## See Also

* [NGINX Proxy Manager Documentation](https://nginxproxymanager.com/setup/)
* [NGINX ngx_http_auth_request_module Module Documentation](https://nginx.org/en/docs/http/ngx_http_auth_request_module.html)
* [Forwarded Headers]

[NGINX Proxy Manager]: https://nginxproxymanager.com/
[NGINX]: https://www.nginx.com/
[Forwarded Headers]: fowarded-headers
