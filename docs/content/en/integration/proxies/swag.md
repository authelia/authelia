---
title: "SWAG"
description: "An integration guide for Authelia and the SWAG reverse proxy"
lead: "A guide on integrating Authelia with SWAG."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  integration:
    parent: "proxies"
weight: 351
toc: true
aliases:
  - /i/swag
---

[SWAG] is a reverse proxy supported by __Authelia__. It's an [NGINX] proxy container with bundled configurations to make
your life easier.

*__Important:__ When using these guides it's important to recognize that we cannot provide a guide for every possible
method of deploying a proxy. These guides show a suggested setup only and you need to understand the proxy
configuration and customize it to your needs. To-that-end we include links to the official proxy documentation
throughout this documentation and in the [See Also](#see-also) section.*

## Introduction

As [SWAG] is a [NGINX] proxy with curated configurations, integration of __Authelia__ with [SWAG] is very easy and you
only need to enabled two includes.

*__Note:__ All paths in this guide are the locations inside the container. You will have to either edit the files within
the container or adapt the path to the path you have mounted the relevant container path to.*

## Get Started

It's __*strongly recommended*__ that users setting up *Authelia* for the first time take a look at our
[Get Started](../prologue/get-started.md) guide. This takes you through various steps which are essential to
bootstrapping *Authelia*.

## Requirements

[SWAG] supports the required [NGINX](nginx.md#requirements) requirements for __Authelia__ out-of-the-box.

## Trusted Proxies

*__Important:__ You should read the [Forwarded Headers] section and this section as part of any proxy configuration.
Especially if you have never read it before.*

To configure trusted proxies for [SWAG] see the [NGINX] section on [Trusted Proxies](nginx.md#trusted-proxies).
Adapting this to [SWAG] is beyond the scope of this documentation.

## Prerequisite Steps

These steps must be followed regardless of the choice of [subdomain](#subdomain-steps) or [subpath](#subpath-steps).

1. Deploy __Authelia__ to your docker network with the `container_name` of `authelia` and ensure it's listening on the
   default port and you have not configured the __Authelia__ server TLS settings.

## Subdomain Steps

In the server configuration for the application you want to protect:

1. Edit the `/config/nginx/proxy-confs/` file for the application you wish to protect.
2. Uncomment the `#include /config/nginx/authelia-server.conf;` line which should be within the `server` block
   but not inside any `location` blocks.
3. Uncomment the `#include /config/nginx/authelia-location.conf;` line which should be within the applications
   `location` block.

### Example

```nginx
server {
    listen 443 ssl;
    listen [::]:443 ssl;

    server_name heimdall.*;

    include /config/nginx/ssl.conf;

    client_max_body_size 0;

    # Authelia: Step 1.
    include /config/nginx/authelia-server.conf;

    location / {
        # Authelia: Step 2.
        include /config/nginx/authelia-location.conf;

        include /config/nginx/proxy.conf;
        resolver 127.0.0.11 valid=30s;
        set $upstream_app heimdall;
        set $upstream_port 443;
        set $upstream_proto https;
        proxy_pass $upstream_proto://$upstream_app:$upstream_port;
    }
}
```

## Subpath Steps

*__Note:__ Steps 1 and 2 only need to be done once, even if you wish to protect multiple applications.*

1. Edit `/config/nginx/proxy-confs/default`.
2. Uncomment the `#include /config/nginx/authelia-server.conf;` line.
3. Edit the `/config/nginx/proxy-confs/` file for the application you wish to protect.
4. Uncomment the `#include /config/nginx/authelia-location.conf;` line which should be within the applications
   `location` block.

### Example

```nginx
location ^~ /bazarr/ {
    # Authelia: Step 4.
    include /config/nginx/authelia-location.conf;

    include /config/nginx/proxy.conf;
    resolver 127.0.0.11 valid=30s;
    set $upstream_app bazarr;
    set $upstream_port 6767;
    set $upstream_proto http;
    proxy_pass $upstream_proto://$upstream_app:$upstream_port;

    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "Upgrade";
}
```

## See Also

* [Authelia NGINX Integration Documentation](nginx.md)
* [LinuxServer.io Setting Up Authelia With SWAG Documentation / Blog Post](https://www.linuxserver.io/blog/2020-08-26-setting-up-authelia)
* [NGINX ngx_http_auth_request_module Module Documentation](https://nginx.org/en/docs/http/ngx_http_auth_request_module.html)
* [Forwarded Headers]

[SWAG]: https://docs.linuxserver.io/general/swag
[NGINX]: https://www.nginx.com/
[Forwarded Headers]: fowarded-headers
